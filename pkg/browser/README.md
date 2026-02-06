# Browser Pool - Yüksek Performanslı Chrome Instance Yönetimi

Bu paket, ErosHit projesi için yüksek performanslı bir Browser Pool implementasyonu sağlar. Object Pool pattern kullanarak Chrome instance'larını yeniden kullanır.

## 🎯 Problem ve Çözüm

### Mevcut Sorun
`internal/browser/hit.go` dosyasında her ziyarette yeni Chrome instance'ı (`chromedp.NewContext`) oluşturuluyor:
- Chrome başlatma süresi: ~2-5 saniye
- Yüksek bellek kullanımı
- CPU yoğun işlem
- Ölçeklenebilirlik sorunları

### Çözüm: Browser Pool
- Instance'ları önceden oluşturur ve havuzda tutar
- Her ziyarette mevcut instance'ı yeniden kullanır
- Otomatik cookie/cache temizleme
- Thread-safe kanal bazlı yönetim

## 📊 Performans Karşılaştırması

| Metrik | Eski (Her Ziyaret) | Yeni (Pool) | İyileştirme |
|--------|-------------------|-------------|-------------|
| İlk başlatma | ~3 sn | ~3 sn | - |
| Sonraki ziyaretler | ~3 sn | ~0.1 sn | **30x** |
| Bellek/instance | ~150MB | ~150MB | - |
| Max paralellik | Sınırsız (kaynak tükenmesi) | 10-20 (kontrollü) | **Daha stabil** |

## 🚀 Hızlı Başlangıç

### Temel Kullanım

```go
package main

import (
    "context"
    "log"
    "time"
    
    "eroshit/pkg/browser"
    "github.com/chromedp/chromedp"
)

func main() {
    // Pool oluştur
    config := browser.DefaultPoolConfig()
    config.MaxInstances = 10
    config.MinInstances = 2
    
    pool, err := browser.NewBrowserPool(config)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    
    // Instance al
    ctx := context.Background()
    instance, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Kullan
    tabCtx := instance.GetContext()
    err = chromedp.Run(tabCtx,
        chromedp.Navigate("https://example.com"),
        chromedp.WaitReady("body"),
    )
    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Geri ver (otomatik temizlenir)
    pool.Release(instance)
}
```

### PooledHitVisitor ile Entegre Kullanım

```go
// Eski yöntem (her ziyaret yeni Chrome)
// visitor := NewHitVisitor(...)

// Yeni yöntem (pool bazlı)
config := browser.PooledHitVisitorConfig{
    PoolConfig: browser.PoolConfig{
        MaxInstances: 10,
        MinInstances: 2,
        AcquireTimeout: 30 * time.Second,
    },
}

visitor, err := browser.NewPooledHitVisitor(config)
if err != nil {
    log.Fatal(err)
}
defer visitor.Close()

// Ziyaret yap
err = visitor.VisitURL(ctx, browser.VisitOptions{
    URL: "https://example.com",
    UserAgent: "Mozilla/5.0...",
    CanvasFingerprint: true,
    ScrollStrategy: "gradual",
})
```

## ⚙️ Konfigürasyon

### PoolConfig

```go
type PoolConfig struct {
    MaxInstances        int           // Max instance sayısı (default: 10)
    MinInstances        int           // Başlangıç instance sayısı (default: 2)
    AcquireTimeout      time.Duration // Acquire timeout (default: 30s)
    InstanceMaxAge      time.Duration // Instance max yaşam süresi (default: 30m)
    InstanceMaxSessions int32         // Instance başına max oturum (default: 50)
    ProxyURL            string        // Opsiyonel proxy
    ProxyUser           string        // Proxy kullanıcı adı
    ProxyPass           string        // Proxy şifre
    Headless            bool          // Headless mod (default: true)
}
```

## 📈 Metrikler ve İzleme

```go
metrics := pool.GetMetrics()

fmt.Printf("Oluşturulan: %d\n", metrics.TotalCreated)
fmt.Printf("Yok edilen: %d\n", metrics.TotalDestroyed)
fmt.Printf("Yeniden kullanılan: %d\n", metrics.TotalReused)
fmt.Printf("Aktif: %d\n", metrics.CurrentActive)
fmt.Printf("Boşta: %d\n", metrics.CurrentIdle)
fmt.Printf("Bekleme sayısı: %d\n", metrics.AcquireWaits)
```

## 🔒 Thread Safety

Tüm pool operasyonları thread-safe'dir:
- `Acquire()` - Bloklayıcı ama thread-safe
- `Release()` - Non-blocking, thread-safe
- `Reset()` - Instance bazlı, thread-safe

## 🧹 Temizlik ve Reset

### Otomatik Reset (Release'de)
```go
pool.Release(instance) // Cookie ve cache otomatik temizlenir
```

### Manuel Deep Reset
```go
// LocalStorage, IndexedDB, Service Workers dahil derin temizlik
pool.ForceReset(instance)
```

## 🔄 Instance Lifecycle

```
Create → Available → Acquire → In-Use → Release → Reset → Available
                              ↓
                         Recycle (age/sessions exceeded)
                              ↓
                         Destroy → Create New
```

## 🏗️ Mimari

```
┌─────────────────────────────────────────────┐
│           BrowserPool                       │
│  ┌─────────────────────────────────────┐   │
│  │  available (chan)                   │   │
│  │  [Instance1] [Instance2] [...]      │   │
│  └─────────────────────────────────────┘   │
│  ┌─────────────────────────────────────┐   │
│  │  instances (map)                    │   │
│  │  {"id1": Instance1, ...}            │   │
│  └─────────────────────────────────────┘   │
│  ┌─────────────────────────────────────┐   │
│  │  metrics (atomic)                   │   │
│  │  TotalCreated, CurrentActive, ...   │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────┐
│       BrowserInstance                       │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
│  │allocCtx │→ │tabCtx   │→ |chromedp|     │
│  │         │  │         │  │         │     │
│  │Chrome   │  │Tab      │  │Actions  │     │
│  │Process  │  │Context  │  │         │     │
│  └─────────┘  └─────────┘  └─────────┘     │
└─────────────────────────────────────────────┘
```

## 🎛️ Gelişmiş Özellikler

### Çoklu Sekme (Multi-Tab)
```go
instance, _ := pool.Acquire(ctx)
defer pool.Release(instance)

// İlk sekme
tab1 := instance.GetContext()
chromedp.Run(tab1, chromedp.Navigate("https://site1.com"))

// Yeni sekme aynı browser'da
tab2, cancel, _ := instance.CreateNewTab(ctx)
defer cancel()
chromedp.Run(tab2, chromedp.Navigate("https://site2.com"))
```

### Proxy ile Kullanım
```go
config := browser.PoolConfig{
    MaxInstances: 10,
    ProxyURL:     "http://proxy.example.com:8080",
    ProxyUser:    "user",
    ProxyPass:    "pass",
}
```

### Timeout Yönetimi
```go
// Acquire timeout
config.AcquireTimeout = 10 * time.Second

// Ziyaret timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
instance, err := pool.Acquire(ctx)
```

## 🧪 Test

```bash
cd eros-hitbot
go test -v ./pkg/browser/...
```

## 📋 Dosya Yapısı

```
pkg/browser/
├── pool.go              # Ana pool implementasyonu
├── pool_visitor.go      # HitVisitor entegrasyonu
├── pool_example_test.go # Kullanım örnekleri
└── README.md            # Bu dosya
```

## 🔧 Migration Guide (hit.go'dan)

### Eski Kod
```go
func (h *HitVisitor) VisitURL(ctx context.Context, urlStr string) error {
    tabCtx, tabCancel := chromedp.NewContext(h.allocCtx)
    defer tabCancel()
    // ... ziyaret işlemleri
}
```

### Yeni Kod
```go
func (v *PooledHitVisitor) VisitURL(ctx context.Context, opts VisitOptions) error {
    instance, err := v.pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer v.pool.Release(instance) // Otomatik reset
    
    tabCtx := instance.GetContext()
    // ... ziyaret işlemleri
}
```

## ⚠️ Bilinen Sınırlamalar

1. **Instance Lifecycle**: Max 30 dakika veya 50 oturum sonra otomatik yenilenir
2. **Memory**: Her instance ~150MB bellek kullanır
3. **Proxy**: Tüm instance'lar aynı proxy'yi kullanır (pool bazlı)

## 📚 API Referansı

### BrowserPool
- `NewBrowserPool(config PoolConfig) (*BrowserPool, error)`
- `Acquire(ctx context.Context) (*BrowserInstance, error)`
- `Release(instance *BrowserInstance)`
- `Reset(instance *BrowserInstance) error`
- `ForceReset(instance *BrowserInstance) error`
- `GetMetrics() PoolMetrics`
- `Close() error`

### BrowserInstance
- `GetContext() context.Context`
- `GetAllocatorContext() context.Context`
- `CreateNewTab(ctx context.Context) (context.Context, context.CancelFunc, error)`
- `IsInUse() bool`
- `IsHealthy() bool`
- `GetID() string`
- `GetAge() time.Duration`
- `NeedsRecycle(maxAge time.Duration, maxSessions int32) bool`

### PooledHitVisitor
- `NewPooledHitVisitor(config PooledHitVisitorConfig) (*PooledHitVisitor, error)`
- `VisitURL(ctx context.Context, opts VisitOptions) error`
- `GetMetrics() PoolMetrics`
- `Close() error`
