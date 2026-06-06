// Command server runs the LIM REST API (consumer app + admin dashboard).
//
//	go run ./cmd/server         # starts on :8080 with seeded demo data
//
// Configuration is via environment variables; see internal/config and
// .env.example. With ANTHROPIC_API_KEY set, the analysis engine uses Claude and
// falls back to the built-in heuristic on any error.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mosson9/lim/backend/internal/ai"
	"github.com/mosson9/lim/backend/internal/appstore"
	"github.com/mosson9/lim/backend/internal/auth"
	"github.com/mosson9/lim/backend/internal/config"
	"github.com/mosson9/lim/backend/internal/httpapi"
	"github.com/mosson9/lim/backend/internal/seed"
	"github.com/mosson9/lim/backend/internal/store"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the local /healthz endpoint and exit")
	flag.Parse()

	cfg := config.Load()

	// Container healthcheck mode: dial our own port and exit 0/1.
	if *healthcheck {
		runHealthcheck(cfg.Addr)
		return
	}

	st, err := store.Open(cfg.DatabaseURL, cfg.DataFile)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	if st.IsEmpty() {
		log.Printf("seeding store (demo=%v)…", cfg.SeedDemo)
		if err := seed.Run(st, cfg.SeedDemo, cfg.AdminEmail, cfg.AdminPassword); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	authMgr := auth.New(cfg.JWTSecret, cfg.TokenTTL)
	engine := ai.New(cfg.AnthropicKey, cfg.AnthropicModel)
	app := httpapi.NewApp(st, authMgr, engine, cfg.CORSOrigin, cfg.AdminDir)

	// App Store (StoreKit 2) receipt verification.
	var roots [][]byte
	if cfg.AppleRootCert != "" {
		pem, err := os.ReadFile(cfg.AppleRootCert)
		if err != nil {
			log.Fatalf("read apple root cert: %v", err)
		}
		roots = append(roots, pem)
	}
	verifier, err := appstore.New(roots, cfg.AppleBundleID, cfg.AppleEnv)
	if err != nil {
		log.Fatalf("appstore verifier: %v", err)
	}
	app.ConfigureBilling(verifier, cfg.ProductMonthly, cfg.ProductYearly, cfg.AllowMockSubscribe)
	app.ConfigureSecurity(cfg.RateRPM, cfg.AuthRateRPM, cfg.MaxBodyBytes)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		mode := "heuristic"
		if engine.UsesLLM() {
			mode = "claude (" + cfg.AnthropicModel + ")"
		}
		storeKind := "file:" + cfg.DataFile
		if cfg.DatabaseURL != "" {
			storeKind = "postgres"
		}
		log.Printf("LIM API listening on %s · store=%s · ai=%s", cfg.Addr, storeKind, mode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// runHealthcheck probes /healthz on the configured address and exits non-zero on
// failure. Used by the Docker HEALTHCHECK in the distroless image.
func runHealthcheck(addr string) {
	host := addr
	if len(host) > 0 && host[0] == ':' {
		host = "127.0.0.1" + host
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/healthz", host))
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
