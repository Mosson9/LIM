#!/usr/bin/env python3
"""Render the LIM legal/support Markdown into styled, hostable HTML.

Reads docs/legal/*.md, substitutes {{PLACEHOLDERS}} from site/values.json, and
writes a small static site (index + privacy + support) into site/dist/ — ready
to publish on GitHub Pages (see .github/workflows/pages.yml).

    pip install markdown
    python3 site/build.py
"""
import json
import os
import re
from urllib.parse import quote
import markdown

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SITE = os.path.join(ROOT, "site")
DIST = os.path.join(SITE, "dist")
LEGAL = os.path.join(ROOT, "docs", "legal")

# LIM brand favicon: the serif "L" on indigo, inlined so there are no binaries.
_FAVICON_SVG = (
    "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'>"
    "<rect width='64' height='64' rx='14' fill='#34357C'/>"
    "<text x='32' y='47' font-family='Georgia,serif' font-size='42' font-weight='bold'"
    " text-anchor='middle' fill='white'>L</text></svg>"
)
FAVICON = "data:image/svg+xml," + quote(_FAVICON_SVG)

CSS = """
:root{
  --paper:#F4F3EE;--surface:#fff;--ink:#1B1B19;--ink2:#4a4a45;--ink3:#6A6963;
  --hair:#E7E5DD;--indigo:#34357C;--indigo2:#4B4DAE;--sage:#5E7E63;
  --serif:"Newsreader",Georgia,"Songti SC","Noto Serif SC",serif;
  --sans:-apple-system,BlinkMacSystemFont,"Segoe UI","Noto Sans SC",sans-serif;
}
*{box-sizing:border-box}
html{-webkit-text-size-adjust:100%}
body{margin:0;background:var(--paper);color:var(--ink);font-family:var(--sans);
  line-height:1.75;font-size:17px}
.wrap{max-width:760px;margin:0 auto;padding:0 22px}
header.site{border-bottom:1px solid var(--hair);background:rgba(252,251,247,.85);
  backdrop-filter:blur(8px);position:sticky;top:0;z-index:5}
header.site .wrap{display:flex;align-items:center;gap:12px;height:64px}
.mark{width:36px;height:36px;border-radius:10px;background:var(--indigo);color:#fff;
  font-family:var(--serif);font-weight:700;font-size:22px;display:flex;
  align-items:center;justify-content:center;text-decoration:none}
.brand{font-weight:600}.brand small{color:var(--ink3);font-weight:400}
nav.site{margin-left:auto;display:flex;gap:18px;font-size:15px}
nav.site a{color:var(--ink2);text-decoration:none}
nav.site a:hover{color:var(--indigo)}
main{padding:40px 0 64px}
article h1{font-family:var(--serif);font-size:40px;line-height:1.2;margin:.2em 0 .4em;letter-spacing:-.3px}
article h2{font-family:var(--serif);font-size:27px;margin:1.8em 0 .5em;
  padding-top:1.2em;border-top:1px solid var(--hair)}
article h2:first-of-type{border-top:none;padding-top:0}
article h3{font-size:19px;margin:1.4em 0 .3em}
article p,article li{color:var(--ink2)}
article a{color:var(--indigo);text-decoration:none;border-bottom:1px solid rgba(52,53,124,.25)}
article a:hover{color:var(--indigo2)}
article strong{color:var(--ink)}
article hr{border:none;border-top:1px solid var(--hair);margin:2.4em 0}
article blockquote{margin:1.4em 0;padding:12px 18px;background:#fff;
  border:1px solid var(--hair);border-left:3px solid var(--sage);border-radius:10px;
  color:var(--ink3);font-size:15px}
article code{background:#EFEEE7;padding:2px 6px;border-radius:6px;font-size:.9em}
table{border-collapse:collapse;width:100%;margin:1.2em 0;font-size:15px}
th,td{border:1px solid var(--hair);padding:9px 12px;text-align:left}
th{background:#fff}
.hero{font-family:var(--serif);font-size:clamp(34px,6vw,52px);line-height:1.15;
  letter-spacing:-.5px;margin:.4em 0 .3em}
.lede{font-size:19px;color:var(--ink3);max-width:560px}
.cards{display:grid;grid-template-columns:1fr 1fr;gap:18px;margin:36px 0}
@media(max-width:620px){.cards{grid-template-columns:1fr}}
.card{display:block;background:#fff;border:1px solid var(--hair);border-radius:18px;
  padding:24px;text-decoration:none;color:inherit;transition:.15s}
.card:hover{box-shadow:0 8px 28px rgba(27,27,25,.08);transform:translateY(-1px)}
.card h3{margin:.2em 0;font-size:20px}
.card p{margin:.2em 0 0;color:var(--ink3);font-size:15px}
.kicker{font-size:12px;font-weight:700;letter-spacing:1.5px;text-transform:uppercase;color:var(--indigo)}
footer.site{border-top:1px solid var(--hair);color:var(--ink3);font-size:14px;padding:28px 0}
footer.site .wrap{display:flex;flex-wrap:wrap;gap:6px 18px;align-items:center}
footer.site a{color:var(--ink3)}
"""

TEMPLATE = """<!doctype html>
<html lang="{lang}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title>
<meta name="description" content="{desc}">
<link rel="icon" href="{favicon}">
<meta property="og:title" content="{title}">
<meta property="og:type" content="website">
<style>{css}</style>
</head>
<body>
<header class="site"><div class="wrap">
  <a class="mark" href="./index.html">L</a>
  <span class="brand">{app}</span>
  <nav class="site">
    <a href="./index.html">Home</a>
    <a href="./privacy.html">Privacy</a>
    <a href="./support.html">Support</a>
  </nav>
</div></header>
<main><div class="wrap">{body}</div></main>
<footer class="site"><div class="wrap">
  <span>© {year} {company}</span><span>·</span>
  <a href="./privacy.html">Privacy</a><a href="./support.html">Support</a>
  <a href="{website}">{website}</a>
</div></footer>
</body>
</html>
"""


def load_values():
    with open(os.path.join(SITE, "values.json"), encoding="utf-8") as f:
        return json.load(f)


def substitute(text, values):
    def repl(m):
        return values.get(m.group(1), m.group(0))
    return re.sub(r"\{\{([A-Z_]+)\}\}", repl, text)


def normalize_md(text):
    """Make hand-authored Markdown render reliably:
    - turn "· " pseudo-bullets into real list items;
    - ensure a blank line precedes a list that follows a text line (Markdown
      otherwise treats it as a lazy paragraph continuation)."""
    text = re.sub(r"(?m)^(\s*)·\s", r"\1- ", text)
    text = re.sub(r"(?m)^((?:(?![-*]\s)\S).*)\n([-*]\s)", r"\1\n\n\2", text)
    return text


def render_md(path, values):
    with open(path, encoding="utf-8") as f:
        text = f.read()
    text = re.sub(r"^<!--.*?-->\s*", "", text, flags=re.DOTALL)  # drop leading HTML comment
    text = substitute(text, values)
    text = normalize_md(text)
    return markdown.markdown(text, extensions=["extra", "sane_lists"])


def page(title, desc, body, values, lang="en"):
    return TEMPLATE.format(
        lang=lang, title=title, desc=desc, app=values["APP_NAME"],
        company=values["COMPANY"], website=values["WEBSITE"],
        year=values["EFFECTIVE_DATE"][:4], css=CSS, favicon=FAVICON, body=body)


def index_body(values):
    app = values["APP_NAME"]
    return f"""
<p class="kicker">Less is More</p>
<h1 class="hero">{app}</h1>
<p class="lede">An anti-consumerism companion — before you buy something you’re
unsure about, ask LIM. 反消费主义的 AI 伙伴：买之前，先问问自己。</p>
<div class="cards">
  <a class="card" href="./privacy.html">
    <p class="kicker">Legal</p><h3>Privacy Policy · 隐私政策</h3>
    <p>What we collect, why, and your choices.</p></a>
  <a class="card" href="./support.html">
    <p class="kicker">Help</p><h3>Support · 支持中心</h3>
    <p>FAQ, subscriptions, account & contact.</p></a>
</div>
<p style="color:var(--ink3);font-size:15px">Questions? Email
<a href="mailto:{values['SUPPORT_EMAIL']}">{values['SUPPORT_EMAIL']}</a>.</p>
"""


def main():
    values = load_values()
    os.makedirs(DIST, exist_ok=True)

    pages = {
        "index.html": page(values["APP_NAME"], "LIM — Less is More", index_body(values), values),
        "privacy.html": page(
            f"Privacy Policy · {values['APP_NAME']}", "Privacy Policy / 隐私政策",
            render_md(os.path.join(LEGAL, "privacy-policy.md"), values), values),
        "support.html": page(
            f"Support · {values['APP_NAME']}", "Support / 支持中心",
            render_md(os.path.join(LEGAL, "support.md"), values), values),
    }
    # Prevent Jekyll from reprocessing the artifact on Pages.
    open(os.path.join(DIST, ".nojekyll"), "w").close()
    for name, html in pages.items():
        with open(os.path.join(DIST, name), "w", encoding="utf-8") as f:
            f.write(html)
        print("wrote", os.path.join("site", "dist", name))


if __name__ == "__main__":
    main()
