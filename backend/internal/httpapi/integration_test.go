package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/mosson9/lim/backend/internal/ai"
	"github.com/mosson9/lim/backend/internal/auth"
	"github.com/mosson9/lim/backend/internal/httpapi"
	"github.com/mosson9/lim/backend/internal/seed"
	"github.com/mosson9/lim/backend/internal/store"
)

// newServer builds a fully-wired test server backed by a temp-file store,
// seeded with catalogue content and a bootstrap admin (no demo data).
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	st, err := store.OpenFile(filepath.Join(t.TempDir(), "data.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := seed.Run(st, false, "admin@lim.app", "admin123"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	app := httpapi.NewApp(st, auth.New("test-secret", time.Hour), ai.New("", ""), "*", "")
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	return srv
}

// do is a tiny JSON HTTP helper for tests.
func do(t *testing.T, srv *httptest.Server, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, srv.URL+path, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	data, _ := io.ReadAll(resp.Body)
	if len(data) > 0 && data[0] == '{' {
		_ = json.Unmarshal(data, &out)
	}
	return resp.StatusCode, out
}

// TestFullUserFlow walks the core journey: register → analyze → resist → stats.
func TestFullUserFlow(t *testing.T) {
	srv := newServer(t)

	// Register a new user.
	code, body := do(t, srv, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "flow@test.com", "password": "secret1", "name": "测试君",
	})
	if code != http.StatusOK {
		t.Fatalf("register: got %d, body %v", code, body)
	}
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatal("register: missing token")
	}

	// Analyse a clear impulse buy.
	code, dec := do(t, srv, "POST", "/api/v1/analyze", token, map[string]any{
		"item": "限量联名球鞋", "price": 1599, "cat": "fashion",
		"reason": "直播间打折，已经有好几双了，就是想要",
	})
	if code != http.StatusOK {
		t.Fatalf("analyze: got %d, body %v", code, dec)
	}
	if dec["verdict"] != "resist" {
		t.Errorf("analyze: expected resist verdict, got %v (impulse %v)", dec["verdict"], dec["impulse"])
	}
	if dec["status"] != "pending" {
		t.Errorf("analyze: expected pending status, got %v", dec["status"])
	}
	decID, _ := dec["id"].(string)

	// Resolve it as resisted → savings credited.
	code, resolved := do(t, srv, "POST", "/api/v1/decisions/"+decID+"/resolve", token, map[string]any{"action": "resist"})
	if code != http.StatusOK {
		t.Fatalf("resolve: got %d, body %v", code, resolved)
	}
	if resolved["status"] != "resisted" {
		t.Errorf("resolve: expected resisted, got %v", resolved["status"])
	}

	// Stats should reflect the saved amount.
	code, stats := do(t, srv, "GET", "/api/v1/stats", token, nil)
	if code != http.StatusOK {
		t.Fatalf("stats: got %d", code)
	}
	if got := stats["total_saved"].(float64); got != 1599 {
		t.Errorf("stats: total_saved = %v, want 1599", got)
	}
	if got := stats["resist_count"].(float64); got != 1 {
		t.Errorf("stats: resist_count = %v, want 1", got)
	}
}

// TestDailyQuotaAndPlus checks the free daily limit gates analysis until Plus.
func TestDailyQuotaAndPlus(t *testing.T) {
	srv := newServer(t)
	_, body := do(t, srv, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "quota@test.com", "password": "secret1", "name": "Q",
	})
	token := body["token"].(string)

	// Default free limit is 3 — the 4th analysis must be blocked with 402.
	for i := 0; i < 3; i++ {
		code, _ := do(t, srv, "POST", "/api/v1/analyze", token, map[string]any{
			"item": "item", "price": 100, "cat": "other",
		})
		if code != http.StatusOK {
			t.Fatalf("analyze %d: got %d, want 200", i, code)
		}
	}
	code, _ := do(t, srv, "POST", "/api/v1/analyze", token, map[string]any{"item": "item", "price": 100, "cat": "other"})
	if code != http.StatusPaymentRequired {
		t.Errorf("4th analyze: got %d, want 402", code)
	}

	// Subscribe → Plus → analysis allowed again.
	if code, _ := do(t, srv, "POST", "/api/v1/subscription/subscribe", token, map[string]any{"plan": "year"}); code != http.StatusOK {
		t.Fatalf("subscribe: got %d", code)
	}
	if code, _ := do(t, srv, "POST", "/api/v1/analyze", token, map[string]any{"item": "item", "price": 100, "cat": "other"}); code != http.StatusOK {
		t.Errorf("post-subscribe analyze: got %d, want 200", code)
	}
}

// TestAdminRBAC verifies admin endpoints require the admin role.
func TestAdminRBAC(t *testing.T) {
	srv := newServer(t)

	// A normal user is forbidden.
	_, ub := do(t, srv, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "rbac@test.com", "password": "secret1", "name": "U",
	})
	if code, _ := do(t, srv, "GET", "/api/v1/admin/overview", ub["token"].(string), nil); code != http.StatusForbidden {
		t.Errorf("user → admin: got %d, want 403", code)
	}

	// The seeded admin succeeds.
	_, ab := do(t, srv, "POST", "/api/v1/auth/login", "", map[string]any{
		"email": "admin@lim.app", "password": "admin123",
	})
	code, ov := do(t, srv, "GET", "/api/v1/admin/overview", ab["token"].(string), nil)
	if code != http.StatusOK {
		t.Fatalf("admin → overview: got %d", code)
	}
	if _, ok := ov["kpis"]; !ok {
		t.Error("admin overview: missing kpis")
	}

	// Unauthenticated is rejected.
	if code, _ := do(t, srv, "GET", "/api/v1/me", "", nil); code != http.StatusUnauthorized {
		t.Errorf("no token → /me: got %d, want 401", code)
	}
}

// TestWishlistCoolingOff parks a decision and resolves it from the wishlist.
func TestWishlistCoolingOff(t *testing.T) {
	srv := newServer(t)
	_, b := do(t, srv, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "wish@test.com", "password": "secret1", "name": "W",
	})
	token := b["token"].(string)

	_, dec := do(t, srv, "POST", "/api/v1/analyze", token, map[string]any{
		"item": "机械键盘", "price": 899, "cat": "digital", "reason": "想换个手感",
	})
	decID := dec["id"].(string)

	code, _ := do(t, srv, "POST", "/api/v1/wishlist", token, map[string]any{
		"item": "机械键盘", "price": 899, "cat": "digital", "impulse": 60, "decision_id": decID,
	})
	if code != http.StatusCreated {
		t.Fatalf("add wishlist: got %d", code)
	}

	code, list := do(t, srv, "GET", "/api/v1/wishlist", token, nil)
	if code != http.StatusOK {
		t.Fatalf("list wishlist: got %d", code)
	}
	_ = list // shape covered by client; here we just resolve below

	// Resolve the wishlist item directly via its id requires fetching it; the
	// list endpoint returns an array, so re-list and resolve the first.
	var arr []map[string]any
	{
		req, _ := http.NewRequest("GET", srv.URL+"/api/v1/wishlist", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		_ = json.Unmarshal(data, &arr)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 wishlist item, got %d", len(arr))
	}
	wid := arr[0]["id"].(string)
	if code, _ := do(t, srv, "POST", "/api/v1/wishlist/"+wid+"/resolve", token, map[string]any{"action": "resist"}); code != http.StatusOK {
		t.Fatalf("resolve wishlist: got %d", code)
	}

	// The source decision should now be resisted and savings credited.
	_, stats := do(t, srv, "GET", "/api/v1/stats", token, nil)
	if got := stats["total_saved"].(float64); got != 899 {
		t.Errorf("total_saved = %v, want 899", got)
	}
}
