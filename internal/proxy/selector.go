package proxy

import (
	"sync"
	"time"

	pkgproxy "eroshit/pkg/proxy"
)

// SelectorManager iç proxy paketi için selector yöneticisi
type SelectorManager struct {
	pool      *LivePool
	selector  pkgproxy.Selector
	metrics   *pkgproxy.MetricsCollector
	mu        sync.RWMutex
	countries []string // GeoSelector için
}

// NewSelectorManager yeni selector manager oluşturur
func NewSelectorManager(pool *LivePool, selectorType string) *SelectorManager {
	sm := &SelectorManager{
		pool:    pool,
		metrics: pkgproxy.NewMetricsCollector(),
	}
	
	// Seçici tipine göre oluştur
	sm.SetSelectorType(selectorType)
	
	return sm
}

// SetSelectorType seçici tipini değiştirir
func (sm *SelectorManager) SetSelectorType(selectorType string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Geo selector için özel kontrol
	if selectorType == "geo" {
		sm.selector = pkgproxy.NewGeoSelector(sm.countries)
	} else {
		sm.selector = pkgproxy.NewSelectorFromString(selectorType)
	}
}

// SetGeoCountries geo selector için tercih edilen ülkeleri ayarlar
func (sm *SelectorManager) SetGeoCountries(countries []string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.countries = countries
	
	// Eğer geo selector aktifse güncelle
	if geoSel, ok := sm.selector.(*pkgproxy.GeoSelector); ok {
		geoSel.SetPreferredCountries(countries)
	}
}

// GetProxy bir proxy seçer
func (sm *SelectorManager) GetProxy() *ProxyConfig {
	sm.mu.RLock()
	selector := sm.selector
	metrics := sm.metrics
	sm.mu.RUnlock()
	
	if selector == nil || sm.pool == nil {
		// Fallback: livepool'un kendi GetNext metodu
		if sm.pool != nil {
			return sm.pool.GetNext()
		}
		return nil
	}
	
	// Adapter kullanarak LivePool'u LivePoolAccessor'a dönüştür
	adapter := &livePoolAdapter{pool: sm.pool}
	selected := selector.Select(adapter, metrics)
	
	if selected == nil {
		return nil
	}
	
	return &ProxyConfig{
		Host:     selected.Host,
		Port:     selected.Port,
		Username: selected.Username,
		Password: selected.Password,
		Protocol: selected.Protocol,
	}
}

// RecordResult proxy kullanım sonucunu kaydeder
func (sm *SelectorManager) RecordResult(proxy *ProxyConfig, success bool, responseTime time.Duration, err error) {
	if proxy == nil {
		return
	}
	
	result := &pkgproxy.Result{
		Success:      success,
		ResponseTime: responseTime,
		Error:        err,
		Timestamp:    time.Now(),
	}
	
	sm.mu.RLock()
	selector := sm.selector
	metrics := sm.metrics
	sm.mu.RUnlock()
	
	// pkg/proxy ProxyConfig'e dönüştür
	pkgProxy := &pkgproxy.ProxyConfig{
		Host:     proxy.Host,
		Port:     proxy.Port,
		Username: proxy.Username,
		Password: proxy.Password,
		Protocol: proxy.Protocol,
	}
	
	// Seçiciye bildir
	if selector != nil {
		selector.UpdateMetrics(pkgProxy, result)
	}
	
	// Metrikleri kaydet
	if metrics != nil {
		metrics.RecordResult(proxy.Key(), result)
	}
}

// GetMetrics tüm metrikleri döner
func (sm *SelectorManager) GetMetrics() map[string]*pkgproxy.ProxyMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.metrics == nil {
		return nil
	}
	return sm.metrics.GetAllMetrics()
}

// ResetMetrics metrikleri sıfırlar
func (sm *SelectorManager) ResetMetrics() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.metrics != nil {
		sm.metrics.Reset()
	}
}

// CurrentSelectorName mevcut seçici adını döner
func (sm *SelectorManager) CurrentSelectorName() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.selector == nil {
		return "round-robin"
	}
	return sm.selector.Name()
}

// ==================== ADAPTER ====================

// livePoolAdapter LivePool'u pkg/proxy.LivePoolAccessor'a uyarlar
type livePoolAdapter struct {
	pool *LivePool
}

// Snapshot LivePool'dan proxy listesini alır
func (a *livePoolAdapter) Snapshot() []*pkgproxy.LiveProxy {
	if a.pool == nil {
		return nil
	}
	
	liveProxies := a.pool.Snapshot()
	result := make([]*pkgproxy.LiveProxy, len(liveProxies))
	
	for i, lp := range liveProxies {
		result[i] = &pkgproxy.LiveProxy{
			Host:     lp.Host,
			Port:     lp.Port,
			Username: lp.Username,
			Password: lp.Password,
			Protocol: lp.Protocol,
			SpeedMs:  lp.SpeedMs,
			Country:  lp.Country,
		}
	}
	
	return result
}

// Count LivePool'daki proxy sayısını döner
func (a *livePoolAdapter) Count() int {
	if a.pool == nil {
		return 0
	}
	return a.pool.Count()
}

// ==================== SERVICE INTEGRATION ====================

// SelectorEnabledService selector manager'lı proxy seçimi yapan service wrapper
type SelectorEnabledService struct {
	*Service
	selectorManager *SelectorManager
}

// GetProxyWithSelector selector kullanarak proxy alır
func (s *SelectorEnabledService) GetProxyWithSelector() *ProxyConfig {
	if s.selectorManager != nil {
		return s.selectorManager.GetProxy()
	}
	// Fallback: normal GetNext
	if s.Service != nil && s.LivePool != nil {
		return s.LivePool.GetNext()
	}
	return nil
}

// RecordProxyResult proxy kullanım sonucunu kaydeder
func (s *SelectorEnabledService) RecordProxyResult(proxy *ProxyConfig, success bool, responseTime time.Duration) {
	if s.selectorManager != nil {
		var err error
		if !success {
			err = proxyError{msg: "proxy request failed"}
		}
		s.selectorManager.RecordResult(proxy, success, responseTime, err)
	}
}

// GetSelectorManager selector manager'ı döner
func (s *SelectorEnabledService) GetSelectorManager() *SelectorManager {
	return s.selectorManager
}

type proxyError struct {
	msg string
}

func (e proxyError) Error() string {
	return e.msg
}

// InitializeSelector selector manager'ı başlatır
func (s *SelectorEnabledService) InitializeSelector(selectorType string, geoCountries []string) {
	if s.Service == nil || s.LivePool == nil {
		return
	}
	
	s.selectorManager = NewSelectorManager(s.LivePool, selectorType)
	if len(geoCountries) > 0 {
		s.selectorManager.SetGeoCountries(geoCountries)
	}
}

// ==================== SERVICE CONSTRUCTOR ====================

// NewSelectorEnabledService yeni selector-enabled service oluşturur
func NewSelectorEnabledService(selectorType string, geoCountries []string) *SelectorEnabledService {
	svc := &SelectorEnabledService{
		Service: NewService(),
	}
	svc.InitializeSelector(selectorType, geoCountries)
	return svc
}
