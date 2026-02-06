package proxy

import (
	"fmt"
	"testing"
	"time"

	pkgproxy "eroshit/pkg/proxy"
)

// TestSelectorIntegration tam entegrasyon testi
func TestSelectorIntegration(t *testing.T) {
	// LivePool oluştur
	pool := NewLivePool()

	// Test proxy'leri ekle
	proxies := []*LiveProxy{
		{ProxyConfig: &ProxyConfig{Host: "fast-us.com", Port: 8080}, Country: "US", SpeedMs: 50},
		{ProxyConfig: &ProxyConfig{Host: "slow-us.com", Port: 8080}, Country: "US", SpeedMs: 500},
		{ProxyConfig: &ProxyConfig{Host: "fast-de.com", Port: 8080}, Country: "DE", SpeedMs: 60},
		{ProxyConfig: &ProxyConfig{Host: "slow-de.com", Port: 8080}, Country: "DE", SpeedMs: 600},
	}

	for _, p := range proxies {
		pool.Add(p)
	}

	// Tüm selector tiplerini test et
	selectorTypes := []string{
		"round-robin",
		"random",
		"least-used",
		"fastest",
		"success-rate",
		"geo",
		"weighted",
	}

	for _, st := range selectorTypes {
		t.Run(st, func(t *testing.T) {
			manager := NewSelectorManager(pool, st)

			// Geo selector için ülke ayarla
			if st == "geo" {
				manager.SetGeoCountries([]string{"US"})
			}

			// Birkaç proxy seç ve sonuç kaydet
			for i := 0; i < 10; i++ {
				proxy := manager.GetProxy()
				if proxy == nil {
					t.Fatalf("Expected proxy for selector type %s", st)
				}

				// Sonuç kaydet
				success := i%3 != 0 // Bazı başarısızlıklar ekle
				manager.RecordResult(proxy, success, 100*time.Millisecond, nil)
			}

			// Metrikleri kontrol et
			metrics := manager.GetMetrics()
			if len(metrics) == 0 {
				t.Error("Expected metrics to be recorded")
			}

			t.Logf("Selector %s: recorded metrics for %d proxies", st, len(metrics))
		})
	}
}

// TestGeoSelectorIntegration geo selector entegrasyon testi
func TestGeoSelectorIntegration(t *testing.T) {
	pool := NewLivePool()

	// Farklı ülkelerden proxy'ler ekle
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "us1.com", Port: 8080},
		Country:     "US",
		SpeedMs:     100,
	})
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "us2.com", Port: 8080},
		Country:     "US",
		SpeedMs:     150,
	})
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "de1.com", Port: 8080},
		Country:     "DE",
		SpeedMs:     80,
	})
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "fr1.com", Port: 8080},
		Country:     "FR",
		SpeedMs:     120,
	})

	// US selector'ü oluştur
	manager := NewSelectorManager(pool, "geo")
	manager.SetGeoCountries([]string{"US"})

	// Sadece US proxy'leri seçilmeli
	for i := 0; i < 50; i++ {
		p := manager.GetProxy()
		if p == nil {
			t.Fatal("Expected proxy, got nil")
		}

		// Seçilen proxy'nin ülkesini doğrula
		found := false
		for _, lp := range pool.Snapshot() {
			if lp.Host == p.Host && lp.Country == "US" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected US proxy, got %s", p.Host)
		}
	}

	t.Log("Geo selector correctly filtered for US proxies")
}

// TestFastestSelectorIntegration fastest selector entegrasyon testi
func TestFastestSelectorIntegration(t *testing.T) {
	pool := NewLivePool()

	// Hız farklı proxy'ler ekle
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "slow.com", Port: 8080},
		SpeedMs:     500,
	})
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "fast.com", Port: 8080},
		SpeedMs:     50,
	})
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "medium.com", Port: 8080},
		SpeedMs:     200,
	})

	// Fastest selector
	manager := NewSelectorManager(pool, "fastest")

	// En hızlıyı seç
	selected := manager.GetProxy()
	if selected == nil {
		t.Fatal("Expected proxy, got nil")
	}

	if selected.Host != "fast.com" {
		t.Errorf("Expected fastest proxy (fast.com), got %s", selected.Host)
	}

	t.Log("Fastest selector correctly selected the fastest proxy")
}

// TestSelectorManagerUsage selector manager kullanım örneği testi
func TestSelectorManagerUsage(t *testing.T) {
	// LivePool oluştur
	pool := NewLivePool()
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "proxy1.com", Port: 8080},
		Country:     "US",
		SpeedMs:     100,
	})
	pool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "proxy2.com", Port: 8080},
		Country:     "DE",
		SpeedMs:     50,
	})

	// Weighted selector ile manager oluştur
	manager := NewSelectorManager(pool, "weighted")

	// Proxy seç
	selected := manager.GetProxy()
	if selected == nil {
		t.Fatal("Expected proxy, got nil")
	}
	t.Logf("Selected: %s:%d", selected.Host, selected.Port)

	// Kullanım sonucunu kaydet
	manager.RecordResult(selected, true, 120*time.Millisecond, nil)

	// Metrikleri kontrol et
	metrics := manager.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected metrics to be recorded")
	}
}

// TestNewSelectorEnabledService selector-enabled service kullanım örneği testi
func TestNewSelectorEnabledService(t *testing.T) {
	// Geo selector ile service oluştur (US ve DE proxy'leri tercih et)
	service := NewSelectorEnabledService("geo", []string{"US", "DE"})

	// Proxy'leri havuza ekle
	service.LivePool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "us-proxy.com", Port: 8080},
		Country:     "US",
		SpeedMs:     100,
	})
	service.LivePool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "de-proxy.com", Port: 8080},
		Country:     "DE",
		SpeedMs:     50,
	})
	service.LivePool.Add(&LiveProxy{
		ProxyConfig: &ProxyConfig{Host: "fr-proxy.com", Port: 8080},
		Country:     "FR",
		SpeedMs:     200,
	})

	// Selector kullanarak proxy al
	selected := service.GetProxyWithSelector()
	if selected == nil {
		t.Fatal("Expected proxy, got nil")
	}

	// US veya DE olmalı
	if selected.Host != "us-proxy.com" && selected.Host != "de-proxy.com" {
		t.Errorf("Expected US or DE proxy, got %s", selected.Host)
	}

	t.Logf("Selected proxy: %s", selected.Host)

	// Kullanım sonucunu kaydet
	service.RecordProxyResult(selected, true, 100*time.Millisecond)
}

// BenchmarkSelectors selector performans karşılaştırması
func BenchmarkSelectors(b *testing.B) {
	// Test pool oluştur
	pool := NewLivePool()
	for i := 0; i < 100; i++ {
		pool.Add(&LiveProxy{
			ProxyConfig: &ProxyConfig{
				Host: fmt.Sprintf("proxy%d.com", i),
				Port: 8080,
			},
			Country: fmt.Sprintf("C%d", i%10),
			SpeedMs: int64(50 + i*10),
		})
	}

	selectorTypes := map[string]string{
		"RoundRobin":  "round-robin",
		"Random":      "random",
		"LeastUsed":   "least-used",
		"Fastest":     "fastest",
		"SuccessRate": "success-rate",
		"Geo":         "geo",
		"Weighted":    "weighted",
	}

	for name, st := range selectorTypes {
		b.Run(name, func(b *testing.B) {
			manager := NewSelectorManager(pool, st)
			if st == "geo" {
				manager.SetGeoCountries([]string{"C0", "C1"})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = manager.GetProxy()
			}
		})
	}
}

// BenchmarkWeightedSelectorWithMetrics metrikli weighted selector benchmark'ı
func BenchmarkWeightedSelectorWithMetrics(b *testing.B) {
	pool := NewLivePool()
	metrics := pkgproxy.NewMetricsCollector()

	// Proxy'leri ekle
	for i := 0; i < 100; i++ {
		pool.Add(&LiveProxy{
			ProxyConfig: &ProxyConfig{
				Host: fmt.Sprintf("proxy%d.com", i),
				Port: 8080,
			},
			SpeedMs: int64(50 + i*10),
		})
	}

	// Bazı metrikler ekle
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("proxy%d.com:8080", i)
		for j := 0; j < 10; j++ {
			metrics.RecordResult(key, &pkgproxy.Result{
				Success:      j%3 != 0,
				ResponseTime: time.Duration(50+i*10) * time.Millisecond,
				Timestamp:    time.Now(),
			})
		}
	}

	selector := pkgproxy.NewWeightedSelector()
	adapter := &livePoolAdapter{pool: pool}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = selector.Select(adapter, metrics)
	}
}
