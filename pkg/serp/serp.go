package serp

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// SERPConfig SERP CTR manipülasyonu yapılandırması
type SERPConfig struct {
	Keywords       []string // Aranacak kelimeler
	TargetDomain   string   // Hedef domain (tıklanacak site)
	MaxPages       int      // Maksimum arama sayfası (1-5)
	SearchEngine   string   // "google", "bing", "yandex"
	ClickDelay     time.Duration // Tıklamadan önce bekleme
	ScrollBehavior string   // "natural", "fast", "slow"
	Language       string   // Arama dili (tr, en, de, vb.)
	Country        string   // Ülke kodu (tr, us, de, vb.)
}

// SERPResult SERP işlem sonucu
type SERPResult struct {
	Keyword      string
	TargetFound  bool
	TargetPage   int      // Hangi sayfada bulundu (1-5)
	TargetRank   int      // Sayfadaki sıralama (1-10)
	ClickedURL   string
	SearchURL    string
	Error        error
	Duration     time.Duration
}

// SERPClicker SERP CTR manipülatörü
type SERPClicker struct {
	config SERPConfig
	mu     sync.Mutex
	rng    *rand.Rand
}

// NewSERPClicker yeni SERP clicker oluşturur
func NewSERPClicker(config SERPConfig) *SERPClicker {
	if config.MaxPages <= 0 {
		config.MaxPages = 3
	}
	if config.MaxPages > 5 {
		config.MaxPages = 5
	}
	if config.SearchEngine == "" {
		config.SearchEngine = "google"
	}
	if config.ClickDelay <= 0 {
		config.ClickDelay = 2 * time.Second
	}
	if config.ScrollBehavior == "" {
		config.ScrollBehavior = "natural"
	}
	if config.Language == "" {
		config.Language = "tr"
	}
	if config.Country == "" {
		config.Country = "tr"
	}
	
	return &SERPClicker{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetRandomKeyword rastgele bir anahtar kelime döner
func (s *SERPClicker) GetRandomKeyword() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if len(s.config.Keywords) == 0 {
		return s.config.TargetDomain
	}
	return s.config.Keywords[s.rng.Intn(len(s.config.Keywords))]
}

// BuildSearchURL arama URL'si oluşturur
func (s *SERPClicker) BuildSearchURL(keyword string, page int) string {
	encoded := url.QueryEscape(keyword)
	
	switch s.config.SearchEngine {
	case "google":
		baseURL := fmt.Sprintf("https://www.google.com/search?q=%s&hl=%s&gl=%s", 
			encoded, s.config.Language, s.config.Country)
		if page > 1 {
			start := (page - 1) * 10
			baseURL += fmt.Sprintf("&start=%d", start)
		}
		return baseURL
		
	case "bing":
		baseURL := fmt.Sprintf("https://www.bing.com/search?q=%s&setlang=%s", 
			encoded, s.config.Language)
		if page > 1 {
			first := (page-1)*10 + 1
			baseURL += fmt.Sprintf("&first=%d", first)
		}
		return baseURL
		
	case "yandex":
		baseURL := fmt.Sprintf("https://yandex.com/search/?text=%s&lr=%s", 
			encoded, s.config.Country)
		if page > 1 {
			baseURL += fmt.Sprintf("&p=%d", page-1)
		}
		return baseURL
		
	default:
		return fmt.Sprintf("https://www.google.com/search?q=%s", encoded)
	}
}

// SearchAndClick arama yapıp hedef siteyi bulur ve tıklar
func (s *SERPClicker) SearchAndClick(ctx context.Context, keyword string) *SERPResult {
	start := time.Now()
	result := &SERPResult{
		Keyword:     keyword,
		TargetFound: false,
	}
	
	// Her sayfa için ara
	for page := 1; page <= s.config.MaxPages; page++ {
		searchURL := s.BuildSearchURL(keyword, page)
		result.SearchURL = searchURL
		
		// Sayfaya git
		if err := chromedp.Run(ctx, chromedp.Navigate(searchURL)); err != nil {
			result.Error = fmt.Errorf("arama sayfasına gidilemedi: %w", err)
			result.Duration = time.Since(start)
			return result
		}
		
		// Sayfa yüklenmesini bekle
		if err := chromedp.Run(ctx, chromedp.Sleep(2*time.Second)); err != nil {
			result.Error = err
			result.Duration = time.Since(start)
			return result
		}
		
		// Doğal scroll davranışı
		if err := s.naturalScroll(ctx); err != nil {
			// Scroll hatası kritik değil, devam et
			_ = err
		}
		
		// Arama sonuçlarını bul
		found, rank, clickURL, err := s.findAndClickTarget(ctx, page)
		if err != nil {
			result.Error = err
			result.Duration = time.Since(start)
			return result
		}
		
		if found {
			result.TargetFound = true
			result.TargetPage = page
			result.TargetRank = rank
			result.ClickedURL = clickURL
			result.Duration = time.Since(start)
			return result
		}
		
		// Sonraki sayfaya geçmeden önce bekle
		if page < s.config.MaxPages {
			time.Sleep(time.Duration(1000+s.rng.Intn(2000)) * time.Millisecond)
		}
	}
	
	result.Duration = time.Since(start)
	return result
}

// findAndClickTarget hedef siteyi bulur ve tıklar
func (s *SERPClicker) findAndClickTarget(ctx context.Context, page int) (found bool, rank int, clickURL string, err error) {
	var nodes []*cdp.Node
	
	// Google arama sonuçlarını seç
	selector := s.getResultSelector()
	
	if err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll)); err != nil {
		return false, 0, "", fmt.Errorf("sonuçlar bulunamadı: %w", err)
	}
	
	targetDomain := strings.ToLower(s.config.TargetDomain)
	targetDomain = strings.TrimPrefix(targetDomain, "www.")
	
	for i := range nodes {
		// Link'i bul
		var href string
		if err := chromedp.Run(ctx, chromedp.AttributeValue(fmt.Sprintf("%s:nth-child(%d) a", selector, i+1), "href", &href, nil)); err != nil {
			continue
		}
		
		// URL'yi parse et
		parsedURL, err := url.Parse(href)
		if err != nil {
			continue
		}
		
		host := strings.ToLower(parsedURL.Host)
		host = strings.TrimPrefix(host, "www.")
		
		// Hedef domain'i kontrol et
		if strings.Contains(host, targetDomain) || strings.Contains(targetDomain, host) {
			// Tıklamadan önce bekle (doğal davranış)
			s.mu.Lock()
			delay := s.config.ClickDelay + time.Duration(s.rng.Intn(1500))*time.Millisecond
			s.mu.Unlock()
			time.Sleep(delay)
			
			// Sonuca scroll yap
			if err := chromedp.Run(ctx, chromedp.ScrollIntoView(fmt.Sprintf("%s:nth-child(%d)", selector, i+1))); err != nil {
				// Scroll hatası kritik değil
				_ = err
			}
			
			// Tıkla
			if err := chromedp.Run(ctx, chromedp.Click(fmt.Sprintf("%s:nth-child(%d) a", selector, i+1))); err != nil {
				return false, 0, "", fmt.Errorf("tıklama hatası: %w", err)
			}
			
			// Sayfa yüklenmesini bekle
			if err := chromedp.Run(ctx, chromedp.Sleep(3*time.Second)); err != nil {
				// Bekleme hatası kritik değil
				_ = err
			}
			
			return true, i + 1, href, nil
		}
	}
	
	return false, 0, "", nil
}

// getResultSelector arama motoruna göre sonuç seçicisi döner
func (s *SERPClicker) getResultSelector() string {
	switch s.config.SearchEngine {
	case "google":
		return "div.g"
	case "bing":
		return "li.b_algo"
	case "yandex":
		return "li.serp-item"
	default:
		return "div.g"
	}
}

// naturalScroll doğal scroll davranışı simüle eder
func (s *SERPClicker) naturalScroll(ctx context.Context) error {
	s.mu.Lock()
	behavior := s.config.ScrollBehavior
	s.mu.Unlock()
	
	var scrollSteps int
	var scrollDelay time.Duration
	
	switch behavior {
	case "fast":
		scrollSteps = 2
		scrollDelay = 200 * time.Millisecond
	case "slow":
		scrollSteps = 6
		scrollDelay = 800 * time.Millisecond
	default: // natural
		scrollSteps = 4
		scrollDelay = 400 * time.Millisecond
	}
	
	for i := 0; i < scrollSteps; i++ {
		scrollAmount := 200 + s.rng.Intn(300)
		script := fmt.Sprintf("window.scrollBy(0, %d)", scrollAmount)
		
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			return err
		}
		
		// Rastgele bekleme
		jitter := time.Duration(s.rng.Intn(200)) * time.Millisecond
		time.Sleep(scrollDelay + jitter)
	}
	
	return nil
}

// SimulatePageVisit hedef sayfada doğal davranış simüle eder
func (s *SERPClicker) SimulatePageVisit(ctx context.Context, dwellTime time.Duration) error {
	if dwellTime <= 0 {
		dwellTime = 15 * time.Second
	}
	
	start := time.Now()
	
	// Sayfa yüklenmesini bekle
	if err := chromedp.Run(ctx, chromedp.WaitReady("body")); err != nil {
		return err
	}
	
	// Doğal scroll
	for time.Since(start) < dwellTime {
		// Scroll
		scrollAmount := 100 + s.rng.Intn(200)
		script := fmt.Sprintf("window.scrollBy(0, %d)", scrollAmount)
		
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			// Scroll hatası kritik değil
			_ = err
		}
		
		// Rastgele bekleme
		waitTime := time.Duration(500+s.rng.Intn(1500)) * time.Millisecond
		time.Sleep(waitTime)
		
		// Bazen yukarı scroll
		if s.rng.Float64() < 0.2 {
			upScroll := 50 + s.rng.Intn(100)
			script := fmt.Sprintf("window.scrollBy(0, -%d)", upScroll)
			_ = chromedp.Run(ctx, chromedp.Evaluate(script, nil))
		}
	}
	
	return nil
}

// GetSearchEngines desteklenen arama motorlarını döner
func GetSearchEngines() []string {
	return []string{"google", "bing", "yandex"}
}

// GetScrollBehaviors desteklenen scroll davranışlarını döner
func GetScrollBehaviors() []string {
	return []string{"natural", "fast", "slow"}
}
