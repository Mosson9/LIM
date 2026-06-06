package httpapi

import (
	"net/http"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
)

// subscriptionResp describes the caller's current entitlement.
type subscriptionResp struct {
	Plan      models.Plan          `json:"plan"`
	IsPlus    bool                 `json:"is_plus"`
	PlusUntil *time.Time           `json:"plus_until,omitempty"`
	Plans     []models.PlanOption  `json:"plans"`
	Perks     []models.PlusPerk    `json:"perks"`
}

func (a *App) handleSubscription(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	writeJSON(w, http.StatusOK, subscriptionResp{
		Plan:      u.Plan,
		IsPlus:    u.IsPlus(),
		PlusUntil: u.PlusUntil,
		Plans:     a.store.Plans(),
		Perks:     a.store.Perks(),
	})
}

type subscribeReq struct {
	Plan string `json:"plan"` // "month" | "year"
}

// handleSubscribe simulates a successful purchase: it grants the entitlement,
// extends PlusUntil, and records a transaction for the revenue dashboard.
//
// In production this endpoint would verify an App Store / StoreKit receipt
// (see docs/DEPLOYMENT.md) before granting; the grant logic is otherwise identical.
func (a *App) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if !a.allowMock {
		writeError(w, http.StatusForbidden, "模拟开通已禁用，请使用 /subscription/verify 提交 App Store 凭证")
		return
	}
	u := userFromCtx(r)
	var req subscribeReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}

	var (
		plan     models.Plan
		amount   int
		extend   time.Duration
		planName string
	)
	switch req.Plan {
	case "month":
		plan, amount, extend, planName = models.PlanMonth, 18, 30*24*time.Hour, "月度会员"
	case "year":
		plan, amount, extend, planName = models.PlanYear, 98, 365*24*time.Hour, "年度会员"
	default:
		writeError(w, http.StatusBadRequest, "plan 必须为 month 或 year")
		return
	}

	now := time.Now()
	base := now
	if u.PlusUntil != nil && u.PlusUntil.After(now) {
		base = *u.PlusUntil // stack onto remaining time
	}
	until := base.Add(extend)
	u.Plan = plan
	u.PlusUntil = &until
	if err := a.store.UpdateUser(u); err != nil {
		writeError(w, http.StatusInternalServerError, "开通失败")
		return
	}
	_ = a.store.AddTransaction(&models.Transaction{
		UserID: u.ID, UserName: u.Name, Plan: planName, Amount: amount,
		Status: "success", CreatedAt: now,
	})

	writeJSON(w, http.StatusOK, subscriptionResp{
		Plan: u.Plan, IsPlus: u.IsPlus(), PlusUntil: u.PlusUntil,
		Plans: a.store.Plans(), Perks: a.store.Perks(),
	})
}

type verifyReq struct {
	// JWS is StoreKit 2's `Transaction.jwsRepresentation` from the client.
	JWS string `json:"jws"`
}

// handleVerifySubscription validates an Apple-signed StoreKit 2 transaction and
// grants the entitlement based on the product and its expiry. This is the
// production purchase path (the mock /subscribe is dev-only).
func (a *App) handleVerifySubscription(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if a.appstore == nil {
		writeError(w, http.StatusNotImplemented, "未配置 App Store 校验")
		return
	}
	var req verifyReq
	if err := decodeJSON(r, &req); err != nil || req.JWS == "" {
		writeError(w, http.StatusBadRequest, "缺少 jws 凭证")
		return
	}

	txn, err := a.appstore.Verify(req.JWS)
	if err != nil {
		writeError(w, http.StatusBadRequest, "凭证校验失败："+err.Error())
		return
	}

	var (
		plan     models.Plan
		amount   int
		planName string
	)
	switch txn.ProductID {
	case a.productMonthly:
		plan, amount, planName = models.PlanMonth, 18, "月度会员"
	case a.productYearly:
		plan, amount, planName = models.PlanYear, 98, "年度会员"
	default:
		writeError(w, http.StatusBadRequest, "未知的商品："+txn.ProductID)
		return
	}

	// Entitlement runs until Apple's expiry (fall back to a sane window if absent).
	until := txn.ExpiresDate
	if until.IsZero() {
		if plan == models.PlanYear {
			until = time.Now().Add(365 * 24 * time.Hour)
		} else {
			until = time.Now().Add(30 * 24 * time.Hour)
		}
	}
	if until.Before(time.Now()) {
		writeError(w, http.StatusBadRequest, "订阅已过期")
		return
	}

	u.Plan = plan
	u.PlusUntil = &until
	if err := a.store.UpdateUser(u); err != nil {
		writeError(w, http.StatusInternalServerError, "开通失败")
		return
	}
	_ = a.store.AddTransaction(&models.Transaction{
		UserID: u.ID, UserName: u.Name, Plan: planName, Amount: amount,
		Status: "success", CreatedAt: time.Now(),
	})

	writeJSON(w, http.StatusOK, subscriptionResp{
		Plan: u.Plan, IsPlus: u.IsPlus(), PlusUntil: u.PlusUntil,
		Plans: a.store.Plans(), Perks: a.store.Perks(),
	})
}
