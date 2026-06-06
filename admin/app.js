/* LIM Admin — a zero-build single-page app over the Go admin API.
   No frameworks: a tiny DOM helper + fetch. Mirrors the admin prototype. */
(() => {
  "use strict";

  const API = window.LIM_API_BASE;
  const TOKEN_KEY = "lim.admin.token";
  const root = document.getElementById("root");

  // ---------- tiny DOM helper ----------
  // h('div.card', {onclick}, child, child, ...) — tag supports .class and #id.
  function h(spec, props, ...kids) {
    const [tag, ...cls] = spec.split(".");
    const m = tag.split("#");
    const el = document.createElement(m[0] || "div");
    if (m[1]) el.id = m[1];
    if (cls.length) el.className = cls.join(" ");
    if (props) for (const [k, v] of Object.entries(props)) {
      if (v == null) continue;
      if (k === "class") el.className += " " + v;
      else if (k === "html") el.innerHTML = v;
      else if (k === "style") el.setAttribute("style", v);
      else if (k.startsWith("on")) el.addEventListener(k.slice(2), v);
      else el.setAttribute(k, v);
    }
    for (const kid of kids.flat()) {
      if (kid == null || kid === false) continue;
      el.appendChild(typeof kid === "object" ? kid : document.createTextNode(String(kid)));
    }
    return el;
  }
  const clear = (n) => { while (n.firstChild) n.removeChild(n.firstChild); };

  // ---------- formatting ----------
  const fmt = (n) => Number(n || 0).toLocaleString("zh-CN");
  const yuan = (n) => "¥" + fmt(n);
  function relTime(iso) {
    if (!iso) return "—";
    const d = (Date.now() - new Date(iso).getTime()) / 1000;
    if (d < 60) return "刚刚";
    if (d < 3600) return Math.floor(d / 60) + " 分钟前";
    if (d < 86400) return Math.floor(d / 3600) + " 小时前";
    return Math.floor(d / 86400) + " 天前";
  }
  function toast(msg) {
    const t = h("div.toast", null, msg);
    document.body.appendChild(t);
    setTimeout(() => t.remove(), 2200);
  }

  // ---------- API ----------
  let token = localStorage.getItem(TOKEN_KEY);
  async function api(method, path, body) {
    const res = await fetch(API + path, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: "Bearer " + token } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!res.ok) {
      let msg = "请求失败 (" + res.status + ")";
      try { msg = (await res.json()).error || msg; } catch (e) {}
      throw new Error(msg);
    }
    return res.status === 204 ? null : res.json();
  }

  // ---------- meta ----------
  function verdictMeta(v) {
    return v === "buy" ? { label: "建议买入", pill: "pill-indigo" }
      : v === "pause" ? { label: "冷静一下", pill: "pill-warn" }
        : { label: "建议不买", pill: "pill-saved" };
  }
  function planMeta(p) {
    return p === "year" ? { label: "年度会员", pill: "pill-indigo" }
      : p === "month" ? { label: "月度会员", pill: "pill-indigo" }
        : { label: "免费版", pill: "pill-spent" };
  }
  function statusMeta(s) {
    return ({ active: { l: "活跃", d: "#5E7E63" }, new: { l: "新用户", d: "#34357C" },
      dormant: { l: "沉睡", d: "#C0824F" }, churned: { l: "流失", d: "#B5524E" } })[s] || { l: s, d: "#9C9A92" };
  }
  const DIM_LABELS = { need: "需求", alt: "替代", emo: "情感", value: "长期价值", money: "经济", env: "环境" };
  const ICONS = {
    grid: '<rect x="4" y="4" width="7" height="7" rx="2"/><rect x="13" y="4" width="7" height="7" rx="2"/><rect x="4" y="13" width="7" height="7" rx="2"/><rect x="13" y="13" width="7" height="7" rx="2"/>',
    users: '<circle cx="9" cy="8.5" r="3.2"/><path d="M3.5 19c.5-3 2.8-4.8 5.5-4.8s5 1.8 5.5 4.8"/><path d="M15.5 5.5a3.2 3.2 0 0 1 0 6"/>',
    spark: '<path d="M12 3v3M12 18v3M3 12h3M18 12h3"/><path d="M12 7.5c.6 2.6 1.9 3.9 4.5 4.5-2.6.6-3.9 1.9-4.5 4.5-.6-2.6-1.9-3.9-4.5-4.5 2.6-.6 3.9-1.9 4.5-4.5Z"/>',
    card: '<rect x="3" y="5.5" width="18" height="13" rx="3"/><path d="M3 9.5h18"/>',
    sliders: '<path d="M5 6h14M5 12h14M5 18h14"/><circle cx="9" cy="6" r="2.2" fill="#fff"/><circle cx="15" cy="12" r="2.2" fill="#fff"/><circle cx="8" cy="18" r="2.2" fill="#fff"/>',
    tag: '<path d="M4 4h7l9 9-7 7-9-9V4Z"/><circle cx="8" cy="8" r="1.4"/>',
    image: '<rect x="3.5" y="4.5" width="17" height="15" rx="3"/><circle cx="9" cy="10" r="1.8"/><path d="m4 17 5-4 4 3 3-2 4 3"/>',
    bell: '<path d="M6.5 10a5.5 5.5 0 0 1 11 0c0 5 2 6.5 2 6.5h-15s2-1.5 2-6.5Z"/><path d="M10 19.5a2 2 0 0 0 4 0"/>',
    search: '<circle cx="11" cy="11" r="6.3"/><path d="m16 16 4 4"/>',
    up: '<path d="M12 19V5M6 11l6-6 6 6"/>', down: '<path d="M12 5v14M6 13l6 6 6-6"/>',
    plus: '<path d="M12 5v14M5 12h14"/>', close: '<path d="M6 6l12 12M18 6 6 18"/>',
    chevR: '<path d="M9 5l7 7-7 7"/>', download: '<path d="M12 4v11M7 11l5 5 5-5"/><path d="M5 20h14"/>',
    logout: '<path d="M14 7V5a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2v-2"/><path d="M10 12h11M18 9l3 3-3 3"/>',
    check: '<path d="M5 12.5 10 17l9-10"/>',
  };
  function icon(name, size = 18, stroke = "currentColor", sw = 1.7) {
    return h("span", { html: `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="${stroke}" stroke-width="${sw}" stroke-linecap="round" stroke-linejoin="round">${ICONS[name] || ""}</svg>`, style: "display:inline-flex" });
  }

  // ---------- charts (inline SVG) ----------
  function areaChart(series, labels, { color = "#34357C", fill = "rgba(52,53,124,.08)", height = 210 } = {}) {
    const W = 620, H = height, pl = 8, pr = 8, pt = 14, pb = 26;
    const max = Math.max(...series, 1) * 1.15, n = series.length;
    const x = (i) => pl + (i / (n - 1)) * (W - pl - pr);
    const y = (v) => pt + (1 - v / max) * (H - pt - pb);
    const line = series.map((v, i) => (i ? "L" : "M") + x(i) + " " + y(v)).join(" ");
    const area = line + ` L${x(n - 1)} ${H - pb} L${x(0)} ${H - pb} Z`;
    const grid = [0.25, 0.5, 0.75, 1].map(g => `<line x1="${pl}" x2="${W - pr}" y1="${pt + (1 - g) * (H - pt - pb)}" y2="${pt + (1 - g) * (H - pt - pb)}" stroke="#EFEDE6"/>`).join("");
    const dots = series.map((v, i) => `<circle cx="${x(i)}" cy="${y(v)}" r="3" fill="#fff" stroke="${color}" stroke-width="2"/>`).join("");
    const labs = (labels || []).map((l, i) => l ? `<text x="${x(i)}" y="${H - 7}" text-anchor="middle" font-size="11" fill="#9C9A92">${l}</text>` : "").join("");
    return h("div", { html: `<svg viewBox="0 0 ${W} ${H}" width="100%" height="${height}" preserveAspectRatio="none" style="overflow:visible">${grid}<path d="${area}" fill="${fill}"/><path d="${line}" fill="none" stroke="${color}" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"/>${dots}${labs}</svg>` });
  }
  function donut(segments, { size = 140, thick = 20, centerTop = "", centerSub = "" } = {}) {
    const total = segments.reduce((a, s) => a + s.value, 0) || 1;
    const r = size / 2 - thick / 2, c = 2 * Math.PI * r;
    let off = 0;
    const arcs = segments.map(s => {
      const len = (s.value / total) * c;
      const el = `<circle cx="${size / 2}" cy="${size / 2}" r="${r}" fill="none" stroke="${s.color}" stroke-width="${thick}" stroke-dasharray="${len} ${c - len}" stroke-dashoffset="${-off}"/>`;
      off += len; return el;
    }).join("");
    return h("div", { style: `position:relative;width:${size}px;height:${size}px;flex-shrink:0` },
      h("div", { html: `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" style="transform:rotate(-90deg)"><circle cx="${size / 2}" cy="${size / 2}" r="${r}" fill="none" stroke="#EFEDE6" stroke-width="${thick}"/>${arcs}</svg>` }),
      h("div", { style: "position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center" },
        h("div.display", { style: "font-size:24px" }, centerTop),
        h("div.faint", { style: "font-size:11px" }, centerSub)));
  }
  function radar(dims, size = 230) {
    const data = Object.keys(DIM_LABELS).map(k => ({ label: DIM_LABELS[k], value: dims[k] || 0 }));
    const cx = size / 2, cy = size / 2, R = size / 2 - 32, n = data.length, max = 5;
    const pt = (i, r) => { const a = -Math.PI / 2 + i * 2 * Math.PI / n; return [cx + Math.cos(a) * r, cy + Math.sin(a) * r]; };
    const rings = [1, 2, 3, 4, 5].map(g => `<polygon points="${data.map((_, i) => pt(i, R * g / 5).join(",")).join(" ")}" fill="none" stroke="#EFEDE6"/>`).join("");
    const spokes = data.map((_, i) => { const [x, y] = pt(i, R); return `<line x1="${cx}" y1="${cy}" x2="${x}" y2="${y}" stroke="#EFEDE6"/>`; }).join("");
    const poly = data.map((d, i) => pt(i, R * d.value / max).join(",")).join(" ");
    const labels = data.map((d, i) => { const [x, y] = pt(i, R + 16); return `<text x="${x}" y="${y + 4}" text-anchor="middle" font-size="11" fill="#6A6963">${d.label}</text>`; }).join("");
    return h("div", { html: `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">${rings}${spokes}<polygon points="${poly}" fill="rgba(52,53,124,.14)" stroke="#34357C" stroke-width="2"/>${labels}</svg>` });
  }

  // ============================================================
  //  LOGIN
  // ============================================================
  function renderLogin(err) {
    root.className = "login-wrap";
    clear(root);
    const email = h("input.input", { type: "email", value: "admin@lim.app" });
    const pw = h("input.input", { type: "password", value: "admin123" });
    const errBox = h("div.err", null, err || "");
    const submit = async () => {
      errBox.textContent = "";
      try {
        const r = await api("POST", "/api/v1/auth/login", { email: email.value.trim(), password: pw.value });
        if (r.user.role !== "admin") throw new Error("该账号不是管理员");
        token = r.token; localStorage.setItem(TOKEN_KEY, token);
        renderShell();
      } catch (e) { errBox.textContent = e.message; }
    };
    root.appendChild(h("div.login-card", null,
      h("div.brand", { style: "padding-bottom:8px" }, h("div.brand-mark", null, h("span", null, "L")),
        h("div", null, h("div.display", { style: "font-size:18px" }, "LIM"),
          h("div.faint", { style: "font-size:11px;margin-top:3px;letter-spacing:.5px" }, "Less is More · 管理台"))),
      h("div", { style: "font-size:20px;font-weight:600;margin:18px 0 4px" }, "登录后台"),
      h("div.faint", { style: "font-size:13px" }, "请使用管理员账号登录"),
      h("label", null, "邮箱"), email,
      h("label", null, "密码"), pw,
      errBox,
      h("button.btn.btn-indigo", { style: "width:100%;margin-top:22px", onclick: submit }, "登录"),
      h("div.faint", { style: "font-size:12px;text-align:center;margin-top:16px" }, "演示账号 admin@lim.app / admin123")));
    pw.addEventListener("keydown", (e) => { if (e.key === "Enter") submit(); });
  }

  // ============================================================
  //  SHELL
  // ============================================================
  const NAV = [
    { sec: "概览" },
    { key: "dashboard", icon: "grid", label: "数据总览" },
    { sec: "用户与决策" },
    { key: "users", icon: "users", label: "用户管理" },
    { key: "decisions", icon: "spark", label: "AI 决策记录", badge: "实时" },
    { key: "subscriptions", icon: "card", label: "订阅与营收" },
    { sec: "AI 与配置" },
    { key: "ai", icon: "sliders", label: "AI 维度 / Prompt" },
    { key: "categories", icon: "tag", label: "记账分类" },
    { key: "skins", icon: "image", label: "图标皮肤" },
    { sec: "运营" },
    { key: "push", icon: "bell", label: "推送 / 运营" },
  ];
  const PAGES = {
    dashboard: { title: "数据总览", render: pageDashboard },
    users: { title: "用户管理", render: pageUsers },
    decisions: { title: "AI 决策记录", render: pageDecisions },
    subscriptions: { title: "订阅与营收", render: pageSubscriptions },
    ai: { title: "AI 维度 / Prompt", render: pageAI },
    categories: { title: "记账分类", render: pageCategories },
    skins: { title: "图标皮肤", render: pageSkins },
    push: { title: "推送 / 运营", render: pagePush },
  };

  let contentEl;
  function renderShell() {
    root.className = "";
    clear(root);
    const side = h("aside.sidebar", null,
      h("div.brand", null, h("div.brand-mark", null, h("span", null, "L")),
        h("div", null, h("div.display", { style: "font-size:18px;line-height:1" }, "LIM"),
          h("div", { style: "font-size:11px;color:var(--ink-3);margin-top:3px;letter-spacing:.5px" }, "Less is More · 管理台"))),
      h("nav", { style: "flex:1;overflow-y:auto" }, NAV.map(n => n.sec
        ? h("div.nav-sec", null, n.sec)
        : h("button.nav-item", { "data-key": n.key, onclick: () => go(n.key) },
          icon(n.icon, 18), h("span", null, n.label),
          n.badge ? h("span.nav-badge", { style: "background:var(--saved)" }, n.badge) : null))),
      h("div", { style: "border-top:1px solid var(--hairline);padding-top:14px;margin-top:8px" },
        h("button.nav-item", { onclick: logout }, icon("logout", 18), h("span", null, "退出登录"))));

    contentEl = h("div.content");
    const top = h("header.topbar", null,
      h("div.search", { style: "width:240px" }, icon("search", 17, "#9C9A92"), h("input", { placeholder: "搜索用户 / 决策 ID…" })),
      h("div.grow"),
      h("div.row", { style: "gap:10px" }, h("div.avatar", { style: "width:34px;height:34px" }, "管"),
        h("div", { style: "line-height:1.3" }, h("div", { style: "font-size:13px;font-weight:600" }, "运营管理员"),
          h("div", { style: "font-size:11px;color:var(--ink-3)" }, "admin@lim.app"))));

    root.appendChild(h("div.admin", null, side, h("div.main", null, top, contentEl)));
    go(location.hash.replace(/^#\/?/, "") || "dashboard");
  }

  function go(key) {
    if (!PAGES[key]) key = "dashboard";
    location.hash = "#/" + key;
    document.querySelectorAll(".nav-item[data-key]").forEach(b =>
      b.classList.toggle("on", b.getAttribute("data-key") === key));
    clear(contentEl);
    const page = h("div.anim-page");
    contentEl.appendChild(page);
    PAGES[key].render(page).catch(e => page.appendChild(errorCard(e)));
  }
  window.addEventListener("hashchange", () => {
    const key = location.hash.replace(/^#\/?/, "");
    if (PAGES[key] && contentEl) go(key);
  });
  function logout() { localStorage.removeItem(TOKEN_KEY); token = null; renderLogin(); }
  function errorCard(e) {
    return h("div.card", { style: "color:var(--danger)" }, "加载失败：" + e.message);
  }
  function pageHead(kicker, title, sub, actions) {
    return h("div.page-head", null,
      h("div", null, h("div.kicker", null, kicker), h("h1.page-title", null, title), h("div.page-sub", null, sub)),
      actions ? h("div.row", { style: "gap:10px" }, actions) : null);
  }
  function kpiCard(k) {
    return h("div.card", null,
      h("div.kpi-label", null, k.label),
      h("div.kpi-val.tnum", null, k.val),
      h("div.row.between", { style: "margin-top:12px" },
        k.delta ? h("span.kpi-delta", { class: k.dir === "down" ? "down" : "up" }, icon(k.dir === "down" ? "down" : "up", 13), k.delta) : h("span"),
        h("span.faint", { style: "font-size:12px" }, k.sub || "")));
  }

  // ============================================================
  //  DASHBOARD
  // ============================================================
  async function pageDashboard(page) {
    const d = await api("GET", "/api/v1/admin/overview");
    page.appendChild(pageHead("概览 · Overview", "数据总览", "实时聚合自线上数据库"));

    const kpis = h("div.grid", { style: "grid-template-columns:repeat(4,1fr);margin-bottom:18px" }, d.kpis.map(kpiCard));
    page.appendChild(kpis);

    const totalV = d.verdict_split.reduce((a, s) => a + s.value, 0) || 1;
    const noBuy = Math.round((d.verdict_split[0].value / totalV) * 100);
    page.appendChild(h("div.grid", { style: "grid-template-columns:1.7fr 1fr;margin-bottom:18px" },
      h("div.card", null,
        h("div.row.between", { style: "margin-bottom:18px" },
          h("div", null, h("div", { style: "font-size:15px;font-weight:600" }, "日活跃用户趋势"),
            h("div.faint", { style: "font-size:12.5px;margin-top:3px" }, "近 30 天 DAU")),
          h("span.pill.pill-saved", null, icon("up", 12), "较上月 +8.6%")),
        areaChart(d.dau_series, d.dau_labels, { height: 210 })),
      h("div.card", null,
        h("div", { style: "font-size:15px;font-weight:600;margin-bottom:6px" }, "AI 建议分布"),
        h("div.faint", { style: "font-size:12.5px;margin-bottom:14px" }, "本月共 " + fmt(totalV) + " 次决策"),
        h("div.row", { style: "gap:22px" },
          donut(d.verdict_split, { centerTop: noBuy + "%", centerSub: "选择不买" }),
          h("div.grow", null, legend(d.verdict_split))))));

    page.appendChild(h("div.grid", { style: "grid-template-columns:1fr 1fr;margin-bottom:18px" },
      h("div.card", null, h("div", { style: "font-size:15px;font-weight:600;margin-bottom:14px" }, "咨询转化漏斗"),
        h("div", { style: "display:flex;flex-direction:column;gap:10px" }, d.funnel.map((f, i) => h("div", null,
          h("div.row.between", { style: "font-size:12.5px;margin-bottom:5px" }, h("span.muted", null, f.step),
            h("span.tnum", null, fmt(f.value) + " · " + Math.round(f.pct * 100) + "%")),
          h("div.bar-track", null, h("div.bar-fill", { style: `width:${f.pct * 100}%;background:var(--indigo);opacity:${1 - i * 0.13}` })))))),
      h("div.card", null, h("div", { style: "font-size:15px;font-weight:600;margin-bottom:14px" }, "咨询品类分布"),
        h("div", { style: "display:flex;flex-direction:column;gap:12px" }, (d.cat_dist || []).slice(0, 6).map(c => {
          const max = d.cat_dist[0].value || 1;
          return h("div.row", { style: "gap:12px" }, h("span.muted", { style: "width:72px;font-size:12.5px" }, c.label),
            h("div.grow.bar-track", { style: "height:10px" }, h("div.bar-fill", { style: `width:${c.value / max * 100}%;background:${c.color}` })),
            h("span.tnum.faint", { style: "font-size:12px;width:42px;text-align:right" }, fmt(c.value)));
        })))));

    page.appendChild(h("div.card-pad0", null,
      h("div.row.between", { style: "padding:18px 22px" }, h("div", { style: "font-size:15px;font-weight:600" }, "实时决策流"),
        h("button.btn.btn-ghost.btn-sm", { onclick: () => go("decisions") }, "查看全部", icon("chevR", 14))),
      decisionTable(d.recent, false)));
  }
  function legend(items) {
    return h("div", { style: "display:flex;flex-direction:column;gap:11px" }, items.map(s =>
      h("div.row.between", { style: "font-size:13px" },
        h("span.row", { style: "gap:8px" }, h("span.pill-dot", { style: "background:" + s.color }), s.label),
        h("span.tnum.muted", { style: "font-weight:600" }, fmt(s.value)))));
  }

  // ============================================================
  //  shared: decisions table
  // ============================================================
  function impulseBar(value) {
    const c = value >= 62 ? "var(--saved)" : value >= 45 ? "var(--warn)" : "var(--indigo)";
    return h("span.row", { style: "gap:9px" },
      h("span", { style: "width:54px;height:6px;background:var(--paper-2);border-radius:999px;display:inline-block" },
        h("span", { style: `display:block;width:${value}%;height:100%;background:${c};border-radius:999px` })),
      h("span.tnum", { style: `font-size:12.5px;font-weight:600;color:${c}` }, value));
  }
  function decisionTable(list, withId) {
    const head = withId ? ["决策 ID", "用户", "想买的东西", "价格", "冲动指数", "AI 建议", "时间", ""]
      : ["用户", "想买的东西", "价格", "冲动指数", "AI 建议", "时间"];
    return h("table.tbl", null,
      h("thead", null, h("tr", null, head.map(th => h("th", null, th)))),
      h("tbody", null, list.map(d => {
        const vm = verdictMeta(d.verdict);
        const cells = [];
        if (withId) cells.push(h("td.tnum.faint", { style: "font-size:12.5px" }, d.id));
        cells.push(h("td", null, h("span.row", { style: "gap:9px" }, h("span.avatar", { style: "width:28px;height:28px;font-size:13px" }, (d.user || "?")[0]), d.user)));
        cells.push(h("td", null, d.item));
        cells.push(h("td.tnum", null, yuan(d.price)));
        cells.push(h("td", null, impulseBar(d.impulse)));
        cells.push(h("td", null, h("span.pill", { class: vm.pill }, vm.label)));
        cells.push(h("td.faint", null, relTime(d.created_at)));
        if (withId) cells.push(h("td", null, icon("chevR", 15, "#C2C0B6")));
        return h("tr", { onclick: () => decisionDrawer(d) }, cells);
      })));
  }

  // ============================================================
  //  DECISIONS
  // ============================================================
  let decFilter = "all";
  async function pageDecisions(page) {
    const d = await api("GET", "/api/v1/admin/decisions?filter=" + decFilter);
    const seg = h("div.seg", null, [["all", "全部"], ["resist", "建议不买"], ["pause", "冷静一下"], ["buy", "建议买入"]].map(([k, l]) =>
      h("button", { class: decFilter === k ? "on" : "", onclick: () => { decFilter = k; go("decisions"); } }, l)));
    page.appendChild(pageHead("AI 决策 · Decisions", "AI 决策记录", "每一次「买 / 不买」背后的六维分析", seg));
    page.appendChild(h("div.grid", { style: "grid-template-columns:repeat(4,1fr);margin-bottom:18px" },
      kpiCard({ label: "决策总数", val: fmt(d.count), sub: "当前筛选" }),
      kpiCard({ label: "平均冲动指数", val: d.avg_impulse, sub: "越高越该克制" }),
      kpiCard({ label: "今日帮省下", val: yuan(d.saved_today), sub: "克制金额" }),
      kpiCard({ label: "不买率", val: pctOf(d.decisions, x => x.verdict === "resist"), sub: "建议不买占比" })));
    page.appendChild(h("div.card-pad0", null, decisionTable(d.decisions, true)));
  }
  function pctOf(list, pred) {
    if (!list.length) return "0%";
    return Math.round(list.filter(pred).length / list.length * 100) + "%";
  }
  function decisionDrawer(d) {
    const vm = verdictMeta(d.verdict);
    const close = () => { mask.remove(); drawer.remove(); };
    const mask = h("div.drawer-mask", { onclick: close });
    const rows = Object.keys(DIM_LABELS).map(k => h("div.row", { style: "gap:12px" },
      h("span.muted", { style: "width:64px;font-size:12.5px" }, DIM_LABELS[k]),
      h("div.grow.bar-track", null, h("div.bar-fill", { style: `width:${(d.dims[k] || 0) / 5 * 100}%;background:var(--indigo)` })),
      h("span.tnum.faint", { style: "width:28px;text-align:right;font-size:12.5px" }, (d.dims[k] || 0).toFixed(1))));
    const drawer = h("div.drawer", null,
      h("div", { style: "padding:22px 24px;border-bottom:1px solid var(--hairline)" },
        h("div.row.between", { style: "margin-bottom:16px" }, h("span.kicker", null, "决策详情 · " + d.id),
          h("button.btn.btn-ghost.btn-sm", { style: "padding:7px", onclick: close }, icon("close", 16))),
        h("div", { style: "font-size:19px;font-weight:600" }, d.item),
        h("div.row", { style: "gap:10px;margin-top:8px" }, h("span.tnum.muted", null, yuan(d.price)), h("span.faint", null, "·"),
          h("span.muted", null, d.user || ""), h("span.faint", null, "·"), h("span.faint", null, relTime(d.created_at)))),
      h("div", { style: "padding:22px 24px" },
        h("div.card", { style: "display:flex;align-items:center;gap:20px;margin-bottom:18px" },
          gauge(d.impulse),
          h("div", null, h("div.faint", { style: "font-size:12px" }, "冲动指数"),
            h("div.display.tnum", { style: "font-size:38px;line-height:1" }, d.impulse),
            h("span.pill", { class: vm.pill, style: "margin-top:8px" }, vm.label))),
        h("div.card", { style: "display:flex;flex-direction:column;align-items:center;margin-bottom:18px" },
          h("div", { style: "width:100%;font-size:14px;font-weight:600;margin-bottom:6px" }, "六维分析"), radar(d.dims)),
        d.message ? h("div.card", { style: "margin-bottom:18px;background:var(--indigo-tint)" },
          h("div.kicker", { style: "margin-bottom:8px" }, "LIM 回复"),
          h("div", { style: "font-size:14px;line-height:1.7;color:var(--indigo-ink)" }, d.message)) : null,
        h("div", { style: "font-size:14px;font-weight:600;margin-bottom:10px" }, "各维度评分"),
        h("div", { style: "display:flex;flex-direction:column;gap:9px" }, rows)));
    document.body.appendChild(mask); document.body.appendChild(drawer);
  }
  function gauge(value, size = 92) {
    const r = size / 2 - 9, c = 2 * Math.PI * r, len = value / 100 * c;
    const col = value >= 62 ? "#5E7E63" : value >= 45 ? "#C0824F" : "#34357C";
    return h("div", { html: `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" style="transform:rotate(-90deg)"><circle cx="${size / 2}" cy="${size / 2}" r="${r}" fill="none" stroke="#EFEDE6" stroke-width="9"/><circle cx="${size / 2}" cy="${size / 2}" r="${r}" fill="none" stroke="${col}" stroke-width="9" stroke-linecap="round" stroke-dasharray="${len} ${c - len}"/></svg>` });
  }

  // ============================================================
  //  USERS
  // ============================================================
  let userFilter = "all", userQuery = "";
  async function pageUsers(page) {
    const list = await api("GET", "/api/v1/admin/users?filter=" + userFilter + "&q=" + encodeURIComponent(userQuery));
    page.appendChild(pageHead("用户 · Users", "用户管理", "共 " + list.length + " 名用户（当前筛选）",
      h("button.btn.btn-ghost.btn-sm", null, icon("download", 15), "导出 CSV")));
    const searchInput = h("input", { placeholder: "搜索姓名 / ID / 城市", value: userQuery });
    searchInput.addEventListener("keydown", e => { if (e.key === "Enter") { userQuery = searchInput.value; go("users"); } });
    const seg = h("div.seg", null, [["all", "全部"], ["plus", "付费"], ["free", "免费"], ["risk", "流失风险"]].map(([k, l]) =>
      h("button", { class: userFilter === k ? "on" : "", onclick: () => { userFilter = k; go("users"); } }, l)));
    page.appendChild(h("div.card-pad0", null,
      h("div.row.between", { style: "padding:16px 18px;border-bottom:1px solid var(--hairline);gap:14px" },
        h("div.search", { style: "width:300px" }, icon("search", 16, "#9C9A92"), searchInput), seg),
      h("table.tbl", null,
        h("thead", null, h("tr", null, ["用户", "城市", "订阅", "咨询次数", "累计省下", "克制次数", "状态", "最近活跃"].map(t => h("th", null, t)))),
        h("tbody", null, list.map(u => {
          const pm = planMeta(u.plan), sm = statusMeta(u.status);
          return h("tr", { onclick: () => userDrawer(u.id) },
            h("td", null, h("span.row", { style: "gap:11px" }, h("span.avatar", null, (u.name || "?")[0]),
              h("span", null, h("div", { style: "font-weight:600" }, u.name), h("div.faint.tnum", { style: "font-size:11.5px" }, u.id)))),
            h("td.muted", null, u.city || "—"),
            h("td", null, h("span.pill", { class: pm.pill }, pm.label)),
            h("td.tnum", null, u.consults),
            h("td.tnum", { style: "color:var(--saved-deep);font-weight:600" }, yuan(u.saved)),
            h("td.tnum", null, u.resist),
            h("td", null, h("span.row", { style: "gap:7px" }, h("span.pill-dot", { style: "background:" + sm.d }), sm.l)),
            h("td.faint", null, u.last_seen));
        })))));
  }
  async function userDrawer(id) {
    const close = () => { mask.remove(); drawer.remove(); };
    const mask = h("div.drawer-mask", { onclick: close });
    const drawer = h("div.drawer", null, h("div.card", { style: "margin:24px;border:none" }, "加载中…"));
    document.body.appendChild(mask); document.body.appendChild(drawer);
    let data;
    try { data = await api("GET", "/api/v1/admin/users/" + id); }
    catch (e) { clear(drawer); drawer.appendChild(errorCard(e)); return; }
    const u = data.user, st = data.stats, pm = planMeta(u.plan), sm = statusMeta(userStatusGuess(u));
    clear(drawer);
    const stat = (l, v, c) => h("div.card", { style: "padding:14px 16px" }, h("div.faint", { style: "font-size:12px" }, l),
      h("div.tnum", { style: "font-size:19px;font-weight:600;margin-top:5px;color:" + (c || "var(--ink)") }, v));
    drawer.appendChild(h("div", { style: "padding:22px 24px;border-bottom:1px solid var(--hairline)" },
      h("div.row.between", { style: "margin-bottom:18px" }, h("span.kicker", null, "用户详情"),
        h("button.btn.btn-ghost.btn-sm", { style: "padding:7px", onclick: close }, icon("close", 16))),
      h("div.row", { style: "gap:14px" }, h("span.avatar", { style: "width:54px;height:54px;font-size:24px;border-radius:15px" }, (u.name || "?")[0]),
        h("div.grow", null, h("div", { style: "font-size:20px;font-weight:600" }, u.name),
          h("div.faint.tnum", { style: "font-size:12.5px;margin-top:2px" }, u.id + " · " + (u.city || "—"))),
        h("span.pill", { class: pm.pill }, pm.label))));
    drawer.appendChild(h("div", { style: "padding:20px 24px" },
      h("div.grid", { style: "grid-template-columns:1fr 1fr;gap:12px;margin-bottom:20px" },
        stat("累计省下", yuan(st.total_saved), "var(--saved-deep)"), stat("咨询次数", st.resist_count + st.buy_count),
        stat("克制次数", st.resist_count), stat("克制率", st.restraint_rate + "%")),
      h("div.row.between", { style: "margin-bottom:12px" }, h("span", { style: "font-size:14px;font-weight:600" }, "状态"),
        h("span.row", { style: "gap:7px;font-size:13px" }, h("span.pill-dot", { style: "background:" + sm.d }), sm.l)),
      h("hr.hr", { style: "margin:4px 0 18px" }),
      h("div", { style: "font-size:14px;font-weight:600;margin-bottom:12px" }, "最近决策"),
      h("div", { style: "display:flex;flex-direction:column;gap:10px" }, (data.decisions || []).slice(0, 5).map(dec => {
        const vm = verdictMeta(dec.verdict);
        return h("div.card", { style: "padding:13px 16px;cursor:pointer", onclick: () => decisionDrawer({ ...dec, user: u.name }) },
          h("div.row.between", null, h("span", { style: "font-weight:500;font-size:13.5px" }, dec.item), h("span.pill", { class: vm.pill }, vm.label)),
          h("div.row.between", { style: "margin-top:8px" }, h("span.tnum.faint", { style: "font-size:12.5px" }, yuan(dec.price) + " · 冲动 " + dec.impulse),
            h("span.faint", { style: "font-size:12px" }, relTime(dec.created_at))));
      }))));
  }
  function userStatusGuess(u) {
    const idle = (Date.now() - new Date(u.last_active_at).getTime()) / 86400000;
    if ((Date.now() - new Date(u.created_at).getTime()) / 86400000 < 7) return "new";
    if (idle > 21) return "churned"; if (idle > 7) return "dormant"; return "active";
  }

  // ============================================================
  //  SUBSCRIPTIONS
  // ============================================================
  async function pageSubscriptions(page) {
    const d = await api("GET", "/api/v1/admin/subscriptions");
    page.appendChild(pageHead("营收 · Revenue", "订阅与营收", "LIM Plus 会员 · 月度 ¥18 / 年度 ¥98",
      h("button.btn.btn-ghost.btn-sm", null, icon("download", 15), "财务报表")));
    page.appendChild(h("div.grid", { style: "grid-template-columns:repeat(4,1fr);margin-bottom:18px" },
      d.kpis.map(k => kpiCard({ label: k.label, val: k.val, sub: k.sub }))));
    const totalPlan = d.plan_split.reduce((a, s) => a + s.value, 0);
    page.appendChild(h("div.grid", { style: "grid-template-columns:1.7fr 1fr;margin-bottom:18px" },
      h("div.card", null, h("div.row.between", { style: "margin-bottom:18px" },
        h("div", null, h("div", { style: "font-size:15px;font-weight:600" }, "月度经常性营收 MRR"),
          h("div.faint", { style: "font-size:12.5px;margin-top:3px" }, "近 12 个月")),
        h("span.pill.pill-saved", null, icon("up", 12), "稳步增长")),
        areaChart(d.mrr_series, d.mrr_labels, { height: 210 })),
      h("div.card", null, h("div", { style: "font-size:15px;font-weight:600;margin-bottom:14px" }, "套餐分布"),
        h("div.row", { style: "gap:22px" }, donut(d.plan_split, { centerTop: fmt(totalPlan), centerSub: "付费用户" }),
          h("div.grow", null, legend(d.plan_split))))));
    page.appendChild(h("div.card-pad0", null,
      h("div.row.between", { style: "padding:18px 22px" }, h("div", { style: "font-size:15px;font-weight:600" }, "最近交易")),
      h("table.tbl", null, h("thead", null, h("tr", null, ["交易号", "用户", "套餐", "金额", "状态", "时间"].map(t => h("th", null, t)))),
        h("tbody", null, d.transactions.map(t => h("tr", null,
          h("td.tnum.faint", { style: "font-size:12.5px" }, t.id),
          h("td", null, h("span.row", { style: "gap:9px" }, h("span.avatar", { style: "width:28px;height:28px;font-size:13px" }, (t.user_name || "?")[0]), t.user_name)),
          h("td.muted", null, t.plan), h("td.tnum", { style: "font-weight:600" }, yuan(t.amount)),
          h("td", null, h("span.pill", { class: t.status === "success" ? "pill-saved" : "pill-danger" }, t.status === "success" ? "成功" : "已退款")),
          h("td.faint", null, relTime(t.created_at))))))));
  }

  // ============================================================
  //  AI CONFIG
  // ============================================================
  async function pageAI(page) {
    const cfg = await api("GET", "/api/v1/admin/ai-config");
    const state = JSON.parse(JSON.stringify(cfg));
    const save = async () => {
      try { await api("PUT", "/api/v1/admin/ai-config", state); toast("已发布更改，将影响线上分析"); }
      catch (e) { toast("保存失败：" + e.message); }
    };
    page.appendChild(pageHead("模型 · AI Engine", "AI 维度与 Prompt 配置", "调整六维权重与系统提示词 · 改动将影响线上分析",
      h("button.btn.btn-indigo.btn-sm", { onclick: save }, icon("check", 15, "#fff"), "发布更改")));

    const dimRows = state.dims.map((dm, i) => {
      const weightTag = h("span.pill.pill-indigo.tnum", null, "×" + dm.weight.toFixed(1));
      const slider = h("input", { type: "range", min: "0.5", max: "2", step: "0.1", value: dm.weight, style: "width:100%;accent-color:#34357C" });
      slider.addEventListener("input", () => { dm.weight = parseFloat(slider.value); weightTag.textContent = "×" + dm.weight.toFixed(1); });
      const tog = h("span.toggle", { class: dm.enabled ? "on" : "" }, h("b"));
      tog.addEventListener("click", () => { dm.enabled = !dm.enabled; tog.classList.toggle("on", dm.enabled); wrap.style.opacity = dm.enabled ? 1 : 0.45; slider.disabled = !dm.enabled; });
      const wrap = h("div", { style: "opacity:" + (dm.enabled ? 1 : 0.45) },
        h("div.row.between", { style: "margin-bottom:8px" },
          h("span.row", { style: "gap:10px" }, h("span", { style: "font-weight:600;font-size:14px" }, dm.label),
            h("span.faint", { style: "font-size:12px" }, dm.desc)),
          h("span.row", { style: "gap:12px" }, weightTag, tog)),
        slider);
      return wrap;
    });

    const promptArea = h("textarea.input", { style: "min-height:230px;line-height:1.7;font-size:13.5px;resize:vertical" }, state.prompt_template);
    promptArea.addEventListener("input", () => { state.prompt_template = promptArea.value; });

    const personaBtns = ["温柔朋友", "理性顾问", "犀利毒舌"].map(p =>
      h("button.btn.btn-sm", { class: state.persona === p ? "btn-indigo" : "btn-ghost", style: "flex:1",
        onclick: (e) => { state.persona = p; page.querySelectorAll("[data-persona]").forEach(b => b.className = "btn btn-sm " + (b.getAttribute("data-persona") === p ? "btn-indigo" : "btn-ghost")); },
        "data-persona": p }, p));

    page.appendChild(h("div.grid", { style: "grid-template-columns:1fr 1fr;align-items:start" },
      h("div.card", null, h("div", { style: "font-size:15px;font-weight:600;margin-bottom:4px" }, "六维权重"),
        h("div.faint", { style: "font-size:12.5px;margin-bottom:18px" }, "权重越高，该维度对冲动指数影响越大"),
        h("div", { style: "display:flex;flex-direction:column;gap:18px" }, dimRows)),
      h("div", { style: "display:flex;flex-direction:column;gap:18px" },
        h("div.card", null, h("div.row.between", { style: "margin-bottom:12px" },
          h("span", { style: "font-size:15px;font-weight:600" }, "系统 Prompt 模板"),
          h("span.pill.pill-spent", null, (cfg.model || "heuristic") + " · " + state.persona)), promptArea,
          h("div.faint", { style: "font-size:12px;margin-top:10px" }, "变量：", h("code", null, "{{item}} {{price}} {{reason}} {{budget}} {{members}}"))),
        h("div.card", null, h("div", { style: "font-size:14px;font-weight:600;margin-bottom:14px" }, "语气人设"),
          h("div.row", { style: "gap:10px" }, personaBtns),
          h("div.row.between", { style: "margin-top:18px;font-size:13px" }, h("span.muted", null, "免费版每日咨询上限"),
            numStepper(state.free_daily_limit, v => state.free_daily_limit = v)),
          h("div.row.between", { style: "margin-top:12px;font-size:13px" }, h("span.muted", null, "冷静期默认时长（小时）"),
            numStepper(state.cooling_hours, v => state.cooling_hours = v))))));
  }
  function numStepper(value, onChange) {
    const span = h("span.tnum", { style: "font-weight:600;min-width:28px;text-align:center;display:inline-block" }, value);
    const mk = (delta, label) => h("button.btn.btn-ghost.btn-sm", { style: "padding:4px 10px", onclick: () => { value = Math.max(0, value + delta); span.textContent = value; onChange(value); } }, label);
    return h("span.row", { style: "gap:8px" }, mk(-1, "−"), span, mk(1, "+"));
  }

  // ============================================================
  //  CATEGORIES
  // ============================================================
  async function pageCategories(page) {
    const cats = await api("GET", "/api/v1/admin/categories");
    const state = JSON.parse(JSON.stringify(cats));
    const save = async () => {
      try { await api("PUT", "/api/v1/admin/categories", state); toast("分类已保存"); }
      catch (e) { toast("保存失败：" + e.message); }
    };
    page.appendChild(pageHead("内容 · Content", "记账分类与图标", "管理消费分类 · 免费与 Plus 解锁",
      h("button.btn.btn-indigo.btn-sm", { onclick: save }, icon("check", 15, "#fff"), "保存更改")));
    page.appendChild(h("div.grid", { style: "grid-template-columns:repeat(4,1fr)" }, state.map(c => {
      const tog = h("span.toggle", { class: c.free ? "" : "on", title: "Plus 专属" }, h("b"));
      tog.addEventListener("click", () => { c.free = !c.free; tog.classList.toggle("on", !c.free); });
      return h("div.card", { style: "text-align:center;position:relative" },
        h("div", { style: `width:56px;height:56px;border-radius:16px;background:${c.color}18;display:flex;align-items:center;justify-content:center;margin:4px auto 14px` },
          icon(c.icon === "spark" ? "spark" : c.icon === "tag" ? "tag" : "grid", 26, c.color)),
        h("div", { style: "font-weight:600;font-size:14.5px" }, c.label),
        h("div.faint", { style: "font-size:12px;margin-top:4px" }, "本月 " + fmt(c.count) + " 笔"),
        h("div.row", { style: "gap:8px;margin-top:14px;justify-content:center;align-items:center" },
          h("span.faint", { style: "font-size:12px" }, "Plus 专属"), tog));
    })));
  }

  // ============================================================
  //  SKINS
  // ============================================================
  async function pageSkins(page) {
    const skins = await api("GET", "/api/v1/admin/skins");
    page.appendChild(pageHead("品牌 · Brand", "App 图标皮肤", "桌面图标主题 · 经典与暖纸免费 · 其余 Plus 专属"));
    page.appendChild(h("div.grid", { style: "grid-template-columns:repeat(4,1fr)" }, skins.map(s =>
      h("div.card", { style: "text-align:center;position:relative" },
        s.free ? null : h("span.pill.pill-indigo", { style: "position:absolute;top:14px;right:14px" }, "Plus"),
        h("div", { style: `width:76px;height:76px;border-radius:20px;background:${s.bg};display:flex;align-items:center;justify-content:center;margin:8px auto 16px;box-shadow:var(--shadow-md)` },
          h("span.display", { style: "font-size:38px;color:" + (s.dark ? "#34357C" : "#fff") }, "L")),
        h("div", { style: "font-weight:600;font-size:14.5px" }, s.name),
        h("div.faint", { style: "font-size:12px;margin-top:4px" }, "使用率 " + (s.used || "—"))))));
  }

  // ============================================================
  //  PUSH
  // ============================================================
  async function pagePush(page) {
    const pushes = await api("GET", "/api/v1/admin/push");
    const sm = { scheduled: { l: "待发送", p: "pill-warn" }, sent: { l: "已发送", p: "pill-saved" }, draft: { l: "草稿", p: "pill-spent" } };
    const create = () => pushForm(() => go("push"));
    page.appendChild(pageHead("运营 · Engagement", "推送与运营", "分群触达 · 克制成就感召回",
      h("button.btn.btn-indigo.btn-sm", { onclick: create }, icon("plus", 15, "#fff"), "新建推送")));
    page.appendChild(h("div.card-pad0", null, h("table.tbl", null,
      h("thead", null, h("tr", null, ["推送文案", "目标人群", "触达", "打开率", "状态", "时间"].map(t => h("th", null, t)))),
      h("tbody", null, pushes.map(p => h("tr", null,
        h("td", null, h("span.row", { style: "gap:11px" },
          h("span", { style: "width:34px;height:34px;border-radius:10px;background:var(--indigo-soft);display:flex;align-items:center;justify-content:center" }, icon("bell", 17, "#34357C")),
          h("span", { style: "font-weight:500" }, p.title))),
        h("td", null, h("span.pill.pill-spent", null, p.segment)),
        h("td.tnum", null, fmt(p.sent)),
        h("td.tnum", { style: "font-weight:600;color:" + (p.status === "sent" ? "var(--saved-deep)" : "var(--ink-3)") }, p.status === "draft" ? "—" : p.open_rate),
        h("td", null, h("span.pill", { class: (sm[p.status] || sm.draft).p }, (sm[p.status] || sm.draft).l)),
        h("td.faint", null, p.schedule || relTime(p.created_at)))))))); }
  function pushForm(done) {
    const close = () => { mask.remove(); drawer.remove(); };
    const mask = h("div.drawer-mask", { onclick: close });
    const title = h("input.input", { placeholder: "例如：本周你已省下 ¥860" });
    const seg = h("input.input", { placeholder: "目标人群，例如：连续克制 ≥3 天", value: "全量用户" });
    const status = h("select.input", null, h("option", { value: "draft" }, "草稿"), h("option", { value: "scheduled" }, "待发送"), h("option", { value: "sent" }, "已发送"));
    const submit = async () => {
      if (!title.value.trim()) { toast("请填写推送文案"); return; }
      try {
        await api("POST", "/api/v1/admin/push", { title: title.value.trim(), segment: seg.value.trim(), status: status.value, schedule: "刚刚创建" });
        toast("推送已创建"); close(); done();
      } catch (e) { toast("创建失败：" + e.message); }
    };
    const drawer = h("div.drawer", null, h("div", { style: "padding:24px" },
      h("div.row.between", { style: "margin-bottom:20px" }, h("span.kicker", null, "新建推送"),
        h("button.btn.btn-ghost.btn-sm", { style: "padding:7px", onclick: close }, icon("close", 16))),
      h("label.label", null, "推送文案"), title,
      h("label.label", { style: "margin-top:14px" }, "目标人群"), seg,
      h("label.label", { style: "margin-top:14px" }, "状态"), status,
      h("button.btn.btn-indigo", { style: "width:100%;margin-top:22px", onclick: submit }, "创建推送")));
    document.body.appendChild(mask); document.body.appendChild(drawer);
  }

  // ---------- boot ----------
  if (token) renderShell(); else renderLogin();
})();
