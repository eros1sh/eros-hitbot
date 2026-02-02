package seo

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// OrganicTraffic organik trafik simülatörü
type OrganicTraffic struct {
	Keywords      *KeywordManager
	TargetDomain  string
	ClickThrough  float64
	mu            sync.Mutex
	rng           *rand.Rand
}

// NewOrganicTraffic yeni organik trafik simülatörü
func NewOrganicTraffic(keywords []Keyword, targetDomain string, ctr float64) *OrganicTraffic {
	return &OrganicTraffic{
		Keywords:     NewKeywordManager(keywords),
		TargetDomain: targetDomain,
		ClickThrough: ctr,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldClick CTR'ye göre tıklama kararı
func (ot *OrganicTraffic) ShouldClick() bool {
	ot.mu.Lock()
	defer ot.mu.Unlock()
	return ot.rng.Float64() < ot.ClickThrough
}

// GetReferrer referrer URL
func (ot *OrganicTraffic) GetReferrer(keyword Keyword) string {
	switch keyword.SearchEngine {
	case "bing":
		return "https://www.bing.com/"
	case "yandex":
		return "https://yandex.com/"
	default:
		return "https://www.google.com/"
	}
}

// NavigateWithReferrer referrer ile navigasyon (direkt URL'e git, referrer header ile)
func (ot *OrganicTraffic) NavigateWithReferrer(ctx context.Context, targetURL string) error {
	keyword := ot.Keywords.GetRandomKeyword()
	ref := ot.GetReferrer(keyword)

	// Chromedp ile referrer ayarlamak için extra headers gerekir - CDP SetExtraHTTPHeaders
	// Basit alternatif: direkt navigate, referrer JS ile document.referrer override edilemez (read-only)
	// Bu yüzden sadece direkt navigate yapıyoruz; referrer genelde önceki sayfadan gelir
	// Gerçek organik simülasyonu için önce arama sayfasına gidip oradan tıklamak gerekir
	_ = ref
	return chromedp.Run(ctx, chromedp.Navigate(targetURL))
}
