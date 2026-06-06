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

// planForProduct maps a StoreKit product id to a plan, monthly-equivalent
// amount, and display name. ok is false for unknown products.
func (a *App) planForProduct(productID string) (plan models.Plan, amount int, name string, ok bool) {
	switch productID {
	case a.productMonthly:
		return models.PlanMonth, 18, "月度会员", true
	case a.productYearly:
		return models.PlanYear, 98, "年度会员", true
	default:
		return models.PlanFree, 0, "", false
	}
}

// entitlementUntil returns Apple's expiry, or a sane fallback window if absent.
func entitlementUntil(expires time.Time, plan models.Plan) time.Time {
	if !expires.IsZero() {
		return expires
	}
	if plan == models.PlanYear {
		return time.Now().Add(365 * 24 * time.Hour)
	}
	return time.Now().Add(30 * 24 * time.Hour)
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

	plan, amount, planName, ok := a.planForProduct(txn.ProductID)
	if !ok {
		writeError(w, http.StatusBadRequest, "未知的商品："+txn.ProductID)
		return
	}

	until := entitlementUntil(txn.ExpiresDate, plan)
	if until.Before(time.Now()) {
		writeError(w, http.StatusBadRequest, "订阅已过期")
		return
	}

	u.Plan = plan
	u.PlusUntil = &until
	// Link the account to its subscription for renewal/refund notifications.
	if txn.OriginalTransactionID != "" {
		u.AppleOriginalTransactionID = txn.OriginalTransactionID
	}
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

type notificationReq struct {
	// SignedPayload is the App Store Server Notification V2 JWS.
	SignedPayload string `json:"signedPayload"`
}

// handleAppStoreNotification receives App Store Server Notifications V2 (renewals,
// expirations, refunds, …). It is public — authenticity comes from the JWS
// signature, not a bearer token. The account is located via the subscription's
// originalTransactionId (stored at purchase time). Always returns 200 once the
// payload is verified so Apple doesn't retry for cases we intentionally ignore.
func (a *App) handleAppStoreNotification(w http.ResponseWriter, r *http.Request) {
	if a.appstore == nil {
		writeError(w, http.StatusNotImplemented, "未配置 App Store 校验")
		return
	}
	var req notificationReq
	if err := decodeJSON(r, &req); err != nil || req.SignedPayload == "" {
		writeError(w, http.StatusBadRequest, "缺少 signedPayload")
		return
	}
	n, err := a.appstore.VerifyNotification(req.SignedPayload)
	if err != nil {
		// 400 → Apple retries (could be transient / misconfiguration).
		writeError(w, http.StatusBadRequest, "通知校验失败："+err.Error())
		return
	}
	if n.Transaction == nil || n.Transaction.OriginalTransactionID == "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ignored": "no transaction"})
		return
	}
	u, err := a.store.GetUserByOriginalTransactionID(n.Transaction.OriginalTransactionID)
	if err != nil {
		// Unknown subscription (e.g. account not linked yet) — acknowledge.
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ignored": "unknown subscription"})
		return
	}

	now := time.Now()
	switch n.Type {
	case "SUBSCRIBED", "DID_RENEW", "DID_RECOVER", "OFFER_REDEEMED", "RESUBSCRIBE":
		if plan, amount, planName, ok := a.planForProduct(n.Transaction.ProductID); ok {
			until := entitlementUntil(n.Transaction.ExpiresDate, plan)
			u.Plan = plan
			u.PlusUntil = &until
			_ = a.store.UpdateUser(u)
			if n.Type == "DID_RENEW" {
				_ = a.store.AddTransaction(&models.Transaction{
					UserID: u.ID, UserName: u.Name, Plan: planName + "（续订）",
					Amount: amount, Status: "success", CreatedAt: now,
				})
			}
		}
	case "EXPIRED", "REVOKE", "GRACE_PERIOD_EXPIRED":
		past := now.Add(-time.Minute)
		u.Plan = models.PlanFree
		u.PlusUntil = &past
		_ = a.store.UpdateUser(u)
	case "REFUND":
		past := now.Add(-time.Minute)
		u.Plan = models.PlanFree
		u.PlusUntil = &past
		_ = a.store.UpdateUser(u)
		if _, amount, planName, ok := a.planForProduct(n.Transaction.ProductID); ok {
			_ = a.store.AddTransaction(&models.Transaction{
				UserID: u.ID, UserName: u.Name, Plan: planName + "（退款）",
				Amount: amount, Status: "refund", CreatedAt: now,
			})
		}
	default:
		// DID_CHANGE_RENEWAL_STATUS, PRICE_INCREASE, etc. — no entitlement change.
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "type": n.Type})
}
