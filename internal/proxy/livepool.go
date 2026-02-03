package proxy

import (
	"sync"
	"sync/atomic"
	"time"
)

// LivePool sadece çalışan proxy'leri tutar; başarısız olanlar silinir
type LivePool struct {
	mu      sync.RWMutex
	list    []*LiveProxy
	next    uint32 // round-robin
	added   int64  // toplam eklenen (checker'dan veya unchecked)
	removed int64  // başarısız diye silinen
}

// NewLivePool boş canlı havuz oluşturur
func NewLivePool() *LivePool {
	return &LivePool{list: make([]*LiveProxy, 0, 256)}
}

// Clear havuzu temizler (GitHub vb. ile yeniden doldurmadan önce)
func (p *LivePool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.list = p.list[:0]
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
func (p *LivePool) Add(live *LiveProxy) {
	if live == nil || live.ProxyConfig == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	key := live.Key()
	for _, existing := range p.list {
		if existing.Key() == key {
			return
		}
	}
	p.list = append(p.list, live)
	atomic.AddInt64(&p.added, 1)
}

// Remove proxy'yi havuzdan kaldırır (başarısız kullanım sonrası)
func (p *LivePool) Remove(proxy *ProxyConfig) {
	if proxy == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	key := proxy.Key()
	for i, lp := range p.list {
		if lp.Key() == key {
			p.list = append(p.list[:i], p.list[i+1:]...)
			atomic.AddInt64(&p.removed, 1)
			return
		}
	}
}

// GetNext round-robin sıradaki proxy'yi döner (hitter için)
func (p *LivePool) GetNext() *ProxyConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	n := len(p.list)
	if n == 0 {
		return nil
	}
	idx := atomic.AddUint32(&p.next, 1) % uint32(n)
	if idx >= uint32(len(p.list)) {
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
