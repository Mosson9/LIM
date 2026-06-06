package ai

import (
	"testing"

	"github.com/mosson9/lim/backend/internal/models"
)

func testConfig() models.AIConfig {
	return models.AIConfig{
		Dims: []models.DimensionConfig{
			{Key: "need", Weight: 1.3, Enabled: true},
			{Key: "alt", Weight: 1.1, Enabled: true},
			{Key: "emo", Weight: 1.2, Enabled: true},
			{Key: "value", Weight: 1.0, Enabled: true},
			{Key: "money", Weight: 1.4, Enabled: true},
			{Key: "env", Weight: 0.8, Enabled: true},
		},
		Persona: "温柔朋友",
	}
}

func TestHeuristicResistsImpulseBuy(t *testing.T) {
	// "已经有" + "直播" + "打折" + expensive vs a 6000 budget => strong resist signal.
	in := Input{Item: "限量联名球鞋", Price: 1599, Cat: "fashion", Reason: "直播间打折，已经有好几双了，就是想要"}
	p := models.Profile{MonthBudget: 6000, Members: 3}
	r := Heuristic(in, p, testConfig())

	if r.Verdict != models.VerdictResist {
		t.Fatalf("expected resist, got %s (impulse=%d, dims=%+v)", r.Verdict, r.Impulse, r.Dims)
	}
	if r.Impulse < 62 {
		t.Errorf("expected impulse >= 62 for a clear impulse buy, got %d", r.Impulse)
	}
	if r.Message == "" || r.Note == "" {
		t.Errorf("expected non-empty message/note")
	}
}

func TestHeuristicApprovesNecessity(t *testing.T) {
	// Genuine, durable, frequent, cheap necessity => buy.
	in := Input{Item: "通勤雨衣", Price: 89, Cat: "other", Reason: "旧的坏了，每天通勤必需，长期耐用"}
	p := models.Profile{MonthBudget: 6000, Members: 3}
	r := Heuristic(in, p, testConfig())

	if r.Verdict != models.VerdictBuy {
		t.Fatalf("expected buy, got %s (impulse=%d, dims=%+v)", r.Verdict, r.Impulse, r.Dims)
	}
	if r.Impulse >= 45 {
		t.Errorf("expected impulse < 45 for a necessity, got %d", r.Impulse)
	}
}

func TestDimensionsClampedToRange(t *testing.T) {
	in := Input{Item: "x", Price: 100000, Cat: "other"} // huge price tanks money score
	r := Heuristic(in, models.Profile{MonthBudget: 6000}, testConfig())
	for name, v := range map[string]float64{
		"need": r.Dims.Need, "alt": r.Dims.Alt, "emo": r.Dims.Emo,
		"value": r.Dims.Value, "money": r.Dims.Money, "env": r.Dims.Env,
	} {
		if v < 1 || v > 5 {
			t.Errorf("dimension %s out of range: %v", name, v)
		}
	}
	if r.Impulse < 4 || r.Impulse > 96 {
		t.Errorf("impulse out of range: %d", r.Impulse)
	}
}

func TestDisabledDimensionIgnored(t *testing.T) {
	cfg := testConfig()
	// Disable money — the dominant axis for an over-budget item — and the verdict
	// should soften relative to the all-enabled config.
	in := Input{Item: "相机", Price: 5000, Cat: "digital"}
	p := models.Profile{MonthBudget: 6000}

	withMoney := Heuristic(in, p, cfg)
	cfg.Dims[4].Enabled = false // money
	withoutMoney := Heuristic(in, p, cfg)

	if withoutMoney.Impulse > withMoney.Impulse {
		t.Errorf("disabling the weak money axis should not raise impulse: with=%d without=%d",
			withMoney.Impulse, withoutMoney.Impulse)
	}
}
