# App Store screenshots · 应用市场图

Five marketing screenshots for submitting **LIM** to the App Store, generated
from [`generate.py`](generate.py) (SVG → PNG via cairosvg, CJK via WenQuanYi).

| # | File | Story |
|---|------|-------|
| 1 | `appstore_1.png` | 反消费主义 AI 伙伴 — 买之前先问问自己（首页/本月已省下） |
| 2 | `appstore_2.png` | 六维分析 — AI 从六个角度看清这次心动（冲动指数 + 雷达） |
| 3 | `appstore_3.png` | 省下的每一笔都长成一棵树（成长/累计省下） |
| 4 | `appstore_4.png` | 24 小时冷静期 — 心动时先放进心愿单 |
| 5 | `appstore_5.png` | LIM Plus — 解锁无限次 AI 咨询 |

`contact_sheet.png` is a combined preview of all five.

## Specs

- **Size:** 1290 × 2796 px (portrait) — the iPhone 6.7"/6.9" display size, which
  App Store Connect accepts as the primary required set (it also covers 6.5").
- **Format:** opaque RGB PNG. App Store screenshots must not be transparent.
- Up to 10 screenshots per device size; these 5 are ordered for the carousel.

## Regenerate / customize

```bash
pip install cairosvg pillow
python3 generate.py            # writes appstore_1..5 .svg and .png
```

To also produce the 6.5" set (1242 × 2688), change `W, H` at the top of
`generate.py` and re-run. Headlines, copy, colors and the recreated screens are
all defined inline and easy to edit. Colors are the app's design tokens
(paper `#F4F3EE`, ink `#1B1B19`, indigo `#34357C`, sage `#5E7E63`, clay `#C0824F`,
gold `#E9C36B`).

> The device mockups are faithful recreations of the real screens for marketing;
> for pixel-exact frames you can instead capture the running SwiftUI app in the
> simulator and drop those into the same headline layout.
