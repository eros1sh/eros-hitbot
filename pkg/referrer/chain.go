package referrer

import (
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

// ReferrerSource referrer kaynağı
type ReferrerSource struct {
	Type     string
	URL      string
	Platform string
}

// ReferrerConfig referrer dağılım yapılandırması
type ReferrerConfig struct {
	GooglePercent   int
	BingPercent     int
	DirectPercent   int
	SocialPercent   int
	InternalPercent int
	Keywords        []string
	SocialPlatforms []string
}

// ReferrerChain referrer zinciri yöneticisi
type ReferrerChain struct {
	config       *ReferrerConfig
	targetDomain string
	visitedPages []string
	mu           sync.Mutex
}

var defaultKeywords = []string{
	"web development", "seo tools", "analytics",
	"website optimization", "digital marketing",
}

var defaultSocialPlatforms = []string{
	"facebook", "twitter", "linkedin", "instagram", "reddit",
}

// NewReferrerChain yeni referrer zinciri oluşturur
func NewReferrerChain(targetDomain string, config *ReferrerConfig) *ReferrerChain {
	if config == nil {
		config = &ReferrerConfig{
			GooglePercent:   40,
			BingPercent:     10,
			DirectPercent:   20,
			SocialPercent:   10,
			InternalPercent: 20,
			Keywords:        defaultKeywords,
			SocialPlatforms: defaultSocialPlatforms,
		}
	}
	return &ReferrerChain{
		config:       config,
		targetDomain: targetDomain,
		visitedPages: make([]string, 0, 20),
	}
}

// Generate yeni referrer kaynağı üretir
func (r *ReferrerChain) Generate() *ReferrerSource {
	r.mu.Lock()
	roll := rand.Intn(100)
	cumulative := 0

	cumulative += r.config.GooglePercent
	if roll < cumulative {
		r.mu.Unlock()
		return &ReferrerSource{
			Type:     "search",
			URL:      r.generateSearchReferrer("google"),
			Platform: "google",
		}
	}

	cumulative += r.config.BingPercent
	if roll < cumulative {
		r.mu.Unlock()
		return &ReferrerSource{
			Type:     "search",
			URL:      r.generateSearchReferrer("bing"),
			Platform: "bing",
		}
	}

	cumulative += r.config.SocialPercent
	if roll < cumulative {
		r.mu.Unlock()
		return &ReferrerSource{
			Type:     "social",
			URL:      r.generateSocialReferrer(),
			Platform: r.selectRandomSocial(),
		}
	}

	cumulative += r.config.InternalPercent
	if roll < cumulative && len(r.visitedPages) > 0 {
		pages := make([]string, len(r.visitedPages))
		copy(pages, r.visitedPages)
		internal := r.generateInternalReferrerUnlocked(pages)
		r.mu.Unlock()
		return &ReferrerSource{
			Type:     "internal",
			URL:      internal,
			Platform: "internal",
		}
	}

	r.mu.Unlock()
	return &ReferrerSource{Type: "direct", URL: "", Platform: "direct"}
}

func (r *ReferrerChain) generateSearchReferrer(engine string) string {
	keyword := r.selectRandomKeyword()
	encoded := url.QueryEscape(keyword)
	templates := map[string]string{
		"google":      "https://www.google.com/search?q=%s",
		"bing":        "https://www.bing.com/search?q=%s",
		"duckduckgo":  "https://duckduckgo.com/?q=%s",
		"yandex":      "https://yandex.com/search/?text=%s",
	}
	tpl, ok := templates[engine]
	if !ok {
		tpl = templates["google"]
	}
	return fmt.Sprintf(tpl, encoded)
}

func (r *ReferrerChain) generateSocialReferrer() string {
	platform := r.selectRandomSocial()
	referrers := map[string][]string{
		"facebook":  {"https://www.facebook.com/", "https://m.facebook.com/"},
		"twitter":   {"https://t.co/", "https://twitter.com/"},
		"linkedin":  {"https://www.linkedin.com/feed/", "https://www.linkedin.com/"},
		"instagram": {"https://www.instagram.com/"},
		"reddit":    {"https://www.reddit.com/r/webdev/", "https://www.reddit.com/"},
		"pinterest": {"https://www.pinterest.com/pin/"},
	}
	urls, ok := referrers[platform]
	if !ok {
		return "https://www.facebook.com/"
	}
	return urls[rand.Intn(len(urls))]
}

func (r *ReferrerChain) generateInternalReferrerUnlocked(pages []string) string {
	if len(pages) == 0 {
		return ""
	}
	if rand.Float64() < 0.7 {
		return pages[len(pages)-1]
	}
	return pages[rand.Intn(len(pages))]
}

func (r *ReferrerChain) selectRandomKeyword() string {
	r.mu.Lock()
	kws := r.config.Keywords
	r.mu.Unlock()

	if len(kws) == 0 {
		return r.targetDomain
	}
	if rand.Float64() < 0.5 {
		return r.targetDomain
	}
	return kws[rand.Intn(len(kws))]
}

func (r *ReferrerChain) selectRandomSocial() string {
	r.mu.Lock()
	platforms := r.config.SocialPlatforms
	r.mu.Unlock()

	if len(platforms) == 0 {
		return "facebook"
	}
	return platforms[rand.Intn(len(platforms))]
}

// AddVisitedPage ziyaret edilen sayfayı ekler
func (r *ReferrerChain) AddVisitedPage(pageURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.visitedPages = append(r.visitedPages, pageURL)
	if len(r.visitedPages) > 10 {
		r.visitedPages = r.visitedPages[len(r.visitedPages)-10:]
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
