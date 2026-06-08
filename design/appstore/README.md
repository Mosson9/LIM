# App Store screenshots · 应用市场图

Five marketing screenshots for submitting **LIM** to the App Store, in
**zh-Hans and en**, generated from [`generate.py`](generate.py) (SVG → PNG via
cairosvg; CJK via WenQuanYi, Latin headlines in serif).

| # | zh-Hans | en | Story |
|---|---------|----|-------|
| 1 | `appstore_1.png` | `appstore_en_1.png` | 买之前先问问自己 / Before you buy, ask yourself（首页·已省下） |
| 2 | `appstore_2.png` | `appstore_en_2.png` | 六维分析 / Six angles（冲动指数 + 雷达） |
| 3 | `appstore_3.png` | `appstore_en_3.png` | 省下的每一笔都长成一棵树 / grows into a tree |
| 4 | `appstore_4.png` | `appstore_en_4.png` | 24 小时冷静期 / Park it and sleep on it |
| 5 | `appstore_5.png` | `appstore_en_5.png` | LIM Plus — 无限次咨询 / Unlimited AI |

`contact_sheet.png` / `contact_sheet_en.png` are combined previews. App Store
Connect lets you upload a localized screenshot set per language.

## The full package (20 images)

Both device sizes × both languages × 5 shots — the complete set you can upload
to App Store Connect:

| Size | zh-Hans | en |
|------|---------|----|
| **6.7"/6.9"** · 1290×2796 (primary, required) | `appstore_1..5.png` | `appstore_en_1..5.png` |
| **6.5"** · 1242×2688 (optional) | `appstore_65_1..5.png` | `appstore_en_65_1..5.png` |

- **Format:** opaque RGB PNG (App Store screenshots must not be transparent).
- Up to 10 screenshots per device size per language; these 5 are carousel-ordered.

## Regenerate / customize

```bash
pip install cairosvg pillow
python3 generate.py            # writes all 20 PNGs (+ source SVGs)
```

Sizes are listed in `SIZES`, copy in the bilingual `TR` table, and each screen is
recreated inline — all easy to edit. Colors are the app's design tokens
(paper `#F4F3EE`, ink `#1B1B19`, indigo `#34357C`, sage `#5E7E63`, clay `#C0824F`,
gold `#E9C36B`).

> The device mockups are faithful recreations of the real screens for marketing;
> for pixel-exact frames you can instead capture the running SwiftUI app in the
> simulator and drop those into the same headline layout.
