# LIM · backend (Go)

The REST API behind **LIM (Less is More)** — it powers both the iOS app and the
admin dashboard. It runs the six-dimension purchase analysis, tracks decisions,
savings, the cooling-off wishlist, subscriptions, and exposes admin analytics.

- **Language:** Go 1.25+ (standard-library HTTP, minimal deps)
- **Storage:** two interchangeable `store.Store` implementations — a zero-config
  JSON-file snapshot store (default) and a real Postgres store
  (`LIM_DATABASE_URL`). Both seed themselves and persist across restarts.
- **Auth:** JWT (HS256) + bcrypt
- **AI:** deterministic six-dimension heuristic by default; optional Claude
  (Anthropic) when `ANTHROPIC_API_KEY` is set, with graceful fallback

Full references: [`../docs/API.md`](../docs/API.md),
[`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md),
[`../docs/DATA_MODEL.md`](../docs/DATA_MODEL.md),
[`../docs/DEPLOYMENT.md`](../docs/DEPLOYMENT.md).

## Quick start

```bash
cd backend
make run            # or: go run ./cmd/server
# → LIM API listening on :8080 · store=lim-data.json · ai=heuristic
```

That's it — the store seeds itself on first boot (demo users, decisions,
transactions, catalogue, default AI config).

```bash
# health
curl localhost:8080/healthz

# log in as a seeded demo user and analyse a purchase
TOKEN=$(curl -s localhost:8080/api/v1/auth/login \
  -d '{"email":"lin@lim.app","password":"password"}' | jq -r .token)

curl -s localhost:8080/api/v1/analyze -H "Authorization: Bearer $TOKEN" \
  -d '{"item":"索尼降噪耳机","price":2299,"cat":"digital","reason":"直播打折，已经有一副了"}' | jq
```

### Demo credentials

| Role  | Email          | Password   |
|-------|----------------|------------|
| Admin | `admin@lim.app`| `admin123` |
| User  | `lin@lim.app`  | `password` |

(Other seeded users: `su@`, `chen@`, `jiang@`, `zhou@`, … all `@lim.app` /
`password`.)

## Make targets

```
make run         Run the API server (seeded demo data)
make build       Build the binary into bin/
make test        Run the test suite
make vet         go vet
make docker      Build the Docker image
make seed-reset  Delete the local data file to force a fresh seed
```

## Configuration

All via environment variables (see [`.env.example`](.env.example)); every value
has a default so it runs with zero config.

| Variable | Default | Purpose |
|---|---|---|
| `LIM_ADDR` | `:8080` | listen address |
| `LIM_DATA_FILE` | `lim-data.json` | JSON snapshot store path (file store) |
| `LIM_DATABASE_URL` | — | use Postgres instead of the file store |
| `LIM_JWT_SECRET` | `dev-secret-change-me` | **change in production** |
| `LIM_TOKEN_TTL_HOURS` | `720` | access-token lifetime |
| `LIM_CORS_ORIGIN` | `*` | allowed origin for the admin web app |
| `LIM_SEED_DEMO` | `true` | seed demo data on first boot |
| `LIM_ADMIN_EMAIL` / `LIM_ADMIN_PASSWORD` | `admin@lim.app` / `admin123` | bootstrap admin |
| `ANTHROPIC_API_KEY` | — | enable the Claude-backed engine |
| `LIM_ANTHROPIC_MODEL` | `claude-sonnet-4-6` | model id |

## Docker

```bash
docker compose up --build      # serves :8080, data persisted in a volume
```

## Storage backends

Everything behind the HTTP layer talks to the `store.Store` interface
(`internal/store/interface.go`), so the backing store is a drop-in choice:

- **File store (default)** — in-memory state snapshotted to JSON
  (`internal/store/store.go`). Zero setup; great for dev and single-node.
- **Postgres** — set `LIM_DATABASE_URL` (`internal/store/postgres.go`). The
  schema is created on boot; entities are stored as JSONB with promoted columns
  for lookups; catalogue/admin/demo data seed on first run exactly as with the
  file store.

```bash
LIM_DATABASE_URL="postgres://lim:lim@localhost:5432/lim?sslmode=disable" make run
```

No handler, seeder, or test changes are needed to switch — they depend only on
the interface.

## Layout

```
backend/
├── cmd/server/main.go         # entry point, graceful shutdown, healthcheck flag
└── internal/
    ├── config/                # env-driven configuration
    ├── models/                # domain entities (the shared contract)
    ├── store/                 # Store interface + JSON-file implementation
    ├── ai/                    # six-dimension engine + optional Claude client
    ├── auth/                  # JWT + bcrypt
    ├── httpapi/               # router, middleware, handlers (app + admin)
    └── seed/                  # first-boot catalogue + demo data
```

## How the analysis works

`internal/ai/engine.go` ports the prototype's heuristic: each of the six axes
(need / alt / emo / value / money / env) starts neutral at 3, keyword signals and
a price-to-budget ratio nudge the scores, and the admin-weighted average folds
into a 0–100 **impulse index** (clamped to 4–96). The verdict bands are
`< 45 → buy`, `45–61 → pause`, `≥ 62 → resist`. Admins tune the weights, persona,
prompt, daily free limit and cooling-off window live via `PUT /api/v1/admin/ai-config`.
