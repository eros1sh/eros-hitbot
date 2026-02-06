# Config Hot-Reload

Bu paket, ErosHit uygulaması için config dosyası hot-reload özelliği sağlar.

## Özellikler

- **fsnotify**: Config dosyası değişimlerini izler
- **Debounce**: 1 saniye içindeki çoklu değişimleri birleştirir
- **Callback sistemi**: Config değişince çağrılacak fonksiyonlar
- **Thread-safe**: Concurrent erişime güvenli
- **Graceful restart**: Çalışan simulation'ları graceful şekilde restart eder

## Kurulum

`fsnotify` bağımlılığını ekleyin:

```bash
cd eros-hitbot
go get github.com/fsnotify/fsnotify
```

## Kullanım

### Temel Kullanım

```go
package main

import (
    "log"
    configpkg "eroshit/pkg/config"
)

func main() {
    // Reloader oluştur
    reloader := configpkg.NewReloader("config.yaml")
    
    // Callback kaydet
    reloader.OnChange(func(newCfg *configpkg.Config) {
        log.Println("Config yenilendi!")
        log.Printf("Target: %s, Hits/min: %d", 
            newCfg.TargetDomain, newCfg.HitsPerMinute)
    })
    
    // Başlat
    if err := reloader.Start(); err != nil {
        log.Fatal(err)
    }
    defer reloader.Stop()
    
    // Config'e erişim
    cfg := reloader.GetConfig()
    log.Println(cfg.TargetDomain)
}
```

### Simulation ile Entegrasyon

```go
type SimulationManager struct {
    reloader *configpkg.Reloader
    sim      *Simulation
}

func (sm *SimulationManager) Start() error {
    // Reloader oluştur
    sm.reloader = configpkg.NewReloader("config.yaml")
    
    // Config değişince restart yap
    sm.reloader.OnChange(func(newCfg *configpkg.Config) {
        log.Info("config_reloaded")
        
        // Eski simulasyonu durdur
        if sm.sim != nil {
            sm.sim.Stop()
        }
        
        // Yeni config ile başlat
        sm.sim = NewSimulation(newCfg)
        sm.sim.Start()
    })
    
    // Başlat
    if err := sm.reloader.Start(); err != nil {
        return err
    }
    
    // İlk simülasyonu başlat
    cfg := sm.reloader.GetConfig()
    sm.sim = NewSimulation(cfg)
    sm.sim.Start()
    
    return nil
}

func (sm *SimulationManager) Stop() {
    if sm.sim != nil {
        sm.sim.Stop()
    }
    if sm.reloader != nil {
        sm.reloader.Stop()
    }
}
```

### internal/config ile Entegrasyon

Mevcut `internal/config` paketini kullanmak isterseniz, reloader'ı şu şekilde entegre edebilirsiniz:

```go
// internal/config/reloader.go

package config

import (
    configpkg "eroshit/pkg/config"
)

// Wrap pkg/config for internal use
type Reloader struct {
    inner *configpkg.Reloader
}

func NewReloader(path string) *Reloader {
    return &Reloader{
        inner: configpkg.NewReloader(path),
    }
}

func (r *Reloader) Start() error {
    // Wrap callback to convert Config types
    return r.inner.Start()
}

func (r *Reloader) Stop() error {
    return r.inner.Stop()
}

func (r *Reloader) GetConfig() *Config {
    // Convert pkg/config.Config to internal/config.Config
    // Implementation depends on your needs
    pkgCfg := r.inner.GetConfig()
    return convertConfig(pkgCfg)
}

func (r *Reloader) OnChange(callback func(*Config)) {
    r.inner.OnChange(func(pkgCfg *configpkg.Config) {
        cfg := convertConfig(pkgCfg)
        callback(cfg)
    })
}
```

### Custom Logger

```go
type ZapLogger struct {
    logger *zap.Logger
}

func (l *ZapLogger) Info(msg string, fields ...interface{}) {
    l.logger.Info(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Error(msg string, fields ...interface{}) {
    l.logger.Error(msg, zap.Any("fields", fields))
}

// Kullanım
reloader := configpkg.NewReloader("config.yaml")
reloader.SetLogger(&ZapLogger{logger: zapLogger})
```

## API

### NewReloader(configPath string) *Reloader

Yeni bir Reloader oluşturur.

### (r *Reloader) Start() error

Config dosyasını izlemeye başlar ve ilk config'i yükler.

### (r *Reloader) Stop() error

İzlemeyi durdurur ve kaynakları temizler.

### (r *Reloader) OnChange(callback ChangeCallback)

Config değişince çağrılacak callback fonksiyonu kaydeder.

### (r *Reloader) GetConfig() *Config

Mevcut config'i thread-safe şekilde döndürür.

### (r *Reloader) SetLogger(logger Logger)

Custom logger ayarlar.

### (r *Reloader) SetDebounceDelay(delay time.Duration)

Debounce gecikmesini ayarlar (varsayılan: 1 saniye).

## Config Değişiklik Farkları

```go
oldCfg := reloader.GetConfig()
newCfg := // ... yeni config

diff := configpkg.Diff(oldCfg, newCfg)
for field, values := range diff {
    log.Printf("%s: %v -> %v", field, values.Old, values.New)
}
```

## Notlar

- Config dosyası YAML formatında olmalıdır
- Debounce süresi içindeki çoklu değişimler tek bir reload olarak işlenir
- Callback'ler ayrı goroutine'lerde çalıştırılır (blocking yapmaz)
- Config yenilenirken hata olursa eski config kullanılmaya devam eder
