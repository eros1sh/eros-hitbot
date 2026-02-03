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
	ScrollStrategy        string   `json:"scrollStrategy"`
	SendScrollEvent       bool     `json:"sendScrollEvent"`
	UseSitemap            bool     `json:"useSitemap"`
	SitemapHomepageWeight int      `json:"sitemapHomepageWeight"`
	Keywords              []string `json:"keywords"`
	UsePublicProxy        bool     `json:"usePublicProxy"`
	ProxySourceURLs       []string `json:"proxySourceURLs"`
	GitHubRepos           []string `json:"githubRepos"`
	CheckerWorkers        int      `json:"checkerWorkers"`
	// Private proxy listesi
	PrivateProxies        []PrivateProxyJSON `json:"privateProxies"`
	UsePrivateProxy       bool               `json:"usePrivateProxy"`
	// Yeni alanlar
	DeviceType        string   `json:"deviceType"`
	DeviceBrands      []string `json:"deviceBrands"`
	ReferrerKeyword   string   `json:"referrerKeyword"`
	ReferrerEnabled   bool     `json:"referrerEnabled"`
}

// PrivateProxyJSON JSON formatında private proxy
type PrivateProxyJSON struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Protocol string `json:"protocol"`
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

	// Private proxy'leri dönüştür
	var privateProxies []PrivateProxy
	for _, pp := range j.PrivateProxies {
		if pp.Host != "" && pp.Port > 0 {
			privateProxies = append(privateProxies, PrivateProxy{
				Host:     pp.Host,
				Port:     pp.Port,
				User:     pp.User,
				Pass:     pp.Pass,
				Protocol: pp.Protocol,
			})
		}
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
		ScrollStrategy:        j.ScrollStrategy,
		SendScrollEvent:       j.SendScrollEvent,
		UseSitemap:            j.UseSitemap,
		SitemapHomepageWeight: j.SitemapHomepageWeight,
		Keywords:              j.Keywords,
		UsePublicProxy:        j.UsePublicProxy,
		ProxySourceURLs:       j.ProxySourceURLs,
		GitHubRepos:           j.GitHubRepos,
		CheckerWorkers:        j.CheckerWorkers,
		ProxyHost:          j.ProxyHost,
		ProxyPort:          j.ProxyPort,
		ProxyUser:          j.ProxyUser,
		ProxyPass:          j.ProxyPass,
		GtagID:             j.FallbackGAID,
		// Private proxy alanları
		PrivateProxies:    privateProxies,
		UsePrivateProxy:   j.UsePrivateProxy,
		// Yeni alanlar
		DeviceType:        j.DeviceType,
		DeviceBrands:      j.DeviceBrands,
		ReferrerKeyword:   j.ReferrerKeyword,
		ReferrerEnabled:   j.ReferrerEnabled,
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
