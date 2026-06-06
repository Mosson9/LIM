<div align="center">

# LIM · Less is More

**反消费主义 · An anti-consumerism companion**

When you're about to buy something you're not sure you need, you open LIM and ask.
An AI weighs it across **six dimensions**, shows you an **impulse index (冲动指数)**,
and gently helps you decide — *buy*, *sleep on it*, or *let it go*. Every time you
resist becomes real money saved and a tree that slowly grows.

*Native iOS app (SwiftUI) · Go backend · Admin dashboard*

</div>

---

## What's in here

This repository is a complete, runnable implementation derived from the LIM
product mind-map, the App prototype, and the Admin prototype:

| Path | What it is |
|------|------------|
| [`backend/`](backend) | **Go REST API** — auth, the six-dimension analysis engine, decisions, savings, cooling-off wishlist, subscriptions, and all admin analytics. Runs with zero external services. |
| [`ios/`](ios) | **SwiftUI app** — all 14 screens (onboarding → ask → analyzing → result → growth → history → plus …), pixel-faithful to the prototype's design system. |
| [`docs/`](docs) | **Documentation** — architecture, full API reference, data model, and deployment guide. |
| [`design/`](design) | Extracted CSS design tokens from the original prototypes, kept for traceability. |

## The idea (from the mind-map)

```
想买点东西，但不确定是否真的需要？
        │
        ▼
  打开 LIM，问问 AI
        │
        ▼
  AI 基于你的画像，从六个维度分析：
   需求 · 替代 · 情感 · 长期价值 · 经济 · 环境
        │
        ▼
  冲动指数 (0–100)  +  最终建议
        │
   ┌────┼─────────────┐
   ▼    ▼             ▼
  买    冷静 24 小时    不买
 记账   心愿单         省下的钱 → 成长的树
```

- **Six dimensions** (`需求/替代/情感/长期价值/经济/环境`) each score 1–5; the
  weighted average becomes a 0–100 impulse index and a verdict
  (`可以拥有它 / 再想想 / 也许不必买`).
- **Resisting** credits savings and grows a tree (the original *Less* logic).
- **Pausing** parks the item in a **24-hour cooling-off** wishlist.
- **Buying** records a clean, deliberate purchase.
- **LIM Plus** subscription unlocks unlimited consultations, faster AI, app-icon
  skins and more — including the mind-map's wink: *the one purchase LIM will
  always recommend.* 😄
- An **admin dashboard** API surfaces DAU, AI-verdict split, conversion funnel,
  revenue/MRR, user management and live AI-config tuning.

## Quick start

### 1. Backend (Go 1.25+)

```bash
cd backend
make run            # → http://localhost:8080, seeds demo data on first boot
```

```bash
# smoke test
curl localhost:8080/healthz
TOKEN=$(curl -s localhost:8080/api/v1/auth/login \
  -d '{"email":"lin@lim.app","password":"password"}' | jq -r .token)
curl -s localhost:8080/api/v1/stats -H "Authorization: Bearer $TOKEN" | jq
```

Demo logins: **admin** `admin@lim.app` / `admin123`, **user** `lin@lim.app` / `password`.

### 2. iOS app (Xcode 15+, iOS 16+)

```bash
cd ios
brew install xcodegen     # one-time
xcodegen generate
open LIM.xcodeproj         # Run on an iPhone simulator
```

The app defaults to `http://localhost:8080`; sign in with a demo account or
register a new one.

### 3. Optional — real AI via Claude

```bash
cd backend
ANTHROPIC_API_KEY=sk-ant-... make run   # engine uses Claude, falls back to the heuristic
```

## Architecture at a glance

```
┌─────────────────┐        REST / JWT        ┌──────────────────────────┐
│  iOS app         │ ───────────────────────▶ │  Go API (cmd/server)     │
│  (SwiftUI)       │ ◀─────────────────────── │  httpapi → store → models│
└─────────────────┘                           │  ai engine · auth · seed │
                                               └────────────┬─────────────┘
┌─────────────────┐        REST / JWT                       │ JSON snapshot
│  Admin web app   │ ──────────────────────────────────────▶│  (swappable for
│  (prototype)     │   GET /api/v1/admin/*                   │   Postgres)
└─────────────────┘                                          ▼
                                          optional ──▶ Anthropic Claude
```

The backend depends only on a small `store.Store` surface, so the JSON-file
store can be replaced with Postgres without touching any handler. See
[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — system design, layering, request lifecycle, diagrams
- [API reference](docs/API.md) — every endpoint, request/response, curl examples
- [Data model](docs/DATA_MODEL.md) — entities, enums, and how stats are derived
- [Deployment](docs/DEPLOYMENT.md) — local, Docker, production hardening, Postgres migration, iOS build
- [Backend README](backend/README.md) · [iOS README](ios/README.md)

## Tech & design

- **Backend:** Go 1.25, standard-library `net/http` routing, `golang-jwt`,
  `bcrypt`, pure-Go — no database server required to run.
- **iOS:** SwiftUI, iOS 16+, async/await `URLSession`, custom `Shape`/`Canvas`
  charts — no third-party Swift packages.
- **Design system:** paper `#F4F3EE`, ink `#1B1B19`, indigo accent `#34357C`,
  sage savings `#5E7E63`, clay high-impulse `#C0824F`; Noto Sans SC + serif
  display numerals. Ported 1:1 from the prototypes.

## License

Provided as a reference implementation for the LIM concept.
