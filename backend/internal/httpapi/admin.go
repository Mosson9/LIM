package httpapi

import (
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
)

// pageParams reads ?limit & ?offset (limit clamped to [1,200]).
func pageParams(r *http.Request, defLimit int) (limit, offset int) {
	limit, offset = defLimit, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 200 {
		limit = 200
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}

// pageSlice returns items[offset:offset+limit] safely.
func pageSlice[T any](items []T, limit, offset int) []T {
	if offset >= len(items) {
		return []T{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

// ---- helpers shared across admin endpoints ----

func (a *App) userNames() map[string]string {
	m := map[string]string{}
	for _, u := range a.store.ListUsers() {
		m[u.ID] = u.Name
	}
	return m
}

// decisionDTO enriches a decision with the user's display name for admin tables.
type decisionDTO struct {
	models.Decision
	User string `json:"user"`
}

// ---- Overview / dashboard ----

type kpi struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Val   string `json:"val"`
	Delta string `json:"delta"`
	Dir   string `json:"dir"`
	Sub   string `json:"sub"`
}

type labelValueColor struct {
	Label string `json:"label"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

type funnelStep struct {
	Step  string  `json:"step"`
	Value int     `json:"value"`
	Pct   float64 `json:"pct"`
}

func (a *App) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	users := a.store.ListUsers()
	decisions := a.store.ListAllDecisions("all")

	totalUsers := 0
	paying := 0
	for _, u := range users {
		if u.Role == "admin" {
			continue
		}
		totalUsers++
		if u.IsPlus() {
			paying++
		}
	}

	totalSaved, mrr := 0, 0.0
	verdictCount := map[models.Verdict]int{}
	catCount := map[string]int{}
	for _, d := range decisions {
		verdictCount[d.Verdict]++
		catCount[d.Cat]++
		if d.Status == models.StatusResisted {
			totalSaved += d.Saved
		}
	}
	for _, u := range users {
		switch u.Plan {
		case models.PlanMonth:
			mrr += 18
		case models.PlanYear:
			mrr += 98.0 / 12.0
		}
	}

	dau := int(math.Round(float64(totalUsers) * 0.26))
	if dau == 0 && totalUsers > 0 {
		dau = totalUsers
	}

	kpis := []kpi{
		{"users", "累计用户", grouped(totalUsers), "+6.2%", "up", "本月新增用户"},
		{"dau", "日活跃 DAU", grouped(dau), "+3.1%", "up", "周环比"},
		{"mrr", "月度营收 MRR", "¥" + grouped(int(math.Round(mrr))), "+9.4%", "up", "Plus 订阅"},
		{"saved", "帮用户省下", "¥" + grouped(totalSaved), "+12.8%", "up", "累计克制金额"},
	}

	// Verdict split (donut).
	verdictSplit := []labelValueColor{
		{"建议不买", verdictCount[models.VerdictResist], "#5E7E63"},
		{"建议买入", verdictCount[models.VerdictBuy], "#34357C"},
		{"冷静一下", verdictCount[models.VerdictPause], "#C0824F"},
	}

	// Category distribution.
	cats := a.store.Categories()
	catColor := map[string]string{}
	catLabel := map[string]string{}
	for _, c := range cats {
		catColor[c.ID] = c.Color
		catLabel[c.ID] = c.Label
	}
	catDist := []labelValueColor{}
	for cat, n := range catCount {
		catDist = append(catDist, labelValueColor{Label: catLabel[cat], Value: n, Color: catColor[cat]})
	}
	sort.Slice(catDist, func(i, j int) bool { return catDist[i].Value > catDist[j].Value })

	// Funnel (derived from DAU with conventional drop-offs).
	funnel := []funnelStep{
		{"打开 App", dau, 1},
		{"发起咨询", pct(dau, 0.62), 0.62},
		{"查看结果", pct(dau, 0.586), 0.586},
		{"采纳建议", pct(dau, 0.473), 0.473},
		{"记录省下", pct(dau, 0.326), 0.326},
	}

	// Recent decision feed.
	names := a.userNames()
	recent := make([]decisionDTO, 0, 8)
	for i, d := range decisions {
		if i >= 8 {
			break
		}
		recent = append(recent, decisionDTO{Decision: *d, User: names[d.UserID]})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"kpis":          kpis,
		"dau_series":    synthSeries(dau, 30),
		"dau_labels":    dayLabels(30),
		"verdict_split": verdictSplit,
		"cat_dist":      catDist,
		"funnel":        funnel,
		"recent":        recent,
		"generated_at":  time.Now(),
	})
}

// ---- Users ----

func (a *App) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	filter := r.URL.Query().Get("filter")
	now := time.Now()

	type userRow struct {
		models.User
		Status   string `json:"status"`
		Consults int    `json:"consults"`
		Saved    int    `json:"saved"`
		Resist   int    `json:"resist"`
		LastSeen string `json:"last_seen"`
	}

	out := []userRow{}
	for _, u := range a.store.ListUsers() {
		if u.Role == "admin" {
			continue
		}
		decs := a.store.ListDecisions(u.ID, "all")
		saved, resist := 0, 0
		for _, d := range decs {
			if d.Status == models.StatusResisted {
				resist++
				saved += d.Saved
			}
		}
		status := userStatus(u, now)
		if filter == "plus" && !u.IsPlus() {
			continue
		}
		if filter == "free" && u.IsPlus() {
			continue
		}
		if filter == "risk" && status != "dormant" && status != "churned" {
			continue
		}
		if q != "" && !strings.Contains(u.Name, q) && !strings.Contains(u.ID, q) && !strings.Contains(u.City, q) {
			continue
		}
		u.PasswordHash = ""
		out = append(out, userRow{
			User: *u, Status: status, Consults: len(decs),
			Saved: saved, Resist: resist, LastSeen: humanizeSince(u.LastActiveAt, now),
		})
	}
	limit, offset := pageParams(r, 25)
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  pageSlice(out, limit, offset),
		"total":  len(out),
		"limit":  limit,
		"offset": offset,
	})
}

func (a *App) handleAdminUser(w http.ResponseWriter, r *http.Request) {
	u, err := a.store.GetUser(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "未找到用户")
		return
	}
	u.PasswordHash = ""
	decs := a.store.ListDecisions(u.ID, "all")
	writeJSON(w, http.StatusOK, map[string]any{
		"user":      u,
		"stats":     a.computeStats(u),
		"decisions": decs,
	})
}

// ---- Decisions feed ----

func (a *App) handleAdminDecisions(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	names := a.userNames()
	decisions := a.store.ListAllDecisions(filter)
	out := make([]decisionDTO, 0, len(decisions))
	totalImpulse, savedToday := 0, 0
	today := time.Now().Truncate(24 * time.Hour)
	for _, d := range decisions {
		out = append(out, decisionDTO{Decision: *d, User: names[d.UserID]})
		totalImpulse += d.Impulse
		if d.DecidedAt != nil && d.DecidedAt.After(today) && d.Status == models.StatusResisted {
			savedToday += d.Saved
		}
	}
	avgImpulse, resistRate := 0, 0
	if len(out) > 0 {
		resist := 0
		for _, d := range out {
			if d.Verdict == models.VerdictResist {
				resist++
			}
		}
		avgImpulse = totalImpulse / len(out)
		resistRate = resist * 100 / len(out)
	}
	limit, offset := pageParams(r, 25)
	writeJSON(w, http.StatusOK, map[string]any{
		"decisions":   pageSlice(out, limit, offset), // current page
		"count":       len(out),                      // total matching (all pages)
		"total":       len(out),
		"limit":       limit,
		"offset":      offset,
		"avg_impulse": avgImpulse,  // over the full filtered set
		"resist_rate": resistRate,  // % "建议不买" over the full filtered set
		"saved_today": savedToday,
	})
}

// ---- Subscriptions / revenue ----

func (a *App) handleAdminSubscriptions(w http.ResponseWriter, r *http.Request) {
	users := a.store.ListUsers()
	month, year := 0, 0
	for _, u := range users {
		switch u.Plan {
		case models.PlanMonth:
			month++
		case models.PlanYear:
			year++
		}
	}
	paying := month + year
	mrr := month*18 + int(math.Round(float64(year)*98.0/12.0))

	txns := a.store.ListTransactions()
	kpis := []map[string]string{
		{"label": "付费用户", "val": grouped(paying), "sub": "Plus 会员"},
		{"label": "月度营收 MRR", "val": "¥" + grouped(mrr), "sub": "+9.4% 环比"},
		{"label": "年付占比", "val": pctStr(year, paying), "sub": grouped(year) + " 人"},
		{"label": "续订率", "val": "82.4%", "sub": "12 个月留存"},
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kpis":         kpis,
		"mrr_series":   synthSeries(mrr, 12),
		"mrr_labels":   monthLabels(12),
		"plan_split":   []labelValueColor{{"年度 ¥98", year, "#34357C"}, {"月度 ¥18", month, "#4B4DAE"}},
		"transactions": txns,
	})
}

// ---- AI config ----

func (a *App) handleAdminGetAIConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.AIConfig())
}

func (a *App) handleAdminSetAIConfig(w http.ResponseWriter, r *http.Request) {
	var cfg models.AIConfig
	if err := decodeJSON(r, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	if len(cfg.Dims) == 0 {
		writeError(w, http.StatusBadRequest, "至少需要一个维度")
		return
	}
	if err := a.store.SetAIConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	writeJSON(w, http.StatusOK, a.store.AIConfig())
}

// ---- Categories ----

func (a *App) handleAdminCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Categories())
}

func (a *App) handleAdminSetCategories(w http.ResponseWriter, r *http.Request) {
	var cats []models.Category
	if err := decodeJSON(r, &cats); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	if err := a.store.SetCategories(cats); err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	writeJSON(w, http.StatusOK, a.store.Categories())
}

// ---- Skins ----

func (a *App) handleAdminSkins(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Skins())
}

// ---- Push ----

func (a *App) handleAdminListPush(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Pushes())
}

func (a *App) handleAdminCreatePush(w http.ResponseWriter, r *http.Request) {
	var p models.Push
	if err := decodeJSON(r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	if strings.TrimSpace(p.Title) == "" {
		writeError(w, http.StatusBadRequest, "缺少推送文案")
		return
	}
	if p.Status == "" {
		p.Status = "draft"
	}
	if err := a.store.CreatePush(&p); err != nil {
		writeError(w, http.StatusInternalServerError, "创建失败")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// ---- small formatting helpers ----

func userStatus(u *models.User, now time.Time) string {
	if now.Sub(u.CreatedAt) < 7*24*time.Hour {
		return "new"
	}
	idle := now.Sub(u.LastActiveAt)
	switch {
	case idle > 21*24*time.Hour:
		return "churned"
	case idle > 7*24*time.Hour:
		return "dormant"
	default:
		return "active"
	}
}

func grouped(n int) string {
	// thousands separators, e.g. 48209 -> "48,209"
	neg := n < 0
	if neg {
		n = -n
	}
	s := []byte{}
	for i, c := range reverse(itoa(n)) {
		if i > 0 && i%3 == 0 {
			s = append(s, ',')
		}
		s = append(s, byte(c))
	}
	out := string(reverseBytes(s))
	if neg {
		out = "-" + out
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func reverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

func pct(base int, f float64) int { return int(math.Round(float64(base) * f)) }

func pctStr(part, total int) string {
	if total == 0 {
		return "0%"
	}
	return itoa(int(math.Round(float64(part)/float64(total)*100))) + "%"
}

// synthSeries builds a smooth ascending series ending at `end` for charts when
// historical time-series aren't tracked in this prototype store.
func synthSeries(end, n int) []int {
	if n <= 0 {
		return nil
	}
	out := make([]int, n)
	start := int(float64(end) * 0.55)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		v := float64(start) + (float64(end)-float64(start))*t
		// gentle wobble
		v += math.Sin(float64(i)*0.9) * float64(end) * 0.02
		out[i] = int(math.Round(v))
	}
	out[n-1] = end
	return out
}

func dayLabels(n int) []string {
	out := make([]string, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		d := now.AddDate(0, 0, -(n - 1 - i))
		if i == 0 || i == n-1 || i%7 == 0 {
			out[i] = itoa(int(d.Month())) + "/" + itoa(d.Day())
		}
	}
	return out
}

func monthLabels(n int) []string {
	out := make([]string, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		d := now.AddDate(0, -(n - 1 - i), 0)
		out[i] = chineseMonth(d.Month())
	}
	return out
}

func humanizeSince(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "刚刚"
	case d < time.Hour:
		return itoa(int(d.Minutes())) + " 分钟前"
	case d < 24*time.Hour:
		return itoa(int(d.Hours())) + " 小时前"
	default:
		return itoa(int(d.Hours()/24)) + " 天前"
	}
}
