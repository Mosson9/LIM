# LIM Data Model · 数据模型

Every domain entity in `internal/models/models.go`, with field names, JSON tags,
types, and meaning. These structs *are* the API payloads — the iOS `Models/` layer
should mirror them (Codable). Persistence and the derived statistics are explained
at the end. See [API.md](./API.md) for the endpoints and [ARCHITECTURE.md](./ARCHITECTURE.md)
for the product flow.

> **Conventions.** Money is integer CNY (元). Timestamps are `time.Time`
> (serialised RFC3339). A `json:"-"` tag means the field is **never** serialised.
> `omitempty` means the key is omitted when the value is empty/zero/nil.

---

## Enums

```go
type Verdict string        // buy | pause | resist
type DecisionStatus string // pending | resisted | bought | wishlist
type Plan string           // free | month | year
```

| Enum | Value | 中文 | Meaning |
|------|-------|------|---------|
| **Verdict** | `buy` | 可以拥有它 | Reasonable purchase — go ahead. |
| | `pause` | 再想想 / 冷静一下 | Not necessary; enter cooling-off. |
| | `resist` | 也许不必买 | Mostly emotion; skip it. |
| **DecisionStatus** | `pending` | — | Analysed, awaiting the user's choice. |
| | `resisted` | 忍住了 | Counts as savings (`saved = price`). |
| | `bought` | 买了 | Counts as spending. |
| | `wishlist` | — | Parked in the 24h cooling-off list. |
| **Plan** | `free` | 免费 | Daily AI quota applies. |
| | `month` | 月度 | Plus, ¥18/月. |
| | `year` | 年度 | Plus, ¥98/年. |

---

## User

An authenticated account. `password_hash` is never serialised. The financial
`profile` is embedded.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | Friendly id, `U-#####`. |
| Email | `email` | string | Login email (stored/looked up lower-case). |
| Name | `name` | string | Display name (defaults to `LIM 用户`). |
| PasswordHash | `-` | string | bcrypt hash. **Never serialised.** |
| Role | `role` | string | `user` or `admin`. |
| City | `city` | string `omitempty` | Optional city (demo data). |
| CreatedAt | `created_at` | time | Account creation. |
| LastActiveAt | `last_active_at` | time | Last login / activity. |
| Profile | `profile` | Profile | Embedded financial profile (see below). |
| Plan | `plan` | Plan | `free` / `month` / `year`. |
| PlusUntil | `plus_until` | *time `omitempty` | Entitlement expiry (nil = none). |
| AIUsedToday | `ai_used_today` | int | Consults used today (free-quota counter). |
| AIUsedDate | `-` | string | `YYYY-MM-DD` bucket for the counter. **Not serialised.** |
| AppIcon | `app_icon` | string | Selected skin id (default `classic`). |
| Onboarded | `onboarded` | bool | Whether onboarding is complete. |

**Behaviour.** `IsPlus()` returns true when the plan is `month`/`year` **and**
(`plus_until` is nil or in the future). The daily counter resets when
`ai_used_today`'s date bucket no longer matches today.

---

## Profile

Household economics collected at onboarding; makes the analysis personal. Embedded
in `User` and accepted standalone by `PUT /api/v1/me/profile`.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| MonthlyIncome | `monthly_income` | int | 个人月收入. |
| HouseholdIncome | `household_income` | int | 家庭月收入. |
| Members | `members` | int | 家庭成员数 (clamped to ≥ 1). |
| Debt | `debt` | int | 每月固定负债 / 月供. |
| MonthBudget | `month_budget` | int | 可支配预算. If ≤ 0 on update, derived as `max(1000, (monthly_income - debt) / 3)`. |

`month_budget` feeds the heuristic's price/budget ratio (and the LLM prompt).

---

## Dimensions

The six 1–5 sub-scores. **Higher = more justified** on that axis, so a low total
drives a high impulse index. Embedded in `Decision.dims` and in the analysis
`Result`.

| Field | JSON | Type | 中文 | Axis |
|-------|------|------|------|------|
| Need | `need` | float | 需求 | Is it actually necessary? |
| Alt | `alt` | float | 替代 | Could an existing/rented/shared item do? |
| Emo | `emo` | float | 情感 | Genuine want vs. ads/pressure/mood? |
| Value | `value` | float | 长期价值 | Frequency, lifespan, lasting value. |
| Money | `money` | float | 经济 | Fits the budget without crowding out? |
| Env | `env` | float | 环境 | Production/transport/use footprint. |

Scores are snapped to the nearest 0.5 within [1, 5] by the engine.

---

## Decision

One "should I buy this?" enquiry and its outcome. The central entity.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | `D-#####`. |
| UserID | `user_id` | string | Owner. |
| Item | `item` | string | What's being considered. |
| Price | `price` | int | Price in 元 (`> 0`). |
| Cat | `cat` | string | Category id (defaults to `other`). |
| Reason | `reason` | string `omitempty` | User's stated motivation. |
| Dims | `dims` | Dimensions | The six sub-scores. |
| Impulse | `impulse` | int | 0–100 冲动指数. |
| Verdict | `verdict` | Verdict | `buy` / `pause` / `resist`. |
| Message | `message` | string | LIM's natural-language note (persona-styled). |
| Note | `note` | string | Short one-line summary for list rows. |
| Saved | `saved` | int | Amount saved if resisted (= price; else 0). |
| Status | `status` | DecisionStatus | `pending` → `resisted`/`bought`/`wishlist`. |
| CreatedAt | `created_at` | time | When analysed. |
| DecidedAt | `decided_at` | *time `omitempty` | When resolved (nil while pending). |

**Lifecycle.** `/analyze` creates it `pending`. `/decisions/{id}/resolve` (or the
wishlist resolve path) sets `status`, `saved`, and `decided_at`. A `pause` verdict
moving to the wishlist sets `status = wishlist`; resolving the wishlist item later
flips it to `resisted`/`bought`.

---

## WishlistItem

A purchase parked in the 24h cooling-off period.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | UUID. |
| UserID | `user_id` | string | Owner. |
| DecisionID | `decision_id` | string `omitempty` | Source decision (if any). |
| Item | `item` | string | Item name. |
| Price | `price` | int | Price in 元. |
| Cat | `cat` | string | Category id. |
| Impulse | `impulse` | int | Impulse index carried over. |
| AddedAt | `added_at` | time | When parked. |
| ExpiresAt | `expires_at` | time | `added_at + cooling_hours` (default 24h). |

**Progress (computed, not stored).** `WishlistItem.Progress(now)` returns the
`0..1` fraction of the window elapsed. The wishlist endpoints wrap each item in a
**view** that adds three derived fields — see
[API.md → Wishlist](./API.md#wishlist-24h-cooling-off):

| Derived field | JSON | Meaning |
|---------------|------|---------|
| Progress | `progress` | `0..1` of the cooling-off window elapsed. |
| Expired | `expired` | Window fully elapsed. |
| RemainSec | `remain_sec` | Seconds left (≥ 0). |

---

## Transaction

A subscription payment (mock gateway / receipt placeholder). Backs the revenue
dashboard.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | `T-####`. |
| UserID | `user_id` | string | Payer (may be empty for seeded demo rows). |
| UserName | `user_name` | string | Display name (denormalised for the table). |
| Plan | `plan` | string | Plan label, e.g. `年度会员`, `月度续订`. |
| Amount | `amount` | int | Amount in 元. |
| Status | `status` | string | `success` or `refund`. |
| CreatedAt | `created_at` | time | When paid. |

---

## AIConfig

The live configuration of the analysis engine. Read at `GET /api/v1/admin/ai-config`
and `GET /api/v1/meta/dimensions`; replaced wholesale at `PUT /api/v1/admin/ai-config`.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| Dims | `dims` | []DimensionConfig | Per-axis weight/enable + reflective question. |
| PromptTemplate | `prompt_template` | string | LLM system-prompt template (supports `{{item}}`, `{{price}}`, `{{reason}}`, `{{budget}}`, `{{members}}` placeholders). |
| Persona | `persona` | string | `温柔朋友` \| `理性顾问` \| `犀利毒舌` — tone of `message`. |
| Model | `model` | string | Model name shown in the dashboard (default `claude-sonnet-4-6`). |
| FreeDailyLimit | `free_daily_limit` | int | Free-plan `/analyze` calls per day (default 3). |
| CoolingHours | `cooling_hours` | int | Wishlist cooling-off window (default 24). |
| UpdatedAt | `updated_at` | time | Set server-side on save. |

### DimensionConfig

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| Key | `key` | string | `need`/`alt`/`emo`/`value`/`money`/`env`. |
| Label | `label` | string | 中文 label (需求, 替代, …). |
| Desc | `desc` | string | Short description (admin UI). |
| Weight | `weight` | float | Relative weight in the impulse fold. |
| Enabled | `enabled` | bool | Disabled axes are skipped (weight treated as 0). |
| Q | `q` | string | Reflective question shown on the result screen. |

**Default weights (seed):** money 1.4 · need 1.3 · emo 1.2 · alt 1.1 · value 1.0
· env 0.8 — money and need dominate; environment is the gentlest nudge.

---

## Category

A spending category / icon. Served by `/meta/categories` and managed by
`/admin/categories`.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | e.g. `digital`, `fashion`, `home`, `other`. |
| Label | `label` | string | 中文 label (数码电子, …). |
| Color | `color` | string | Hex accent (`#RRGGBB`). |
| Soft | `soft` | string | Hex tint background. |
| Icon | `icon` | string | Icon name (`bolt`, `tag`, `home`, …). |
| Free | `free` | bool | Whether available on the free plan. |
| Count | `count` | int | Analytics: consults this month (seeded sample). |

---

## Skin

A swappable home-screen app icon. Served by `/meta/skins` and `/admin/skins`.
Non-free skins require Plus (enforced by `PUT /api/v1/me/app-icon`).

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | e.g. `classic`, `paper`, `ink`, `sage`. |
| Name | `name` | string | 中文 name (经典靛蓝, 墨黑, …). |
| BG | `bg` | string | Hex background color. |
| Dark | `dark` | bool | Whether the icon art is dark-on-light. |
| Free | `free` | bool | Available on the free plan? |
| Used | `used` | string `omitempty` | Analytics: adoption %, e.g. `48%`. |

---

## PlusPerk

A marketing bullet on the subscription page. Served by `/meta/perks` and in the
subscription response.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| Icon | `icon` | string | Icon name (`spark`, `bolt`, `grid`, `crown`, `leaf`). |
| Title | `title` | string | Headline (每日无限次 AI 咨询, …). |
| Desc | `desc` | string | Supporting line. |

---

## PlanOption

A purchasable subscription tier shown on the paywall. Served by `/meta/plans` and
in the subscription response.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | `month` or `year`. |
| Name | `name` | string | 月度 / 年度. |
| Price | `price` | int | Price in 元 (18 / 98). |
| Per | `per` | string | Period suffix (`/月`, `/年`). |
| Note | `note` | string | Sub-label (随时取消, 省 ¥118 · 最受欢迎). |
| Best | `best` | bool | Highlight as the recommended tier. |

---

## Push

An operations / engagement campaign. Served by `/admin/push` (list) and created by
`POST /admin/push`.

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| ID | `id` | string | `P-###`. |
| Title | `title` | string | Campaign copy (required). |
| Segment | `segment` | string | Target audience (活跃用户, 沉睡用户, …). |
| Sent | `sent` | int | Recipients (analytics). |
| OpenRate | `open_rate` | string | Open rate, e.g. `38.2%`. |
| Status | `status` | string | `scheduled` \| `sent` \| `draft` (default `draft`). |
| Schedule | `schedule` | string | Human schedule (每周日 20:00, 今天 10:00). |
| CreatedAt | `created_at` | time | When created. |

---

## <a id="derived-statistics"></a>Derived statistics

> **Stats are computed, not stored.** There is no "stats" table. `GET /api/v1/stats`
> (and the admin per-user view) calls `computeStats(user)`, which walks the user's
> **resolved** decisions (`ListDecisions(uid, "all")` — excludes `pending`) and
> derives everything live. The response shape (`statsResp`):

| Field | JSON | Type | How it's derived |
|-------|------|------|------------------|
| MonthSaved | `month_saved` | int | Σ `saved` of `resisted` decisions whose effective date is in the current year-month. |
| MonthSpent | `month_spent` | int | Σ `price` of `bought` decisions in the current year-month. |
| TotalSaved | `total_saved` | int | Σ `saved` over all `resisted` decisions. |
| TotalSpent | `total_spent` | int | Σ `price` over all `bought` decisions. |
| ResistCount | `resist_count` | int | Count of `resisted` decisions. |
| BuyCount | `buy_count` | int | Count of `bought` decisions. |
| Streak | `streak` | int | **Days since the most recent purchase** (连续克制天数). If no purchase yet, days since the earliest decision, or the account creation date if there are none. |
| TreeStage | `tree_stage` | int | `0..6` from `total_saved` (thresholds below). |
| RestraintRate | `restraint_rate` | int | `resist_count / (resist_count + buy_count) * 100`, floored. |
| ByMonth | `by_month` | []monthBucket | Trailing **5** months, oldest→newest. |
| TopCats | `top_cats` | []catSaved | Up to **4** categories with the most savings. |

The "effective date" of a decision is `decided_at` when present, else `created_at`.

### Tree stages (省钱树 / growth tree)

Cumulative savings map onto six growth stages. Stage = the number of thresholds
the `total_saved` has reached:

| `total_saved` (元) | `tree_stage` |
|--------------------|--------------|
| `< 500` | 0 |
| `≥ 500` | 1 |
| `≥ 2,000` | 2 |
| `≥ 5,000` | 3 |
| `≥ 9,000` | 4 |
| `≥ 14,000` | 5 |
| `≥ 20,000` | 6 |

### Month buckets

`by_month` is an array of 5 objects, one per month from four months ago to the
current month:

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| Label | `label` | string | Chinese month label (`1月`…`12月`). |
| Spent | `spent` | int | Σ `price` of `bought` decisions that month. |
| Saved | `saved` | int | Σ `saved` of `resisted` decisions that month. |

### Top categories

`top_cats` is up to 4 entries, sorted by `saved` descending, joined with category
metadata for display:

| Field | JSON | Type | Meaning |
|-------|------|------|---------|
| Cat | `cat` | string | Category id. |
| Label | `label` | string | Category 中文 label (falls back to the id). |
| Color | `color` | string | Category hex accent. |
| Saved | `saved` | int | Total savings attributed to this category. |

### Admin aggregates

Admin endpoints derive their own roll-ups the same way (from store records):
verdict split, category distribution, MRR (`月度 ×18 + 年度 ×98/12`), paying-user
counts (excluding admins), `avg_impulse`, `saved_today`, etc. A handful of
dashboard figures — DAU, the DAU/MRR time-series, funnel drop-offs, 续订率 — are
**synthetic** placeholders, because the prototype store doesn't retain historical
time-series. See [API.md → admin/overview](./API.md#get-apiv1adminoverview--admin).

---

## Persistence

All of the above is persisted by `internal/store`. The default implementation is a
**file-backed, concurrency-safe** store:

* In-memory state (`map`s keyed by id for users/decisions/wishlist/transactions/
  pushes, plus slices for the catalogue singletons and per-type sequence counters)
  marshalled to a single indented JSON file (`LIM_DATA_FILE`).
* A **snapshot is written on every mutation**, via an atomic temp-file + rename, so
  a crash can't corrupt the file. Reads/writes are guarded by a `sync.RWMutex`, and
  every getter returns a **copy** so callers can't mutate stored state by accident.
* On startup the file is loaded and an `email → id` index is rebuilt; an empty
  store triggers seeding (catalogue + bootstrap admin + optional demo data).

The serialised top-level shape (`state`):

```json
{
  "users": { "U-20413": { "...User..." : "" } },
  "decisions": { "D-88001": { "...Decision..." : "" } },
  "wishlist": { "<uuid>": { "...WishlistItem..." : "" } },
  "transactions": { "T-5501": { "...Transaction..." : "" } },
  "pushes": { "P-201": { "...Push..." : "" } },
  "ai_config": { "...AIConfig..." : "" },
  "categories": [ "...Category..." ],
  "skins": [ "...Skin..." ],
  "plans": [ "...PlanOption..." ],
  "perks": [ "...PlusPerk..." ],
  "decision_seq": 10, "txn_seq": 6, "user_seq": 0, "push_seq": 5
}
```

Because the HTTP layer depends only on the `store.Store` method surface, this JSON
engine can be swapped for Postgres without touching handlers, models, or routes —
see [DEPLOYMENT.md → Migrating to Postgres](./DEPLOYMENT.md#migrating-to-postgres).
