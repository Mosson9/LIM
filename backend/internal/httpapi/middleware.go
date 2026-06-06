package httpapi

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
	"github.com/mosson9/lim/backend/internal/store"
)

type ctxKey string

const userCtxKey ctxKey = "user"

// withCORS adds permissive CORS headers for the admin web app and handles
// preflight requests.
func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", a.corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logging logs each request with its status and latency.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(sr, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sr.status, time.Since(start).Round(time.Millisecond))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// authenticate is middleware that requires a valid bearer token and injects the
// resolved user into the request context.
func (a *App) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.userFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "未授权，请重新登录")
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, u)
		next(w, r.WithContext(ctx))
	}
}

// requireAdmin is middleware that additionally requires the admin role.
func (a *App) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return a.authenticate(func(w http.ResponseWriter, r *http.Request) {
		u := userFromCtx(r)
		if u == nil || u.Role != "admin" {
			writeError(w, http.StatusForbidden, "需要管理员权限")
			return
		}
		next(w, r)
	})
}

// userFromRequest resolves the bearer token to a user.
func (a *App) userFromRequest(r *http.Request) (*models.User, error) {
	h := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer"))
	if token == "" {
		return nil, store.ErrNotFound
	}
	claims, err := a.auth.Parse(token)
	if err != nil {
		return nil, err
	}
	return a.store.GetUser(claims.UserID)
}

// userFromCtx returns the authenticated user stored on the request context.
func userFromCtx(r *http.Request) *models.User {
	u, _ := r.Context().Value(userCtxKey).(*models.User)
	return u
}
