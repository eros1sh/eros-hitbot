package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

// ConfigJSON config.json dosya formatı (agents.json, operaagent ile uyumlu)
type ConfigJSON struct {
	ProxyHost           string   `json:"PROXY_HOST"`
	ProxyPort           int      `json:"PROXY_PORT"`
	ProxyUser           string   `json:"PROXY_USER"`
	ProxyPass           string   `json:"PROXY_PASS"`
	Lisans              string   `json:"LISANS"`
	TargetQueries       []string `json:"targetQueries"`
	TargetDomain        string   `json:"targetDomain"`
	FallbackGAID        string   `json:"fallbackGAID"`
	MaxPages            int      `json:"maxPages"`
	DurationMinutes     int      `json:"durationMinutes"`
	HitsPerMinute       int      `json:"hitsPerMinute"`
	MaxConcurrentVisits int      `json:"maxConcurrentVisits"`
	OutputDir           string   `json:"outputDir"`
	ExportFormat        string   `json:"exportFormat"`
	CanvasFingerprint   bool     `json:"canvasFingerprint"`
	ScrollStrategy      string   `json:"scrollStrategy"`
	SendScrollEvent     bool     `json:"sendScrollEvent"`
	Keywords            []string `json:"keywords"`
}

// LoadFromJSON config.json'dan yükler; Config'e dönüştürür
func LoadFromJSON(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var j ConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}

	cfg := &Config{
		TargetDomain:       j.TargetDomain,
		MaxPages:           j.MaxPages,
		DurationMinutes:    j.DurationMinutes,
		HitsPerMinute:      j.HitsPerMinute,
		OutputDir:          j.OutputDir,
		ExportFormat:       j.ExportFormat,
		MaxConcurrentVisits: j.MaxConcurrentVisits,
		CanvasFingerprint:  j.CanvasFingerprint,
		ScrollStrategy:     j.ScrollStrategy,
		SendScrollEvent:    j.SendScrollEvent,
		Keywords:           j.Keywords,
		ProxyHost:          j.ProxyHost,
		ProxyPort:          j.ProxyPort,
		ProxyUser:          j.ProxyUser,
		ProxyPass:          j.ProxyPass,
		GtagID:             j.FallbackGAID,
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./reports"
	}
	if cfg.ExportFormat == "" {
		cfg.ExportFormat = "both"
	}
	if len(cfg.Keywords) == 0 && len(j.TargetQueries) > 0 {
		cfg.Keywords = j.TargetQueries
	}

	if j.ProxyHost != "" && j.ProxyPort > 0 {
		cfg.ProxyEnabled = true
		cfg.ProxyURL = buildProxyURL(j.ProxyHost, j.ProxyPort, j.ProxyUser, j.ProxyPass)
	}

	cfg.ApplyDefaults()
	cfg.ComputeDerived()
	return cfg, nil
}

func buildProxyURL(host string, port int, user, pass string) string {
	if host == "" || port <= 0 {
		return ""
	}
	hostPort := fmt.Sprintf("%s:%d", host, port)
	if user != "" || pass != "" {
		userInfo := url.UserPassword(user, pass)
		return fmt.Sprintf("http://%s@%s", userInfo.String(), hostPort)
	}
	return fmt.Sprintf("http://%s", hostPort)
}
