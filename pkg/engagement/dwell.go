package engagement

import (
	"context"
	"math/rand"
	"strings"
	"time"
)

// DwellTime sayfa bekleme süresi
type DwellTime struct {
	MinSeconds int
	MaxSeconds int
	PageType   string
}

// Calculate sayfa tipine göre bekleme süresi
func (dt DwellTime) Calculate() time.Duration {
	var base int
	switch dt.PageType {
	case "article":
		base = 45 + rand.Intn(90)
	case "product":
		base = 30 + rand.Intn(60)
	case "list":
		base = 15 + rand.Intn(30)
	default:
		base = 20 + rand.Intn(40)
	}
	if base < dt.MinSeconds {
		base = dt.MinSeconds
	}
	if dt.MaxSeconds > 0 && base > dt.MaxSeconds {
		base = dt.MaxSeconds
	}
	return time.Duration(base) * time.Second
}

// Wait bekleme + aktivite simülasyonu
func (dt DwellTime) Wait(ctx context.Context) error {
	duration := dt.Calculate()
	chunks := 3 + rand.Intn(5)
	chunkDuration := duration / time.Duration(chunks)
	for i := 0; i < chunks; i++ {
		select {
		case <-time.After(chunkDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// DetectPageType URL'den sayfa tipi tahmini
func DetectPageType(url string) string {
	url = strings.ToLower(url)
	if strings.Contains(url, "/blog/") || strings.Contains(url, "/article/") || strings.Contains(url, "/post/") {
		return "article"
	}
	if strings.Contains(url, "/product/") || strings.Contains(url, "/urun/") || strings.Contains(url, "/item/") {
		return "product"
	}
	if strings.Contains(url, "/category/") || strings.Contains(url, "/list/") || strings.Contains(url, "/search") {
		return "list"
	}
	return "general"
}
