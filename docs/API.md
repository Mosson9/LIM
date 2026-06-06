# LIM REST API Reference · 接口文档

The LIM backend exposes one REST API that serves **both** the consumer iOS app
and the admin dashboard. This is the complete reference, derived directly from
`internal/httpapi/router.go` (the source of truth) and the handler structs.

* **Base URL:** `http://localhost:8080` in development; your host in production.
* **API prefix:** all application routes live under `/api/v1`. Health is at the root.
* **Content type:** request and response bodies are JSON (`application/json; charset=utf-8`).
* **Unknown fields:** request bodies are decoded with `DisallowUnknownFields()` —
  sending an unexpected JSON key returns **400**.
* **Errors:** non-2xx responses use the envelope `{"error": "<message>"}`. Messages
  are Chinese (user-facing).

---

## Authentication

LIM uses stateless **JWT bearer tokens** (HS256). Obtain one from
`POST /api/v1/auth/register` or `POST /api/v1/auth/login`, then send it on every
protected request:

```
Authorization: Bearer <token>
```

Tokens carry the user id and role and expire after `LIM_TOKEN_TTL_HOURS`
(default **720h / 30 days**). Passwords are hashed with bcrypt.

### Auth levels

| Level | Requirement | Failure |
|-------|-------------|---------|
| **public** | none | — |
| **user**   | valid bearer token | `401 {"error":"未授权，请重新登录"}` |
| **admin**  | valid bearer token **and** `role == "admin"` | `401` if no/invalid token, `403 {"error":"需要管理员权限"}` if non-admin |

### Daily free-quota (402)

Free-plan users may call `POST /api/v1/analyze` at most `AIConfig.FreeDailyLimit`
times per calendar day (default **3**). The counter resets at local midnight. When
the quota is exhausted the endpoint returns:

```
402 Payment Required
{"error":"今日免费咨询次数已用完，升级 Plus 可无限使用"}
```

Plus users (`plan` is `month` or `year` with an active entitlement) are unlimited.
The same `402` pattern guards Plus-only icon skins on `PUT /api/v1/me/app-icon`
(`{"error":"该图标为 Plus 专属"}`).

### Demo credentials (seeded)

When `LIM_SEED_DEMO=true` (the default), the store is seeded on first boot:

| Account | Email | Password | Role / Plan |
|---------|-------|----------|-------------|
| Admin | `admin@lim.app` | `admin123` | admin · 年度 |
| 林一 (Lin) | `lin@lim.app` | `password` | user · free (has demo decisions + live wishlist) |
| 苏晚 (Su) | `su@lim.app` | `password` | user · 年度 (Plus) |
| 陈默 (Chen) | `chen@lim.app` | `password` | user · 月度 (Plus) |
| 江予安, 周禾, 吴桐, 郑清, 何笙, 许念, 梁知秋 | `jiang@/zhou@/wu@/zheng@/he@/xu@/liang@lim.app` | `password` | user (mixed plans) |

> Admin email/password are overridable via `LIM_ADMIN_EMAIL` / `LIM_ADMIN_PASSWORD`.
> **Change them for any non-local deployment.**

---

## Route index

| Area | Method & path | Auth |
|------|---------------|------|
| Health | `GET /healthz` | public |
| Auth | `POST /api/v1/auth/register` | public |
| Auth | `POST /api/v1/auth/login` | public |
| Meta | `GET /api/v1/meta/categories` | public |
| Meta | `GET /api/v1/meta/skins` | public |
| Meta | `GET /api/v1/meta/plans` | public |
| Meta | `GET /api/v1/meta/perks` | public |
| Meta | `GET /api/v1/meta/dimensions` | public |
| Me | `GET /api/v1/me` | user |
| Me | `PUT /api/v1/me/profile` | user |
| Me | `POST /api/v1/me/onboard` | user |
| Me | `PUT /api/v1/me/app-icon` | user |
| Decisions | `POST /api/v1/analyze` | user |
| Decisions | `GET /api/v1/decisions` | user |
| Decisions | `GET /api/v1/decisions/{id}` | user |
| Decisions | `POST /api/v1/decisions/{id}/resolve` | user |
| Wishlist | `GET /api/v1/wishlist` | user |
| Wishlist | `POST /api/v1/wishlist` | user |
| Wishlist | `POST /api/v1/wishlist/{id}/resolve` | user |
| Wishlist | `DELETE /api/v1/wishlist/{id}` | user |
| Stats | `GET /api/v1/stats` | user |
| Subscription | `GET /api/v1/subscription` | user |
| Subscription | `POST /api/v1/subscription/verify` | user |
| Subscription | `POST /api/v1/subscription/subscribe` | user (dev mock) |
| Subscription | `POST /api/v1/appstore/notifications` | public (JWS-signed) |
| Admin | `GET /api/v1/admin/overview` | admin |
| Admin | `GET /api/v1/admin/users` | admin |
| Admin | `GET /api/v1/admin/users/{id}` | admin |
| Admin | `GET /api/v1/admin/decisions` | admin |
| Admin | `GET /api/v1/admin/subscriptions` | admin |
| Admin | `GET /api/v1/admin/ai-config` | admin |
| Admin | `PUT /api/v1/admin/ai-config` | admin |
| Admin | `GET /api/v1/admin/categories` | admin |
| Admin | `PUT /api/v1/admin/categories` | admin |
| Admin | `GET /api/v1/admin/skins` | admin |
| Admin | `GET /api/v1/admin/push` | admin |
| Admin | `POST /api/v1/admin/push` | admin |

### Enums

| Enum | Values |
|------|--------|
| `verdict` | `buy` (可以拥有它) · `pause` (再想想) · `resist` (也许不必买) |
| `status` (decision) | `pending` · `resisted` (忍住了) · `bought` (买了) · `wishlist` |
| `plan` | `free` · `month` · `year` |
| `action` (resolve) | `resist` · `buy` |

---

## Health

### `GET /healthz` · public

Liveness probe. Also reports whether the LLM engine is active.

```bash
curl -s http://localhost:8080/healthz
```

```json
{ "status": "ok", "llm": false }
```

`llm` is `true` only when `ANTHROPIC_API_KEY` is configured.

---

## Auth

### `POST /api/v1/auth/register` · public

Create an account and receive a token. Email is lower-cased & trimmed; password
must be ≥ 6 chars and the email must contain `@`. `name` defaults to `LIM 用户`.
An optional `profile` may be supplied (see [Profile](./DATA_MODEL.md#profile)).

Request:

```json
{
  "email": "new@lim.app",
  "password": "secret123",
  "name": "新用户",
  "profile": {
    "monthly_income": 18000,
    "household_income": 32000,
    "members": 3,
    "debt": 4200,
    "month_budget": 6000
  }
}
```

Response `200` — `{ "token": "<jwt>", "user": <User> }`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "U-20001",
    "email": "new@lim.app",
    "name": "新用户",
    "role": "user",
    "created_at": "2026-06-06T10:00:00Z",
    "last_active_at": "2026-06-06T10:00:00Z",
    "profile": { "monthly_income": 18000, "household_income": 32000, "members": 3, "debt": 4200, "month_budget": 6000 },
    "plan": "free",
    "ai_used_today": 0,
    "app_icon": "classic",
    "onboarded": false
  }
}
```

Errors: `400` invalid email/password; `409 {"error":"该邮箱已注册"}` duplicate email.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"new@lim.app","password":"secret123","name":"新用户"}'
```

### `POST /api/v1/auth/login` · public

Exchange credentials for a token. Updates the user's last-active timestamp.

Request:

```json
{ "email": "lin@lim.app", "password": "password" }
```

Response `200` — same `{ "token", "user" }` shape as register.
Error: `401 {"error":"邮箱或密码错误"}`.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"lin@lim.app","password":"password"}'
```

> Tip — capture the token for subsequent calls:
> ```bash
> TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
>   -H 'Content-Type: application/json' \
>   -d '{"email":"lin@lim.app","password":"password"}' | jq -r .token)
> ```

---

## Meta / Catalogue

Public, read-only catalogue content used to populate the app. All five are simple
arrays.

### `GET /api/v1/meta/categories` · public

Returns the spending categories (`[]Category`).

```bash
curl -s http://localhost:8080/api/v1/meta/categories
```

```json
[
  { "id": "digital", "label": "数码电子", "color": "#4B4DAE", "soft": "#ECECF6", "icon": "bolt", "free": true, "count": 3240 },
  { "id": "fashion", "label": "服饰穿搭", "color": "#9A6FB0", "soft": "#F0EAF4", "icon": "tag", "free": true, "count": 2680 }
]
```

### `GET /api/v1/meta/skins` · public

Returns the swappable app-icon skins (`[]Skin`). `free: false` skins require Plus.

```json
[
  { "id": "classic", "name": "经典靛蓝", "bg": "#34357C", "dark": false, "free": true, "used": "48%" },
  { "id": "ink", "name": "墨黑", "bg": "#1B1B19", "dark": false, "free": false, "used": "12%" }
]
```

### `GET /api/v1/meta/plans` · public

Returns purchasable subscription tiers (`[]PlanOption`).

```json
[
  { "id": "month", "name": "月度", "price": 18, "per": "/月", "note": "随时取消", "best": false },
  { "id": "year",  "name": "年度", "price": 98, "per": "/年", "note": "省 ¥118 · 最受欢迎", "best": true }
]
```

### `GET /api/v1/meta/perks` · public

Returns Plus marketing bullets (`[]PlusPerk`).

```json
[
  { "icon": "spark", "title": "每日无限次 AI 咨询", "desc": "免费版每天 3 次，Plus 想问就问" }
]
```

### `GET /api/v1/meta/dimensions` · public

Returns the six reflective questions shown on the result screen, sourced from the
live AI config. Each item is `{ key, label, q }`:

```bash
curl -s http://localhost:8080/api/v1/meta/dimensions
```

```json
[
  { "key": "need",  "label": "需求", "q": "这是生活中必需的吗？没有它，日子会变得困难或不便吗？" },
  { "key": "alt",   "label": "替代", "q": "你是否已经拥有类似的东西？能不能靠租借、修理或共享来满足？" },
  { "key": "emo",   "label": "情感", "q": "购买的动机是什么？是真的喜欢，还是广告、社交压力或一时情绪？" },
  { "key": "value", "label": "长期价值", "q": "它的使用频率和寿命如何？能长期为你带来价值吗？" },
  { "key": "money", "label": "经济", "q": "价格和你的预算匹配吗？会挤占其他更重要的开支吗？" },
  { "key": "env",   "label": "环境", "q": "它的生产、运输和使用对环境意味着什么？符合你的价值观吗？" }
]
```

---

## Me / Profile

All require a user token.

### `GET /api/v1/me` · user

Returns the current user (`User`). Refreshes the daily AI-usage counter if the
date rolled over. `password_hash` is never serialised.

```bash
curl -s http://localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"
```

```json
{
  "id": "U-20413", "email": "lin@lim.app", "name": "林一", "role": "user",
  "city": "上海", "created_at": "2025-04-07T...", "last_active_at": "2026-06-06T...",
  "profile": { "monthly_income": 18000, "household_income": 32000, "members": 3, "debt": 4200, "month_budget": 6000 },
  "plan": "free", "ai_used_today": 0, "app_icon": "classic", "onboarded": true
}
```

### `PUT /api/v1/me/profile` · user

Update the financial profile. `members` is clamped to ≥ 1; if `month_budget` ≤ 0
the server derives a default as `max(1000, (monthly_income - debt) / 3)`.
Returns the updated `User`.

Request (`Profile`):

```json
{ "monthly_income": 20000, "household_income": 35000, "members": 2, "debt": 5000, "month_budget": 7000 }
```

```bash
curl -s -X PUT http://localhost:8080/api/v1/me/profile \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"monthly_income":20000,"household_income":35000,"members":2,"debt":5000,"month_budget":7000}'
```

### `POST /api/v1/me/onboard` · user

Mark onboarding complete (`onboarded = true`). No body required. Returns the
updated `User`.

```bash
curl -s -X POST http://localhost:8080/api/v1/me/onboard -H "Authorization: Bearer $TOKEN"
```

### `PUT /api/v1/me/app-icon` · user

Set the home-screen app icon. The icon id must exist in `/meta/skins`; Plus-only
skins (`free: false`) require an active subscription. Returns the updated `User`.

Request:

```json
{ "app_icon": "ink" }
```

Errors: `400 {"error":"未知图标"}`; `402 {"error":"该图标为 Plus 专属"}`.

```bash
curl -s -X PUT http://localhost:8080/api/v1/me/app-icon \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"app_icon":"paper"}'
```

---

## Analyze & Decisions

The core flow. All require a user token.

### `POST /api/v1/analyze` · user

Run the six-dimension analysis and persist a **pending** `Decision`. Counts
against the daily free quota (see [402 behavior](#daily-free-quota-402)).

Request (`analyzeReq`): `item` and `price` are required (`price > 0`); `cat`
defaults to `"other"`; `reason` is optional but improves the analysis.

```json
{ "item": "索尼 WH-1000XM5 耳机", "price": 2299, "cat": "digital", "reason": "通勤想要降噪，但已经有一副耳机了" }
```

Response `200` — the created `Decision` with `status: "pending"`:

```json
{
  "id": "D-88011",
  "user_id": "U-20413",
  "item": "索尼 WH-1000XM5 耳机",
  "price": 2299,
  "cat": "digital",
  "reason": "通勤想要降噪，但已经有一副耳机了",
  "dims": { "need": 3, "alt": 1.5, "emo": 3, "value": 3, "money": 1, "env": 3 },
  "impulse": 72,
  "verdict": "resist",
  "message": "我懂那种心动。但把六个角度摊开看……",
  "note": "你已经有类似的东西，或能租借共享。",
  "saved": 0,
  "status": "pending",
  "created_at": "2026-06-06T10:05:00Z"
}
```

Errors: `400` missing item/price; `402` quota exhausted (free plan).

```bash
curl -s -X POST http://localhost:8080/api/v1/analyze \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"item":"索尼 WH-1000XM5 耳机","price":2299,"cat":"digital","reason":"通勤降噪"}'
```

> The score is produced by the Claude engine when configured, otherwise by the
> deterministic heuristic. See [ARCHITECTURE.md](./ARCHITECTURE.md#2-the-product-concept--产品理念)
> for how the six dimensions fold into the impulse index and verdict.

### `GET /api/v1/decisions` · user

List the caller's decisions, newest first. Optional `?filter=` query:

| `filter` | Returns |
|----------|---------|
| `all` / omitted | all *resolved* decisions (excludes `pending`) |
| `resist` | only `resisted` |
| `buy` | only `bought` |
| `pending` | only `pending` |

```bash
curl -s "http://localhost:8080/api/v1/decisions?filter=resist" -H "Authorization: Bearer $TOKEN"
```

Returns `[]Decision`.

### `GET /api/v1/decisions/{id}` · user

Fetch one decision by id (must belong to the caller). Returns `Decision`.
Error: `404 {"error":"未找到该决策"}`.

```bash
curl -s http://localhost:8080/api/v1/decisions/D-88011 -H "Authorization: Bearer $TOKEN"
```

### `POST /api/v1/decisions/{id}/resolve` · user

Finalise a pending decision. `resist` credits savings (`saved = price`, status
`resisted`); `buy` records spending (`saved = 0`, status `bought`). Sets
`decided_at`. Returns the updated `Decision`.

Request (`resolveReq`):

```json
{ "action": "resist" }
```

Errors: `400 {"error":"action 必须为 resist 或 buy"}`; `404` not found / not owner.

```bash
curl -s -X POST http://localhost:8080/api/v1/decisions/D-88011/resolve \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"action":"resist"}'
```

---

## Wishlist (24h cooling-off)

The `pause` path. All require a user token. List/add/resolve responses use a
**decorated** view that adds live countdown fields to the stored item:

| Field | Type | Meaning |
|-------|------|---------|
| `progress` | float | `0..1` fraction of the cooling-off window elapsed |
| `expired` | bool | window fully elapsed |
| `remain_sec` | int | seconds left (≥ 0) |

(plus all `WishlistItem` fields: `id`, `user_id`, `decision_id`, `item`, `price`,
`cat`, `impulse`, `added_at`, `expires_at`).

### `GET /api/v1/wishlist` · user

List the caller's cooling-off items, soonest-expiring first. Returns
`[]wishlistView`.

```bash
curl -s http://localhost:8080/api/v1/wishlist -H "Authorization: Bearer $TOKEN"
```

```json
[
  {
    "id": "5f1c...", "user_id": "U-20413", "decision_id": "",
    "item": "机械键盘 HHKB", "price": 899, "cat": "digital", "impulse": 64,
    "added_at": "2026-06-05T17:00:00Z", "expires_at": "2026-06-06T17:00:00Z",
    "progress": 0.71, "expired": false, "remain_sec": 25200
  }
]
```

### `POST /api/v1/wishlist` · user

Park an item in the cooling-off list. `expires_at` is set to
`now + AIConfig.CoolingHours` (default 24h). If `decision_id` is supplied, that
decision is marked `wishlist`. Returns `201` with the `wishlistView`.

Request (`addWishlistReq`): `item` and `price` required; `cat` defaults to
`"other"`; `impulse` and `decision_id` optional.

```json
{ "item": "机械键盘 HHKB", "price": 899, "cat": "digital", "impulse": 64, "decision_id": "D-88012" }
```

```bash
curl -s -X POST http://localhost:8080/api/v1/wishlist \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"item":"机械键盘 HHKB","price":899,"cat":"digital","impulse":64}'
```

### `POST /api/v1/wishlist/{id}/resolve` · user

Resolve a cooling-off item after the window. `resist` → savings, `buy` → spend.
If the item is linked to a decision, that decision is updated to match; otherwise
a resolved decision is synthesised so it appears in history & stats. The wishlist
entry is then deleted.

Request (`resolveReq`):

```json
{ "action": "resist" }
```

Response `200`:

```json
{ "ok": true, "action": "resist", "saved": true }
```

Errors: `400` bad action; `404` not found / not owner.

```bash
curl -s -X POST http://localhost:8080/api/v1/wishlist/5f1c.../resolve \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"action":"buy"}'
```

### `DELETE /api/v1/wishlist/{id}` · user

Remove a wishlist item without recording a decision. Returns `{ "ok": true }`.
Error: `404` not found / not owner.

```bash
curl -s -X DELETE http://localhost:8080/api/v1/wishlist/5f1c... -H "Authorization: Bearer $TOKEN"
```

---

## Stats

### `GET /api/v1/stats` · user

Returns the statistics payload backing the Home and Growth screens. **All values
are derived on the fly** from the caller's resolved decisions — nothing here is
stored. See [DATA_MODEL.md → Derived statistics](./DATA_MODEL.md#derived-statistics)
for the exact computation (month/total saved & spent, streak as days-since-last-
purchase, tree stage thresholds, 5-month buckets, top categories).

```bash
curl -s http://localhost:8080/api/v1/stats -H "Authorization: Bearer $TOKEN"
```

```json
{
  "month_saved": 868,
  "month_spent": 88,
  "total_saved": 3967,
  "total_spent": 616,
  "resist_count": 3,
  "buy_count": 2,
  "streak": 20,
  "tree_stage": 2,
  "restraint_rate": 60,
  "by_month": [
    { "label": "2月", "spent": 0, "saved": 0 },
    { "label": "3月", "spent": 0, "saved": 0 },
    { "label": "4月", "spent": 0, "saved": 0 },
    { "label": "5月", "spent": 88, "saved": 1399 },
    { "label": "6月", "spent": 0, "saved": 2568 }
  ],
  "top_cats": [
    { "cat": "fashion", "label": "服饰穿搭", "color": "#9A6FB0", "saved": 1399 },
    { "cat": "home", "label": "家居好物", "color": "#5E7E63", "saved": 599 }
  ]
}
```

---

## Subscription

All require a user token.

### `GET /api/v1/subscription` · user

Returns the caller's entitlement plus the catalogue of plans & perks.

```bash
curl -s http://localhost:8080/api/v1/subscription -H "Authorization: Bearer $TOKEN"
```

```json
{
  "plan": "free",
  "is_plus": false,
  "plans": [ { "id": "month", "name": "月度", "price": 18, "per": "/月", "note": "随时取消", "best": false }, { "id": "year", "name": "年度", "price": 98, "per": "/年", "note": "省 ¥118 · 最受欢迎", "best": true } ],
  "perks": [ { "icon": "spark", "title": "每日无限次 AI 咨询", "desc": "免费版每天 3 次，Plus 想问就问" } ]
}
```

`plus_until` (RFC3339) is included when the user has an entitlement.

### `POST /api/v1/subscription/verify` · user

**Production purchase path.** Validates an Apple StoreKit 2 signed transaction and
grants the entitlement until the transaction's expiry. The server verifies the
JWS certificate chain against Apple's root (configured via `LIM_APPLE_ROOT_CERT`),
checks the bundle id / environment, maps the `productId` to a plan, sets
`plan` + `plus_until`, and records a `Transaction`. Returns the same shape as
`GET /subscription`.

Request (`verifyReq`): `jws` is StoreKit 2's `Transaction.jwsRepresentation`.

```json
{ "jws": "eyJhbGciOiJFUzI1NiIsIng1YyI6Wy4uLl19.eyJ0cmFuc2FjdGlvbklkIjoi..." }
```

Errors: `400 {"error":"缺少 jws 凭证"}`, `400 {"error":"凭证校验失败：…"}`,
`400 {"error":"未知的商品：…"}`, `501 {"error":"未配置 App Store 校验"}`.

```bash
curl -s -X POST http://localhost:8080/api/v1/subscription/verify \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"jws":"<StoreKit2 jwsRepresentation>"}'
```

### `POST /api/v1/appstore/notifications` · public (JWS-signed)

App Store Server Notifications **V2** webhook. Register this URL in App Store
Connect. No bearer token — authenticity comes from the JWS signature. The server
verifies the `signedPayload` (and its nested `signedTransactionInfo`), finds the
account via `originalTransactionId`, and updates the entitlement:

- `SUBSCRIBED` / `DID_RENEW` / `DID_RECOVER` / `RESUBSCRIBE` / `OFFER_REDEEMED` →
  extend Plus to the new expiry (renewals also record a transaction).
- `EXPIRED` / `REVOKE` / `GRACE_PERIOD_EXPIRED` → downgrade to free.
- `REFUND` → downgrade to free and record a refund transaction.

Request: `{ "signedPayload": "<App Store notification JWS>" }`. Returns `200` once
the payload is verified (unknown subscriptions are acknowledged and ignored);
`400` on a malformed/unverifiable payload (Apple will retry).

### `POST /api/v1/subscription/subscribe` · user

**Dev-only mock** purchase (gated by `LIM_ALLOW_MOCK_SUBSCRIBE`, default on):
grants the entitlement, extends `plus_until` (stacking onto remaining time), and
records a success `Transaction`. Returns the same shape as `GET /subscription`.
Production should disable this (`LIM_ALLOW_MOCK_SUBSCRIBE=false`) and use
`/subscription/verify`.

Request (`subscribeReq`): `plan` must be `month` (¥18, +30 days) or `year`
(¥98, +365 days).

```json
{ "plan": "year" }
```

Errors: `400 {"error":"plan 必须为 month 或 year"}`,
`403 {"error":"模拟开通已禁用…"}` when mock is disabled.

```bash
curl -s -X POST http://localhost:8080/api/v1/subscription/subscribe \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"plan":"year"}'
```

---

## Admin

All admin routes require a token whose user has `role == "admin"` (otherwise
`403`). Sign in as `admin@lim.app` / `admin123` to obtain one:

```bash
ADMIN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@lim.app","password":"admin123"}' | jq -r .token)
```

### `GET /api/v1/admin/overview` · admin

The dashboard home. Aggregates KPIs (累计用户 / DAU / MRR / 帮用户省下), a verdict
donut split, category distribution, a conversion funnel, a synthetic DAU series
for charts, and the 8 most recent decisions (each enriched with the user's name).

```bash
curl -s http://localhost:8080/api/v1/admin/overview -H "Authorization: Bearer $ADMIN"
```

```json
{
  "kpis": [ { "key": "users", "label": "累计用户", "val": "10", "delta": "+6.2%", "dir": "up", "sub": "本月新增用户" } ],
  "dau_series": [ 1, 1, 2, "...30 values..." ],
  "dau_labels": [ "5/8", "", "..." ],
  "verdict_split": [ { "label": "建议不买", "value": 5, "color": "#5E7E63" }, { "label": "建议买入", "value": 3, "color": "#34357C" }, { "label": "冷静一下", "value": 2, "color": "#C0824F" } ],
  "cat_dist": [ { "label": "数码电子", "value": 3, "color": "#4B4DAE" } ],
  "funnel": [ { "step": "打开 App", "value": 3, "pct": 1 }, { "step": "发起咨询", "value": 2, "pct": 0.62 } ],
  "recent": [ { "...Decision fields...": "", "user": "苏晚" } ],
  "generated_at": "2026-06-06T10:00:00Z"
}
```

> Admin (`role == "admin"`) users are **excluded** from the user/paying counts.
> Several figures (DAU = 26% of users, funnel drop-offs, the DAU/MRR time-series,
> 续订率 82.4%) are **synthetic** — the prototype store does not track historical
> time-series. Treat them as illustrative.

### `GET /api/v1/admin/users` · admin

List users (excluding admins) with computed activity columns: `status`
(`new`/`active`/`dormant`/`churned`), `consults`, `saved`, `resist`, `last_seen`.
Supports `?q=` (matches name / id / city) and `?filter=`
(`plus` / `free` / `risk` — risk = dormant or churned).

```bash
curl -s "http://localhost:8080/api/v1/admin/users?filter=plus" -H "Authorization: Bearer $ADMIN"
```

### `GET /api/v1/admin/users/{id}` · admin

Full detail for one user: `{ "user": <User>, "stats": <statsResp>, "decisions": [<Decision>] }`.
Error: `404 {"error":"未找到用户"}`.

```bash
curl -s http://localhost:8080/api/v1/admin/users/U-20413 -H "Authorization: Bearer $ADMIN"
```

### `GET /api/v1/admin/decisions` · admin

Global decision feed (newest first), each enriched with `user` (name). Optional
`?filter=` by verdict (`buy`/`pause`/`resist`) or `all`. Includes aggregate
`count`, `avg_impulse`, `saved_today`.

```json
{ "decisions": [ { "...Decision...": "", "user": "苏晚" } ], "count": 10, "avg_impulse": 58, "saved_today": 0 }
```

```bash
curl -s "http://localhost:8080/api/v1/admin/decisions?filter=resist" -H "Authorization: Bearer $ADMIN"
```

### `GET /api/v1/admin/subscriptions` · admin

Revenue dashboard: paying-user KPIs, MRR, a synthetic 12-month MRR series,
plan split (年度/月度), and the transaction ledger (`[]Transaction`).

```bash
curl -s http://localhost:8080/api/v1/admin/subscriptions -H "Authorization: Bearer $ADMIN"
```

```json
{
  "kpis": [ { "label": "付费用户", "val": "5", "sub": "Plus 会员" } ],
  "mrr_series": [ "...12 values..." ],
  "mrr_labels": [ "7月", "8月", "..." ],
  "plan_split": [ { "label": "年度 ¥98", "value": 4, "color": "#34357C" }, { "label": "月度 ¥18", "value": 2, "color": "#4B4DAE" } ],
  "transactions": [ { "id": "T-5501", "user_name": "何笙", "plan": "年度会员", "amount": 98, "status": "success", "created_at": "..." } ]
}
```

### `GET /api/v1/admin/ai-config` · admin

Return the live engine configuration (`AIConfig`): the six `dims` (key/label/desc/
weight/enabled/q), `prompt_template`, `persona`, `model`, `free_daily_limit`,
`cooling_hours`, `updated_at`.

```bash
curl -s http://localhost:8080/api/v1/admin/ai-config -H "Authorization: Bearer $ADMIN"
```

### `PUT /api/v1/admin/ai-config` · admin

Replace the engine configuration. Body is a full `AIConfig` (at least one
dimension required). `updated_at` is set server-side. Returns the saved config.
Changing weights/persona/prompt/quota takes effect on the next `/analyze`.

Error: `400 {"error":"至少需要一个维度"}` if `dims` is empty.

```bash
curl -s -X PUT http://localhost:8080/api/v1/admin/ai-config \
  -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"dims":[{"key":"need","label":"需求","desc":"...","weight":1.5,"enabled":true,"q":"..."}],"persona":"理性顾问","model":"claude-sonnet-4-6","free_daily_limit":5,"cooling_hours":24}'
```

### `GET /api/v1/admin/categories` · admin

Return the category catalogue (`[]Category`) — same shape as `/meta/categories`.

```bash
curl -s http://localhost:8080/api/v1/admin/categories -H "Authorization: Bearer $ADMIN"
```

### `PUT /api/v1/admin/categories` · admin

Replace the category catalogue. Body is `[]Category`. Returns the saved list.

```bash
curl -s -X PUT http://localhost:8080/api/v1/admin/categories \
  -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '[{"id":"digital","label":"数码电子","color":"#4B4DAE","soft":"#ECECF6","icon":"bolt","free":true,"count":0}]'
```

### `GET /api/v1/admin/skins` · admin

Return the app-icon skin catalogue (`[]Skin`) — same shape as `/meta/skins`.
(There is **no** admin write endpoint for skins; they are seed-managed.)

```bash
curl -s http://localhost:8080/api/v1/admin/skins -H "Authorization: Bearer $ADMIN"
```

### `GET /api/v1/admin/push` · admin

List push/engagement campaigns (`[]Push`), newest first.

```bash
curl -s http://localhost:8080/api/v1/admin/push -H "Authorization: Bearer $ADMIN"
```

### `POST /api/v1/admin/push` · admin

Create a campaign. `title` is required; `status` defaults to `draft`. Returns
`201` with the created `Push` (assigned a `P-###` id).

Request (`Push`):

```json
{ "title": "本周你已省下 ¥860", "segment": "活跃用户", "schedule": "每周日 20:00", "status": "scheduled", "sent": 0, "open_rate": "" }
```

Error: `400 {"error":"缺少推送文案"}` if `title` is blank.

```bash
curl -s -X POST http://localhost:8080/api/v1/admin/push \
  -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"title":"本周你已省下 ¥860","segment":"活跃用户","schedule":"每周日 20:00","status":"scheduled"}'
```

---

## Conventions summary

* **IDs:** users `U-#####`, decisions `D-#####`, transactions `T-####`, pushes
  `P-###`, wishlist items are UUIDs.
* **Timestamps:** RFC3339 (`created_at`, `decided_at`, `expires_at`, …).
* **Money:** integer CNY (元); no fractional currency.
* **Ownership:** user-scoped resources verify the caller owns the record; a
  mismatch returns `404` (not `403`) so existence isn't leaked.
* **CORS:** controlled by `LIM_CORS_ORIGIN` (default `*`); `OPTIONS` preflight is
  answered with `204`.
