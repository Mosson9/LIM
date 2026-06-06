#!/usr/bin/env python3
"""Generate LIM App Store screenshots (1290x2796, 6.7" portrait).

Each image: a marketing headline + an on-brand device mockup recreating a key
screen. Renders SVG -> PNG with cairosvg (CJK via WenQuanYi Zen Hei).

    pip install cairosvg
    python3 generate.py
"""
import math
import os
import cairosvg

W, H = 1290, 2796
OUT = os.path.dirname(os.path.abspath(__file__))

# ---- palette (ported from the app design tokens) ----
PAPER = "#F4F3EE"; PAPER2 = "#EFEEE7"; SURFACE = "#FFFFFF"
INK = "#1B1B19"; INK2 = "#6A6963"; INK3 = "#9C9A92"; HAIR = "#E7E5DD"
INDIGO = "#34357C"; INDIGO_SOFT = "#ECECF6"; INDIGO_TINT = "#F3F3FA"; INDIGO_INK = "#2A2B5E"
SAVED = "#5E7E63"; SAVED_DEEP = "#46613F"; SAVED_SOFT = "#E9EEE7"
WARN = "#C0824F"; WARN_SOFT = "#F4E9DD"; DANGER = "#B5524E"; DANGER_SOFT = "#F6E5E4"
GOLD = "#E9C36B"; SPENT_SOFT = "#EEEDE7"

CJK = "WenQuanYi Zen Hei"
SERIF = "DejaVu Serif"
SANS = "DejaVu Sans"


def esc(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def text(x, y, s, size, fill, font=CJK, anchor="start", spacing=0, weight="normal", opacity=1):
    sp = f' letter-spacing="{spacing}"' if spacing else ""
    op = f' opacity="{opacity}"' if opacity != 1 else ""
    return (f'<text x="{x}" y="{y}" font-family="{font}" font-size="{size}" '
            f'fill="{fill}" text-anchor="{anchor}" font-weight="{weight}"{sp}{op}>{esc(s)}</text>')


def rrect(x, y, w, h, r, fill, stroke=None, sw=1, opacity=1):
    st = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
    op = f' opacity="{opacity}"' if opacity != 1 else ""
    return f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="{r}" ry="{r}" fill="{fill}"{st}{op}/>'


def circle(cx, cy, r, fill, opacity=1):
    op = f' opacity="{opacity}"' if opacity != 1 else ""
    return f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{fill}"{op}/>'


def polar(cx, cy, r, deg):
    a = math.radians(deg)
    return cx + r * math.cos(a), cy + r * math.sin(a)


def arc_path(cx, cy, r, start, end, sw, color, cap="round"):
    x1, y1 = polar(cx, cy, r, start)
    x2, y2 = polar(cx, cy, r, end)
    large = 1 if (end - start) > 180 else 0
    d = f"M {x1:.1f} {y1:.1f} A {r} {r} 0 {large} 1 {x2:.1f} {y2:.1f}"
    return f'<path d="{d}" fill="none" stroke="{color}" stroke-width="{sw}" stroke-linecap="{cap}"/>'


# ---------- device frame ----------
def device(idx, content, screen_bg=PAPER, dark_home=False):
    """Return SVG for a phone at fixed position containing `content` (drawn in
    screen-local coordinates with origin at the screen's top-left)."""
    pw, ph = 900, 1880
    px, py = (W - pw) // 2, 720
    b = 22
    sx, sy, sw, sh = px + b, py + b, pw - 2 * b, ph - 2 * b
    cid = f"screen{idx}"
    parts = []
    # bezel
    parts.append(rrect(px, py, pw, ph, 86, INK))
    parts.append(rrect(px - 1, py - 1, pw + 2, ph + 2, 87, "none", stroke="#000", sw=2))
    # screen clip + bg
    parts.append(f'<clipPath id="{cid}"><rect x="{sx}" y="{sy}" width="{sw}" height="{sh}" rx="60" ry="60"/></clipPath>')
    parts.append(f'<g clip-path="url(#{cid})">')
    parts.append(rrect(sx, sy, sw, sh, 60, screen_bg))
    # translate content into screen space
    parts.append(f'<g transform="translate({sx},{sy})">')
    parts.append(content(sw, sh))
    parts.append("</g>")
    # status bar
    sb = "#fff" if dark_home else INK
    parts.append(text(sx + 54, sy + 60, "9:41", 30, sb, font=SANS, weight="bold"))
    parts.append(rrect(sx + sw - 150, sy + 40, 28, 20, 4, sb))   # signal-ish
    parts.append(rrect(sx + sw - 112, sy + 40, 26, 20, 4, sb))   # wifi-ish
    parts.append(rrect(sx + sw - 74, sy + 40, 40, 20, 6, "none", stroke=sb, sw=2))  # battery
    parts.append(rrect(sx + sw - 70, sy + 44, 28, 12, 3, sb))
    # dynamic island
    parts.append(rrect(sx + sw / 2 - 65, sy + 36, 130, 38, 19, "#000"))
    parts.append("</g>")
    # home indicator
    hi = "#fff" if dark_home else "rgba(0,0,0,0.28)"
    parts.append(rrect(sx + sw / 2 - 70, sy + sh - 26, 140, 7, 4, hi))
    return "".join(parts)


# ---------- shared screen pieces ----------
def chip(x, y, w, label, fg, bg, size=26):
    return rrect(x, y, w, 52, 26, bg) + text(x + w / 2, y + 35, label, size, fg, anchor="middle")


def card(x, y, w, h, fill=SURFACE):
    return rrect(x, y, w, h, 34, fill, stroke=HAIR, sw=1.5)


# ---------- screen recreations ----------
def screen_home(w, h):
    p = []
    p.append(text(70, 150, "晚上好，林一", 30, INK3))
    p.append(text(70, 200, "少一点，反而更好", 40, INK, weight="bold"))
    # hero savings card
    p.append(card(56, 250, w - 112, 360))
    p.append(text(104, 330, "本月已省下", 28, INK3, spacing=3))
    p.append(text(104, 470, "¥4,733", 130, SAVED_DEEP, font=SERIF))
    # little tree, top-right of card
    p.append(tree_svg(w - 230, 300, 0.85))
    p.append(chip(104, 520, 230, "忍住 14 次", SAVED_DEEP, SAVED_SOFT, 26))
    p.append(chip(348, 520, 280, "连续 5 天克制", INDIGO_INK, INDIGO_SOFT, 26))
    # CTA
    p.append(rrect(56, 660, w - 112, 108, 28, INDIGO))
    p.append(text(w / 2, 726, "我想买点东西，该不该买？", 34, "#fff", anchor="middle"))
    p.append(text(w / 2, 818, "今日还可咨询 1 次 · Plus 无限次", 24, INK3, anchor="middle"))
    # recent decisions
    p.append(text(70, 920, "最近的决定", 30, INK, weight="bold"))
    rows = [("索尼 WH-1000XM5 耳机", "今天 14:20", "省 ¥2,299", SAVED_SOFT, SAVED_DEEP, INDIGO_SOFT, INDIGO),
            ("空气炸锅", "昨天 20:10", "买了", SPENT_SOFT, INK2, SAVED_SOFT, SAVED),
            ("第三件设计师马克杯", "5月27日", "省 ¥168", SAVED_SOFT, SAVED_DEEP, WARN_SOFT, WARN)]
    p.append(card(56, 950, w - 112, 380))
    yy = 1000
    for i, (item, date, badge, bbg, bfg, icbg, icfg) in enumerate(rows):
        p.append(rrect(96, yy, 64, 64, 16, icbg))
        p.append(circle(128, yy + 32, 13, icfg))
        p.append(text(184, yy + 30, item, 28, INK))
        p.append(text(184, yy + 64, date, 22, INK3))
        bw = 150 if "省" in badge else 110
        p.append(chip(w - 96 - bw, yy + 6, bw, badge, bfg, bbg, 24))
        if i < 2:
            p.append(f'<line x1="184" y1="{yy+96}" x2="{w-96}" y2="{yy+96}" stroke="{HAIR}" stroke-width="1.5"/>')
        yy += 118
    # tab bar
    p.append(tabbar(w, h))
    return "".join(p)


def tabbar(w, h):
    p = []
    by = h - 150
    p.append(rrect(40, by, w - 80, 110, 40, SURFACE, stroke=HAIR, sw=1.5))
    labels = [("首页", INDIGO), ("成长", INK3), ("历史", INK3), ("我的", INK3)]
    xs = [130, 300, w - 300, w - 130]
    for (lab, col), x in zip(labels, xs):
        p.append(circle(x, by + 42, 5, col))
        p.append(text(x, by + 86, lab, 22, col, anchor="middle"))
    # FAB
    p.append(circle(w / 2, by + 6, 56, INDIGO))
    p.append(text(w / 2, by + 22, "✦", 48, "#fff", anchor="middle"))
    return "".join(p)


def tree_svg(cx, baseY, scale=1.0):
    s = scale
    p = []
    p.append(f'<ellipse cx="{cx}" cy="{baseY}" rx="{120*s}" ry="22*s" fill="{SAVED_SOFT}"/>'.replace("22*s", str(22 * s)))
    # trunk
    p.append(f'<line x1="{cx}" y1="{baseY}" x2="{cx}" y2="{baseY-150*s}" stroke="#8B7355" stroke-width="{12*s}" stroke-linecap="round"/>')
    for dx, dy, r, col in [(0, -210, 78, SAVED), (-70, -170, 55, "#6E8B6A"), (70, -175, 52, "#557A57"), (20, -255, 46, "#6E8B6A")]:
        p.append(circle(cx + dx * s, baseY + dy * s, r * s, col, opacity=0.95))
    return "".join(p)


def screen_result(w, h):
    p = []
    p.append(text(w / 2, 150, "LIM 的建议", 34, INK, anchor="middle", weight="bold"))
    p.append(text(w / 2, 230, "索尼降噪耳机  ·  ¥2,299", 30, INK2, anchor="middle"))
    # impulse dial
    cx, cy, r = w / 2, 560, 200
    p.append(arc_path(cx, cy, r, 150, 390, 26, PAPER2))
    val = 72
    end = 150 + 240 * val / 100
    p.append(arc_path(cx, cy, r, 150, end, 26, DANGER))
    p.append(text(cx, cy + 30, "72", 150, INK, font=SERIF, anchor="middle"))
    p.append(text(cx, cy + 110, "冲动", 30, DANGER, anchor="middle", spacing=4))
    p.append(text(w / 2, 700, "冲动指数", 24, INK3, anchor="middle", spacing=3))
    # verdict pill
    p.append(rrect(w / 2 - 200, 740, 400, 80, 40, DANGER_SOFT))
    p.append(text(w / 2, 792, "也许不必买", 36, DANGER, anchor="middle"))
    # AI message
    p.append(card(56, 870, w - 112, 250, INDIGO_TINT))
    p.append(rrect(104, 910, 52, 52, 14, INDIGO))
    p.append(text(118, 947, "✦", 30, "#fff"))
    p.append(text(176, 945, "LIM", 28, INDIGO_INK, weight="bold"))
    msg = ["我懂那种心动。但把六个角度摊开看，", "它更多是当下的情绪在说话。", "也许，你已经拥有得够多了。"]
    for i, line in enumerate(msg):
        p.append(text(104, 1010 + i * 44, line, 28, INDIGO_INK))
    # radar
    p.append(card(56, 1160, w - 112, 540))
    p.append(text(104, 1230, "六维分析", 30, INK, weight="bold"))
    p.append(text(w - 104, 1230, "分数越高越值得", 24, INK3, anchor="end"))
    radar(p, w / 2, 1470, 200, [2, 1.5, 1.5, 3, 2, 3])
    return "".join(p)


def radar(p, cx, cy, R, vals):
    labels = ["需求", "替代", "情感", "价值", "经济", "环境"]
    n = 6
    def pt(i, rr):
        a = -math.pi / 2 + i * 2 * math.pi / n
        return cx + math.cos(a) * rr, cy + math.sin(a) * rr
    for g in range(1, 6):
        pts = " ".join(f"{x:.1f},{y:.1f}" for x, y in (pt(i, R * g / 5) for i in range(n)))
        p.append(f'<polygon points="{pts}" fill="none" stroke="{HAIR}" stroke-width="1.5"/>')
    for i in range(n):
        x, y = pt(i, R)
        p.append(f'<line x1="{cx}" y1="{cy}" x2="{x:.1f}" y2="{y:.1f}" stroke="{HAIR}" stroke-width="1.5"/>')
    pts = " ".join(f"{x:.1f},{y:.1f}" for x, y in (pt(i, R * vals[i] / 5) for i in range(n)))
    p.append(f'<polygon points="{pts}" fill="rgba(52,53,124,0.14)" stroke="{INDIGO}" stroke-width="3"/>')
    for i in range(n):
        x, y = pt(i, R * vals[i] / 5)
        p.append(circle(x, y, 6, INDIGO))
    for i in range(n):
        x, y = pt(i, R + 46)
        p.append(text(x, y + 8, labels[i], 24, INK2, anchor="middle"))


def screen_growth(w, h):
    p = []
    p.append(text(70, 170, "我的成长", 30, SAVED_DEEP, spacing=4, weight="bold"))
    # tree card
    p.append(card(56, 210, w - 112, 720, SURFACE))
    p.append(tree_svg(w / 2, 560, 1.5))
    p.append(text(w / 2, 660, "累计省下，种出一棵", 26, INK3, anchor="middle"))
    p.append(text(w / 2, 770, "¥18,640", 110, SAVED_DEEP, font=SERIF, anchor="middle"))
    # goal bar
    p.append(text(104, 840, "距离下一阶段「成树」", 24, INK3))
    p.append(text(w - 104, 840, "75%", 26, SAVED_DEEP, anchor="end", weight="bold"))
    p.append(rrect(104, 860, w - 208, 16, 8, SAVED_SOFT))
    p.append(rrect(104, 860, (w - 208) * 0.75, 16, 8, SAVED))
    # stat tiles
    tiles = [("忍住次数", "14", "次", SAVED_DEEP), ("理性购买", "6", "次", INK),
             ("连续克制", "5", "天", INDIGO), ("克制率", "70", "%", WARN)]
    tw = (w - 112 - 28) / 2
    for i, (lab, val, unit, col) in enumerate(tiles):
        tx = 56 + (i % 2) * (tw + 28)
        ty = 960 + (i // 2) * 230
        p.append(card(tx, ty, tw, 200))
        p.append(text(tx + 40, ty + 70, lab, 26, INK3))
        p.append(text(tx + 40, ty + 160, val, 76, col, font=SERIF))
        p.append(text(tx + 40 + len(val) * 46 + 14, ty + 160, unit, 30, INK3))
    return "".join(p)


def screen_wishlist(w, h):
    p = []
    p.append(text(w / 2, 150, "冷静期 · 心愿单", 34, INK, anchor="middle", weight="bold"))
    # explainer
    p.append(card(56, 220, w - 112, 200, INDIGO_TINT))
    expl = ["还想要的东西，先在这里放 24 小时。", "很多冲动，睡一觉就过去了。", "到期我会提醒你，再决定一次。"]
    for i, line in enumerate(expl):
        p.append(text(104, 290 + i * 44, line, 27, INDIGO_INK))
    items = [("机械键盘 HHKB", "¥899", 64, "剩 7 小时", 0.71),
             ("露营折叠椅", "¥269", 58, "剩 22 小时", 0.08)]
    yy = 470
    for name, price, imp, remain, prog in items:
        p.append(card(56, yy, w - 112, 320))
        p.append(rrect(104, yy + 44, 90, 90, 22, INDIGO_SOFT))
        p.append(circle(149, yy + 89, 22, INDIGO))
        p.append(text(220, yy + 80, name, 32, INK))
        p.append(text(220, yy + 130, price, 28, INK2))
        p.append(chip(360, yy + 102, 150, f"冲动 {imp}", WARN, WARN_SOFT, 24))
        # countdown ring
        rc_x, rc_y = w - 150, yy + 95
        p.append(arc_path(rc_x, rc_y, 40, 0, 359.9, 8, PAPER2))
        p.append(arc_path(rc_x, rc_y, 40, -90, -90 + 360 * prog, 8, INDIGO))
        p.append(text(220, yy + 210, remain, 26, INDIGO, weight="bold"))
        # actions
        p.append(rrect(104, yy + 240, (w - 112 - 80) / 2, 64, 18, SAVED))
        p.append(text(104 + (w - 192) / 4, yy + 282, "不买了", 26, "#fff", anchor="middle"))
        p.append(rrect(104 + (w - 192) / 2 + 40, yy + 240, (w - 112 - 80) / 2, 64, 18, SURFACE, stroke=HAIR, sw=1.5))
        p.append(text(104 + (w - 192) * 3 / 4 + 40, yy + 282, "仍然要买", 26, INK2, anchor="middle"))
        yy += 360
    return "".join(p)


def screen_plus(w, h):
    p = []
    bg = "#1B1B19"
    p.append(rrect(0, 0, w, h, 0, bg))
    p.append(rrect(w / 2 - 56, 150, 112, 112, 30, "rgba(233,195,107,0.16)"))
    p.append(text(w / 2, 235, "♛", 70, GOLD, anchor="middle"))
    p.append(text(w / 2, 360, "LIM Plus", 80, "#fff", font=SERIF, anchor="middle"))
    p.append(text(w / 2, 430, "让克制更轻松一点", 30, "rgba(255,255,255,0.6)", anchor="middle"))
    perks = [("无限次 AI 咨询", "免费版每天 3 次，Plus 想问就问"),
             ("更快的 AI 响应", "优先算力，分析快人一步"),
             ("更换 App 图标", "9 款主题图标，点亮你的桌面"),
             ("专属成长皮肤", "樱花 · 银杏 · 极光")]
    yy = 520
    for title, desc in perks:
        p.append(rrect(70, yy, 64, 64, 16, "rgba(255,255,255,0.07)"))
        p.append(text(102, yy + 44, "✦", 32, GOLD, anchor="middle"))
        p.append(text(160, yy + 30, title, 32, "#fff"))
        p.append(text(160, yy + 70, desc, 24, "rgba(255,255,255,0.5)"))
        yy += 110
    # plans
    py = yy + 30
    pw2 = (w - 140 - 30) / 2
    p.append(rrect(70, py, pw2, 220, 26, "rgba(255,255,255,0.04)", stroke="rgba(255,255,255,0.12)", sw=2))
    p.append(text(70 + 40, py + 60, "月度", 28, "rgba(255,255,255,0.6)"))
    p.append(text(70 + 40, py + 140, "¥18", 64, "#fff", font=SERIF))
    p.append(text(70 + 40, py + 185, "/月", 26, "rgba(255,255,255,0.5)"))
    bx = 70 + pw2 + 30
    p.append(rrect(bx, py, pw2, 220, 26, "rgba(233,195,107,0.12)", stroke=GOLD, sw=2.5))
    p.append(rrect(bx + pw2 - 170, py - 18, 150, 40, 20, GOLD))
    p.append(text(bx + pw2 - 95, py + 9, "最受欢迎", 22, INK, anchor="middle"))
    p.append(text(bx + 40, py + 60, "年度", 28, "rgba(255,255,255,0.6)"))
    p.append(text(bx + 40, py + 140, "¥98", 64, "#fff", font=SERIF))
    p.append(text(bx + 40, py + 185, "省 ¥118", 26, GOLD))
    # CTA
    cy2 = py + 270
    p.append(rrect(70, cy2, w - 140, 104, 28, GOLD))
    p.append(text(w / 2, cy2 + 66, "¥98 / 年 开启 Plus", 36, INK, anchor="middle"))
    return "".join(p)


# ---------- compose a full screenshot ----------
def screenshot(idx, kicker, head_lines, sub, content, bg_top, bg_bot,
               head_fill=INK, kicker_fill=INDIGO, sub_fill=INK2, screen_bg=PAPER, dark_home=False):
    g = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
    g.append(f'<defs><linearGradient id="bg{idx}" x1="0" y1="0" x2="0" y2="1">'
             f'<stop offset="0" stop-color="{bg_top}"/><stop offset="1" stop-color="{bg_bot}"/></linearGradient></defs>')
    g.append(rrect(0, 0, W, H, 0, f"url(#bg{idx})"))
    # headline block
    g.append(text(96, 215, kicker, 34, kicker_fill, spacing=6, weight="bold"))
    y = 320
    for line in head_lines:
        g.append(text(96, y, line, 92, head_fill, font=CJK, weight="bold"))
        y += 110
    g.append(text(96, y + 8, sub, 36, sub_fill))
    # device
    g.append(device(idx, content, screen_bg=screen_bg, dark_home=dark_home))
    g.append("</svg>")
    return "".join(g)


SHOTS = [
    dict(idx=1, kicker="反消费主义 · AI 伙伴", head_lines=["买之前，", "先问问自己。"],
         sub="心动的那一刻，LIM 陪你停下来想清楚", content=screen_home,
         bg_top="#FBFAF6", bg_bot=PAPER2),
    dict(idx=2, kicker="六维分析", head_lines=["AI 从六个角度，", "看清这次心动。"],
         sub="需求 · 替代 · 情感 · 价值 · 经济 · 环境", content=screen_result,
         bg_top=INDIGO_TINT, bg_bot=PAPER, kicker_fill=INDIGO),
    dict(idx=3, kicker="看得见的克制", head_lines=["省下的每一笔，", "都长成一棵树。"],
         sub="忍住的欲望，变成实实在在的积累", content=screen_growth,
         bg_top=SAVED_SOFT, bg_bot=PAPER, kicker_fill=SAVED_DEEP),
    dict(idx=4, kicker="24 小时冷静期", head_lines=["心动时，", "先放进冷静期。"],
         sub="很多冲动，睡一觉就过去了", content=screen_wishlist,
         bg_top=INDIGO_TINT, bg_bot=PAPER, kicker_fill=INDIGO),
    dict(idx=5, kicker="LIM PLUS", head_lines=["解锁无限次", "AI 咨询。"],
         sub="更快的 AI、专属皮肤，陪你一路成长", content=screen_plus,
         bg_top="#26264F", bg_bot="#1B1B19", head_fill="#fff", kicker_fill=GOLD,
         sub_fill="rgba(255,255,255,0.6)", screen_bg="#1B1B19", dark_home=True),
]


def main():
    for s in SHOTS:
        svg = screenshot(**s)
        base = os.path.join(OUT, f"appstore_{s['idx']}")
        with open(base + ".svg", "w", encoding="utf-8") as f:
            f.write(svg)
        cairosvg.svg2png(bytestring=svg.encode("utf-8"), write_to=base + ".png",
                         output_width=W, output_height=H)
        print("wrote", base + ".png")


if __name__ == "__main__":
    main()
