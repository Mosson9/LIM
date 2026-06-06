// Package config loads runtime configuration from the environment.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all tunables for the server. Everything has a sensible default
// so the binary runs with zero configuration for local development.
type Config struct {
	Addr          string        // listen address, e.g. ":8080"
	DataFile      string        // JSON snapshot path for the file store
	DatabaseURL   string        // if set, use Postgres instead of the file store
	JWTSecret     string        // HMAC secret for signing tokens
	TokenTTL      time.Duration // access-token lifetime
	CORSOrigin    string        // allowed origin for the admin web app ("*" for any)
	AdminDir      string        // if set, serve the admin web app from this dir at /admin/
	SeedDemo      bool          // seed demo users/decisions on first boot
	AdminEmail    string        // bootstrap admin account
	AdminPassword string        // bootstrap admin password

	// Optional LLM (Anthropic Claude) integration. When AnthropicKey is set the
	// engine asks Claude for the six-dimension analysis; otherwise it falls back
	// to the built-in deterministic heuristic.
	AnthropicKey   string
	AnthropicModel string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		Addr:           env("LIM_ADDR", ":8080"),
		DataFile:       env("LIM_DATA_FILE", "lim-data.json"),
		DatabaseURL:    env("LIM_DATABASE_URL", ""),
		JWTSecret:      env("LIM_JWT_SECRET", "dev-secret-change-me"),
		TokenTTL:       time.Duration(envInt("LIM_TOKEN_TTL_HOURS", 720)) * time.Hour,
		CORSOrigin:     env("LIM_CORS_ORIGIN", "*"),
		AdminDir:       env("LIM_ADMIN_DIR", ""),
		SeedDemo:       envBool("LIM_SEED_DEMO", true),
		AdminEmail:     env("LIM_ADMIN_EMAIL", "admin@lim.app"),
		AdminPassword:  env("LIM_ADMIN_PASSWORD", "admin123"),
		AnthropicKey:   env("ANTHROPIC_API_KEY", ""),
		AnthropicModel: env("LIM_ANTHROPIC_MODEL", "claude-sonnet-4-6"),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
