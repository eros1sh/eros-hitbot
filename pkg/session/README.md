# Advanced Session Management

ErosHit için gelişmiş session yönetimi modülü. Cookie persistence, LocalStorage/SessionStorage/IndexedDB persistence, returning visitor simülasyonu ve canvas fingerprint desteği sunar.

## Özellikler

- **Cookie Persistence**: HTTP cookie'lerini disk'e kaydetme ve yükleme
- **LocalStorage Persistence**: Tarayıcı localStorage verilerini saklama
- **SessionStorage Persistence**: SessionStorage verilerini saklama
- **IndexedDB Persistence**: IndexedDB verilerini saklama
- **Returning Visitor Simülasyonu**: Belirli oranda mevcut session'ları yeniden kullanma
- **Session Fingerprint**: Canvas + WebGL + Audio fingerprint üretimi
- **TTL ve Yaşam Döngüsü**: Otomatik eski session temizliği
- **Şifreleme**: AES-GCM ile session verisi şifreleme (opsiyonel)
- **Export/Import**: Session'ları taşınabilir formatta dışa/içe aktarma

## Kurulum

Config dosyasına (`config/config.yaml`) aşağıdaki ayarları ekleyin:

```yaml
# Advanced Session Management
session_persistence: true              # Session persistence aktif
session_storage_path: "./sessions"     # Session kayıt dizini
session_encryption: true               # Session şifreleme aktif
session_encryption_key: ""             # Şifreleme anahtarı (boş = otomatik)
session_ttl_hours: 168                 # 7 gün TTL
session_indexeddb_persist: true        # IndexedDB persistence
session_canvas_fingerprint: true       # Canvas fingerprint kullan
returning_visitor_rate: 30             # Returning visitor oranı (%)
```

## Hızlı Başlangıç

```go
package main

import (
    "log"
    "time"
    "eroshit/pkg/session"
)

func main() {
    // Session manager oluştur
    cfg := session.SessionManagerConfig{
        StoragePath:          "./sessions",
        TTL:                  168 * time.Hour,  // 7 gün
        Encrypt:              true,
        EncryptionKey:        "my-secret-key",
        ReturningVisitorRate: 30,              // %30 returning visitor
    }

    sm, err := session.NewSessionManager(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer sm.Close()

    // Yeni session oluştur
    sess := sm.CreateSession()
    log.Printf("Yeni session: %s", sess.ID)
    log.Printf("Canvas fingerprint: %s", sess.CanvasFingerprint)

    // Cookie kaydet
    cookies := []*http.Cookie{
        {Name: "session_id", Value: "abc123", Domain: "example.com"},
    }
    sm.SaveCookies(sess.ID, cookies)

    // LocalStorage kaydet
    localStorage := map[string]string{
        "theme": "dark",
        "lang":  "tr",
    }
    sm.SaveLocalStorage(sess.ID, localStorage)
}
```

## API Referansı

### Session Yapısı

```go
type Session struct {
    ID                string            // Benzersiz session ID
    Cookies           []*http.Cookie    // HTTP cookie'leri
    LocalStorage      map[string]string // localStorage verisi
    SessionStorage    map[string]string // sessionStorage verisi
    IndexedDB         map[string]any    // IndexedDB verisi
    CanvasFingerprint string            // Canvas/WebGL fingerprint
    UserAgent         string            // User-Agent string
    ScreenResolution  string            // Ekran çözünürlüğü
    Timezone          string            // Zaman dilimi
    Language          string            // Dil
    CreatedAt         time.Time         // Oluşturulma zamanı
    LastUsedAt        time.Time         // Son kullanım zamanı
    VisitCount        int               // Ziyaret sayısı
    IsReturning       bool              // Returning visitor flag
}
```

### SessionManager Metodları

#### Session Oluşturma

```go
// Yeni session oluştur
session := sm.CreateSession()

// Returning visitor session oluştur
session := sm.CreateReturningSession()

// Mevcut session döndür veya yeni oluştur (%30 olasılıkla returning)
session := sm.GetOrCreateSession()
```

#### Session Yükleme

```go
// ID ile session getir
session := sm.GetSession(id)

// Disk'ten session yükle (TTL kontrolü ile)
session, err := sm.LoadSession(id)

// Rastgele mevcut session getir (returning visitor için)
session := sm.GetRandomExistingSession()
```

#### Veri Saklama

```go
// Cookie kaydet
sm.SaveCookies(sessionID, []*http.Cookie{...})
cookies, err := sm.GetCookies(sessionID)

// LocalStorage kaydet
sm.SaveLocalStorage(sessionID, map[string]string{"key": "value"})
data, err := sm.GetLocalStorage(sessionID)

// SessionStorage kaydet
sm.SaveSessionStorage(sessionID, map[string]string{"key": "value"})
data, err := sm.GetSessionStorage(sessionID)

// IndexedDB kaydet
sm.SaveIndexedDB(sessionID, map[string]any{"data": ...})
data, err := sm.GetIndexedDB(sessionID)
```

#### Session Yaşam Döngüsü

```go
// Session sil
err := sm.DeleteSession(id)

// Süresi dolmuş session'ları temizle
err := sm.CleanupExpired()

// Aktif session sayısı
count := sm.GetSessionCount()

// Tüm session'ları getir
sessions := sm.GetAllSessions()
```

#### İstatistikler

```go
stats := sm.GetStats()
fmt.Printf("Toplam: %d\n", stats.TotalSessions)
fmt.Printf("Returning: %d\n", stats.ReturningCount)
fmt.Printf("Yeni: %d\n", stats.NewCount)
fmt.Printf("Ort. Ziyaret: %.1f\n", stats.AvgVisitCount)
fmt.Printf("Ort. Yaş: %v\n", stats.AvgSessionAge)
```

#### Export/Import

```go
// Session'ı dışa aktar (base64 encoded)
encoded, err := sm.ExportSession(id)

// Session'ı içe aktar (yeni ID atanır)
imported, err := sm.ImportSession(encoded)
```

## JavaScript Entegrasyonu

Tarayıcıdan session verisi çıkarmak veya geri yüklemek için JavaScript helper fonksiyonları:

```go
// Canvas fingerprint al
fpScript := session.GetFingerprintJS()
// chromedp.Evaluate(fpScript, &fingerprint)

// localStorage çıkar
lsScript := session.GetLocalStorageJS()
// chromedp.Evaluate(lsScript, &localStorageData)

// localStorage geri yükle
restoreScript := session.SetLocalStorageJS(map[string]string{
    "theme": "dark",
})
// chromedp.Evaluate(restoreScript, nil)
```

## Güvenlik

### Şifreleme

Session verileri AES-256-GCM ile şifrelenebilir:

```go
cfg := session.SessionManagerConfig{
    Encrypt:       true,
    EncryptionKey: "gizli-anahtar-buraya", // Boş = otomatik üret
}
```

### TTL (Time To Live)

Session'lar otomatik olarak belirtilen süre sonra geçersiz olur:

```go
cfg := session.SessionManagerConfig{
    TTL: 168 * time.Hour, // 7 gün
}
```

## Chromedp Entegrasyonu

```go
import (
    "github.com/chromedp/chromedp"
)

func applySession(ctx context.Context, sm *session.SessionManager, sessionID string) error {
    // Session'ı yükle
    sess, err := sm.LoadSession(sessionID)
    if err != nil {
        return err
    }

    // Cookie'leri uygula
    for _, cookie := range sess.Cookies {
        // network.SetCookies kullan
    }

    // localStorage geri yükle
    if len(sess.LocalStorage) > 0 {
        script := session.SetLocalStorageJS(sess.LocalStorage)
        if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
            return err
        }
    }

    // SessionStorage geri yükle
    if len(sess.SessionStorage) > 0 {
        script := session.SetSessionStorageJS(sess.SessionStorage)
        if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
            return err
        }
    }

    // Canvas fingerprint uygula
    cf := canvas.GenerateFingerprint()
    if err := cf.InjectCanvasNoise(ctx); err != nil {
        return err
    }

    return nil
}
```

## Test

```bash
cd eros-hitbot
go test ./pkg/session/... -v
```

## Lisans

Bu modül ErosHit projesi kapsamında geliştirilmiştir.
