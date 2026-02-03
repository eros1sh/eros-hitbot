package proxy

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Service fetch + checker + live pool orkestrasyonu
type Service struct {
	LivePool   *LivePool
	mu         sync.Mutex
	queueCount int32 // fetcher'dan gelen toplam (checker'a giden)
	checking   int32 // 1 = checker çalışıyor
	done       int32 // checker'ın işlediği
	cancel     context.CancelFunc
}

// NewService yeni proxy servisi oluşturur
func NewService() *Service {
	return &Service{LivePool: NewLivePool()}
}

// Status proxy servis durumu (API için)
type Status struct {
	QueueCount   int   `json:"queue_count"`
	LiveCount    int   `json:"live_count"`
	Checking     bool  `json:"checking"`
	CheckedDone  int32 `json:"checked_done"`
	AddedTotal   int64 `json:"added_total"`
	RemovedTotal int64 `json:"removed_total"`
}

// Status anlık durum
func (s *Service) Status() Status {
	added, removed := s.LivePool.AddedRemoved()
	return Status{
		QueueCount:   int(atomic.LoadInt32(&s.queueCount)),
		LiveCount:    s.LivePool.Count(),
		Checking:     atomic.LoadInt32(&s.checking) == 1,
		CheckedDone:  atomic.LoadInt32(&s.done),
		AddedTotal:   added,
		RemovedTotal: removed,
	}
}

// FetchAndCheck kaynaklardan çeker, checker ile test eder; çalışanları LivePool'a ekler
func (s *Service) FetchAndCheck(ctx context.Context, sources []string, checkerWorkers int, onLog func(string)) {
	s.mu.Lock()
	if atomic.LoadInt32(&s.checking) == 1 {
		s.mu.Unlock()
		if onLog != nil {
			onLog("Checker zaten çalışıyor.")
		}
		return
	}
	atomic.StoreInt32(&s.checking, 1)
	atomic.StoreInt32(&s.done, 0)
	atomic.StoreInt32(&s.queueCount, 0)
	s.mu.Unlock()

	if onLog != nil {
		onLog("Proxy listeleri çekiliyor...")
	}
	fetcher := NewFetcher(sources)
	queue, err := fetcher.FetchAll(ctx)
	if err != nil {
		atomic.StoreInt32(&s.checking, 0)
		if onLog != nil {
			onLog("Fetch hatası: " + err.Error())
		}
		return
	}
	atomic.StoreInt32(&s.queueCount, int32(len(queue)))
	if onLog != nil {
		onLog("Toplam " + strconv.Itoa(len(queue)) + " proxy alındı. Test ediliyor (Checker Workers: " + strconv.Itoa(checkerWorkers) + ")...")
	}

	checker := NewChecker(checkerWorkers)
	liveChan := make(chan *LiveProxy, 256)
	go func() {
		checker.RunSlice(ctx, queue, liveChan)
		atomic.StoreInt32(&s.checking, 0)
	}()

	for lp := range liveChan {
		atomic.AddInt32(&s.done, 1)
		s.LivePool.Add(lp)
	}
	if onLog != nil {
		onLog("Proxy testi bitti. Canlı: " + strconv.Itoa(s.LivePool.Count()))
	}
}

// FetchAndCheckBackground arka planda fetch+check başlatır; cancel için context döner
func (s *Service) FetchAndCheckBackground(sources []string, checkerWorkers int, onLog func(string)) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()
	go func() {
		s.FetchAndCheck(ctx, sources, checkerWorkers, onLog)
		s.mu.Lock()
		s.cancel = nil
		s.mu.Unlock()
	}()
	return ctx
}

// StopCheck checker'ı durdurur
func (s *Service) StopCheck() {
	s.mu.Lock()
	c := s.cancel
	s.mu.Unlock()
	if c != nil {
		c()
	}
}

// FetchFromGitHubNoCheck GitHub repo URL'lerinden tüm .txt dosyalarını indirir,
// tek listede birleştirir ve canlılık testi yapmadan hepsini havuza ekler.
// Başarısız proxy'ler kullanım sırasında (simulator tarafında) havuzdan silinir.
func (s *Service) FetchFromGitHubNoCheck(ctx context.Context, repoURLs []string, onLog func(string)) (int, error) {
	if onLog != nil {
		onLog("GitHub repo'larından dosyalar indiriliyor (test yok)...")
	}
	client := &http.Client{
		Timeout: 45 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    20,
			IdleConnTimeout: 20 * time.Second,
		},
	}
	list, err := FetchFromGitHubRepos(ctx, repoURLs, client)
	if err != nil {
		if onLog != nil {
			onLog("GitHub fetch hatası: " + err.Error())
		}
		return 0, err
	}
	if len(list) == 0 {
		if onLog != nil {
			onLog("Hiç proxy satırı bulunamadı.")
		}
		return 0, nil
	}
	s.LivePool.Clear()
	for _, cfg := range list {
		s.LivePool.AddUnchecked(cfg)
	}
	if onLog != nil {
		onLog("Havuza " + strconv.Itoa(len(list)) + " proxy eklendi (test yok). Başarısızlar kullanımda silinecek.")
	}
	return len(list), nil
}

