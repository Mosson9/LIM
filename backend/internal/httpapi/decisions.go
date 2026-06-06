package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/mosson9/lim/backend/internal/ai"
	"github.com/mosson9/lim/backend/internal/models"
)

type analyzeReq struct {
	Item   string `json:"item"`
	Price  int    `json:"price"`
	Cat    string `json:"cat"`
	Reason string `json:"reason"`
}

// handleAnalyze runs the six-dimension analysis and persists a *pending*
// decision. The app then calls /decisions/{id}/resolve once the user acts.
//
// Free users are limited to AIConfig.FreeDailyLimit analyses per day.
func (a *App) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	a.refreshDailyQuota(u)

	cfg := a.store.AIConfig()
	if !u.IsPlus() && cfg.FreeDailyLimit > 0 && u.AIUsedToday >= cfg.FreeDailyLimit {
		writeError(w, http.StatusPaymentRequired, "今日免费咨询次数已用完，升级 Plus 可无限使用")
		return
	}

	var req analyzeReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	req.Item = strings.TrimSpace(req.Item)
	if req.Item == "" || req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "请填写想买的东西和价格")
		return
	}
	if req.Cat == "" {
		req.Cat = "other"
	}

	res := a.engine.Analyze(r.Context(), ai.Input{
		Item: req.Item, Price: req.Price, Cat: req.Cat, Reason: req.Reason,
	}, u.Profile, cfg)

	dec := &models.Decision{
		UserID:  u.ID,
		Item:    req.Item,
		Price:   req.Price,
		Cat:     req.Cat,
		Reason:  req.Reason,
		Dims:    res.Dims,
		Impulse: res.Impulse,
		Verdict: res.Verdict,
		Message: res.Message,
		Note:    res.Note,
		Status:  models.StatusPending,
	}
	if err := a.store.CreateDecision(dec); err != nil {
		writeError(w, http.StatusInternalServerError, "保存分析失败")
		return
	}

	// Count the consultation against the daily quota.
	u.AIUsedToday++
	u.AIUsedDate = time.Now().Format("2006-01-02")
	_ = a.store.UpdateUser(u)

	writeJSON(w, http.StatusOK, dec)
}

func (a *App) handleListDecisions(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	filter := r.URL.Query().Get("filter")
	writeJSON(w, http.StatusOK, a.store.ListDecisions(u.ID, filter))
}

func (a *App) handleGetDecision(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	d, err := a.store.GetDecision(r.PathValue("id"))
	if err != nil || d.UserID != u.ID {
		writeError(w, http.StatusNotFound, "未找到该决策")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

type resolveReq struct {
	// Action is the user's final choice: "resist" (忍住了) or "buy" (买了).
	Action string `json:"action"`
}

// handleResolveDecision finalises a pending decision. Resisting credits savings;
// buying records spending.
func (a *App) handleResolveDecision(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	d, err := a.store.GetDecision(r.PathValue("id"))
	if err != nil || d.UserID != u.ID {
		writeError(w, http.StatusNotFound, "未找到该决策")
		return
	}
	var req resolveReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	now := time.Now()
	switch req.Action {
	case "resist":
		d.Status = models.StatusResisted
		d.Saved = d.Price
	case "buy":
		d.Status = models.StatusBought
		d.Saved = 0
	default:
		writeError(w, http.StatusBadRequest, "action 必须为 resist 或 buy")
		return
	}
	d.DecidedAt = &now
	if err := a.store.UpdateDecision(d); err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// resolvePending is a small helper used by the wishlist flow to mark the source
// decision as resolved when an item leaves cooling-off.
func (a *App) resolvePending(decisionID, userID string, status models.DecisionStatus) {
	if decisionID == "" {
		return
	}
	d, err := a.store.GetDecision(decisionID)
	if err != nil || d.UserID != userID {
		return
	}
	now := time.Now()
	d.Status = status
	if status == models.StatusResisted {
		d.Saved = d.Price
	} else {
		d.Saved = 0
	}
	d.DecidedAt = &now
	_ = a.store.UpdateDecision(d)
}
