package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
)

// wishlistView decorates a stored item with computed fields for the UI.
type wishlistView struct {
	models.WishlistItem
	Progress  float64 `json:"progress"`   // 0..1 of the cooling-off window elapsed
	Expired   bool    `json:"expired"`    // window fully elapsed
	RemainSec int64   `json:"remain_sec"` // seconds left (>=0)
}

func toWishlistView(w *models.WishlistItem, now time.Time) wishlistView {
	remain := int64(w.ExpiresAt.Sub(now).Seconds())
	if remain < 0 {
		remain = 0
	}
	return wishlistView{
		WishlistItem: *w,
		Progress:     w.Progress(now),
		Expired:      !now.Before(w.ExpiresAt),
		RemainSec:    remain,
	}
}

func (a *App) handleListWishlist(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	now := time.Now()
	items := a.store.ListWishlist(u.ID)
	out := make([]wishlistView, 0, len(items))
	for _, it := range items {
		out = append(out, toWishlistView(it, now))
	}
	writeJSON(w, http.StatusOK, out)
}

type addWishlistReq struct {
	Item       string `json:"item"`
	Price      int    `json:"price"`
	Cat        string `json:"cat"`
	Impulse    int    `json:"impulse"`
	DecisionID string `json:"decision_id,omitempty"`
}

func (a *App) handleAddWishlist(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	var req addWishlistReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	req.Item = strings.TrimSpace(req.Item)
	if req.Item == "" || req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "缺少必要字段")
		return
	}
	if req.Cat == "" {
		req.Cat = "other"
	}
	hours := a.store.AIConfig().CoolingHours
	if hours <= 0 {
		hours = 24
	}
	now := time.Now()
	item := &models.WishlistItem{
		UserID:     u.ID,
		DecisionID: req.DecisionID,
		Item:       req.Item,
		Price:      req.Price,
		Cat:        req.Cat,
		Impulse:    req.Impulse,
		AddedAt:    now,
		ExpiresAt:  now.Add(time.Duration(hours) * time.Hour),
	}
	if err := a.store.AddWishlist(item); err != nil {
		writeError(w, http.StatusInternalServerError, "加入心愿单失败")
		return
	}
	// Mark the originating decision as parked in cooling-off.
	a.resolvePending(req.DecisionID, u.ID, models.StatusWishlist)
	writeJSON(w, http.StatusCreated, toWishlistView(item, now))
}

// handleResolveWishlist finalises a cooling-off item: "resist" credits savings,
// "buy" records a purchase. The linked decision (if any) is updated to match.
func (a *App) handleResolveWishlist(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	item, err := a.store.GetWishlistItem(r.PathValue("id"))
	if err != nil || item.UserID != u.ID {
		writeError(w, http.StatusNotFound, "未找到心愿单条目")
		return
	}
	var req resolveReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}

	var status models.DecisionStatus
	switch req.Action {
	case "resist":
		status = models.StatusResisted
	case "buy":
		status = models.StatusBought
	default:
		writeError(w, http.StatusBadRequest, "action 必须为 resist 或 buy")
		return
	}

	if item.DecisionID != "" {
		a.resolvePending(item.DecisionID, u.ID, status)
	} else {
		// No source decision (added manually) — synthesise a resolved record so
		// it shows up in history and stats.
		now := time.Now()
		saved := 0
		if status == models.StatusResisted {
			saved = item.Price
		}
		_ = a.store.CreateDecision(&models.Decision{
			UserID: u.ID, Item: item.Item, Price: item.Price, Cat: item.Cat,
			Verdict: models.VerdictPause, Impulse: item.Impulse,
			Note: "冷静期后的决定", Saved: saved, Status: status,
			CreatedAt: now, DecidedAt: &now,
		})
	}
	if err := a.store.DeleteWishlist(item.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "action": req.Action, "saved": status == models.StatusResisted})
}

func (a *App) handleDeleteWishlist(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	item, err := a.store.GetWishlistItem(r.PathValue("id"))
	if err != nil || item.UserID != u.ID {
		writeError(w, http.StatusNotFound, "未找到心愿单条目")
		return
	}
	if err := a.store.DeleteWishlist(item.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
