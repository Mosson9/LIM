#!/usr/bin/env python3
"""Generate the LIM app icon (1024x1024) and 9 themed variants.

The mark is a serif "L" (Less is More) on a brand color, with a subtle top sheen
— matching the in-app brand mark and the "更换 App 图标 / themed icons" feature.

Masters are full-bleed, square, opaque (no rounded corners / alpha) per Apple's
1024 App Store icon rules. A rounded preview sheet is also produced.

    pip install cairosvg pillow
    python3 generate_icons.py
"""
import os
import cairosvg
from PIL import Image, ImageDraw

OUT = os.path.dirname(os.path.abspath(__file__))
SIZE = 1024
SERIF = "DejaVu Serif"

# id, name, background, L color is ink when `dark` else white (matches app skins)
SKINS = [
    ("classic", "经典靛蓝 · Classic", "#34357C", False),
    ("paper",   "暖纸 · Paper",       "#E7E2D6", True),
    ("ink",     "墨黑 · Ink",         "#1B1B19", False),
    ("sage",    "山涧绿 · Sage",      "#5E7E63", False),
    ("clay",    "陶土 · Clay",        "#C0824F", False),
    ("plum",    "紫藤 · Plum",        "#9A6FB0", False),
    ("sky",     "晴空 · Sky",         "#5B86B0", False),
    ("rose",    "胭脂 · Rose",        "#C76B8E", False),
    ("mono",    "雾白 · Mono",        "#F4F3EE", True),
]
INK = "#1B1B19"


def mix(hex1, hex2, t):
    a = [int(hex1[i:i + 2], 16) for i in (1, 3, 5)]
    b = [int(hex2[i:i + 2], 16) for i in (1, 3, 5)]
    c = [round(a[i] + (b[i] - a[i]) * t) for i in range(3)]
    return "#%02x%02x%02x" % tuple(c)


def icon_svg(bg, dark):
    top = mix(bg, "#FFFFFF", 0.13)      # lighter top
    bottom = mix(bg, "#000000", 0.05)   # slightly deeper bottom
    glyph = INK if dark else "#FFFFFF"
    # baseline rule + glyph color accents adapt to light/dark backgrounds
    rule = ("#1B1B19" if dark else "#FFFFFF")
    rule_op = 0.10 if dark else 0.22
    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="{SIZE}" height="{SIZE}" viewBox="0 0 {SIZE} {SIZE}">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0" stop-color="{top}"/>
      <stop offset="1" stop-color="{bottom}"/>
    </linearGradient>
    <radialGradient id="sheen" cx="0.5" cy="0.12" r="0.9">
      <stop offset="0" stop-color="#FFFFFF" stop-opacity="{0.10 if not dark else 0.05}"/>
      <stop offset="0.6" stop-color="#FFFFFF" stop-opacity="0"/>
    </radialGradient>
  </defs>
  <rect width="{SIZE}" height="{SIZE}" fill="url(#bg)"/>
  <rect width="{SIZE}" height="{SIZE}" fill="url(#sheen)"/>
  <text x="512" y="730" font-family="{SERIF}" font-size="640" font-weight="bold"
        fill="{glyph}" text-anchor="middle">L</text>
  <rect x="396" y="772" width="232" height="12" rx="6" fill="{rule}" opacity="{rule_op}"/>
</svg>'''


def rounded(im, radius):
    mask = Image.new("L", im.size, 0)
    d = ImageDraw.Draw(mask)
    d.rounded_rectangle([0, 0, im.size[0], im.size[1]], radius=radius, fill=255)
    out = Image.new("RGBA", im.size, (0, 0, 0, 0))
    out.paste(im, (0, 0), mask)
    return out


def main():
    masters = {}
    for sid, name, bg, dark in SKINS:
        path = os.path.join(OUT, f"icon_{sid}.png")
        cairosvg.svg2png(bytestring=icon_svg(bg, dark).encode(), write_to=path,
                         output_width=SIZE, output_height=SIZE)
        masters[sid] = path
        print("wrote", path)

    # Primary App Store icon = classic.
    Image.open(masters["classic"]).convert("RGB").save(os.path.join(OUT, "AppIcon-1024.png"))
    print("wrote AppIcon-1024.png (= classic)")

    # Rounded 3x3 preview sheet on a neutral board.
    cell, pad, lbl = 300, 46, 40
    cols, rows = 3, 3
    W = cols * cell + pad * (cols + 1)
    H = rows * (cell + lbl) + pad * (rows + 1)
    board = Image.new("RGB", (W, H), (217, 215, 206))
    try:
        from PIL import ImageFont
        font = ImageFont.truetype("/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc", 26)
    except Exception:
        font = None
    draw = ImageDraw.Draw(board)
    for i, (sid, name, bg, dark) in enumerate(SKINS):
        r, c = divmod(i, cols)
        x = pad + c * (cell + pad)
        y = pad + r * (cell + lbl + pad)
        ic = Image.open(masters[sid]).convert("RGBA").resize((cell, cell))
        ic = rounded(ic, int(cell * 0.225))
        board.paste(ic, (x, y), ic)
        if font:
            tw = draw.textlength(name, font=font)
            draw.text((x + (cell - tw) / 2, y + cell + 6), name, fill=(60, 58, 52), font=font)
    board.save(os.path.join(OUT, "icons_preview.png"))
    print("wrote icons_preview.png", board.size)


if __name__ == "__main__":
    main()
