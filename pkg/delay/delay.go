package delay

import (
	"context"
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

// TokenBucket dakikada HPM istek sınırı; başta burst = capacity (tüm slotları hemen doldur).
type TokenBucket struct {
	ch   chan struct{}
	stop func()
}

// NewTokenBucket hitsPerMinute hızında token doldurur; capacity kadar burst (başta hemen N istek).
func NewTokenBucket(ctx context.Context, hitsPerMinute, capacity int) *TokenBucket {
	if capacity <= 0 {
		capacity = 1
	}
	if hitsPerMinute <= 0 {
		hitsPerMinute = 60
	}
	ch := make(chan struct{}, 256) // burst + refill için yeterli buffer
	for i := 0; i < capacity; i++ {
		ch <- struct{}{}
	}
	refillInterval := time.Minute / time.Duration(hitsPerMinute)
	ctx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(refillInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}()
	return &TokenBucket{ch: ch, stop: cancel}
}

// Take bir token alır; ctx iptal olursa hemen döner.
func (tb *TokenBucket) Take(ctx context.Context) error {
	select {
	case <-tb.ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop refill goroutine'ini durdurur.
func (tb *TokenBucket) Stop() {
	if tb.stop != nil {
		tb.stop()
	}
}
