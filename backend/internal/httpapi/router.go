// Package httpapi implements LIM's REST API: the consumer app endpoints plus the
// admin dashboard endpoints, all served from one mux.
package httpapi

import (
	"net/http"

	"github.com/mosson9/lim/backend/internal/ai"
	"github.com/mosson9/lim/backend/internal/appstore"
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

	// Billing (StoreKit 2 verification).
	appstore       *appstore.Verifier
	productMonthly string
	productYearly  string
	allowMock      bool

	// Security.
	globalLimiter *rateLimiter
	authLimiter   *rateLimiter
	maxBody       int64
}

// ConfigureSecurity sets the per-IP rate limits and max request body size.
// globalRPM/authRPM are requests-per-minute per client IP; maxBody is in bytes.
func (a *App) ConfigureSecurity(globalRPM, authRPM int, maxBody int64) {
	a.globalLimiter = newRateLimiter(globalRPM, maxInt(globalRPM/4, 20))
	a.authLimiter = newRateLimiter(authRPM, maxInt(authRPM/2, 5))
	a.maxBody = maxBody
}

// ConfigureBilling wires App Store receipt verification and the product→plan
// mapping. allowMock keeps the dev-only mock /subscribe endpoint available.
func (a *App) ConfigureBilling(v *appstore.Verifier, monthly, yearly string, allowMock bool) {
	a.appstore = v
	a.productMonthly = monthly
	a.productYearly = yearly
	a.allowMock = allowMock
}

// NewApp constructs the API application. adminDir, when non-empty, serves the
// static admin web app from that directory at /admin/. Security defaults (rate
// limits, body cap) are applied here and can be overridden via ConfigureSecurity.
func NewApp(s store.Store, a *auth.Manager, e *ai.Engine, corsOrigin, adminDir string) *App {
	app := &App{store: s, auth: a, engine: e, corsOrigin: corsOrigin, adminDir: adminDir}
	app.ConfigureSecurity(240, 20, 1<<20) // 240 rpm global, 20 rpm auth, 1 MiB body
	return app
}

// Handler builds the fully-wired http.Handler (routes + middleware).
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	// --- Health & API docs ---
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "llm": a.engine.UsesLLM()})
	})
	mux.HandleFunc("GET /openapi.yaml", a.handleOpenAPISpec)
	mux.HandleFunc("GET /docs", a.handleDocs)

	// --- Auth (public, stricter per-IP rate limit to slow brute force) ---
	mux.HandleFunc("POST /api/v1/auth/register", a.limit(a.authLimiter, a.handleRegister))
	mux.HandleFunc("POST /api/v1/auth/login", a.limit(a.authLimiter, a.handleLogin))

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
	mux.HandleFunc("POST /api/v1/subscription/verify", a.authenticate(a.handleVerifySubscription))
	// App Store Server Notifications V2 webhook (public; authenticity = JWS signature).
	mux.HandleFunc("POST /api/v1/appstore/notifications", a.handleAppStoreNotification)

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

	// Middleware chain (outermost first): recover → security headers → CORS
	// (handles OPTIONS early) → logging → body cap → per-IP rate limit → routes.
	var h http.Handler = mux
	h = a.rateLimitAll(h)
	h = limitBody(a.maxBody, h)
	h = logging(h)
	h = a.withCORS(h)
	h = securityHeaders(h)
	h = requestID(h)
	h = recoverMiddleware(h)
	return h
}

// rateLimitAll enforces the global per-IP limiter across every route.
func (a *App) rateLimitAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.globalLimiter != nil && !a.globalLimiter.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "10")
			writeError(w, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			return
		}
		next.ServeHTTP(w, r)
	})
}
