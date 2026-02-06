package proxy

import (
	"sync"
	"sync/atomic"
	"time"
)

// LivePool sadece çalışan proxy'leri tutar; başarısız olanlar silinir
// PERFORMANCE FIX: Map eklendi O(1) lookup için
type LivePool struct {
	mu      sync.RWMutex
	list    []*LiveProxy
	index   map[string]int // PERFORMANCE: key -> list index mapping for O(1) lookup
	next    uint32         // round-robin
	added   int64          // toplam eklenen (checker'dan veya unchecked)
	removed int64          // başarısız diye silinen
}

// NewLivePool boş canlı havuz oluşturur
func NewLivePool() *LivePool {
	return &LivePool{
		list:  make([]*LiveProxy, 0, 256),
		index: make(map[string]int, 256), // PERFORMANCE: Pre-allocate map
	}
}

// Clear havuzu temizler (GitHub vb. ile yeniden doldurmadan önce)
func (p *LivePool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.list = p.list[:0]
	// PERFORMANCE FIX: Map'i de temizle
	p.index = make(map[string]int, 256)
	atomic.StoreUint32(&p.next, 0)
}

// AddUnchecked test edilmemiş proxy'yi havuza ekler; kullanımda başarısız olursa Remove ile silinir
func (p *LivePool) AddUnchecked(cfg *ProxyConfig) {
	if cfg == nil {
		return
	}
	lp := &LiveProxy{
		ProxyConfig: cfg,
		Country:     "",
		SpeedMs:     0,
		CheckedAt:   time.Now(),
	}
	p.Add(lp)
}

// Add çalışan proxy'yi havuza ekler
// PERFORMANCE FIX: O(1) lookup için map kullan
func (p *LivePool) Add(live *LiveProxy) {
	if live == nil || live.ProxyConfig == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	key := live.Key()
	// PERFORMANCE FIX: O(1) map lookup instead of O(n) slice iteration
	if _, exists := p.index[key]; exists {
		return
	}
	p.index[key] = len(p.list)
	p.list = append(p.list, live)
	atomic.AddInt64(&p.added, 1)
}

// Remove proxy'yi havuzdan kaldırır (başarısız kullanım sonrası)
// PERFORMANCE FIX: O(1) lookup için map kullan
func (p *LivePool) Remove(proxy *ProxyConfig) {
	if proxy == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	key := proxy.Key()
	// PERFORMANCE FIX: O(1) map lookup
	idx, exists := p.index[key]
	if !exists {
		return
	}
	// Remove from list (swap with last element for O(1) removal)
	lastIdx := len(p.list) - 1
	if idx != lastIdx {
		// Swap with last element
		p.list[idx] = p.list[lastIdx]
		// Update index for swapped element
		p.index[p.list[idx].Key()] = idx
	}
	p.list = p.list[:lastIdx]
	delete(p.index, key)
	atomic.AddInt64(&p.removed, 1)
}

// GetNext round-robin sıradaki proxy'yi döner (hitter için)
// SECURITY FIX: Race condition düzeltildi - Write lock kullanılıyor
// çünkü atomic.AddUint32 ile list erişimi arasında tutarlılık gerekli
func (p *LivePool) GetNext() *ProxyConfig {
	p.mu.Lock() // SECURITY FIX: RLock yerine Lock kullan - atomic op + slice access
	defer p.mu.Unlock()
	n := len(p.list)
	if n == 0 {
		return nil
	}
	// SECURITY FIX: Modulo işlemini lock içinde yap, list değişemez
	idx := int(p.next) % n
	p.next++
	// Bounds check - defensive programming
	if idx < 0 || idx >= len(p.list) {
		idx = 0
	}
	return p.list[idx].ProxyConfig
}

// Snapshot canlı proxy listesinin kopyasını döner
func (p *LivePool) Snapshot() []*LiveProxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*LiveProxy, len(p.list))
	copy(out, p.list)
	return out
}

// Count havuzdaki canlı proxy sayısı
func (p *LivePool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.list)
}

// AddedRemoved toplam eklenen ve kaldırılan sayıları döner
func (p *LivePool) AddedRemoved() (added, removed int64) {
	return atomic.LoadInt64(&p.added), atomic.LoadInt64(&p.removed)
}

// ExportTxt canlı listeyi http://host:port satırları olarak döner
func (p *LivePool) ExportTxt() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.list) == 0 {
		return nil
	}
	var buf []byte
	for _, lp := range p.list {
		buf = append(buf, lp.ProxyConfig.ToURLString()...)
		buf = append(buf, '\n')
	}
	return buf
}

// LiveProxyWithCountry harita için ülke bilgili kayıt
type LiveProxyWithCountry struct {
	Proxy   string `json:"proxy"`
	Country string `json:"country"`
	SpeedMs int64  `json:"speed_ms"`
}

// SnapshotForAPI API için ülke/hız bilgili liste
func (p *LivePool) SnapshotForAPI() []LiveProxyWithCountry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]LiveProxyWithCountry, len(p.list))
	for i, lp := range p.list {
		out[i] = LiveProxyWithCountry{
			Proxy:   lp.Key(),
			Country: lp.Country,
			SpeedMs: lp.SpeedMs,
		}
	}
	return out
}
