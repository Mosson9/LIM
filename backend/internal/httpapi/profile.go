package httpapi

import (
	"net/http"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
)

// refreshDailyQuota resets the per-day AI usage counter when the date rolls over.
// It mutates the passed-in user and persists if a reset occurred.
func (a *App) refreshDailyQuota(u *models.User) {
	today := time.Now().Format("2006-01-02")
	if u.AIUsedDate != today {
		u.AIUsedDate = today
		u.AIUsedToday = 0
		_ = a.store.UpdateUser(u)
	}
}

func (a *App) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	var p models.Profile
	if err := decodeJSON(r, &p); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	// Clamp to sane non-negative values.
	if p.Members < 1 {
		p.Members = 1
	}
	if p.MonthBudget <= 0 {
		// Derive a default disposable budget from income net of debt if unset.
		p.MonthBudget = maxInt(1000, (p.MonthlyIncome-p.Debt)/3)
	}
	u.Profile = p
	if err := a.store.UpdateUser(u); err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	u.PasswordHash = ""
	writeJSON(w, http.StatusOK, u)
}

func (a *App) handleOnboard(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	u.Onboarded = true
	if err := a.store.UpdateUser(u); err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	u.PasswordHash = ""
	writeJSON(w, http.StatusOK, u)
}

type appIconReq struct {
	AppIcon string `json:"app_icon"`
}

func (a *App) handleSetAppIcon(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	var req appIconReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	// Validate the icon exists; Plus-only skins require an active subscription.
	var ok bool
	for _, s := range a.store.Skins() {
		if s.ID == req.AppIcon {
			ok = true
			if !s.Free && !u.IsPlus() {
				writeError(w, http.StatusPaymentRequired, "该图标为 Plus 专属")
				return
			}
		}
	}
	if !ok {
		writeError(w, http.StatusBadRequest, "未知图标")
		return
	}
	u.AppIcon = req.AppIcon
	if err := a.store.UpdateUser(u); err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	u.PasswordHash = ""
	writeJSON(w, http.StatusOK, u)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
