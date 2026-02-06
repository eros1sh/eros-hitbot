package server

import (
	"encoding/json"
	"net/http"

	"eroshit/pkg/i18n"
)

// handleI18N returns translations for the web UI
func (s *Server) handleI18N(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get locale from query parameter, default to "tr"
	locale := r.URL.Query().Get("locale")
	if locale != "en" {
		locale = "tr"
	}

	// Get all web translations
	translations := i18n.GetAllWebTranslations(locale)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"locale":       locale,
		"translations": translations,
	})
}

// handleI18NKeys returns specific translation keys
func (s *Server) handleI18NKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Locale string   `json:"locale"`
		Keys   []string `json:"keys"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Locale != "en" {
		req.Locale = "tr"
	}

	result := make(map[string]string)
	for _, key := range req.Keys {
		result[key] = i18n.WebT(req.Locale, key)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
