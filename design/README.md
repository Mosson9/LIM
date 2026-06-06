# design/

Design tokens extracted from the original LIM prototypes, kept for traceability
so the implementation can be checked against the source of truth.

| File | Source |
|------|--------|
| `app.reference.css`   | Design-system `:root` tokens + component CSS decoded from the **App** prototype (`LIM_App_standalone.html`). |
| `admin.reference.css` | Design-system tokens + component CSS decoded from the **Admin** prototype (`LIM_Admin_standalone.html`). |

These were the basis for:

- the iOS design system in [`../ios/Sources/LIM/Theme/Theme.swift`](../ios/Sources/LIM/Theme/Theme.swift), and
- the seeded catalogue (categories, skins, plans, AI dimensions) in
  [`../backend/internal/seed/seed.go`](../backend/internal/seed/seed.go).

The full prototype HTML files are large self-contained bundles (React + base64
font assets); only the meaningful CSS is retained here.
