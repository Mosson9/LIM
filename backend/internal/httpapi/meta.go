package httpapi

import "net/http"

func (a *App) handleCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Categories())
}

func (a *App) handleSkins(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Skins())
}

func (a *App) handlePlans(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Plans())
}

func (a *App) handlePerks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Perks())
}

// handleDimensions exposes the six reflective questions (label + question text)
// the app shows on the result screen. Sourced from the live AI config.
func (a *App) handleDimensions(w http.ResponseWriter, r *http.Request) {
	type dim struct {
		Key   string `json:"key"`
		Label string `json:"label"`
		Q     string `json:"q"`
	}
	cfg := a.store.AIConfig()
	out := make([]dim, 0, len(cfg.Dims))
	for _, d := range cfg.Dims {
		out = append(out, dim{Key: d.Key, Label: d.Label, Q: d.Q})
	}
	writeJSON(w, http.StatusOK, out)
}
