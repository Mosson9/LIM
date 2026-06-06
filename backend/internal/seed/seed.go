// Package seed populates a fresh store with the catalogue content (categories,
// skins, plans, AI config) plus, optionally, the demo users and decisions used
// by the prototype so the app and admin dashboard look alive out of the box.
package seed

import (
	"time"

	"github.com/mosson9/lim/backend/internal/auth"
	"github.com/mosson9/lim/backend/internal/models"
	"github.com/mosson9/lim/backend/internal/store"
)

// Run seeds catalogue content always, and demo data when demo is true. It also
// ensures a bootstrap admin account exists.
func Run(s store.Store, demo bool, adminEmail, adminPassword string) error {
	if err := seedCatalogue(s); err != nil {
		return err
	}
	if err := ensureAdmin(s, adminEmail, adminPassword); err != nil {
		return err
	}
	if demo {
		if err := seedDemo(s); err != nil {
			return err
		}
	}
	return nil
}

// DefaultAIConfig is the engine configuration shipped on first boot. The weights
// and prompt mirror the admin prototype's "AI 维度 / Prompt" screen.
func DefaultAIConfig() models.AIConfig {
	return models.AIConfig{
		Dims: []models.DimensionConfig{
			{Key: "need", Label: "需求", Desc: "是否生活必需，缺它会否造成困难", Weight: 1.3, Enabled: true,
				Q: "这是生活中必需的吗？没有它，日子会变得困难或不便吗？"},
			{Key: "alt", Label: "替代", Desc: "是否已有同类 / 可租借共享修理", Weight: 1.1, Enabled: true,
				Q: "你是否已经拥有类似的东西？能不能靠租借、修理或共享来满足？"},
			{Key: "emo", Label: "情感", Desc: "动机是真喜欢还是广告社交情绪驱动", Weight: 1.2, Enabled: true,
				Q: "购买的动机是什么？是真的喜欢，还是广告、社交压力或一时情绪？"},
			{Key: "value", Label: "长期价值", Desc: "使用频率与寿命，能否长期带来价值", Weight: 1.0, Enabled: true,
				Q: "它的使用频率和寿命如何？能长期为你带来价值吗？"},
			{Key: "money", Label: "经济", Desc: "价格与预算匹配，是否挤占重要开支", Weight: 1.4, Enabled: true,
				Q: "价格和你的预算匹配吗？会挤占其他更重要的开支吗？"},
			{Key: "env", Label: "环境", Desc: "生产运输使用的环境影响是否符合价值观", Weight: 0.8, Enabled: true,
				Q: "它的生产、运输和使用对环境意味着什么？符合你的价值观吗？"},
		},
		PromptTemplate: "你是 LIM，一位温柔、不评判的理性消费陪伴者。\n" +
			"用户正在犹豫是否购买「{{item}}」，价格 {{price}} 元，动机：{{reason}}。\n" +
			"已知用户画像：月可支配 {{budget}} 元、家庭 {{members}} 人。\n\n" +
			"请从六个维度（需求 / 替代 / 情感 / 长期价值 / 经济 / 环境）逐一评分（1–5），\n" +
			"计算 0–100 的「冲动指数」，并给出「买 / 冷静一下 / 不买」的建议。\n" +
			"语气温柔、像朋友，不说教、不制造焦虑，结尾给一句温暖的话。",
		Persona:        "温柔朋友",
		Model:          "claude-sonnet-4-6",
		FreeDailyLimit: 3,
		CoolingHours:   24,
		UpdatedAt:      time.Now(),
	}
}

func seedCatalogue(s store.Store) error {
	if err := s.SetAIConfig(DefaultAIConfig()); err != nil {
		return err
	}
	cats := []models.Category{
		{ID: "digital", Label: "数码电子", Color: "#4B4DAE", Soft: "#ECECF6", Icon: "bolt", Free: true, Count: 3240},
		{ID: "fashion", Label: "服饰穿搭", Color: "#9A6FB0", Soft: "#F0EAF4", Icon: "tag", Free: true, Count: 2680},
		{ID: "home", Label: "家居好物", Color: "#5E7E63", Soft: "#E9EEE7", Icon: "home", Free: true, Count: 2110},
		{ID: "beauty", Label: "美妆护肤", Color: "#C76B8E", Soft: "#F6E7ED", Icon: "spark", Free: false, Count: 1520},
		{ID: "hobby", Label: "兴趣爱好", Color: "#3E8A8A", Soft: "#E4F0EF", Icon: "heart", Free: false, Count: 1180},
		{ID: "food", Label: "吃喝零食", Color: "#C0824F", Soft: "#F4E9DD", Icon: "wallet", Free: false, Count: 980},
		{ID: "other", Label: "其他", Color: "#8C8A82", Soft: "#EEEDE7", Icon: "grid", Free: true, Count: 740},
	}
	if err := s.SetCategories(cats); err != nil {
		return err
	}
	skins := []models.Skin{
		{ID: "classic", Name: "经典靛蓝", BG: "#34357C", Dark: false, Free: true, Used: "48%"},
		{ID: "paper", Name: "暖纸", BG: "#E7E2D6", Dark: true, Free: true, Used: "18%"},
		{ID: "ink", Name: "墨黑", BG: "#1B1B19", Dark: false, Free: false, Used: "12%"},
		{ID: "sage", Name: "山涧绿", BG: "#5E7E63", Dark: false, Free: false, Used: "9%"},
		{ID: "clay", Name: "陶土", BG: "#C0824F", Dark: false, Free: false, Used: "5%"},
		{ID: "plum", Name: "紫藤", BG: "#9A6FB0", Dark: false, Free: false, Used: "4%"},
		{ID: "sky", Name: "晴空", BG: "#5B86B0", Dark: false, Free: false, Used: "2%"},
		{ID: "rose", Name: "胭脂", BG: "#C76B8E", Dark: false, Free: false, Used: "2%"},
		{ID: "mono", Name: "雾白", BG: "#F4F3EE", Dark: true, Free: false},
	}
	if err := s.SetSkins(skins); err != nil {
		return err
	}
	plans := []models.PlanOption{
		{ID: "month", Name: "月度", Price: 18, Per: "/月", Note: "随时取消"},
		{ID: "year", Name: "年度", Price: 98, Per: "/年", Note: "省 ¥118 · 最受欢迎", Best: true},
	}
	if err := s.SetPlans(plans); err != nil {
		return err
	}
	perks := []models.PlusPerk{
		{Icon: "spark", Title: "每日无限次 AI 咨询", Desc: "免费版每天 3 次，Plus 想问就问"},
		{Icon: "bolt", Title: "更快的 AI 响应", Desc: "优先算力，分析快人一步"},
		{Icon: "grid", Title: "全套记账分类图标", Desc: "解锁 12 个系列、上百枚精致图标"},
		{Icon: "crown", Title: "更换 App 图标", Desc: "9 款主题图标，点亮你的桌面"},
		{Icon: "leaf", Title: "专属成长皮肤", Desc: "樱花树 / 银杏 / 极光等成长场景"},
	}
	return s.SetPerks(perks)
}

func ensureAdmin(s store.Store, email, password string) error {
	if _, err := s.GetUserByEmail(email); err == nil {
		return nil // already exists
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	admin := &models.User{
		ID:           "U-00001",
		Email:        email,
		Name:         "运营管理员",
		PasswordHash: hash,
		Role:         "admin",
		Plan:         models.PlanYear,
		AppIcon:      "classic",
		Onboarded:    true,
	}
	return s.CreateUser(admin)
}

// demoUser is a compact spec for a seeded account.
type demoUser struct {
	id      string
	email   string
	name    string
	city    string
	plan    models.Plan
	joinAgo time.Duration
	lastAgo time.Duration
}

func seedDemo(s store.Store) error {
	now := time.Now()
	hash, _ := auth.HashPassword("password")

	users := []demoUser{
		{"U-20413", "lin@lim.app", "林一", "上海", models.PlanFree, day(425), min(2)},
		{"U-19877", "su@lim.app", "苏晚", "杭州", models.PlanYear, day(470), hour(1)},
		{"U-20890", "chen@lim.app", "陈默", "深圳", models.PlanMonth, day(380), hour(6)},
		{"U-18204", "jiang@lim.app", "江予安", "北京", models.PlanYear, day(560), day(1)},
		{"U-21055", "zhou@lim.app", "周禾", "成都", models.PlanFree, day(360), day(3)},
		{"U-17630", "wu@lim.app", "吴桐", "广州", models.PlanFree, day(615), day(12)},
		{"U-20612", "zheng@lim.app", "郑清", "武汉", models.PlanMonth, day(430), hour(5)},
		{"U-16998", "he@lim.app", "何笙", "南京", models.PlanYear, day(660), hour(3)},
		{"U-21340", "xu@lim.app", "许念", "西安", models.PlanFree, day(12), min(1)},
		{"U-19120", "liang@lim.app", "梁知秋", "苏州", models.PlanFree, day(530), day(28)},
	}
	for _, d := range users {
		u := &models.User{
			ID:           d.id,
			Email:        d.email,
			Name:         d.name,
			City:         d.city,
			PasswordHash: hash,
			Role:         "user",
			Plan:         d.plan,
			AppIcon:      "classic",
			Onboarded:    true,
			CreatedAt:    now.Add(-d.joinAgo),
			LastActiveAt: now.Add(-d.lastAgo),
			Profile: models.Profile{
				MonthlyIncome: 18000, HouseholdIncome: 32000, Members: 3, Debt: 4200, MonthBudget: 6000,
			},
		}
		if u.Plan != models.PlanFree {
			until := now.Add(day(120))
			u.PlusUntil = &until
		}
		if err := s.CreateUser(u); err != nil && err != store.ErrDuplicate {
			return err
		}
	}

	// Decision feed — mirrors the prototype's DECISIONS / HISTORY tables.
	type dseed struct {
		user, uid, item, cat string
		price, impulse       int
		verdict              models.Verdict
		status               models.DecisionStatus
		ago                  time.Duration
		dims                 models.Dimensions
	}
	decisions := []dseed{
		{"苏晚", "U-19877", "索尼 WH-1000XM5 耳机", "digital", 2299, 78, models.VerdictResist, models.StatusResisted, min(2), dm(2, 1.5, 1.5, 3, 2, 3)},
		{"陈默", "U-20890", "空气炸锅", "home", 399, 32, models.VerdictBuy, models.StatusBought, min(8), dm(4.5, 4, 3.5, 4.5, 4, 3)},
		{"江予安", "U-18204", "限量联名球鞋", "fashion", 1399, 85, models.VerdictResist, models.StatusResisted, min(14), dm(1.5, 2, 1, 2.5, 2, 2.5)},
		{"周禾", "U-21055", "机械键盘 HHKB", "digital", 899, 58, models.VerdictPause, models.StatusWishlist, min(22), dm(3, 2, 2.5, 3.5, 3, 3)},
		{"何笙", "U-16998", "瑜伽垫", "hobby", 129, 28, models.VerdictBuy, models.StatusBought, min(31), dm(4, 4.5, 4, 4.5, 4.5, 4)},
		{"郑清", "U-20612", "第三支口红", "beauty", 329, 74, models.VerdictResist, models.StatusResisted, min(46), dm(2, 1.5, 1.5, 2.5, 3, 3)},
		{"林一", "U-20413", "露营折叠椅", "hobby", 269, 56, models.VerdictPause, models.StatusWishlist, hour(1), dm(3, 3, 2.5, 3, 3.5, 3.5)},
		{"许念", "U-21340", "网红小家电", "home", 599, 69, models.VerdictResist, models.StatusResisted, hour(1), dm(2.5, 2, 2, 2.5, 2.5, 3)},
		{"林一", "U-20413", "限量球鞋", "fashion", 1399, 85, models.VerdictResist, models.StatusResisted, day(8), dm(1.5, 2, 1, 2.5, 2, 2.5)},
		{"林一", "U-20413", "绿植 龟背竹", "home", 88, 22, models.VerdictBuy, models.StatusBought, day(20), dm(4, 4, 4, 4.5, 5, 4.5)},
	}
	for _, d := range decisions {
		saved := 0
		if d.status == models.StatusResisted {
			saved = d.price
		}
		decided := now.Add(-d.ago)
		dec := &models.Decision{
			UserID: d.uid, Item: d.item, Price: d.price, Cat: d.cat,
			Dims: d.dims, Impulse: d.impulse, Verdict: d.verdict,
			Message:   "（演示数据）",
			Note:      demoNote(d.verdict),
			Saved:     saved,
			Status:    d.status,
			CreatedAt: decided,
			DecidedAt: &decided,
		}
		if err := s.CreateDecision(dec); err != nil {
			return err
		}
	}

	// A couple of live cooling-off items for 林一 (U-20413).
	wl := []struct {
		item, cat        string
		price, impulse   int
		addedAgo, window time.Duration
	}{
		{"机械键盘 HHKB", "digital", 899, 64, hour(17), day(1)},
		{"露营折叠椅", "hobby", 269, 58, hour(2), day(1)},
	}
	for _, w := range wl {
		added := now.Add(-w.addedAgo)
		if err := s.AddWishlist(&models.WishlistItem{
			UserID: "U-20413", Item: w.item, Cat: w.cat, Price: w.price, Impulse: w.impulse,
			AddedAt: added, ExpiresAt: added.Add(w.window),
		}); err != nil {
			return err
		}
	}

	// Subscription transactions for the revenue dashboard.
	txns := []struct {
		user, plan     string
		amount         int
		status         string
		ago            time.Duration
	}{
		{"何笙", "年度会员", 98, "success", hour(3)},
		{"苏晚", "年度续订", 98, "success", hour(5)},
		{"陈默", "月度会员", 18, "success", hour(8)},
		{"江予安", "年度续订", 98, "success", day(1)},
		{"郑清", "月度续订", 18, "refund", day(1) + hour(3)},
		{"吴桐", "年度会员", 98, "success", day(1) + hour(7)},
	}
	for _, t := range txns {
		if err := s.AddTransaction(&models.Transaction{
			UserName: t.user, Plan: t.plan, Amount: t.amount, Status: t.status,
			CreatedAt: now.Add(-t.ago),
		}); err != nil {
			return err
		}
	}

	// Operations / push campaigns.
	pushes := []struct {
		title, seg, open, status, schedule string
		sent                               int
	}{
		{"本周你已省下 ¥860", "活跃用户", "38.2%", "scheduled", "每周日 20:00", 12640},
		{"心愿单有 1 件冷静期已到", "有心愿单", "52.6%", "sent", "今天 10:00", 3820},
		{"你的省钱树长出新叶子 🌱", "连续克制 ≥3 天", "44.1%", "sent", "昨天", 5210},
		{"Plus 限时 7 折，解锁无限咨询", "免费高活跃", "21.3%", "sent", "5/26", 8940},
		{"好久不见，要不要聊聊最近想买的？", "沉睡用户", "14.8%", "draft", "—", 6120},
	}
	for i, p := range pushes {
		if err := s.CreatePush(&models.Push{
			Title: p.title, Segment: p.seg, OpenRate: p.open, Status: p.status,
			Schedule: p.schedule, Sent: p.sent, CreatedAt: now.Add(-time.Duration(i) * 24 * time.Hour),
		}); err != nil {
			return err
		}
	}
	return nil
}

func demoNote(v models.Verdict) string {
	switch v {
	case models.VerdictBuy:
		return "高频刚需，能切实改善体验。"
	case models.VerdictPause:
		return "可加入心愿单冷静一下再决定。"
	default:
		return "情绪消费信号较强，可暂不购买。"
	}
}

// dm builds a Dimensions value from positional need/alt/emo/value/money/env scores.
func dm(need, alt, emo, value, money, env float64) models.Dimensions {
	return models.Dimensions{Need: need, Alt: alt, Emo: emo, Value: value, Money: money, Env: env}
}

func day(n int) time.Duration  { return time.Duration(n) * 24 * time.Hour }
func hour(n int) time.Duration { return time.Duration(n) * time.Hour }
func min(n int) time.Duration  { return time.Duration(n) * time.Minute }
