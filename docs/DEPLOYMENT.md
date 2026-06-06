# LIM Deployment Guide · 部署指南

How to run, build, and deploy the LIM backend, run the iOS app against it, and
migrate the data store from the bundled JSON file to Postgres. See
[ARCHITECTURE.md](./ARCHITECTURE.md) for the big picture and
[API.md](./API.md) for the REST contract.

---

## Part 1 · Backend (Go)

### Prerequisites

* **Go 1.25+** (the module targets `go 1.25.0`).
* Optional: **Docker** + **Docker Compose** for containerised runs.
* Optional: an **Anthropic API key** to enable the Claude-backed analysis engine.

No database server is required — the default store is a single JSON file.

### Run locally

From `backend/`:

```bash
make run            # = go run ./cmd/server
# or directly:
go run ./cmd/server
```

The server starts on `:8080`, seeds demo data on first boot (because the data
file doesn't exist yet), and logs something like:

```
seeding store (demo=true)…
LIM API listening on :8080 · store=lim-data.json · ai=heuristic
```

Verify it's up:

```bash
curl -s http://localhost:8080/healthz      # {"status":"ok","llm":false}
```

Other useful Make targets:

| Target | Action |
|--------|--------|
| `make build` | Build the binary into `bin/lim-server` (trimmed, stripped). |
| `make test` | `go test ./...` |
| `make vet` | `go vet ./...` |
| `make fmt` | `gofmt -w .` |
| `make tidy` | `go mod tidy` |
| `make docker` | Build the Docker image `lim-backend:latest`. |
| `make seed-reset` | Delete `lim-data.json` to force a fresh seed next run. |
| `make clean` | Remove `bin/` and the data file. |

### Configuration (environment variables)

Everything has a safe default so the binary runs with zero configuration. Copy
`.env.example` to `.env` and adjust. (The binary reads the process environment;
load `.env` via your shell, `docker compose`, or a tool like `direnv`.)

| Variable | Default | Meaning |
|----------|---------|---------|
| `LIM_ADDR` | `:8080` | HTTP listen address. |
| `LIM_DATA_FILE` | `lim-data.json` | Path to the JSON snapshot store (the whole DB). |
| `LIM_JWT_SECRET` | `dev-secret-change-me` | HMAC secret for signing JWTs. **Change in prod.** |
| `LIM_TOKEN_TTL_HOURS` | `720` | Access-token lifetime in hours (30 days). |
| `LIM_CORS_ORIGIN` | `*` | Allowed CORS origin for the admin web app. |
| `LIM_SEED_DEMO` | `true` | Seed demo users/decisions/transactions on first boot. |
| `LIM_ADMIN_EMAIL` | `admin@lim.app` | Bootstrap admin account (created if absent). |
| `LIM_ADMIN_PASSWORD` | `admin123` | Bootstrap admin password. **Change in prod.** |
| `ANTHROPIC_API_KEY` | _(empty)_ | When set, `/analyze` uses Claude (falls back to heuristic on error). |
| `LIM_ANTHROPIC_MODEL` | `claude-sonnet-4-6` | Claude model id used when the key is set. |

> Note: seeding only happens when the store is **empty**. The bootstrap admin is
> ensured on every empty-store boot; the demo content is gated by `LIM_SEED_DEMO`.

### The JSON file store

The entire database lives in one JSON file (`LIM_DATA_FILE`). The store keeps
state in memory and writes an indented JSON snapshot **on every mutation**, using
an atomic temp-file-plus-rename so a crash can't leave a half-written file. On
startup the file is loaded and an email→id index is rebuilt.

* **Local default:** `./lim-data.json` (in the working directory).
* **Docker default:** `/data/lim-data.json` (a mounted volume — see below).
* **Reset:** `make seed-reset` (or delete the file) to re-seed from scratch.
* **Back it up:** in production, snapshot/replicate this file regularly — it *is*
  your database. A `.tmp` sibling may briefly exist during writes.

### Enabling Claude (optional)

Set `ANTHROPIC_API_KEY` (and optionally `LIM_ANTHROPIC_MODEL`). The engine then
calls the Anthropic Messages API (`https://api.anthropic.com/v1/messages`,
`anthropic-version: 2023-06-01`, 30s timeout) for the six-dimension scoring and
**gracefully falls back to the deterministic heuristic on any error** (network,
non-2xx, malformed JSON). `GET /healthz` reports `"llm": true` when active, and
the startup log shows `ai=claude (<model>)`.

```bash
ANTHROPIC_API_KEY=sk-ant-... LIM_ANTHROPIC_MODEL=claude-sonnet-4-6 go run ./cmd/server
```

> The admin prompt template, persona, model name, dimension weights, free quota,
> and cooling hours are all editable at runtime via
> `PUT /api/v1/admin/ai-config` — no redeploy needed. (The `model` field in the
> AI config is informational for the dashboard; the engine uses
> `LIM_ANTHROPIC_MODEL`.)

### Docker

Build and run with Compose (from `backend/`):

```bash
docker compose up --build
```

This builds a multi-stage image (Go 1.25 alpine builder → `distroless/static`
runtime, non-root), publishes `8080:8080`, mounts a named volume `lim-data` at
`/data` for persistence, and wires a container healthcheck that runs the binary's
built-in `-healthcheck` probe against `/healthz`.

Build the image by itself:

```bash
make docker          # docker build -t lim-backend:latest .
# or
docker build -t lim-backend:latest .
```

Edit the `environment:` block in `docker-compose.yml` to set secrets and (optionally)
the Anthropic key. **The committed compose file ships placeholder secrets — change
them before any non-local deployment.**

### Production hardening

Before exposing LIM to real users:

1. **Change `LIM_JWT_SECRET`** to a long random value. Rotating it invalidates all
   existing tokens (users must re-login).
2. **Change `LIM_ADMIN_PASSWORD`** (and ideally `LIM_ADMIN_EMAIL`). Do this
   *before* the first boot — the admin is only created when absent. If you already
   booted with the default, reset the store or update the account.
3. **Set `LIM_SEED_DEMO=false`** so production doesn't get demo users/decisions.
4. **Set `LIM_CORS_ORIGIN`** to your admin SPA's exact origin (e.g.
   `https://admin.lim.app`) instead of `*`.
5. **Terminate TLS** at a reverse proxy (nginx / Caddy / a cloud LB) in front of
   the Go service. The app speaks plain HTTP; the proxy provides HTTPS, and the
   iOS app must talk HTTPS in production (App Transport Security).
6. **Persist & back up the data file.** Put `LIM_DATA_FILE` on durable storage
   (a mounted volume) and snapshot it on a schedule. It is the single source of
   truth.
7. **Consider a real database** for concurrency/scale (next section). The JSON
   store serialises every write behind a single mutex and rewrites the whole file
   each time — perfect for a prototype or small footprint, not for high write
   throughput.
8. **Run a single instance** while on the JSON store. Two processes pointed at the
   same file will clobber each other. Scale out only after migrating to Postgres.

### <a id="migrating-to-postgres"></a>Using Postgres

Postgres is a **first-class, implemented backend** — not a future task. Set one
env var and the server uses it:

```bash
LIM_DATABASE_URL="postgres://lim:lim@localhost:5432/lim?sslmode=disable" make run
# or via docker compose: uncomment the `db` service + LIM_DATABASE_URL line
```

On boot it creates its schema (`CREATE TABLE IF NOT EXISTS …`) and, on first run
(`IsEmpty()`), seeds catalogue/admin/demo data exactly like the file store. No
other configuration is required.

**The seam.** Everything behind the HTTP layer depends only on the
`store.Store` **interface** (`internal/store/interface.go`); `httpapi.App` and
`seed.Run` accept that interface, and `store.Open(databaseURL, dataFile)` picks
the implementation. The two implementations are:

| Implementation | File | Selected when |
|---|---|---|
| `FileStore` (JSON snapshot) | `internal/store/store.go` | `LIM_DATABASE_URL` empty (default) |
| `PostgresStore` | `internal/store/postgres.go` | `LIM_DATABASE_URL` set |

The interface surface (both implementations satisfy it; enforced by
`var _ Store = (*FileStore)(nil)` / `(*PostgresStore)(nil)`):

```
Users:        CreateUser, GetUser, GetUserByEmail, UpdateUser, ListUsers, TouchUser
Decisions:    CreateDecision, GetDecision, UpdateDecision, ListDecisions, ListAllDecisions
Wishlist:     AddWishlist, GetWishlistItem, ListWishlist, DeleteWishlist
Transactions: AddTransaction, ListTransactions
Config/content: AIConfig/SetAIConfig, Categories/SetCategories, Skins/SetSkins,
                Plans/SetPlans, Perks/SetPerks, Pushes/CreatePush
Lifecycle:    IsEmpty, Save
Errors:       ErrNotFound, ErrDuplicate
```

**How `PostgresStore` maps the model** (`database/sql` + `lib/pq`): each entity
is stored as a JSONB `doc` (reusing the models' json tags) with a few promoted
columns for indexed lookups — `users(id, email, password_hash, last_active, doc)`,
`decisions(id, user_id, status, verdict, created_at, doc)`,
`wishlist(id, user_id, expires_at, doc)`, `transactions`, `pushes`, plus a `kv`
table for the singletons (`ai_config`, `categories`, `skins`, `plans`, `perks`)
and a `counters` table that reproduces the friendly id formats (`U-#####`,
`D-#####`, `T-####`, `P-###`). It preserves the file store's semantics:
case-insensitive email lookup, `ErrNotFound`/`ErrDuplicate`, and identical list
ordering. Password hashes live in their own column (User omits them from JSON).

Because the handlers are written against the seam, **none of `internal/httpapi`,
`internal/models`, `internal/ai`, the routes, or the tests change** when you
switch — the HTTP integration tests run against the file store, and the same
calls work against Postgres. Derived statistics (`/stats`, admin aggregates) are
computed in the handlers from the records the store returns — see
[DATA_MODEL.md](./DATA_MODEL.md#derived-statistics).

> Note: a live Postgres is required to integration-test the Postgres path; the
> default file store needs nothing and is what CI exercises.

---

## Part 2 · iOS app (SwiftUI)

> The app is being built in parallel under `ios/`. Today only
> `Sources/LIM/Theme/Theme.swift` exists; the structure below is the planned
> layout. The backend is fully ready to serve it.

### Prerequisites

* **Xcode 15+**.
* **iOS 16+** deployment target (uses modern SwiftUI).
* A running LIM backend (local or remote) for the app to call.
* Optional tooling: **XcodeGen** (`brew install xcodegen`) to generate the
  `.xcodeproj` from `project.yml`.

### Generate / open the project

Two supported workflows:

**A) XcodeGen (recommended for an app target + simulator):**

```bash
brew install xcodegen
cd ios
xcodegen generate        # reads project.yml → produces LIM.xcodeproj
open LIM.xcodeproj
```

**B) Swift Package Manager:** open `ios/Package.swift` directly in Xcode (or run
`open Package.swift`). Good for building/iterating on the library sources.

### Point the app at the backend

Set the API base URL in the networking layer (`Sources/LIM/Networking/`). For a
locally running backend use:

```
http://localhost:8080
```

The app authenticates by POSTing to `/api/v1/auth/login` (or `/register`), storing
the returned JWT, and sending `Authorization: Bearer <token>` on every protected
request. All endpoints and JSON shapes are in [API.md](./API.md); the `Models/`
layer should mirror those shapes (Codable structs matching the `json` tags in
[DATA_MODEL.md](./DATA_MODEL.md)).

### Run in the simulator

1. Start the backend: `cd backend && make run`.
2. Generate/open the iOS project (above).
3. Select an iPhone simulator (iOS 16+) and press **Run** (⌘R).
4. Log in with a seeded account, e.g. `lin@lim.app` / `password`, or register a
   new one.

### App Transport Security (localhost)

iOS blocks plaintext HTTP by default. Talking to `http://localhost:8080` from the
simulator requires an ATS exception in the app's `Info.plist`. Allow local
networking only (do **not** disable ATS globally), for example:

```xml
<key>NSAppTransportSecurity</key>
<dict>
  <key>NSAllowsLocalNetworking</key>
  <true/>
</dict>
```

In production the app must use **HTTPS** to your TLS-terminated backend, so no ATS
exception is needed there. Remove/scope the localhost exception for release builds.

---

## Part 3 · Admin web app

The original admin dashboard is a **React prototype**. The Go backend already
exposes every admin capability it needs under `GET/PUT/POST /api/v1/admin/*`
(see [API.md → Admin](./API.md#admin)), so a production admin SPA can consume the
same API directly:

* **Auth:** log in via `POST /api/v1/auth/login` as an admin account; send the
  JWT as a bearer token. Non-admin tokens get `403` on `/admin/*`.
* **Read endpoints:** `overview`, `users`, `users/{id}`, `decisions`,
  `subscriptions`, `ai-config`, `categories`, `skins`, `push`.
* **Write endpoints:** `ai-config` (PUT), `categories` (PUT), `push` (POST).
* **CORS:** set `LIM_CORS_ORIGIN` to the SPA's origin (the API already emits the
  appropriate CORS headers and answers `OPTIONS` preflight with `204`).

So the admin frontend can be deployed as a static SPA (any host/CDN) pointed at
the production API base URL — no separate admin backend is required. Note a few
dashboard figures are synthetic in the prototype store (see
[API.md → admin/overview](./API.md#get-apiv1adminoverview--admin)).
