package proxy

import (
	"math"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Result proxy kullanım sonucu
type Result struct {
	Success      bool
	ResponseTime time.Duration
	Error        error
	Timestamp    time.Time
}

// ProxyMetrics proxy performans metrikleri
type ProxyMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	LastUsed        time.Time
	AvgResponseTime time.Duration
	LastResult      *Result
}

// SuccessRate başarı oranını döner (0.0 - 1.0)
func (m *ProxyMetrics) SuccessRate() float64 {
	if m.TotalRequests == 0 {
		return 1.0 // Yeni proxy'ler için varsayılan yüksek başarı oranı
	}
	return float64(m.SuccessRequests) / float64(m.TotalRequests)
}

// MetricsCollector tüm proxy'lerin metriklerini toplar
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*ProxyMetrics // key -> metrics
}

// NewMetricsCollector yeni metrik toplayıcı oluşturur
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*ProxyMetrics),
	}
}

// GetMetrics proxy'nin metriklerini döner
func (mc *MetricsCollector) GetMetrics(proxyKey string) *ProxyMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if m, ok := mc.metrics[proxyKey]; ok {
		return m
	}
	return nil
}

// GetOrCreateMetrics proxy'nin metriklerini döner, yoksa oluşturur
func (mc *MetricsCollector) GetOrCreateMetrics(proxyKey string) *ProxyMetrics {
	mc.mu.RLock()
	if m, ok := mc.metrics[proxyKey]; ok {
		mc.mu.RUnlock()
		return m
	}
	mc.mu.RUnlock()
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if m, ok := mc.metrics[proxyKey]; ok {
		return m
	}
	
	m := &ProxyMetrics{}
	mc.metrics[proxyKey] = m
	return m
}

// RecordResult proxy kullanım sonucunu kaydeder
func (mc *MetricsCollector) RecordResult(proxyKey string, result *Result) {
	m := mc.GetOrCreateMetrics(proxyKey)
	
	atomic.AddInt64(&m.TotalRequests, 1)
	if result.Success {
		atomic.AddInt64(&m.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&m.FailedRequests, 1)
	}
	
	m.LastUsed = result.Timestamp
	m.LastResult = result
	
	// Ortalama yanıt süresini güncelle (exponential moving average)
	if result.Success && result.ResponseTime > 0 {
		if m.AvgResponseTime == 0 {
			m.AvgResponseTime = result.ResponseTime
		} else {
			alpha := 0.3
			m.AvgResponseTime = time.Duration(
				float64(m.AvgResponseTime)*(1-alpha) + float64(result.ResponseTime)*alpha,
			)
		}
	}
}

// GetAllMetrics tüm metriklerin kopyasını döner
func (mc *MetricsCollector) GetAllMetrics() map[string]*ProxyMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	out := make(map[string]*ProxyMetrics, len(mc.metrics))
	for k, v := range mc.metrics {
		cp := *v
		out[k] = &cp
	}
	return out
}

// Reset metrikleri sıfırlar
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics = make(map[string]*ProxyMetrics)
}

// LivePoolAccessor LivePool için interface (internal/proxy/livepool.go ile entegrasyon)
type LivePoolAccessor interface {
	Snapshot() []*LiveProxy
	Count() int
}

// LiveProxy livepool'dan gelen proxy yapısı (internal/proxy'den)
type LiveProxy struct {
	Host     string
	Port     int
	Username string
	Password string
	Protocol string
	SpeedMs  int64
	Country  string
}

// Key benzersiz proxy anahtarı
func (lp *LiveProxy) Key() string {
	return lp.Host + ":" + strconv.Itoa(lp.Port)
}

// ToProxyConfig ProxyConfig'e dönüştürür
func (lp *LiveProxy) ToProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		Host:     lp.Host,
		Port:     lp.Port,
		Username: lp.Username,
		Password: lp.Password,
		Protocol: lp.Protocol,
	}
}

// ProxyConfig selector'dan dışarıya dönen proxy yapılandırması
type ProxyConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Protocol string
}

// Key benzersiz proxy anahtarı
func (pc *ProxyConfig) Key() string {
	return pc.Host + ":" + strconv.Itoa(pc.Port)
}

// Selector proxy seçim stratejisi interface'i
type Selector interface {
	// Select bir proxy seçer
	Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig
	// UpdateMetrics proxy kullanım sonrası metrikleri günceller
	UpdateMetrics(proxy *ProxyConfig, result *Result)
	// Name seçici adını döner
	Name() string
}

// BaseSelector tüm selector'ler için ortak yapı
type BaseSelector struct {
	mu              sync.RWMutex
	metrics         *MetricsCollector
	lastUsed        map[string]time.Time
	useCount        map[string]int64
	lastSelectedIdx uint32
}

// NewBaseSelector yeni temel seçici oluşturur
func NewBaseSelector() *BaseSelector {
	return &BaseSelector{
		lastUsed: make(map[string]time.Time),
		useCount: make(map[string]int64),
	}
}

// UpdateMetrics proxy kullanım sonrası metrikleri günceller
func (bs *BaseSelector) UpdateMetrics(proxy *ProxyConfig, result *Result) {
	if proxy == nil || result == nil {
		return
	}
	
	key := proxy.Key()
	bs.mu.Lock()
	bs.lastUsed[key] = result.Timestamp
	bs.useCount[key]++
	bs.mu.Unlock()
	
	if bs.metrics != nil {
		bs.metrics.RecordResult(key, result)
	}
}

// getLastUsed son kullanım zamanını döner
func (bs *BaseSelector) getLastUsed(key string) time.Time {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.lastUsed[key]
}

// getUseCount kullanım sayısını döner
func (bs *BaseSelector) getUseCount(key string) int64 {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.useCount[key]
}

// SetMetricsCollector metrik toplayıcıyı ayarlar
func (bs *BaseSelector) SetMetricsCollector(mc *MetricsCollector) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.metrics = mc
}

// ==================== ROUND ROBIN SELECTOR ====================

// RoundRobinSelector sırayla proxy seçer
type RoundRobinSelector struct {
	*BaseSelector
	counter uint32
}

// NewRoundRobinSelector yeni round-robin seçici oluşturur
func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{
		BaseSelector: NewBaseSelector(),
	}
}

// Name seçici adını döner
func (rr *RoundRobinSelector) Name() string {
	return "round-robin"
}

// Select sıradaki proxy'yi döner
func (rr *RoundRobinSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	idx := atomic.AddUint32(&rr.counter, 1) % uint32(len(proxies))
	return proxies[idx].ToProxyConfig()
}

// ==================== RANDOM SELECTOR ====================

// RandomSelector rastgele proxy seçer
type RandomSelector struct {
	*BaseSelector
	rng *rand.Rand
}

// NewRandomSelector yeni rastgele seçici oluşturur
func NewRandomSelector() *RandomSelector {
	return &RandomSelector{
		BaseSelector: NewBaseSelector(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Name seçici adını döner
func (r *RandomSelector) Name() string {
	return "random"
}

// Select rastgele bir proxy döner
func (r *RandomSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	r.mu.Lock()
	idx := r.rng.Intn(len(proxies))
	r.mu.Unlock()
	
	return proxies[idx].ToProxyConfig()
}

// ==================== LEAST USED SELECTOR ====================

// LeastUsedSelector en az kullanılan proxy'yi seçer
type LeastUsedSelector struct {
	*BaseSelector
}

// NewLeastUsedSelector yeni en az kullanılan seçici oluşturur
func NewLeastUsedSelector() *LeastUsedSelector {
	return &LeastUsedSelector{
		BaseSelector: NewBaseSelector(),
	}
}

// Name seçici adını döner
func (lu *LeastUsedSelector) Name() string {
	return "least-used"
}

// Select en az kullanılan proxy'yi döner
func (lu *LeastUsedSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	var selected *LiveProxy
	minCount := int64(math.MaxInt64)
	
	for _, p := range proxies {
		count := lu.getUseCount(p.Key())
		if count < minCount {
			minCount = count
			selected = p
		}
	}
	
	if selected == nil && len(proxies) > 0 {
		selected = proxies[0]
	}
	
	return selected.ToProxyConfig()
}

// ==================== FASTEST SELECTOR ====================

// FastestSelector en hızlı proxy'yi seçer (SpeedMs)
type FastestSelector struct {
	*BaseSelector
}

// NewFastestSelector yeni en hızlı seçici oluşturur
func NewFastestSelector() *FastestSelector {
	return &FastestSelector{
		BaseSelector: NewBaseSelector(),
	}
}

// Name seçici adını döner
func (f *FastestSelector) Name() string {
	return "fastest"
}

// Select en hızlı proxy'yi döner
func (f *FastestSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	var selected *LiveProxy
	minSpeed := int64(math.MaxInt64)
	
	for _, p := range proxies {
		if p.SpeedMs > 0 && p.SpeedMs < minSpeed {
			minSpeed = p.SpeedMs
			selected = p
		}
	}
	
	// Eğer hiç hızlı proxy yoksa ilkini seç
	if selected == nil && len(proxies) > 0 {
		selected = proxies[0]
	}
	
	return selected.ToProxyConfig()
}

// ==================== SUCCESS RATE SELECTOR ====================

// SuccessRateSelector en başarılı proxy'yi seçer
type SuccessRateSelector struct {
	*BaseSelector
}

// NewSuccessRateSelector yeni başarı oranı seçici oluşturur
func NewSuccessRateSelector() *SuccessRateSelector {
	return &SuccessRateSelector{
		BaseSelector: NewBaseSelector(),
	}
}

// Name seçici adını döner
func (sr *SuccessRateSelector) Name() string {
	return "success-rate"
}

// Select en başarılı proxy'yi döner
func (sr *SuccessRateSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	// Metrics toplayıcı yoksa rastgele seç
	if metrics == nil {
		return proxies[0].ToProxyConfig()
	}
	
	var selected *LiveProxy
	bestRate := -1.0
	
	for _, p := range proxies {
		m := metrics.GetMetrics(p.Key())
		rate := 1.0 // Yeni proxy'ler için varsayılan yüksek başarı oranı
		
		if m != nil && m.TotalRequests > 0 {
			rate = m.SuccessRate()
		}
		
		if rate > bestRate {
			bestRate = rate
			selected = p
		}
	}
	
	if selected == nil && len(proxies) > 0 {
		selected = proxies[0]
	}
	
	return selected.ToProxyConfig()
}

// ==================== GEO SELECTOR ====================

// GeoSelector ülke bazlı proxy seçer
type GeoSelector struct {
	*BaseSelector
	preferredCountries []string
	rng                *rand.Rand
}

// NewGeoSelector yeni geo seçici oluşturur
func NewGeoSelector(countries []string) *GeoSelector {
	return &GeoSelector{
		BaseSelector:       NewBaseSelector(),
		preferredCountries: countries,
		rng:                rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Name seçici adını döner
func (g *GeoSelector) Name() string {
	return "geo"
}

// SetPreferredCountries tercih edilen ülkeleri ayarlar
func (g *GeoSelector) SetPreferredCountries(countries []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.preferredCountries = countries
}

// Select tercih edilen ülkeden bir proxy döner
func (g *GeoSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	g.mu.RLock()
	countries := g.preferredCountries
	g.mu.RUnlock()
	
	// Eğer ülke belirtilmemişse rastgele seç
	if len(countries) == 0 {
		g.mu.Lock()
		idx := g.rng.Intn(len(proxies))
		g.mu.Unlock()
		return proxies[idx].ToProxyConfig()
	}
	
	// Tercih edilen ülkelerden uygun proxy'leri bul
	var candidates []*LiveProxy
	for _, p := range proxies {
		for _, c := range countries {
			if p.Country == c {
				candidates = append(candidates, p)
				break
			}
		}
	}
	
	// Eğer tercih edilen ülkeden proxy yoksa tüm havuzdan rastgele seç
	if len(candidates) == 0 {
		g.mu.Lock()
		idx := g.rng.Intn(len(proxies))
		g.mu.Unlock()
		return proxies[idx].ToProxyConfig()
	}
	
	// Adaylar arasından rastgele seç
	g.mu.Lock()
	idx := g.rng.Intn(len(candidates))
	g.mu.Unlock()
	
	return candidates[idx].ToProxyConfig()
}

// ==================== WEIGHTED SELECTOR ====================

// WeightedSelector tüm metrikleri birleştiren ağırlıklı seçim yapar
// Score = (success_rate * 0.4) + (speed_score * 0.3) + (recency * 0.2) + (random * 0.1)
type WeightedSelector struct {
	*BaseSelector
	rng *rand.Rand
}

// NewWeightedSelector yeni ağırlıklı seçici oluşturur
func NewWeightedSelector() *WeightedSelector {
	return &WeightedSelector{
		BaseSelector: NewBaseSelector(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Name seçici adını döner
func (w *WeightedSelector) Name() string {
	return "weighted"
}

// calculateScore proxy skorunu hesaplar
func (w *WeightedSelector) calculateScore(p *LiveProxy, m *ProxyMetrics, totalProxies int) float64 {
	// 1. Başarı oranı (0.4 ağırlık)
	successRate := 1.0
	if m != nil && m.TotalRequests > 0 {
		successRate = m.SuccessRate()
	}
	
	// 2. Hız skoru (0.3 ağırlık) - Düşük hız yüksek skor
	speedScore := 0.5 // Varsayılan orta hız
	if p.SpeedMs > 0 {
		// Hız skorunu normalize et (0-1 arası)
		// 50ms = 1.0, 5000ms = 0.0
		maxSpeed := float64(5000)
		speedScore = 1.0 - (float64(p.SpeedMs) / maxSpeed)
		if speedScore < 0 {
			speedScore = 0
		}
	}
	
	// 3. Güncellik skoru (0.2 ağırlık) - Son kullanılan daha düşük skor
	recencyScore := 1.0
	if m != nil && !m.LastUsed.IsZero() {
		// Ne kadar uzun süre kullanılmamışsa o kadar yüksek skor
		sinceLastUse := time.Since(m.LastUsed).Minutes()
		if sinceLastUse < 1 {
			recencyScore = 0.1
		} else if sinceLastUse < 5 {
			recencyScore = 0.3
		} else if sinceLastUse < 10 {
			recencyScore = 0.6
		} else if sinceLastUse < 30 {
			recencyScore = 0.8
		} else {
			recencyScore = 1.0
		}
	}
	
	// 4. Rastgele faktör (0.1 ağırlık) - Exploration
	w.mu.Lock()
	randomFactor := w.rng.Float64()
	w.mu.Unlock()
	
	// Ağırlıklı toplam
	score := (successRate * 0.4) + (speedScore * 0.3) + (recencyScore * 0.2) + (randomFactor * 0.1)
	
	return score
}

// Select ağırlıklı skora göre proxy seçer
func (w *WeightedSelector) Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig {
	if pool == nil || pool.Count() == 0 {
		return nil
	}
	
	proxies := pool.Snapshot()
	if len(proxies) == 0 {
		return nil
	}
	
	// Her proxy için skor hesapla
	type scoredProxy struct {
		proxy *LiveProxy
		score float64
	}
	
	scores := make([]scoredProxy, 0, len(proxies))
	for _, p := range proxies {
		var m *ProxyMetrics
		if metrics != nil {
			m = metrics.GetMetrics(p.Key())
		}
		score := w.calculateScore(p, m, len(proxies))
		scores = append(scores, scoredProxy{proxy: p, score: score})
	}
	
	// Toplam skoru hesapla
	var totalScore float64
	for _, s := range scores {
		totalScore += s.score
	}
	
	if totalScore == 0 {
		// Eğer tüm skorlar 0 ise rastgele seç
		w.mu.Lock()
		idx := w.rng.Intn(len(proxies))
		w.mu.Unlock()
		return proxies[idx].ToProxyConfig()
	}
	
	// Ağırlıklı rastgele seçim (roulette wheel selection)
	w.mu.Lock()
	target := w.rng.Float64() * totalScore
	w.mu.Unlock()
	
	var currentSum float64
	for _, s := range scores {
		currentSum += s.score
		if currentSum >= target {
			return s.proxy.ToProxyConfig()
		}
	}
	
	// Fallback: sonuncuyu döndür
	return scores[len(scores)-1].proxy.ToProxyConfig()
}

// ==================== SELECTOR FACTORY ====================

// SelectorType seçici tipi
type SelectorType string

const (
	SelectorRoundRobin   SelectorType = "round-robin"
	SelectorRandom       SelectorType = "random"
	SelectorLeastUsed    SelectorType = "least-used"
	SelectorFastest      SelectorType = "fastest"
	SelectorSuccessRate  SelectorType = "success-rate"
	SelectorGeo          SelectorType = "geo"
	SelectorWeighted     SelectorType = "weighted"
)

// NewSelector seçici tipine göre yeni seçici oluşturur
func NewSelector(selectorType SelectorType) Selector {
	switch selectorType {
	case SelectorRoundRobin:
		return NewRoundRobinSelector()
	case SelectorRandom:
		return NewRandomSelector()
	case SelectorLeastUsed:
		return NewLeastUsedSelector()
	case SelectorFastest:
		return NewFastestSelector()
	case SelectorSuccessRate:
		return NewSuccessRateSelector()
	case SelectorGeo:
		return NewGeoSelector(nil)
	case SelectorWeighted:
		return NewWeightedSelector()
	default:
		return NewRoundRobinSelector()
	}
}

// NewSelectorFromString string'den seçici oluşturur
func NewSelectorFromString(name string) Selector {
	switch name {
	case "round-robin":
		return NewRoundRobinSelector()
	case "random":
		return NewRandomSelector()
	case "least-used":
		return NewLeastUsedSelector()
	case "fastest":
		return NewFastestSelector()
	case "success-rate":
		return NewSuccessRateSelector()
	case "geo":
		return NewGeoSelector(nil)
	case "weighted":
		return NewWeightedSelector()
	default:
		return NewRoundRobinSelector()
	}
}

// ListSelectors mevcut seçici isimlerini listeler
func ListSelectors() []string {
	return []string{
		"round-robin",
		"random",
		"least-used",
		"fastest",
		"success-rate",
		"geo",
		"weighted",
	}
}

// ==================== MANAGER ====================

// SelectorManager proxy seçim yöneticisi
type SelectorManager struct {
	selector Selector
	pool     LivePoolAccessor
	metrics  *MetricsCollector
	mu       sync.RWMutex
}

// NewSelectorManager yeni seçim yöneticisi oluşturur
func NewSelectorManager(selector Selector, pool LivePoolAccessor) *SelectorManager {
	return &SelectorManager{
		selector: selector,
		pool:     pool,
		metrics:  NewMetricsCollector(),
	}
}

// GetProxy bir proxy seçer
func (sm *SelectorManager) GetProxy() *ProxyConfig {
	sm.mu.RLock()
	selector := sm.selector
	pool := sm.pool
	metrics := sm.metrics
	sm.mu.RUnlock()
	
	if selector == nil || pool == nil {
		return nil
	}
	
	return selector.Select(pool, metrics)
}

// RecordResult proxy kullanım sonucunu kaydeder
func (sm *SelectorManager) RecordResult(proxy *ProxyConfig, success bool, responseTime time.Duration, err error) {
	if proxy == nil {
		return
	}
	
	result := &Result{
		Success:      success,
		ResponseTime: responseTime,
		Error:        err,
		Timestamp:    time.Now(),
	}
	
	sm.mu.RLock()
	selector := sm.selector
	metrics := sm.metrics
	sm.mu.RUnlock()
	
	// Seçiciye bildir
	if selector != nil {
		selector.UpdateMetrics(proxy, result)
	}
	
	// Metrik toplayıcıya kaydet
	if metrics != nil {
		metrics.RecordResult(proxy.Key(), result)
	}
}

// SetSelector seçiciyi değiştirir
func (sm *SelectorManager) SetSelector(selector Selector) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selector = selector
}

// GetMetrics tüm metrikleri döner
func (sm *SelectorManager) GetMetrics() map[string]*ProxyMetrics {
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
		return ""
	}
	return sm.selector.Name()
}
