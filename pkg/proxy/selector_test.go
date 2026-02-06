package proxy

import (
	"testing"
	"time"
)

// MockLivePool test için mock LivePool
type MockLivePool struct {
	proxies []*LiveProxy
}

func (m *MockLivePool) Snapshot() []*LiveProxy {
	return m.proxies
}

func (m *MockLivePool) Count() int {
	return len(m.proxies)
}

func createTestProxies() []*LiveProxy {
	return []*LiveProxy{
		{Host: "proxy1.com", Port: 8080, SpeedMs: 100, Country: "US"},
		{Host: "proxy2.com", Port: 8080, SpeedMs: 50, Country: "DE"},
		{Host: "proxy3.com", Port: 8080, SpeedMs: 200, Country: "FR"},
		{Host: "proxy4.com", Port: 8080, SpeedMs: 150, Country: "US"},
	}
}

func TestRoundRobinSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewRoundRobinSelector()
	metrics := NewMetricsCollector()
	
	// İlk seçim
	p1 := selector.Select(pool, metrics)
	if p1 == nil {
		t.Fatal("Expected proxy, got nil")
	}
	if p1.Host != "proxy2.com" { // counter 1'den başlıyor
		t.Errorf("Expected proxy2.com, got %s", p1.Host)
	}
	
	// İkinci seçim
	p2 := selector.Select(pool, metrics)
	if p2 == nil {
		t.Fatal("Expected proxy, got nil")
	}
	if p2.Host != "proxy3.com" {
		t.Errorf("Expected proxy3.com, got %s", p2.Host)
	}
	
	// Döngü testi
	for i := 0; i < 10; i++ {
		p := selector.Select(pool, metrics)
		if p == nil {
			t.Fatal("Expected proxy, got nil")
		}
	}
	
	t.Log("RoundRobinSelector test passed")
}

func TestRandomSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewRandomSelector()
	metrics := NewMetricsCollector()
	
	// Boş pool testi
	emptyPool := &MockLivePool{proxies: []*LiveProxy{}}
	if p := selector.Select(emptyPool, metrics); p != nil {
		t.Error("Expected nil for empty pool")
	}
	
	// Rastgele seçimler
	selected := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p := selector.Select(pool, metrics)
		if p == nil {
			t.Fatal("Expected proxy, got nil")
		}
		selected[p.Host] = true
	}
	
	// En az 2 farklı proxy seçilmiş olmalı (olasılıksal olarak hepsi seçilmiş olabilir)
	if len(selected) < 2 {
		t.Errorf("Expected at least 2 different proxies, got %d", len(selected))
	}
	
	t.Logf("RandomSelector selected %d different proxies", len(selected))
}

func TestFastestSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewFastestSelector()
	metrics := NewMetricsCollector()
	
	// En hızlı proxy speedMs=50 (proxy2.com)
	p := selector.Select(pool, metrics)
	if p == nil {
		t.Fatal("Expected proxy, got nil")
	}
	if p.Host != "proxy2.com" {
		t.Errorf("Expected fastest proxy (proxy2.com), got %s", p.Host)
	}
	
	t.Log("FastestSelector test passed")
}

func TestGeoSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	
	// US proxy'leri seç
	selector := NewGeoSelector([]string{"US"})
	metrics := NewMetricsCollector()
	
	// Birçok seçim yap ve hepsinin US olup olmadığını kontrol et
	selectedCountries := make(map[string]int)
	for i := 0; i < 50; i++ {
		p := selector.Select(pool, metrics)
		if p == nil {
			t.Fatal("Expected proxy, got nil")
		}
		// Ülke bilgisini pool'dan al
		for _, lp := range pool.proxies {
			if lp.Host == p.Host && lp.Port == p.Port {
				selectedCountries[lp.Country]++
				if lp.Country != "US" {
					t.Errorf("Expected US proxy, got %s from %s", lp.Country, p.Host)
				}
				break
			}
		}
	}
	
	// DE ve FR için test
	selectorDE := NewGeoSelector([]string{"DE", "FR"})
	for i := 0; i < 50; i++ {
		p := selectorDE.Select(pool, metrics)
		if p == nil {
			t.Fatal("Expected proxy, got nil")
		}
		for _, lp := range pool.proxies {
			if lp.Host == p.Host && lp.Port == p.Port {
				if lp.Country != "DE" && lp.Country != "FR" {
					t.Errorf("Expected DE or FR proxy, got %s from %s", lp.Country, p.Host)
				}
				break
			}
		}
	}
	
	// Olmayan ülke için rastgele seçim
	selectorXX := NewGeoSelector([]string{"XX"})
	p := selectorXX.Select(pool, metrics)
	if p == nil {
		t.Fatal("Expected proxy for non-existent country (fallback), got nil")
	}
	
	t.Log("GeoSelector test passed")
}

func TestWeightedSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewWeightedSelector()
	metrics := NewMetricsCollector()
	
	// Bazı proxy'lere başarı metrikleri ekle
	for i, p := range pool.proxies {
		result := &Result{
			Success:      i%2 == 0, // proxy1 ve proxy3 başarılı
			ResponseTime: time.Duration(p.SpeedMs) * time.Millisecond,
			Timestamp:    time.Now(),
		}
		metrics.RecordResult(p.Key(), result)
	}
	
	// Çok sayıda seçim yap
	selected := make(map[string]int)
	for i := 0; i < 1000; i++ {
		p := selector.Select(pool, metrics)
		if p == nil {
			t.Fatal("Expected proxy, got nil")
		}
		selected[p.Host]++
	}
	
	// Tüm proxy'lerin seçilmiş olması gerekir (ağırlıklı dağılım var)
	if len(selected) < len(pool.proxies) {
		t.Errorf("Expected all %d proxies to be selected, got %d", len(pool.proxies), len(selected))
	}
	
	// Dağılımı göster
	t.Logf("WeightedSelector distribution: %v", selected)
}

func TestLeastUsedSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewLeastUsedSelector()
	metrics := NewMetricsCollector()
	
	// İlk seçim (hiçbiri kullanılmamış)
	p1 := selector.Select(pool, metrics)
	if p1 == nil {
		t.Fatal("Expected proxy, got nil")
	}
	
	// Sonucu kaydet
	selector.UpdateMetrics(p1, &Result{Success: true, Timestamp: time.Now()})
	
	// İkinci seçim (farklı olmalı)
	p2 := selector.Select(pool, metrics)
	if p2 == nil {
		t.Fatal("Expected proxy, got nil")
	}
	
	// Farklı proxy seçilmeli (en az kullanılan)
	if p1.Host == p2.Host {
		t.Error("Expected different proxy for least-used strategy")
	}
	
	t.Logf("LeastUsedSelector selected: %s then %s", p1.Host, p2.Host)
}

func TestSuccessRateSelector(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewSuccessRateSelector()
	metrics := NewMetricsCollector()
	
	// proxy1 çok başarılı
	for i := 0; i < 10; i++ {
		metrics.RecordResult(pool.proxies[0].Key(), &Result{
			Success:   true,
			Timestamp: time.Now(),
		})
	}
	
	// proxy2 başarısız
	for i := 0; i < 10; i++ {
		metrics.RecordResult(pool.proxies[1].Key(), &Result{
			Success:   false,
			Timestamp: time.Now(),
		})
	}
	
	// En başarılıyı seç
	p := selector.Select(pool, metrics)
	if p == nil {
		t.Fatal("Expected proxy, got nil")
	}
	
	// proxy1 seçilmeli (100% başarı oranı)
	if p.Host != "proxy1.com" {
		t.Errorf("Expected proxy1.com (best success rate), got %s", p.Host)
	}
	
	t.Log("SuccessRateSelector test passed")
}

func TestMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()
	
	// Sonuç kaydet
	mc.RecordResult("proxy1:8080", &Result{
		Success:      true,
		ResponseTime: 100 * time.Millisecond,
		Timestamp:    time.Now(),
	})
	
	mc.RecordResult("proxy1:8080", &Result{
		Success:      true,
		ResponseTime: 150 * time.Millisecond,
		Timestamp:    time.Now(),
	})
	
	mc.RecordResult("proxy1:8080", &Result{
		Success:      false,
		ResponseTime: 0,
		Timestamp:    time.Now(),
	})
	
	// Metrikleri kontrol et
	m := mc.GetMetrics("proxy1:8080")
	if m == nil {
		t.Fatal("Expected metrics, got nil")
	}
	
	if m.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", m.TotalRequests)
	}
	
	if m.SuccessRequests != 2 {
		t.Errorf("Expected 2 success requests, got %d", m.SuccessRequests)
	}
	
	if m.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", m.FailedRequests)
	}
	
	expectedRate := 2.0 / 3.0
	if m.SuccessRate() != expectedRate {
		t.Errorf("Expected success rate %f, got %f", expectedRate, m.SuccessRate())
	}
	
	t.Log("MetricsCollector test passed")
}

func TestNewSelectorFactory(t *testing.T) {
	tests := []struct {
		selectorType SelectorType
		expectedName string
	}{
		{SelectorRoundRobin, "round-robin"},
		{SelectorRandom, "random"},
		{SelectorLeastUsed, "least-used"},
		{SelectorFastest, "fastest"},
		{SelectorSuccessRate, "success-rate"},
		{SelectorGeo, "geo"},
		{SelectorWeighted, "weighted"},
		{SelectorType("unknown"), "round-robin"}, // varsayılan
	}
	
	for _, tt := range tests {
		selector := NewSelector(tt.selectorType)
		if selector.Name() != tt.expectedName {
			t.Errorf("NewSelector(%s) = %s, expected %s", tt.selectorType, selector.Name(), tt.expectedName)
		}
	}
}

func TestNewSelectorFromString(t *testing.T) {
	tests := []struct {
		name         string
		expectedName string
	}{
		{"round-robin", "round-robin"},
		{"random", "random"},
		{"least-used", "least-used"},
		{"fastest", "fastest"},
		{"success-rate", "success-rate"},
		{"geo", "geo"},
		{"weighted", "weighted"},
		{"unknown", "round-robin"}, // varsayılan
	}
	
	for _, tt := range tests {
		selector := NewSelectorFromString(tt.name)
		if selector.Name() != tt.expectedName {
			t.Errorf("NewSelectorFromString(%s) = %s, expected %s", tt.name, selector.Name(), tt.expectedName)
		}
	}
}

func TestSelectorManager(t *testing.T) {
	pool := &MockLivePool{proxies: createTestProxies()}
	selector := NewRoundRobinSelector()
	
	manager := NewSelectorManager(selector, pool)
	
	// Proxy al
	p := manager.GetProxy()
	if p == nil {
		t.Fatal("Expected proxy, got nil")
	}
	
	// Sonuç kaydet
	manager.RecordResult(p, true, 100*time.Millisecond, nil)
	
	// Metrikleri kontrol et
	metrics := manager.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected metrics to be recorded")
	}
	
	// Seçici adını kontrol et
	if manager.CurrentSelectorName() != "round-robin" {
		t.Errorf("Expected selector name 'round-robin', got %s", manager.CurrentSelectorName())
	}
	
	// Seçici değiştir
	manager.SetSelector(NewRandomSelector())
	if manager.CurrentSelectorName() != "random" {
		t.Errorf("Expected selector name 'random', got %s", manager.CurrentSelectorName())
	}
	
	// Metrikleri sıfırla
	manager.ResetMetrics()
	metrics = manager.GetMetrics()
	if len(metrics) != 0 {
		t.Error("Expected metrics to be reset")
	}
	
	t.Log("SelectorManager test passed")
}
