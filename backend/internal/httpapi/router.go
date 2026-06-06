// Package httpapi implements LIM's REST API: the consumer app endpoints plus the
// admin dashboard endpoints, all served from one mux.
package httpapi

import (
	"net/http"

	"github.com/mosson9/lim/backend/internal/ai"
	"github.com/mosson9/lim/backend/internal/auth"
	"github.com/mosson9/lim/backend/internal/store"
)

// App carries the shared dependencies for every handler.
type App struct {
	store      store.Store
	auth       *auth.Manager
	engine     *ai.Engine
	corsOrigin string
	adminDir   string
}

// NewApp constructs the API application. adminDir, when non-empty, serves the
// static admin web app from that directory at /admin/.
func NewApp(s store.Store, a *auth.Manager, e *ai.Engine, corsOrigin, adminDir string) *App {
	return &App{store: s, auth: a, engine: e, corsOrigin: corsOrigin, adminDir: adminDir}
}

// Handler builds the fully-wired http.Handler (routes + middleware).
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	// --- Health ---
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "llm": a.engine.UsesLLM()})
	})

	// --- Auth (public) ---
	mux.HandleFunc("POST /api/v1/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", a.handleLogin)

	// --- Catalogue / meta (public) ---
	mux.HandleFunc("GET /api/v1/meta/categories", a.handleCategories)
	mux.HandleFunc("GET /api/v1/meta/skins", a.handleSkins)
	mux.HandleFunc("GET /api/v1/meta/plans", a.handlePlans)
	mux.HandleFunc("GET /api/v1/meta/perks", a.handlePerks)
	mux.HandleFunc("GET /api/v1/meta/dimensions", a.handleDimensions)

	// --- Current user / profile ---
	mux.HandleFunc("GET /api/v1/me", a.authenticate(a.handleMe))
	mux.HandleFunc("PUT /api/v1/me/profile", a.authenticate(a.handleUpdateProfile))
	mux.HandleFunc("POST /api/v1/me/onboard", a.authenticate(a.handleOnboard))
	mux.HandleFunc("PUT /api/v1/me/app-icon", a.authenticate(a.handleSetAppIcon))

	// --- Core flow: analyse + decisions ---
	mux.HandleFunc("POST /api/v1/analyze", a.authenticate(a.handleAnalyze))
	mux.HandleFunc("GET /api/v1/decisions", a.authenticate(a.handleListDecisions))
	mux.HandleFunc("GET /api/v1/decisions/{id}", a.authenticate(a.handleGetDecision))
	mux.HandleFunc("POST /api/v1/decisions/{id}/resolve", a.authenticate(a.handleResolveDecision))

	// --- Wishlist (24h cooling-off) ---
	mux.HandleFunc("GET /api/v1/wishlist", a.authenticate(a.handleListWishlist))
	mux.HandleFunc("POST /api/v1/wishlist", a.authenticate(a.handleAddWishlist))
	mux.HandleFunc("POST /api/v1/wishlist/{id}/resolve", a.authenticate(a.handleResolveWishlist))
	mux.HandleFunc("DELETE /api/v1/wishlist/{id}", a.authenticate(a.handleDeleteWishlist))

	// --- Stats ---
	mux.HandleFunc("GET /api/v1/stats", a.authenticate(a.handleStats))

	// --- Subscription ---
	mux.HandleFunc("GET /api/v1/subscription", a.authenticate(a.handleSubscription))
	mux.HandleFunc("POST /api/v1/subscription/subscribe", a.authenticate(a.handleSubscribe))

	// --- Admin ---
	mux.HandleFunc("GET /api/v1/admin/overview", a.requireAdmin(a.handleAdminOverview))
	mux.HandleFunc("GET /api/v1/admin/users", a.requireAdmin(a.handleAdminUsers))
	mux.HandleFunc("GET /api/v1/admin/users/{id}", a.requireAdmin(a.handleAdminUser))
	mux.HandleFunc("GET /api/v1/admin/decisions", a.requireAdmin(a.handleAdminDecisions))
	mux.HandleFunc("GET /api/v1/admin/subscriptions", a.requireAdmin(a.handleAdminSubscriptions))
	mux.HandleFunc("GET /api/v1/admin/ai-config", a.requireAdmin(a.handleAdminGetAIConfig))
	mux.HandleFunc("PUT /api/v1/admin/ai-config", a.requireAdmin(a.handleAdminSetAIConfig))
	mux.HandleFunc("GET /api/v1/admin/categories", a.requireAdmin(a.handleAdminCategories))
	mux.HandleFunc("PUT /api/v1/admin/categories", a.requireAdmin(a.handleAdminSetCategories))
	mux.HandleFunc("GET /api/v1/admin/skins", a.requireAdmin(a.handleAdminSkins))
	mux.HandleFunc("GET /api/v1/admin/push", a.requireAdmin(a.handleAdminListPush))
	mux.HandleFunc("POST /api/v1/admin/push", a.requireAdmin(a.handleAdminCreatePush))

	// --- Optional: serve the static admin web app at /admin/ (same origin) ---
	if a.adminDir != "" {
		fs := http.StripPrefix("/admin/", http.FileServer(http.Dir(a.adminDir)))
		mux.Handle("GET /admin/", fs)
		mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
		})
	}

	return logging(a.withCORS(mux))
}
