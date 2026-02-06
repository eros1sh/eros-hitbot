# Smart Proxy Rotation System

Akıllı proxy rotasyon sistemi, 7 farklı strateji ile proxy seçimi yapmanızı sağlar.

## Stratejiler

| Strateji | Açıklama | Kullanım Senaryosu |
|----------|----------|-------------------|
| `round-robin` | Sırayla proxy seçer | Dengeleme, adil dağılım |
| `random` | Rastgele proxy seçer | Anonimlik, varyasyon |
| `least-used` | En az kullanılanı seçer | Yük dengeleme |
| `fastest` | En hızlı yanıt süreli | Performans odaklı |
| `success-rate` | En yüksek başarı oranlı | Güvenilirlik odaklı |
| `geo` | Belirli ülke(ler)den seçer | Coğrafi hedefleme |
| `weighted` | Tüm metrikleri birleştirir | Optimal seçim |

## Weighted Algoritma

```
Score = (success_rate * 0.4) + (speed_score * 0.3) + (recency * 0.2) + (random * 0.1)
```

- **success_rate (0.4)**: Başarı oranı ne kadar yüksekse o kadar iyi
- **speed_score (0.3)**: Hız ne kadar düşükse (hızlıysa) o kadar iyi
- **recency (0.2)**: Ne kadar uzun süre kullanılmamışsa o kadar iyi
- **random (0.1)**: Keşif faktörü, her zaman aynı proxy'yi seçmemek için

## Kurulum

### Config'den Kullanım (config.yaml)

```yaml
# Proxy rotasyon ayarları
proxy_rotation_mode: "weighted"  # round-robin, random, least-used, fastest, success-rate, geo, weighted
proxy_rotation_interval: 1       # Her istekte rotasyon
enable_proxy_rotation: true

# Geo seçici için tercih edilen ülkeler (opsiyonel)
geo_country: "US"  # veya birden fazla: "US,DE,GB"
```

### Kod'dan Kullanım

#### Basit Kullanım

```go
import (
    "eros-hitbot/pkg/proxy"
)

// Selector oluştur
selector := proxy.NewSelectorFromString("weighted")

// Pool ve metrics collector oluştur
pool := createYourPool() // LivePoolAccessor implementasyonu
metrics := proxy.NewMetricsCollector()

// Proxy seç
selectedProxy := selector.Select(pool, metrics)
```

#### Selector Manager ile Kullanım

```go
import (
    internalproxy "eros-hitbot/internal/proxy"
)

// Service oluştur ve selector'ü başlat
service := internalproxy.NewSelectorEnabledService("weighted", []string{"US", "DE"})

// Proxy al
proxy := service.GetProxyWithSelector()

// Kullanım sonucunu kaydet
success := true // veya false
responseTime := 150 * time.Millisecond
service.RecordProxyResult(proxy, success, responseTime)
```

#### Geo Selector ile Kullanım

```go
// Geo seçici oluştur
geoSelector := proxy.NewGeoSelector([]string{"US", "GB", "DE"})

// veya config'den
selector := proxy.NewSelectorFromString("geo")
// Daha sonra ülkeleri ayarla
if geoSel, ok := selector.(*proxy.GeoSelector); ok {
    geoSel.SetPreferredCountries([]string{"US", "GB"})
}
```

#### Manuel Metrik Takibi

```go
metrics := proxy.NewMetricsCollector()

// Sonuç kaydet
result := &proxy.Result{
    Success:      true,
    ResponseTime: 100 * time.Millisecond,
    Timestamp:    time.Now(),
}
metrics.RecordResult("proxy1:8080", result)

// Metrikleri al
allMetrics := metrics.GetAllMetrics()
for proxyKey, m := range allMetrics {
    fmt.Printf("%s: SuccessRate=%.2f, AvgResponseTime=%v\n", 
        proxyKey, m.SuccessRate(), m.AvgResponseTime)
}
```

## API

### Selector Interface

```go
type Selector interface {
    Select(pool LivePoolAccessor, metrics *MetricsCollector) *ProxyConfig
    UpdateMetrics(proxy *ProxyConfig, result *Result)
    Name() string
}
```

### Factory Fonksiyonları

```go
// Tip sabiti ile oluşturma
selector := proxy.NewSelector(proxy.SelectorWeighted)

// String ile oluşturma
selector := proxy.NewSelectorFromString("weighted")

// Doğrudan oluşturma
selector := proxy.NewWeightedSelector()
selector := proxy.NewRoundRobinSelector()
selector := proxy.NewRandomSelector()
selector := proxy.NewLeastUsedSelector()
selector := proxy.NewFastestSelector()
selector := proxy.NewSuccessRateSelector()
selector := proxy.NewGeoSelector([]string{"US", "DE"})
```

### Mevcut Seçicileri Listele

```go
selectors := proxy.ListSelectors()
// ["round-robin", "random", "least-used", "fastest", "success-rate", "geo", "weighted"]
```

## Internal/Proxy Entegrasyonu

`internal/proxy` paketi artık selector desteği içeriyor:

```go
import "eros-hitbot/internal/proxy"

// Yeni selector-enabled service oluştur
service := proxy.NewSelectorEnabledService("weighted", nil)

// veya mevcut service'e selector ekle
baseService := proxy.NewService()
selectorService := &proxy.SelectorEnabledService{Service: baseService}
selectorService.InitializeSelector("fastest", nil)

// Config'den oku
selectorType := cfg.ProxyRotationMode // "weighted"
service.InitializeSelector(selectorType, nil)
```

## Performans

- Tüm selector'ler thread-safe (goroutine-safe) tasarlanmıştır
- MetricsCollector atomik işlemler kullanır
- Weighted selector O(n) karmaşıklığına sahiptir
- Round-robin O(1) karmaşıklığına sahiptir

## Test

```bash
go test -v ./pkg/proxy/...
```
