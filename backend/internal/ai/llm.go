package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mosson9/lim/backend/internal/models"
)

// llmClient talks to the Anthropic Messages API to produce a structured analysis.
type llmClient struct {
	key   string
	model string
	http  *http.Client
}

func newLLMClient(key, model string) *llmClient {
	return &llmClient{key: key, model: model, http: &http.Client{Timeout: 30 * time.Second}}
}

// analyze asks Claude to score the purchase and return strict JSON, which we map
// onto our Result. Any deviation returns an error so the caller can fall back.
func (c *llmClient) analyze(ctx context.Context, in Input, p models.Profile, cfg models.AIConfig) (Result, error) {
	system := buildSystemPrompt(cfg)
	user := fmt.Sprintf(
		"用户正在犹豫是否购买「%s」，价格 %d 元，类别 %s，动机：%s。\n"+
			"用户画像：个人月入 %d、家庭月入 %d、家庭 %d 人、月供 %d、可支配预算 %d。\n\n"+
			"只输出一个 JSON 对象，不要任何解释或 markdown，结构如下：\n"+
			`{"dims":{"need":1-5,"alt":1-5,"emo":1-5,"value":1-5,"money":1-5,"env":1-5},`+
			`"impulse":0-100,"verdict":"buy|pause|resist","message":"给用户的一段温暖的话","note":"一句话总结"}`,
		in.Item, in.Price, in.Cat, fallback(in.Reason, "未说明"),
		p.MonthlyIncome, p.HouseholdIncome, p.Members, p.Debt, p.MonthBudget,
	)

	reqBody := map[string]any{
		"model":      c.model,
		"max_tokens": 700,
		"system":     system,
		"messages": []map[string]any{
			{"role": "user", "content": user},
		},
	}
	buf, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(buf))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.key)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return Result{}, fmt.Errorf("anthropic status %d", resp.StatusCode)
	}

	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Result{}, err
	}
	if len(out.Content) == 0 {
		return Result{}, fmt.Errorf("empty completion")
	}

	raw := extractJSON(out.Content[0].Text)
	var parsed struct {
		Dims    models.Dimensions `json:"dims"`
		Impulse int               `json:"impulse"`
		Verdict string            `json:"verdict"`
		Message string            `json:"message"`
		Note    string            `json:"note"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return Result{}, err
	}

	v := models.Verdict(parsed.Verdict)
	if v != models.VerdictBuy && v != models.VerdictPause && v != models.VerdictResist {
		v = models.VerdictResist
	}
	r := Result{
		Dims:    parsed.Dims,
		Impulse: parsed.Impulse,
		Verdict: v,
		Message: parsed.Message,
		Note:    parsed.Note,
	}
	if r.Message == "" {
		r.Message = Message(v, in.Item, cfg.Persona)
	}
	if r.Note == "" {
		r.Note = note(v, r.Dims)
	}
	return r, nil
}

// buildSystemPrompt fills the admin-editable template's variables that are
// constant for the request and appends the JSON-only instruction.
func buildSystemPrompt(cfg models.AIConfig) string {
	t := cfg.PromptTemplate
	if t == "" {
		t = "你是 LIM，一位温柔、不评判的理性消费陪伴者。"
	}
	return t + "\n\n严格只返回 JSON，不要使用 markdown 代码块。"
}

func fallback(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// extractJSON pulls the first {...} object out of a possibly-fenced reply.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
