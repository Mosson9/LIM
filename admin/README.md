# LIM · admin web app

The management dashboard for **LIM** — a **zero-build** single-page app (vanilla
JS + a tiny DOM helper, no framework, no bundler) that consumes the Go backend's
`/api/v1/admin/*` endpoints. Design ported 1:1 from the admin prototype.

Pages: 数据总览 · 用户管理 · AI 决策记录 · 订阅与营收 · AI 维度/Prompt · 记账分类 ·
图标皮肤 · 推送/运营.

## Run

Pick whichever is convenient — all three work.

### A. Served by the backend (same origin, no CORS) — recommended

```bash
cd ../backend
LIM_ADMIN_DIR=../admin make run
# open http://localhost:8080/admin/
```

### B. Any static file server (uses CORS, which the API enables by default)

```bash
cd admin
python3 -m http.server 5173
# open http://localhost:5173/?api=http://localhost:8080
```

### C. Docker Compose (bundles backend + admin)

```bash
cd ../backend
docker compose up --build
# open http://localhost:8080/admin/
```

Log in with the seeded admin account: **`admin@lim.app` / `admin123`**.

## Configuring the API base URL

`index.html` resolves the API base in this order:

1. `?api=` query parameter, e.g. `…/admin/?api=https://api.lim.app`
2. same origin when served by the backend
3. otherwise `http://localhost:8080`

## What it talks to

| Page | Endpoint(s) |
|------|-------------|
| 数据总览 | `GET /api/v1/admin/overview` |
| 用户管理 | `GET /api/v1/admin/users`, `GET /api/v1/admin/users/{id}` |
| AI 决策记录 | `GET /api/v1/admin/decisions` |
| 订阅与营收 | `GET /api/v1/admin/subscriptions` |
| AI 维度 / Prompt | `GET` / `PUT /api/v1/admin/ai-config` |
| 记账分类 | `GET` / `PUT /api/v1/admin/categories` |
| 图标皮肤 | `GET /api/v1/admin/skins` |
| 推送 / 运营 | `GET` / `POST /api/v1/admin/push` |

Auth is a JWT obtained from `POST /api/v1/auth/login`; the dashboard requires the
`admin` role and stores the token in `localStorage`. See
[`../docs/API.md`](../docs/API.md) for the full contract.

## Files

```
admin/
├── index.html   # shell + API base resolution
├── styles.css   # design tokens (ported from the prototype)
└── app.js       # SPA: API client, router, pages, inline-SVG charts
```
