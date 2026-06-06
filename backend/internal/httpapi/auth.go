package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/mosson9/lim/backend/internal/auth"
	"github.com/mosson9/lim/backend/internal/models"
	"github.com/mosson9/lim/backend/internal/store"
)

type registerReq struct {
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Name     string          `json:"name"`
	Profile  *models.Profile `json:"profile,omitempty"`
}

type authResp struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !strings.Contains(req.Email, "@") || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "邮箱不合法或密码少于 6 位")
		return
	}
	if req.Name == "" {
		req.Name = "LIM 用户"
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "服务器错误")
		return
	}
	u := &models.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hash,
		Role:         "user",
		Plan:         models.PlanFree,
		AppIcon:      "classic",
	}
	if req.Profile != nil {
		u.Profile = *req.Profile
	}
	if err := a.store.CreateUser(u); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "该邮箱已注册")
			return
		}
		writeError(w, http.StatusInternalServerError, "创建用户失败")
		return
	}
	a.issueAndRespond(w, u)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	u, err := a.store.GetUserByEmail(strings.TrimSpace(req.Email))
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "邮箱或密码错误")
		return
	}
	a.store.TouchUser(u.ID)
	a.issueAndRespond(w, u)
}

func (a *App) issueAndRespond(w http.ResponseWriter, u *models.User) {
	token, err := a.auth.Issue(u.ID, u.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	u.PasswordHash = ""
	writeJSON(w, http.StatusOK, authResp{Token: token, User: u})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	a.refreshDailyQuota(u)
	u.PasswordHash = ""
	writeJSON(w, http.StatusOK, u)
}
