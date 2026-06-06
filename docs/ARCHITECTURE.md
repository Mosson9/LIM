# LIM Architecture · 系统架构

> **LIM (Less is More) · 少即是多** — an anti-consumerism companion app.
> 一个温柔、不评判的理性消费陪伴者。Before you buy, ask LIM. LIM scores the
> purchase across six dimensions, gives you an *impulse index* (冲动指数) and a
> verdict, and turns the money you *didn't* spend into savings and a growing tree.

This document explains how the system fits together: the product concept, the
runtime topology, the backend layering, the request lifecycle, and the design
system. For the REST contract see [API.md](./API.md); for entities see
[DATA_MODEL.md](./DATA_MODEL.md); for running it see [DEPLOYMENT.md](./DEPLOYMENT.md).

---

## 1. System overview

```
        ┌──────────────────────┐        HTTPS / JSON          ┌───────────────────────────┐
        │   iOS app (SwiftUI)   │  ───────────────────────▶   │     REST API (Go 1.25)    │
        │   iPhone · iOS 16+    │   Bearer JWT in header      │   net/http std lib mux    │
        │                       │  ◀───────────────────────   │                           │
        └──────────────────────┘                              │   ┌───────────────────┐   │
                                                              │   │  ai.Engine        │   │
        ┌──────────────────────┐        HTTPS / JSON          │   │  six-dimension    │   │
        │  Admin web app (SPA)  │  ───────────────────────▶   │   │  analysis         │   │
        │  React prototype →    │   Bearer JWT (admin role)   │   └─────────┬─────────┘   │
        │  production dashboard │  ◀───────────────────────   │             │ optional    │
        └──────────────────────┘                              │             ▼             │
                                                              │   ┌───────────────────┐   │
                                                              │   │ Anthropic Claude  │◀─ │ ── only when
                                                              │   │ Messages API      │   │   ANTHROPIC_API_KEY
                                                              │   └───────────────────┘   │   is set
                                                              │                           │
                                                              │   ┌───────────────────┐   │
                                                              │   │ store.Store       │   │
                                                              │   │ JSON file snapshot │──▶ lim-data.json
                                                              │   └───────────────────┘   │
                                                              └───────────────────────────┘
```

* **iOS app** — the consumer surface. Pure SwiftUI, talks to the API over REST,
  holds a JWT for the signed-in user. UI strings are Chinese.
* **REST API (Go)** — one binary, one `http.ServeMux`, serving *both* the
  consumer endpoints and the admin dashboard endpoints under `/api/v1`.
* **Storage (`store.Store` interface)** — two implementations: a JSON **file
  store** (default; the whole DB is `lim-data.json`, snapshotted on every write —
  zero external services, survives restarts) and a **Postgres store**
  (`LIM_DATABASE_URL`). Selected at startup; handlers never know which.
* **Claude (optional)** — when `ANTHROPIC_API_KEY` is configured, the analysis
  engine asks Claude for the six-dimension scoring and **gracefully falls back**
  to the built-in deterministic heuristic on any error. With no key, the app is
  fully offline-capable on the server side.

---

## 2. The product concept · 产品理念

The flow is captured in the product mind-map and mirrored 1:1 by the data model.

A user is hesitating over a purchase and asks LIM: **"Should I buy this?" / 我该买它吗？**
The AI scores the item on **six dimensions**, each 1–5 (higher = more justified):

| # | Key | 中文 | Question it answers |
|---|------|------|---------------------|
| 1 | `need`  | 需求 (need)        | Is it actually necessary? 缺它会造成困难吗？ |
| 2 | `alt`   | 替代 (alternative) | Could an existing / rented / shared item do? 能租借、修理或共享吗？ |
| 3 | `emo`   | 情感 (emotion)     | Genuine want, or ads / social pressure / a passing mood? |
| 4 | `value` | 长期价值 (value)    | Frequency, lifespan, lasting value. 长期带来价值吗？ |
| 5 | `money` | 经济 (money)       | Does it fit the budget without crowding out essentials? |
| 6 | `env`   | 环境 (environment) | Production / transport / use footprint vs. your values. |

These fold into a **0–100 impulse index (冲动指数)** via admin-tunable per-dimension
weights — a *lower* justification produces a *higher* impulse. The index maps to a
**verdict**:

| Verdict | 中文 | Meaning | Impulse band (heuristic) |
|---------|------|---------|--------------------------|
| `buy`    | 可以拥有它   | A reasonable purchase — go ahead.        | `< 45` |
| `pause`  | 再想想       | Not necessary; sleep on it (cooling-off). | `45–61` |
| `resist` | 也许不必买   | Mostly emotion; probably skip it.        | `≥ 62` |

What happens next, per verdict:

```
                       ┌─────────── resist 也许不必买 ──▶ 忍住了 / resisted
                       │              saved = price ──▶ 省钱 (savings) ──▶ 🌳 成长树 (growth tree)
   should I buy this?  │
   ──▶ AI analysis ────┼─────────── pause  再想想    ──▶ 心愿单 (wishlist, 24h 冷静期)
        (six dims,     │              cooling-off ──▶ after 24h, decide resist/buy
         impulse,      │
         verdict)      └─────────── buy    可以拥有它 ──▶ 买了 / bought ──▶ 消费记录 (spend record)
```

* **Resist** → the item's price is credited as **savings**; cumulative savings
  grow a **tree** through six stages (see [DATA_MODEL.md](./DATA_MODEL.md#derived-statistics)).
* **Pause** → the item is parked in the **24h cooling-off wishlist**; when the
  window elapses the user revisits and chooses *resist* or *buy*.
* **Buy** → recorded as a **spend**; no savings credited.

Around this core sit two more surfaces:

* **Subscription (Plus)** — free users get a daily quota of AI consultations
  (default **3/day**); Plus removes the cap and unlocks icon skins, growth skins
  and category icons. Tiers: 月度 ¥18/月, 年度 ¥98/年.
* **Admin dashboard** — operational analytics (users, decisions, revenue/MRR,
  verdict & category distributions, funnel), plus live tuning of the AI config
  (weights, persona, prompt, free quota, cooling hours), category & skin
  catalogues, and push campaigns.

---

## 3. Backend layering

The Go backend is a small, layered, dependency-light service. Each layer depends
only on the layer beneath it.

```
   cmd/server/main.go
     │  loads config, opens store, seeds if empty, wires deps, runs http.Server,
     │  graceful shutdown on SIGINT/SIGTERM, -healthcheck probe mode.
     ▼
   internal/httpapi          ← the only HTTP-aware package
     │  router.go      route table (source of truth) + App struct + middleware wiring
     │  middleware.go  CORS · request logging · authenticate (JWT) · requireAdmin
     │  response.go    writeJSON / writeError / decodeJSON (DisallowUnknownFields)
     │  auth.go profile.go meta.go decisions.go wishlist.go stats.go
     │  subscription.go admin.go   ← request/response shapes + handler logic
     ▼
   internal/store            ← persistence seam (interface-like surface)
     │  store.Store: file-backed, concurrency-safe (sync.RWMutex), JSON snapshot
     │  on every mutation; atomic write via temp file + rename.
     ▼
   internal/models           ← pure domain entities + json tags (no behaviour deps)

   side packages used by the handlers:
   internal/ai      Engine.Analyze → Heuristic (deterministic) or Claude (llm.go)
   internal/auth    JWT issue/parse (HS256) + bcrypt password hashing
   internal/config  env-var loading with safe defaults
   internal/seed    catalogue content + bootstrap admin + optional demo data
```

### Dependencies (`go.mod`)

Deliberately minimal — only three direct dependencies:

* `github.com/golang-jwt/jwt/v5` — JWT signing/parsing.
* `golang.org/x/crypto` — bcrypt.
* `github.com/google/uuid` — wishlist item IDs.

Everything else (HTTP routing with method patterns, JSON, file I/O) is the Go
standard library. Routing uses Go 1.22+ `http.ServeMux` method+path patterns
(e.g. `"POST /api/v1/analyze"`, `"GET /api/v1/decisions/{id}"`).

### The store seam · 为什么换数据库很容易

`internal/httpapi` only ever calls methods on the `store.Store` **interface**
(`CreateUser`, `GetDecision`, `ListWishlist`, `SetAIConfig`, …). It never touches
files or SQL directly. Two implementations ship today: the default `FileStore`
keeps state in memory and writes a JSON snapshot on every mutation (so the server
needs **zero external services** yet **survives restarts**), and `PostgresStore`
backs real deployments. `store.Open(databaseURL, dataFile)` picks one from config
— the handlers, models, and routes are identical either way. See
[DEPLOYMENT.md → Using Postgres](./DEPLOYMENT.md#migrating-to-postgres).

---

## 4. Request lifecycle

Every request flows through the same middleware chain (outermost first):

```
   client ─▶ logging ─▶ withCORS ─▶ ServeMux (method+path match)
                                       │
                                       ├── public route ───────────────▶ handler
                                       │
                                       ├── user route  ─▶ authenticate ─▶ handler
                                       │      (Bearer JWT → store.GetUser → ctx)
                                       │
                                       └── admin route ─▶ requireAdmin ─▶ handler
                                              (authenticate + role == "admin")
   handler ─▶ store / ai.Engine ─▶ writeJSON ─▶ (logging records status + latency)
```

* `logging` wraps the response writer to capture the status code, logs
  `METHOD PATH STATUS LATENCY`.
* `withCORS` sets permissive CORS headers (origin from `LIM_CORS_ORIGIN`) and
  short-circuits `OPTIONS` preflight with `204`.
* `authenticate` strips the `Bearer ` prefix, parses+validates the HS256 JWT,
  loads the user from the store, and injects it into the request context.
* `requireAdmin` runs `authenticate` then additionally checks `role == "admin"`,
  returning **403** otherwise.
* Request bodies are decoded with `DisallowUnknownFields()` — unexpected JSON
  keys are rejected with **400**.
* Responses are always JSON; errors use the envelope `{"error": "..."}` with
  Chinese user-facing messages.

### Sequence diagram — analyze → resolve

The signature flow: ask, get a pending decision, then act on it.

```
 iOS app                API (httpapi)            ai.Engine            store
   │                        │                        │                  │
   │ POST /api/v1/analyze   │                        │                  │
   │ {item,price,cat,reason}│                        │                  │
   │───────────────────────▶│ authenticate (JWT)     │                  │
   │                        │ refreshDailyQuota      │                  │
   │                        │ check free daily limit │                  │
   │                        │   (402 if exhausted)   │                  │
   │                        │ Analyze(in,profile,cfg)│                  │
   │                        │───────────────────────▶│ Claude or        │
   │                        │                        │ Heuristic        │
   │                        │   Result{dims,impulse, │                  │
   │                        │◀───────────────────────│ verdict,message} │
   │                        │ CreateDecision(pending)│                  │
   │                        │───────────────────────────────────────────▶│
   │                        │ AIUsedToday++ ; UpdateUser────────────────▶│
   │   200 Decision{id,...} │                        │                  │
   │◀───────────────────────│                        │                  │
   │                        │                        │                  │
   │  …user taps 忍住了 / 买了 …                       │                  │
   │                        │                        │                  │
   │ POST /decisions/{id}/  │                        │                  │
   │ resolve {action}       │                        │                  │
   │───────────────────────▶│ load decision (owner?) │                  │
   │                        │ resist → Saved=Price   │                  │
   │                        │ buy    → Saved=0       │                  │
   │                        │ status, DecidedAt=now  │                  │
   │                        │ UpdateDecision ───────────────────────────▶│
   │   200 Decision{...}    │                        │                  │
   │◀───────────────────────│                        │                  │
   │                        │                        │                  │
   │ GET /api/v1/stats      │  computeStats derives savings/streak/tree  │
   │───────────────────────▶│  from the user's resolved decisions ──────▶│
   │   200 statsResp        │                        │                  │
   │◀───────────────────────│                        │                  │
```

### Sequence diagram — wishlist cooling-off (再想想)

When the verdict is `pause`, the app parks the item for 24h, then resolves it.

```
 iOS app                       API (httpapi)                     store
   │                               │                               │
   │ POST /api/v1/wishlist         │                               │
   │ {item,price,cat,impulse,      │                               │
   │  decision_id}                 │                               │
   │──────────────────────────────▶│ ExpiresAt = now + CoolingHours│
   │                               │ AddWishlist ──────────────────▶│
   │                               │ resolvePending(decision_id,    │
   │                               │   StatusWishlist) ────────────▶│
   │  201 wishlistView{progress,   │                               │
   │      remain_sec, expires_at}  │                               │
   │◀──────────────────────────────│                               │
   │                               │                               │
   │  …24h elapse; client polls…   │                               │
   │ GET /api/v1/wishlist          │ each item decorated with       │
   │──────────────────────────────▶│ progress / expired / remain_sec│
   │  200 [wishlistView]           │ (computed live from clock)     │
   │◀──────────────────────────────│                               │
   │                               │                               │
   │  …window expired, user decides…                               │
   │ POST /wishlist/{id}/resolve   │                               │
   │ {action: resist|buy}          │                               │
   │──────────────────────────────▶│ link decision → resist/bought  │
   │                               │ (or synthesise one if manual)  │
   │                               │ DeleteWishlist ───────────────▶│
   │  200 {ok, action, saved}      │                               │
   │◀──────────────────────────────│                               │
```

---

## 5. Repository layout

The whole repository: a complete Go backend, a SwiftUI app under construction,
and these docs. Files listed under `backend/` are the *actual* tree.

```
LIM/
├── README.md                        # top-level overview (written separately)
├── docs/
│   ├── ARCHITECTURE.md              # ← you are here
│   ├── API.md                       # full REST reference
│   ├── DEPLOYMENT.md                # run/build/deploy + Postgres migration
│   └── DATA_MODEL.md                # every entity + derived stats
│
├── design/
│   ├── app.reference.css            # consumer-app visual reference (web prototype)
│   └── admin.reference.css          # admin dashboard visual reference
│
├── backend/                         # Go REST API — COMPLETE
│   ├── go.mod                       # module github.com/mosson9/lim/backend (go 1.25.0)
│   ├── go.sum
│   ├── Makefile                     # run / build / test / docker / seed-reset …
│   ├── Dockerfile                   # multi-stage → distroless static, nonroot
│   ├── docker-compose.yml           # api service + lim-data volume + healthcheck
│   ├── .env.example                 # all env vars with defaults
│   ├── .gitignore
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # entrypoint: config→store→seed→serve, healthcheck
│   └── internal/
│       ├── ai/
│       │   ├── engine.go            # six-dimension Heuristic + Engine.Analyze
│       │   ├── engine_test.go       # heuristic unit tests
│       │   └── llm.go               # optional Anthropic Claude Messages integration
│       ├── auth/
│       │   └── auth.go              # JWT (HS256) issue/parse + bcrypt
│       ├── config/
│       │   └── config.go            # env-var loading + defaults
│       ├── httpapi/
│       │   ├── router.go            # route table (source of truth) + App + Handler()
│       │   ├── middleware.go        # CORS · logging · authenticate · requireAdmin
│       │   ├── response.go          # writeJSON / writeError / decodeJSON
│       │   ├── auth.go              # register / login / me
│       │   ├── profile.go           # update profile / onboard / app-icon / quota reset
│       │   ├── meta.go              # categories / skins / plans / perks / dimensions
│       │   ├── decisions.go         # analyze + list/get/resolve decisions
│       │   ├── wishlist.go          # 24h cooling-off list
│       │   ├── stats.go             # DERIVED statistics (savings, streak, tree…)
│       │   ├── subscription.go      # entitlement + mock subscribe
│       │   └── admin.go             # overview/users/decisions/revenue/config/push
│       ├── models/
│       │   └── models.go            # all domain entities + json tags + enums
│       ├── seed/
│       │   └── seed.go              # catalogue + bootstrap admin + demo data
│       └── store/
│           ├── interface.go        # Store interface + Open() factory (the seam)
│           ├── store.go            # FileStore — in-memory + JSON snapshot
│           └── postgres.go         # PostgresStore — database/sql + lib/pq
│
├── admin/                          # Admin web app — zero-build vanilla JS SPA
│   ├── index.html                  # shell + API base resolution
│   ├── styles.css                  # design tokens (ported from the prototype)
│   └── app.js                      # router, API client, 8 pages, inline-SVG charts
│
└── ios/                            # SwiftUI app (iOS 16+)
    ├── project.yml                 # XcodeGen project spec → LIM.xcodeproj
    ├── README.md
    ├── Resources/                  # Info.plist (base URL + ATS), Assets.xcassets
    └── Sources/
        └── LIM/
            ├── LIMApp.swift        # @main App entrypoint + root routing
            ├── Theme/              # design tokens + reusable styled components
            ├── Models/             # Codable mirrors of the API JSON
            ├── Networking/         # APIClient (async/await), config
            ├── Store/              # AppModel — session, navigation, data cache
            ├── Components/         # Icon, charts (dial/radar/tree/bars), tab bar
            └── Screens/            # 14 screens: onboarding, ask, analyzing,
                                    #   result, growth, history, plus, me …
```

---

## 6. Design system · 设计系统

极简禅意 · 暖纸白 + 克制黑白灰 + 靛蓝强调 — calm, paper-white surfaces, restrained
neutral inks, an indigo brand accent, and semantic colors for savings, impulse
and danger. Tokens are a faithful port of the web prototype's CSS `:root`
variables and live in `ios/Sources/LIM/Theme/Theme.swift`; the server also emits
hex colors for categories and skins.

### Color tokens

| Token | Hex | Role · 用途 |
|-------|-----|-------------|
| **paper**   | `#F4F3EE` | Primary background — 暖纸白 paper white |
| paper2      | `#EFEEE7` | Secondary background |
| surface     | `#FFFFFF` | Card surface |
| surface2    | `#FBFAF6` | Raised / alt surface |
| **ink**     | `#1B1B19` | Primary text — near-black 墨 |
| ink2        | `#6A6963` | Secondary text |
| ink3        | `#9C9A92` | Tertiary text |
| ink4        | `#C2C0B6` | Disabled / faint |
| hairline    | `#E7E5DD` | Hairline borders / dividers |
| **indigo**  | `#34357C` | Brand / AI / impulse accent — 靛蓝 |
| indigoBright| `#4B4DAE` | Brighter indigo (charts, links) |
| indigoSoft  | `#ECECF6` | Indigo tint surface |
| **saved (sage)** | `#5E7E63` | Savings / growth — 山涧绿 sage |
| savedDeep   | `#46613F` | Deep sage |
| savedSoft   | `#E9EEE7` | Sage tint surface |
| spent       | `#8C8A82` | Spending — neutral grey |
| spentSoft   | `#EEEDE7` | Spending tint |
| **warn (clay)** | `#C0824F` | High impulse — 陶土 warm clay |
| warnSoft    | `#F4E9DD` | Clay tint |
| **danger**  | `#B5524E` | Danger / destructive |
| dangerSoft  | `#F6E5E4` | Danger tint |
| gold        | `#E9C36B` | Plus / premium accent |

### Typography

* **Body / UI:** *Noto Sans SC* (思源黑体) — Chinese-first sans.
* **Display / numerals / pull-quotes:** *Newsreader* paired with *Noto Serif SC*
  (思源宋体) for serif numerals and quotes. The app falls back to the system serif
  (`design: .serif`) when those families aren't bundled.

### Shape

* Corner radii: `radiusS` 14 · `radius` 22 · `radiusL` 30.

These tokens drive the impulse gauge (indigo → clay as the index rises), the
savings/growth visuals (sage), the cooling-off timers, and the Plus surfaces
(gold). The verdict colors used on the admin donut match: resist → sage
`#5E7E63`, buy → indigo `#34357C`, pause → clay `#C0824F`.
