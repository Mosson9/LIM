// Package models defines the core domain entities for the LIM backend.
//
// LIM (Less is More) is an anti-consumerism companion. The data model mirrors
// the product flow captured in the design mind-map: a user asks the AI whether
// they should buy something, the AI scores it across six dimensions and returns
// an "impulse index" plus a verdict (buy / pause / resist). Resisted purchases
// turn into savings and a growing tree; deferred ones land in a 24h cooling-off
// wishlist; confirmed ones are recorded as spending.
package models

import "time"

// Verdict is the final recommendation produced by the analysis engine.
type Verdict string

const (
	VerdictBuy    Verdict = "buy"    // 可以拥有它
	VerdictPause  Verdict = "pause"  // 再想想 / 冷静一下
	VerdictResist Verdict = "resist" // 也许不必买
)

// DecisionStatus tracks where a decision ended up after the user acted on it.
type DecisionStatus string

const (
	StatusPending  DecisionStatus = "pending"  // analysed, awaiting the user's choice
	StatusResisted DecisionStatus = "resisted" // 忍住了 — counts as savings
	StatusBought   DecisionStatus = "bought"   // 买了 — counts as spending
	StatusWishlist DecisionStatus = "wishlist" // moved to the cooling-off list
)

// Plan is a subscription tier.
type Plan string

const (
	PlanFree  Plan = "free"
	PlanMonth Plan = "month"
	PlanYear  Plan = "year"
)

// Dimensions holds the six 1–5 sub-scores. A higher score means the purchase is
// more justified on that axis (so a low total drives a high impulse index).
type Dimensions struct {
	Need  float64 `json:"need"`  // 需求 — is it actually necessary?
	Alt   float64 `json:"alt"`   // 替代 — could an existing/rented/shared item do?
	Emo   float64 `json:"emo"`   // 情感 — driven by genuine want or ads/pressure?
	Value float64 `json:"value"` // 长期价值 — frequency, lifespan, lasting value
	Money float64 `json:"money"` // 经济 — does it fit the budget without crowding out?
	Env   float64 `json:"env"`   // 环境 — production/transport/use footprint
}

// User is an authenticated account.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // "user" or "admin"
	City         string    `json:"city,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`

	// Embedded financial profile (collected during onboarding, step 2 of the map).
	Profile Profile `json:"profile"`

	// Subscription / entitlement.
	Plan         Plan       `json:"plan"`
	PlusUntil    *time.Time `json:"plus_until,omitempty"`
	AIUsedToday  int        `json:"ai_used_today"`
	AIUsedDate   string     `json:"-"` // YYYY-MM-DD bucket for the daily counter
	AppIcon      string     `json:"app_icon"`
	Onboarded    bool       `json:"onboarded"`
}

// IsPlus reports whether the user currently has an active paid entitlement.
func (u *User) IsPlus() bool {
	if u.Plan == PlanFree {
		return false
	}
	if u.PlusUntil == nil {
		return u.Plan == PlanMonth || u.Plan == PlanYear
	}
	return u.PlusUntil.After(time.Now())
}

// Profile captures the household economics that make advice more personal.
type Profile struct {
	MonthlyIncome   int `json:"monthly_income"`   // 个人月收入
	HouseholdIncome int `json:"household_income"` // 家庭月收入
	Members         int `json:"members"`          // 家庭成员数
	Debt            int `json:"debt"`             // 每月固定负债 / 月供
	MonthBudget     int `json:"month_budget"`     // 可支配预算
}

// Decision is one "should I buy this?" enquiry and its outcome.
type Decision struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Item      string         `json:"item"`
	Price     int            `json:"price"`
	Cat       string         `json:"cat"`
	Reason    string         `json:"reason,omitempty"`
	Dims      Dimensions     `json:"dims"`
	Impulse   int            `json:"impulse"` // 0–100 冲动指数
	Verdict   Verdict        `json:"verdict"`
	Message   string         `json:"message"` // LIM's natural-language note
	Note      string         `json:"note"`    // short one-line summary for lists
	Saved     int            `json:"saved"`   // amount saved if resisted (= price)
	Status    DecisionStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	DecidedAt *time.Time     `json:"decided_at,omitempty"`
}

// WishlistItem is a purchase parked in the 24h cooling-off period.
type WishlistItem struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	DecisionID string    `json:"decision_id,omitempty"`
	Item       string    `json:"item"`
	Price      int       `json:"price"`
	Cat        string    `json:"cat"`
	Impulse    int       `json:"impulse"`
	AddedAt    time.Time `json:"added_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// Progress is the 0..1 fraction of the cooling-off window elapsed.
func (w *WishlistItem) Progress(now time.Time) float64 {
	total := w.ExpiresAt.Sub(w.AddedAt).Seconds()
	if total <= 0 {
		return 1
	}
	p := now.Sub(w.AddedAt).Seconds() / total
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

// Transaction records a subscription payment (mock payment gateway).
type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Plan      string    `json:"plan"`
	Amount    int       `json:"amount"`
	Status    string    `json:"status"` // success | refund
	CreatedAt time.Time `json:"created_at"`
}

// --- Configuration / content entities (admin-managed) ---

// DimensionConfig is an admin-tunable weight for one analysis axis.
type DimensionConfig struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Desc    string  `json:"desc"`
	Weight  float64 `json:"weight"`
	Enabled bool    `json:"enabled"`
	// Q is the reflective question shown to the user for this dimension.
	Q string `json:"q"`
}

// AIConfig is the live configuration of the analysis engine.
type AIConfig struct {
	Dims           []DimensionConfig `json:"dims"`
	PromptTemplate string            `json:"prompt_template"`
	Persona        string            `json:"persona"` // 温柔朋友 | 理性顾问 | 犀利毒舌
	Model          string            `json:"model"`
	FreeDailyLimit int               `json:"free_daily_limit"`
	CoolingHours   int               `json:"cooling_hours"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// Category is a spending category / icon.
type Category struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"`
	Soft  string `json:"soft"`
	Icon  string `json:"icon"`
	Free  bool   `json:"free"`
	Count int    `json:"count"` // analytics: consults this month
}

// Skin is a swappable home-screen app icon.
type Skin struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	BG   string `json:"bg"`
	Dark bool   `json:"dark"`
	Free bool   `json:"free"`
	Used string `json:"used,omitempty"` // analytics: adoption %
}

// PlusPerk is a marketing bullet on the subscription page.
type PlusPerk struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// PlanOption is a purchasable subscription tier shown on the paywall.
type PlanOption struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Per   string `json:"per"`
	Note  string `json:"note"`
	Best  bool   `json:"best"`
}

// Push is an operations/engagement campaign.
type Push struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Segment   string    `json:"segment"`
	Sent      int       `json:"sent"`
	OpenRate  string    `json:"open_rate"`
	Status    string    `json:"status"` // scheduled | sent | draft
	Schedule  string    `json:"schedule"`
	CreatedAt time.Time `json:"created_at"`
}
