# Legal & support pages

Markdown templates for the two public pages App Store Connect requires before you
can submit:

| File | App Store Connect field |
|------|-------------------------|
| [`privacy-policy.md`](privacy-policy.md) | App Privacy → **Privacy Policy URL** (required) |
| [`support.md`](support.md) | App Information → **Support URL** (required) |

Both are bilingual (English + 简体中文) and grounded in what the app actually
collects (email, optional financial profile, purchase-decision inputs,
subscription status; optional third-party LLM processing; **no** ads/tracking).

## Use them

1. Replace every `{{PLACEHOLDER}}` (company, contact email, effective date,
   jurisdiction, website, and `{{LLM_PROVIDER}}` only if your deployment uses a
   third-party model for analysis).
2. Have the privacy policy reviewed by counsel — these are starting points, not
   legal advice.
3. Host them at public URLs (e.g. GitHub Pages, your marketing site, or render
   the Markdown to HTML) and paste the URLs into App Store Connect.
4. Make sure the App Privacy questionnaire in App Store Connect matches the data
   types listed in the privacy policy.

> If you enable the Claude-backed analysis engine (`ANTHROPIC_API_KEY`), keep the
> `{{LLM_PROVIDER}}` disclosure — purchase details + profile context are sent to
> the model provider, which is a data-sharing (and possibly cross-border)
> disclosure. If you run the built-in offline heuristic only, you may remove that
> clause.
