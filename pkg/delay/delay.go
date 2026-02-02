package delay

import (
	"math/rand"
	"time"
)

// Jitter ±percent varyasyon ile gecikme
// örn: Base=2s, Percent=20 -> 1.6s - 2.4s arası
func Jitter(base time.Duration, percent float64) time.Duration {
	if percent <= 0 || percent > 100 {
		return base
	}
	delta := float64(base) * (percent / 100)
	min := float64(base) - delta
	max := float64(base) + delta
	if min < 0 {
		min = 0
	}
	ms := min + rand.Float64()*(max-min)
	return time.Duration(ms)
}

// NaturalDelay doğal insan davranışı simülasyonu
// Sayfa yükleme sonrası tipik bekleme: 2-8 saniye
func NaturalDelay() time.Duration {
	base := 3 * time.Second
	return Jitter(base, 80)
}

// RequestInterval istekler arası önerilen minimum aralık
func RequestInterval(hitsPerMinute int) time.Duration {
	if hitsPerMinute <= 0 {
		return 2 * time.Second
	}
	base := time.Minute / time.Duration(hitsPerMinute)
	return Jitter(base, 25)
}

// PageLoadDelay sayfa geçişi simülasyonu (1-4 sn)
func PageLoadDelay() time.Duration {
	base := 2 * time.Second
	return Jitter(base, 50)
}
