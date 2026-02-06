# ErosHit Distributed Mode

Master-Worker mimarisi ile dağıtık mod test sistemi.

## Hızlı Başlangıç

### 1. Master Başlat

```bash
cd eros-hitbot
go run cmd/eroshit/master.go -bind 0.0.0.0:8080 -secret my-secret-key
```

### 2. Worker(lar) Başlat

```bash
# Worker 1
go run cmd/eroshit/worker.go -master http://localhost:8080 -secret my-secret-key -concurrency 5

# Worker 2 (başka terminal)
go run cmd/eroshit/worker.go -master http://localhost:8080 -secret my-secret-key -concurrency 10
```

### 3. Task Gönder

```bash
# Master konsolundan:
submit https://example.com
batch https://example.com 10

# Veya HTTP API:
curl -X POST http://localhost:8080/api/v1/master/task/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-key" \
  -d '{"url": "https://example.com", "session_id": "test-1"}'
```

## API Endpoints

### Master Endpoints

| Endpoint | Method | Auth | Açıklama |
|----------|--------|------|----------|
| `/api/v1/master/status` | GET | No | Master durumu |
| `/api/v1/master/stats` | GET | Yes | İstatistikler |
| `/api/v1/master/workers` | GET | Yes | Worker listesi |
| `/api/v1/master/tasks` | GET | Yes | Task listesi |
| `/api/v1/master/task/submit` | POST | Yes | Task gönder |

### Worker Endpoints

| Endpoint | Method | Auth | Açıklama |
|----------|--------|------|----------|
| `/api/v1/worker/register` | POST | Yes | Kayıt ol |
| `/api/v1/worker/heartbeat` | POST | Yes | Heartbeat gönder |
| `/api/v1/worker/task/request` | POST | Yes | Task iste |
| `/api/v1/worker/task/complete` | POST | Yes | Task tamamla |
| `/api/v1/worker/task/fail` | POST | Yes | Task başarısız |

## Yapılandırma

### Master

```go
master := distributed.NewMaster(distributed.MasterConfig{
    BindAddr:          "0.0.0.0:8080",
    SecretKey:         "secure-token",
    MaxWorkers:        100,
    TaskTimeout:       5 * time.Minute,
    HeartbeatInterval: 10 * time.Second,
})
```

### Worker

```go
worker := distributed.NewWorker(distributed.WorkerConfig{
    MasterURL:      "http://master:8080",
    SecretKey:      "secure-token",
    MaxConcurrency: 10,
    Hostname:       "worker-1",
    Version:        "1.0.0",
}, taskProcessor)
```

## Tek Makinede Test

Aynı makinede test için farklı portlar kullanın:

```bash
# Terminal 1 - Master
go run cmd/eroshit/master.go -bind 0.0.0.0:8080

# Terminal 2 - Worker 1
go run cmd/eroshit/worker.go -master http://localhost:8080 -concurrency 5

# Terminal 3 - Worker 2
go run cmd/eroshit/worker.go -master http://localhost:8080 -concurrency 5

# Terminal 4 - Task gönder
curl -X POST http://localhost:8080/api/v1/master/task/submit \
  -H "Content-Type: application/json" \
  -d '{"url": "https://httpbin.org/get"}'
```

## Master Konsol Komutları

Master çalışırken interaktif komutlar:

- `help` - Yardım
- `status` / `stats` - İstatistikler
- `submit <url>` - Tekil task gönder
- `batch <url> <count>` - Çoklu task gönder
- `workers` - Worker listesi
- `tasks` - Task listesi
- `quit` / `exit` - Çıkış

## Task Yapısı

```go
type Task struct {
    ID        string
    URL       string
    Proxy     *ProxyConfig
    Profile   *BehavioralProfile
    SessionID string
    Status    TaskStatus
    WorkerID  string
    Result    *TaskResult
}
```

## Güvenlik

- `SecretKey` ile worker-master arası kimlik doğrulama
- HTTPS önerilir (production için reverse proxy kullanın)
- IP whitelist desteği eklenebilir
