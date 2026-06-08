# site/ — legal & support pages (GitHub Pages)

Renders the bilingual Markdown in [`../docs/legal`](../docs/legal) into a small,
styled, hostable static site (the Privacy Policy and Support pages App Store
Connect requires), and deploys it to GitHub Pages.

```
site/
├── values.json   # your real values (company, emails, URLs, dates…)
├── build.py      # Markdown -> styled HTML  (needs: pip install markdown)
└── dist/         # generated output: index.html, privacy.html, support.html
```

## Build locally

```bash
pip install markdown
python3 site/build.py          # writes site/dist/*.html
open site/dist/index.html      # preview
```

Edit your real details in [`values.json`](values.json) (company name, contact
emails, effective date, jurisdiction, website, and `LLM_PROVIDER`). The build
substitutes every `{{PLACEHOLDER}}`, drops the template comments, and wraps the
content in an on-brand template (LIM design tokens). `dist/` is committed so the
HTML is ready to host on **any** static host as well.

## Deploy to GitHub Pages (automatic)

1. In the repo: **Settings → Pages → Build and deployment → Source = GitHub
   Actions** (one-time).
2. Merge to `main` (or run the **Deploy legal/support site to GitHub Pages**
   workflow manually under Actions). The workflow
   ([`.github/workflows/pages.yml`](../.github/workflows/pages.yml)) rebuilds and
   publishes `site/dist/`.
3. Your URLs become:
   - Privacy: `https://<user>.github.io/<repo>/privacy.html`
   - Support: `https://<user>.github.io/<repo>/support.html`

   Paste those into App Store Connect (Privacy Policy URL / Support URL).

> **Custom domain:** add a `CNAME` file to `site/dist/` (or set it in Settings →
> Pages) — e.g. `lim.app` — and the build's `.nojekyll` marker keeps Pages from
> reprocessing the output.
