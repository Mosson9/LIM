package httpapi

import (
	"net/http"
	"sort"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
)

// monthBucket is one column of the spent-vs-saved chart.
type monthBucket struct {
	Label string `json:"label"`
	Spent int    `json:"spent"`
	Saved int    `json:"saved"`
}

// catSaved is a per-category savings total.
type catSaved struct {
	Cat   string `json:"cat"`
	Label string `json:"label"`
	Color string `json:"color"`
	Saved int    `json:"saved"`
}

// statsResp is the full statistics payload backing the Home and Growth screens.
type statsResp struct {
	MonthSaved    int           `json:"month_saved"`
	MonthSpent    int           `json:"month_spent"`
	TotalSaved    int           `json:"total_saved"`
	TotalSpent    int           `json:"total_spent"`
	ResistCount   int           `json:"resist_count"`
	BuyCount      int           `json:"buy_count"`
	Streak        int           `json:"streak"`     // 连续克制天数
	TreeStage     int           `json:"tree_stage"` // 0..6
	RestraintRate int           `json:"restraint_rate"`
	ByMonth       []monthBucket `json:"by_month"`
	TopCats       []catSaved    `json:"top_cats"`
}

func (a *App) handleStats(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	writeJSON(w, http.StatusOK, a.computeStats(u))
}

func (a *App) computeStats(u *models.User) statsResp {
	now := time.Now()
	decisions := a.store.ListDecisions(u.ID, "all") // resolved only

	var res statsResp
	curY, curM, _ := now.Date()

	savedByCat := map[string]int{}
	var lastBuy time.Time

	// Build trailing 5-month buckets keyed by year-month.
	buckets := make([]monthBucket, 5)
	bucketIndex := map[string]int{}
	for i := 0; i < 5; i++ {
		m := now.AddDate(0, -(4 - i), 0)
		buckets[i] = monthBucket{Label: chineseMonth(m.Month())}
		bucketIndex[ymKey(m.Year(), m.Month())] = i
	}

	for _, d := range decisions {
		when := d.CreatedAt
		if d.DecidedAt != nil {
			when = *d.DecidedAt
		}
		switch d.Status {
		case models.StatusResisted:
			res.ResistCount++
			res.TotalSaved += d.Saved
			savedByCat[d.Cat] += d.Saved
			if when.Year() == curY && when.Month() == curM {
				res.MonthSaved += d.Saved
			}
			if idx, ok := bucketIndex[ymKey(when.Year(), when.Month())]; ok {
				buckets[idx].Saved += d.Saved
			}
		case models.StatusBought:
			res.BuyCount++
			res.TotalSpent += d.Price
			if when.Year() == curY && when.Month() == curM {
				res.MonthSpent += d.Price
			}
			if idx, ok := bucketIndex[ymKey(when.Year(), when.Month())]; ok {
				buckets[idx].Spent += d.Price
			}
			if when.After(lastBuy) {
				lastBuy = when
			}
		}
	}
	res.ByMonth = buckets

	// Top saved categories (max 4).
	cats := a.store.Categories()
	catMeta := map[string]models.Category{}
	for _, c := range cats {
		catMeta[c.ID] = c
	}
	for cat, saved := range savedByCat {
		m := catMeta[cat]
		label, color := m.Label, m.Color
		if label == "" {
			label = cat
		}
		res.TopCats = append(res.TopCats, catSaved{Cat: cat, Label: label, Color: color, Saved: saved})
	}
	sort.Slice(res.TopCats, func(i, j int) bool { return res.TopCats[i].Saved > res.TopCats[j].Saved })
	if len(res.TopCats) > 4 {
		res.TopCats = res.TopCats[:4]
	}

	// Restraint rate.
	total := res.ResistCount + res.BuyCount
	if total > 0 {
		res.RestraintRate = int(float64(res.ResistCount) / float64(total) * 100)
	}

	// Streak: consecutive days since the most recent purchase.
	if lastBuy.IsZero() {
		// No purchases yet — count days since the account/first decision.
		anchor := u.CreatedAt
		if len(decisions) > 0 {
			anchor = decisions[len(decisions)-1].CreatedAt
		}
		res.Streak = daysBetween(anchor, now)
	} else {
		res.Streak = daysBetween(lastBuy, now)
	}

	res.TreeStage = treeStage(res.TotalSaved)
	return res
}

// treeStage maps cumulative savings onto the 0..6 growth stages.
func treeStage(totalSaved int) int {
	thresholds := []int{500, 2000, 5000, 9000, 14000, 20000}
	stage := 0
	for _, t := range thresholds {
		if totalSaved >= t {
			stage++
		}
	}
	return stage
}

func daysBetween(a, b time.Time) int {
	d := int(b.Sub(a).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

func ymKey(y int, m time.Month) string {
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

func chineseMonth(m time.Month) string {
	names := []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}
	return names[int(m)-1]
}
