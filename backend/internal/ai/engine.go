// Package ai implements LIM's six-dimension purchase analysis.
//
// The default engine is a deterministic, fully-offline heuristic ported from the
// product prototype. It scores a candidate purchase on six axes (need, alt, emo,
// value, money, env), folds them into a 0–100 "impulse index" using the
// admin-configured per-dimension weights, and maps that to a verdict. When an
// Anthropic API key is configured the engine instead delegates to Claude (see
// llm.go) and falls back to the heuristic on any error.
package ai

import (
	"context"
	"math"
	"strings"

	"github.com/mosson9/lim/backend/internal/models"
)

// Input is a purchase to evaluate.
type Input struct {
	Item   string
	Price  int
	Cat    string
	Reason string
}

// Result is the outcome of an analysis.
type Result struct {
	Dims    models.Dimensions `json:"dims"`
	Impulse int               `json:"impulse"`
	Verdict models.Verdict    `json:"verdict"`
	Message string            `json:"message"`
	Note    string            `json:"note"`
}

// Engine produces analyses. It is safe for concurrent use.
type Engine struct {
	llm *llmClient // nil when no API key is configured
}

// New builds an Engine. If anthropicKey is non-empty, Claude is used with a
// graceful fallback to the heuristic.
func New(anthropicKey, model string) *Engine {
	e := &Engine{}
	if anthropicKey != "" {
		e.llm = newLLMClient(anthropicKey, model)
	}
	return e
}

// UsesLLM reports whether a language model backs the engine.
func (e *Engine) UsesLLM() bool { return e.llm != nil }

// Analyze evaluates a purchase against the user's profile and the live AI config.
func (e *Engine) Analyze(ctx context.Context, in Input, p models.Profile, cfg models.AIConfig) Result {
	if e.llm != nil {
		if r, err := e.llm.analyze(ctx, in, p, cfg); err == nil {
			return r
		}
		// fall through to heuristic on error
	}
	return Heuristic(in, p, cfg)
}

// clamp snaps a raw score into the 1–5 range, rounded to the nearest 0.5.
func clamp(v float64) float64 {
	r := math.Round(v*2) / 2
	if r < 1 {
		return 1
	}
	if r > 5 {
		return 5
	}
	return r
}

// Heuristic is the deterministic analysis used offline and as the LLM fallback.
//
// Faithful to the prototype: every axis starts neutral at 3, keyword signals and
// a price/budget ratio nudge the scores, then the weighted average is folded into
// an impulse index where a *lower* justification means a *higher* impulse.
func Heuristic(in Input, p models.Profile, cfg models.AIConfig) Result {
	txt := strings.ToLower(in.Item + " " + in.Reason)
	has := func(ws ...string) bool {
		for _, w := range ws {
			if strings.Contains(txt, w) {
				return true
			}
		}
		return false
	}

	need, alt, emo, value, money, env := 3.0, 3.0, 3.0, 3.0, 3.0, 3.0

	if has("必需", "刚需", "坏了", "没有", "工作", "通勤", "健康") {
		need += 1.5
	}
	if has("想要", "好看", "限量", "联名", "潮", "款", "色") {
		need -= 1
		emo -= 1.2
	}
	if has("第二", "第三", "又", "再买", "已经有", "类似") {
		alt -= 1.8
	}
	if has("租", "借", "二手", "修") {
		alt += 1
	}
	if has("促销", "打折", "直播", "种草", "便宜", "优惠", "凑单") {
		emo -= 1.5
	}
	if has("习惯", "长期", "耐用", "每天", "常用") {
		value += 1.5
	}
	if has("一时", "偶尔", "尝鲜", "跟风") {
		value -= 1.3
	}
	if has("环保", "可降解") {
		env += 1
	}
	if has("快时尚", "一次性", "塑料") {
		env -= 1
	}

	budget := p.MonthBudget
	if budget <= 0 {
		budget = 6000
	}
	ratio := float64(in.Price) / float64(budget)
	switch {
	case ratio > 0.5:
		money -= 2
	case ratio > 0.25:
		money -= 1
	case ratio < 0.05:
		money += 1
	}
	if in.Price > 1500 {
		value -= 0.3
	}

	dims := models.Dimensions{
		Need:  clamp(need),
		Alt:   clamp(alt),
		Emo:   clamp(emo),
		Value: clamp(value),
		Money: clamp(money),
		Env:   clamp(env),
	}

	// Weighted average across enabled dimensions (admin-tunable).
	weights := weightMap(cfg)
	sum, wsum := 0.0, 0.0
	for key, score := range map[string]float64{
		"need": dims.Need, "alt": dims.Alt, "emo": dims.Emo,
		"value": dims.Value, "money": dims.Money, "env": dims.Env,
	} {
		w, ok := weights[key]
		if !ok {
			w = 1
		}
		if w == 0 { // disabled
			continue
		}
		sum += score * w
		wsum += w
	}
	avg := 3.0
	if wsum > 0 {
		avg = sum / wsum
	}

	impulse := int(math.Round((5 - avg) / 4 * 100))
	if impulse < 4 {
		impulse = 4
	}
	if impulse > 96 {
		impulse = 96
	}

	verdict := models.VerdictResist
	switch {
	case impulse < 45:
		verdict = models.VerdictBuy
	case impulse < 62:
		verdict = models.VerdictPause
	}

	return Result{
		Dims:    dims,
		Impulse: impulse,
		Verdict: verdict,
		Message: Message(verdict, in.Item, cfg.Persona),
		Note:    note(verdict, dims),
	}
}

// weightMap returns key -> weight (0 if the dimension is disabled).
func weightMap(cfg models.AIConfig) map[string]float64 {
	m := make(map[string]float64, len(cfg.Dims))
	for _, d := range cfg.Dims {
		if !d.Enabled {
			m[d.Key] = 0
			continue
		}
		m[d.Key] = d.Weight
	}
	return m
}

// Message renders LIM's natural-language reply, honouring the configured persona.
func Message(v models.Verdict, item, persona string) string {
	switch persona {
	case "理性顾问":
		switch v {
		case models.VerdictBuy:
			return "综合六个维度评估，「" + item + "」在你的预算与需求结构内是合理支出，可以购买。"
		case models.VerdictPause:
			return "「" + item + "」并非必要，建议进入 24 小时冷静期后再复核一次决策。"
		default:
			return "数据显示「" + item + "」更多由情绪与外部信号驱动，性价比不足，建议暂不购买。"
		}
	case "犀利毒舌":
		switch v {
		case models.VerdictBuy:
			return "行吧，这个「" + item + "」是真有用，难得，买它。"
		case models.VerdictPause:
			return "「" + item + "」？先放一天。明天还想要再说，多半你就忘了。"
		default:
			return "醒醒，「" + item + "」就是一时上头。你东西已经够多了，把钱留住。"
		}
	default: // 温柔朋友
		switch v {
		case models.VerdictBuy:
			return "我陪你把「" + item + "」从头到尾想了一遍——它是真的能帮到你。如果预算也舒服，那就放心拥有它吧。"
		case models.VerdictPause:
			return "「" + item + "」没有那么必要，但也不是完全不行。要不要给它 24 小时？如果明天你还想要，它依然在。"
		default:
			return "我懂那种心动。但把六个角度摊开看，「" + item + "」更多是当下的情绪在说话。也许，你已经拥有得够多了。"
		}
	}
}

// note picks a short one-liner for history/list rows based on the weakest axis.
func note(v models.Verdict, d models.Dimensions) string {
	if v == models.VerdictBuy {
		return "高频实用，能切实改善生活体验。"
	}
	type ax struct {
		s float64
		t string
	}
	weakest := ax{99, ""}
	for _, a := range []ax{
		{d.Need, "生活中并非必需，缺它影响不大。"},
		{d.Alt, "你已经有类似的东西，或能租借共享。"},
		{d.Emo, "情绪与外部信号驱动较强。"},
		{d.Value, "使用频率低，长期价值有限。"},
		{d.Money, "价格偏高，会挤占其他更重要的开支。"},
		{d.Env, "环境成本较高，可再斟酌。"},
	} {
		if a.s < weakest.s {
			weakest = a
		}
	}
	return weakest.t
}
