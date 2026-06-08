#!/usr/bin/env python3
"""Generate LIM App Store screenshots (1290x2796, 6.7" portrait), zh-Hans + en.

Each image: a marketing headline + an on-brand device mockup recreating a key
screen. Renders SVG -> PNG with cairosvg (CJK via WenQuanYi Zen Hei).

    pip install cairosvg pillow
    python3 generate.py            # writes appstore_1..5 (zh) and appstore_en_1..5 (en)
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

# Active fonts (set per language in generate()).
BODY_FONT = CJK
HEAD_FONT = CJK
LANG = 0  # 0 = zh, 1 = en

# ---- bilingual strings (zh, en) ----
TR = {
    # shot 1 — home
    "k1": ("反消费主义 · AI 伙伴", "ANTI-CONSUMERISM · AI COMPANION"),
    "h1a": ("买之前，", "Before you buy,"),
    "h1b": ("先问问自己。", "ask yourself."),
    "s1": ("心动的那一刻，LIM 陪你停下来想清楚", "The moment you're tempted, LIM pauses with you."),
    "greet": ("晚上好，林一", "Good evening, Lin"),
    "tagline": ("少一点，反而更好", "Less, but better"),
    "saved_month": ("本月已省下", "SAVED THIS MONTH"),
    "resist_n": ("忍住 14 次", "Resisted 14×"),
    "streak": ("连续 5 天克制", "5-day streak"),
    "cta": ("我想买点东西，该不该买？", "Thinking of buying something?"),
    "cta_sub": ("今日还可咨询 1 次 · Plus 无限次", "1 free ask left · Plus is unlimited"),
    "recent": ("最近的决定", "Recent decisions"),
    "item_hp": ("索尼 WH-1000XM5 耳机", "Sony WH-1000XM5"),
    "item_af": ("空气炸锅", "Air fryer"),
    "item_mug": ("第三件设计师马克杯", "3rd designer mug"),
    "d_today": ("今天 14:20", "Today 14:20"),
    "d_yest": ("昨天 20:10", "Yest. 20:10"),
    "d_may": ("5月27日", "May 27"),
    "saved2299": ("省 ¥2,299", "Saved ¥2,299"),
    "bought": ("买了", "Bought"),
    "saved168": ("省 ¥168", "Saved ¥168"),
    "tab_home": ("首页", "Home"), "tab_growth": ("成长", "Growth"),
    "tab_hist": ("历史", "History"), "tab_me": ("我的", "Me"),
    # shot 2 — result
    "k2": ("六维分析", "SIX-DIMENSION ANALYSIS"),
    "h2a": ("AI 从六个角度，", "Six angles to see"),
    "h2b": ("看清这次心动。", "a craving clearly."),
    "s2": ("需求 · 替代 · 情感 · 价值 · 经济 · 环境", "Need · Alternatives · Emotion · Value · Cost · Planet"),
    "lim_take": ("LIM 的建议", "LIM's take"),
    "result_item": ("索尼降噪耳机  ·  ¥2,299", "Sony headphones  ·  ¥2,299"),
    "impulsive": ("冲动", "Impulsive"),
    "impulse_index": ("冲动指数", "IMPULSE INDEX"),
    "maybe_skip": ("也许不必买", "Maybe skip it"),
    "msg_a": ("我懂那种心动。但把六个角度摊开看，", "I get the spark. But across six angles,"),
    "msg_b": ("它更多是当下的情绪在说话。", "this is mostly the moment talking."),
    "msg_c": ("也许，你已经拥有得够多了。", "Maybe you already have enough."),
    "six_dim": ("六维分析", "Six dimensions"),
    "higher_worth": ("分数越高越值得", "Higher = more worth it"),
    "dim_need": ("需求", "Need"), "dim_alt": ("替代", "Alt"), "dim_emo": ("情感", "Emotion"),
    "dim_val": ("价值", "Value"), "dim_money": ("经济", "Cost"), "dim_env": ("环境", "Planet"),
    # shot 3 — growth
    "k3": ("看得见的克制", "VISIBLE RESTRAINT"),
    "h3a": ("省下的每一笔，", "Every yuan saved"),
    "h3b": ("都长成一棵树。", "grows into a tree."),
    "s3": ("忍住的欲望，变成实实在在的积累", "Resisted wants become real, lasting savings."),
    "my_growth": ("我的成长", "MY GROWTH"),
    "grow_caption": ("累计省下，种出一棵", "Total saved — growing a tree"),
    "to_next": ("距离下一阶段「成树」", "To the next stage «Tree»"),
    "t_resist": ("忍住次数", "Resisted"), "t_wise": ("理性购买", "Wise buys"),
    "t_streak": ("连续克制", "Streak"), "t_rate": ("克制率", "Restraint"),
    "u_times": ("次", "×"), "u_days": ("天", "d"), "u_pct": ("%", "%"),
    # shot 4 — wishlist
    "k4": ("24 小时冷静期", "24-HOUR COOLING-OFF"),
    "h4a": ("心动时，", "Tempted? Park it"),
    "h4b": ("先放进冷静期。", "and sleep on it."),
    "s4": ("很多冲动，睡一觉就过去了", "Most impulses pass by morning."),
    "wish_title": ("冷静期 · 心愿单", "Cooling-off · Wishlist"),
    "ex_a": ("还想要的东西，先在这里放 24 小时。", "Park what you still want here for 24 hours."),
    "ex_b": ("很多冲动，睡一觉就过去了。", "Most impulses pass after a night's sleep."),
    "ex_c": ("到期我会提醒你，再决定一次。", "I'll remind you to decide again."),
    "item_kb": ("机械键盘 HHKB", "HHKB keyboard"),
    "item_chair": ("露营折叠椅", "Camping chair"),
    "rem_7h": ("剩 7 小时", "7h left"), "rem_22h": ("剩 22 小时", "22h left"),
    "imp_64": ("冲动 64", "Impulse 64"), "imp_58": ("冲动 58", "Impulse 58"),
    "skip_it": ("不买了", "Skip it"), "buy_anyway": ("仍然要买", "Buy anyway"),
    # shot 5 — plus
    "k5": ("LIM PLUS", "LIM PLUS"),
    "h5a": ("解锁无限次", "Unlimited AI"),
    "h5b": ("AI 咨询。", "consultations."),
    "s5": ("更快的 AI、专属皮肤，陪你一路成长", "Faster AI, exclusive skins, and more."),
    "plus_sub": ("让克制更轻松一点", "Make restraint a little easier"),
    "perk1t": ("无限次 AI 咨询", "Unlimited AI asks"),
    "perk1d": ("免费版每天 3 次，Plus 想问就问", "Free is 3/day; Plus, ask anytime"),
    "perk2t": ("更快的 AI 响应", "Faster AI responses"),
    "perk2d": ("优先算力，分析快人一步", "Priority compute, a step ahead"),
    "perk3t": ("更换 App 图标", "Swap the app icon"),
    "perk3d": ("9 款主题图标，点亮你的桌面", "9 themed icons for your home screen"),
    "perk4t": ("专属成长皮肤", "Exclusive growth skins"),
    "perk4d": ("樱花 · 银杏 · 极光", "Sakura · Ginkgo · Aurora"),
    "monthly": ("月度", "Monthly"), "yearly": ("年度", "Yearly"),
    "popular": ("最受欢迎", "POPULAR"), "save118": ("省 ¥118", "Save ¥118"),
    "per_mo": ("/月", "/mo"),
    "plus_cta": ("¥98 / 年 开启 Plus", "Start Plus · ¥98/yr"),
}


def tr(k):
    return TR[k][LANG]


def esc(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def text(x, y, s, size, fill, font=None, anchor="start", spacing=0, weight="normal", opacity=1):
    if font is None:
        font = BODY_FONT
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
    pw, ph = 900, 1880
    px, py = (W - pw) // 2, 720
    b = 22
    sx, sy, sw, sh = px + b, py + b, pw - 2 * b, ph - 2 * b
    cid = f"screen{idx}"
    p = []
    p.append(rrect(px, py, pw, ph, 86, INK))
    p.append(rrect(px - 1, py - 1, pw + 2, ph + 2, 87, "none", stroke="#000", sw=2))
    p.append(f'<clipPath id="{cid}"><rect x="{sx}" y="{sy}" width="{sw}" height="{sh}" rx="60" ry="60"/></clipPath>')
    p.append(f'<g clip-path="url(#{cid})">')
    p.append(rrect(sx, sy, sw, sh, 60, screen_bg))
    p.append(f'<g transform="translate({sx},{sy})">')
    p.append(content(sw, sh))
    p.append("</g>")
    sb = "#fff" if dark_home else INK
    p.append(text(sx + 54, sy + 60, "9:41", 30, sb, font=SANS, weight="bold"))
    p.append(rrect(sx + sw - 150, sy + 40, 28, 20, 4, sb))
    p.append(rrect(sx + sw - 112, sy + 40, 26, 20, 4, sb))
    p.append(rrect(sx + sw - 74, sy + 40, 40, 20, 6, "none", stroke=sb, sw=2))
    p.append(rrect(sx + sw - 70, sy + 44, 28, 12, 3, sb))
    p.append(rrect(sx + sw / 2 - 65, sy + 36, 130, 38, 19, "#000"))
    p.append("</g>")
    hi = "#fff" if dark_home else "rgba(0,0,0,0.28)"
    p.append(rrect(sx + sw / 2 - 70, sy + sh - 26, 140, 7, 4, hi))
    return "".join(p)


def chip(x, y, w, label, fg, bg, size=26):
    return rrect(x, y, w, 52, 26, bg) + text(x + w / 2, y + 35, label, size, fg, anchor="middle")


def card(x, y, w, h, fill=SURFACE):
    return rrect(x, y, w, h, 34, fill, stroke=HAIR, sw=1.5)


def tree_svg(cx, baseY, scale=1.0):
    s = scale
    p = [f'<ellipse cx="{cx}" cy="{baseY}" rx="{120*s}" ry="{22*s}" fill="{SAVED_SOFT}"/>']
    p.append(f'<line x1="{cx}" y1="{baseY}" x2="{cx}" y2="{baseY-150*s}" stroke="#8B7355" stroke-width="{12*s}" stroke-linecap="round"/>')
    for dx, dy, r, col in [(0, -210, 78, SAVED), (-70, -170, 55, "#6E8B6A"), (70, -175, 52, "#557A57"), (20, -255, 46, "#6E8B6A")]:
        p.append(circle(cx + dx * s, baseY + dy * s, r * s, col, opacity=0.95))
    return "".join(p)


def tabbar(w, h):
    p = []
    by = h - 150
    p.append(rrect(40, by, w - 80, 110, 40, SURFACE, stroke=HAIR, sw=1.5))
    for (key, col), x in zip([("tab_home", INDIGO), ("tab_growth", INK3), ("tab_hist", INK3), ("tab_me", INK3)],
                             [130, 300, w - 300, w - 130]):
        p.append(circle(x, by + 42, 5, col))
        p.append(text(x, by + 86, tr(key), 22, col, anchor="middle"))
    p.append(circle(w / 2, by + 6, 56, INDIGO))
    p.append(text(w / 2, by + 22, "✦", 48, "#fff", anchor="middle"))
    return "".join(p)


def radar(p, cx, cy, R, vals):
    labels = [tr("dim_need"), tr("dim_alt"), tr("dim_emo"), tr("dim_val"), tr("dim_money"), tr("dim_env")]
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


# ---------- screens ----------
def screen_home(w, h):
    p = []
    p.append(text(70, 150, tr("greet"), 30, INK3))
    p.append(text(70, 200, tr("tagline"), 40, INK, weight="bold"))
    p.append(card(56, 250, w - 112, 360))
    p.append(text(104, 330, tr("saved_month"), 28, INK3, spacing=3))
    p.append(text(104, 470, "¥4,733", 130, SAVED_DEEP, font=SERIF))
    p.append(tree_svg(w - 230, 300, 0.85))
    p.append(chip(104, 520, 230, tr("resist_n"), SAVED_DEEP, SAVED_SOFT, 26))
    p.append(chip(348, 520, 300, tr("streak"), INDIGO_INK, INDIGO_SOFT, 26))
    p.append(rrect(56, 660, w - 112, 108, 28, INDIGO))
    p.append(text(w / 2, 726, tr("cta"), 34, "#fff", anchor="middle"))
    p.append(text(w / 2, 818, tr("cta_sub"), 24, INK3, anchor="middle"))
    p.append(text(70, 920, tr("recent"), 30, INK, weight="bold"))
    rows = [(tr("item_hp"), tr("d_today"), tr("saved2299"), SAVED_SOFT, SAVED_DEEP, INDIGO_SOFT, INDIGO, 200),
            (tr("item_af"), tr("d_yest"), tr("bought"), SPENT_SOFT, INK2, SAVED_SOFT, SAVED, 130),
            (tr("item_mug"), tr("d_may"), tr("saved168"), SAVED_SOFT, SAVED_DEEP, WARN_SOFT, WARN, 170)]
    p.append(card(56, 950, w - 112, 380))
    yy = 1000
    for i, (item, date, badge, bbg, bfg, icbg, icfg, bw) in enumerate(rows):
        p.append(rrect(96, yy, 64, 64, 16, icbg))
        p.append(circle(128, yy + 32, 13, icfg))
        p.append(text(184, yy + 30, item, 28, INK))
        p.append(text(184, yy + 64, date, 22, INK3))
        p.append(chip(w - 96 - bw, yy + 6, bw, badge, bfg, bbg, 24))
        if i < 2:
            p.append(f'<line x1="184" y1="{yy+96}" x2="{w-96}" y2="{yy+96}" stroke="{HAIR}" stroke-width="1.5"/>')
        yy += 118
    p.append(tabbar(w, h))
    return "".join(p)


def screen_result(w, h):
    p = []
    p.append(text(w / 2, 150, tr("lim_take"), 34, INK, anchor="middle", weight="bold"))
    p.append(text(w / 2, 230, tr("result_item"), 30, INK2, anchor="middle"))
    cx, cy, r = w / 2, 560, 200
    p.append(arc_path(cx, cy, r, 150, 390, 26, PAPER2))
    p.append(arc_path(cx, cy, r, 150, 150 + 240 * 72 / 100, 26, DANGER))
    p.append(text(cx, cy + 30, "72", 150, INK, font=SERIF, anchor="middle"))
    p.append(text(cx, cy + 110, tr("impulsive"), 30, DANGER, anchor="middle", spacing=4))
    p.append(text(w / 2, 700, tr("impulse_index"), 24, INK3, anchor="middle", spacing=3))
    p.append(rrect(w / 2 - 220, 740, 440, 80, 40, DANGER_SOFT))
    p.append(text(w / 2, 792, tr("maybe_skip"), 36, DANGER, anchor="middle"))
    p.append(card(56, 870, w - 112, 250, INDIGO_TINT))
    p.append(rrect(104, 910, 52, 52, 14, INDIGO))
    p.append(text(118, 947, "✦", 30, "#fff"))
    p.append(text(176, 945, "LIM", 28, INDIGO_INK, weight="bold", font=SANS))
    for i, key in enumerate(["msg_a", "msg_b", "msg_c"]):
        p.append(text(104, 1010 + i * 44, tr(key), 27, INDIGO_INK))
    p.append(card(56, 1160, w - 112, 540))
    p.append(text(104, 1230, tr("six_dim"), 30, INK, weight="bold"))
    p.append(text(w - 104, 1230, tr("higher_worth"), 24, INK3, anchor="end"))
    radar(p, w / 2, 1470, 200, [2, 1.5, 1.5, 3, 2, 3])
    return "".join(p)


def screen_growth(w, h):
    p = []
    p.append(text(70, 170, tr("my_growth"), 30, SAVED_DEEP, spacing=4, weight="bold"))
    p.append(card(56, 210, w - 112, 720, SURFACE))
    p.append(tree_svg(w / 2, 560, 1.5))
    p.append(text(w / 2, 660, tr("grow_caption"), 26, INK3, anchor="middle"))
    p.append(text(w / 2, 770, "¥18,640", 110, SAVED_DEEP, font=SERIF, anchor="middle"))
    p.append(text(104, 840, tr("to_next"), 24, INK3))
    p.append(text(w - 104, 840, "75%", 26, SAVED_DEEP, anchor="end", weight="bold"))
    p.append(rrect(104, 860, w - 208, 16, 8, SAVED_SOFT))
    p.append(rrect(104, 860, (w - 208) * 0.75, 16, 8, SAVED))
    tiles = [(tr("t_resist"), "14", tr("u_times"), SAVED_DEEP), (tr("t_wise"), "6", tr("u_times"), INK),
             (tr("t_streak"), "5", tr("u_days"), INDIGO), (tr("t_rate"), "70", tr("u_pct"), WARN)]
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
    p.append(text(w / 2, 150, tr("wish_title"), 34, INK, anchor="middle", weight="bold"))
    p.append(card(56, 220, w - 112, 200, INDIGO_TINT))
    for i, key in enumerate(["ex_a", "ex_b", "ex_c"]):
        p.append(text(104, 290 + i * 44, tr(key), 26, INDIGO_INK))
    items = [(tr("item_kb"), "¥899", tr("imp_64"), tr("rem_7h"), 0.71),
             (tr("item_chair"), "¥269", tr("imp_58"), tr("rem_22h"), 0.08)]
    yy = 470
    for name, price, imp, remain, prog in items:
        p.append(card(56, yy, w - 112, 320))
        p.append(rrect(104, yy + 44, 90, 90, 22, INDIGO_SOFT))
        p.append(circle(149, yy + 89, 22, INDIGO))
        p.append(text(220, yy + 80, name, 32, INK))
        p.append(text(220, yy + 130, price, 28, INK2))
        p.append(chip(360, yy + 102, 180, imp, WARN, WARN_SOFT, 24))
        rc_x, rc_y = w - 150, yy + 95
        p.append(arc_path(rc_x, rc_y, 40, 0, 359.9, 8, PAPER2))
        p.append(arc_path(rc_x, rc_y, 40, -90, -90 + 360 * prog, 8, INDIGO))
        p.append(text(220, yy + 210, remain, 26, INDIGO, weight="bold"))
        p.append(rrect(104, yy + 240, (w - 112 - 80) / 2, 64, 18, SAVED))
        p.append(text(104 + (w - 192) / 4, yy + 282, tr("skip_it"), 26, "#fff", anchor="middle"))
        p.append(rrect(104 + (w - 192) / 2 + 40, yy + 240, (w - 112 - 80) / 2, 64, 18, SURFACE, stroke=HAIR, sw=1.5))
        p.append(text(104 + (w - 192) * 3 / 4 + 40, yy + 282, tr("buy_anyway"), 26, INK2, anchor="middle"))
        yy += 360
    return "".join(p)


def screen_plus(w, h):
    p = [rrect(0, 0, w, h, 0, "#1B1B19")]
    p.append(rrect(w / 2 - 56, 150, 112, 112, 30, "rgba(233,195,107,0.16)"))
    p.append(text(w / 2, 235, "♛", 70, GOLD, anchor="middle"))
    p.append(text(w / 2, 360, "LIM Plus", 80, "#fff", font=SERIF, anchor="middle"))
    p.append(text(w / 2, 430, tr("plus_sub"), 30, "rgba(255,255,255,0.6)", anchor="middle"))
    perks = [("perk1t", "perk1d"), ("perk2t", "perk2d"), ("perk3t", "perk3d"), ("perk4t", "perk4d")]
    yy = 520
    for tk, dk in perks:
        p.append(rrect(70, yy, 64, 64, 16, "rgba(255,255,255,0.07)"))
        p.append(text(102, yy + 44, "✦", 32, GOLD, anchor="middle"))
        p.append(text(160, yy + 30, tr(tk), 32, "#fff"))
        p.append(text(160, yy + 70, tr(dk), 24, "rgba(255,255,255,0.5)"))
        yy += 110
    py = yy + 30
    pw2 = (w - 140 - 30) / 2
    p.append(rrect(70, py, pw2, 220, 26, "rgba(255,255,255,0.04)", stroke="rgba(255,255,255,0.12)", sw=2))
    p.append(text(110, py + 60, tr("monthly"), 28, "rgba(255,255,255,0.6)"))
    p.append(text(110, py + 140, "¥18", 64, "#fff", font=SERIF))
    p.append(text(110, py + 185, tr("per_mo"), 26, "rgba(255,255,255,0.5)"))
    bx = 70 + pw2 + 30
    p.append(rrect(bx, py, pw2, 220, 26, "rgba(233,195,107,0.12)", stroke=GOLD, sw=2.5))
    p.append(rrect(bx + pw2 - 180, py - 18, 160, 40, 20, GOLD))
    p.append(text(bx + pw2 - 100, py + 9, tr("popular"), 22, INK, anchor="middle"))
    p.append(text(bx + 40, py + 60, tr("yearly"), 28, "rgba(255,255,255,0.6)"))
    p.append(text(bx + 40, py + 140, "¥98", 64, "#fff", font=SERIF))
    p.append(text(bx + 40, py + 185, tr("save118"), 26, GOLD))
    cy2 = py + 270
    p.append(rrect(70, cy2, w - 140, 104, 28, GOLD))
    p.append(text(w / 2, cy2 + 66, tr("plus_cta"), 36, INK, anchor="middle"))
    return "".join(p)


def screenshot(idx, kicker_k, head_ka, head_kb, sub_k, content, bg_top, bg_bot,
               head_fill=INK, kicker_fill=INDIGO, sub_fill=INK2, screen_bg=PAPER, dark_home=False):
    g = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}">']
    g.append(f'<defs><linearGradient id="bg{idx}" x1="0" y1="0" x2="0" y2="1">'
             f'<stop offset="0" stop-color="{bg_top}"/><stop offset="1" stop-color="{bg_bot}"/></linearGradient></defs>')
    g.append(rrect(0, 0, W, H, 0, f"url(#bg{idx})"))
    g.append(text(96, 215, tr(kicker_k), 34, kicker_fill, spacing=6, weight="bold", font=BODY_FONT))
    y = 320
    for k in (head_ka, head_kb):
        g.append(text(96, y, tr(k), 92, head_fill, font=HEAD_FONT, weight="bold"))
        y += 110
    g.append(text(96, y + 8, tr(sub_k), 36, sub_fill))
    g.append(device(idx, content, screen_bg=screen_bg, dark_home=dark_home))
    g.append("</svg>")
    return "".join(g)


SHOTS = [
    dict(idx=1, kicker_k="k1", head_ka="h1a", head_kb="h1b", sub_k="s1", content=screen_home,
         bg_top="#FBFAF6", bg_bot=PAPER2),
    dict(idx=2, kicker_k="k2", head_ka="h2a", head_kb="h2b", sub_k="s2", content=screen_result,
         bg_top=INDIGO_TINT, bg_bot=PAPER, kicker_fill=INDIGO),
    dict(idx=3, kicker_k="k3", head_ka="h3a", head_kb="h3b", sub_k="s3", content=screen_growth,
         bg_top=SAVED_SOFT, bg_bot=PAPER, kicker_fill=SAVED_DEEP),
    dict(idx=4, kicker_k="k4", head_ka="h4a", head_kb="h4b", sub_k="s4", content=screen_wishlist,
         bg_top=INDIGO_TINT, bg_bot=PAPER, kicker_fill=INDIGO),
    dict(idx=5, kicker_k="k5", head_ka="h5a", head_kb="h5b", sub_k="s5", content=screen_plus,
         bg_top="#26264F", bg_bot="#1B1B19", head_fill="#fff", kicker_fill=GOLD,
         sub_fill="rgba(255,255,255,0.6)", screen_bg="#1B1B19", dark_home=True),
]


def generate(lang_idx, lang_suffix, width, height, size_suffix):
    global LANG, BODY_FONT, HEAD_FONT, W, H
    LANG = lang_idx
    W, H = width, height
    BODY_FONT = CJK if lang_idx == 0 else SANS
    HEAD_FONT = CJK if lang_idx == 0 else SERIF
    for s in SHOTS:
        svg = screenshot(**s)
        base = os.path.join(OUT, f"appstore{lang_suffix}{size_suffix}_{s['idx']}")
        with open(base + ".svg", "w", encoding="utf-8") as f:
            f.write(svg)
        cairosvg.svg2png(bytestring=svg.encode("utf-8"), write_to=base + ".png",
                         output_width=W, output_height=H)
        print("wrote", base + ".png")


# App Store device sizes: 6.7"/6.9" (primary, required) and 6.5" (optional).
SIZES = [("", 1290, 2796), ("_65", 1242, 2688)]


def main():
    for size_suffix, w, h in SIZES:
        generate(0, "", w, h, size_suffix)       # zh
        generate(1, "_en", w, h, size_suffix)    # en


if __name__ == "__main__":
    main()
