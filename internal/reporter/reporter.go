package reporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"eroshit/pkg/i18n"
)

// PERFORMANCE FIX: Maximum records to prevent memory exhaustion
const maxRecords = 100000

// HitRecord tek bir istek kaydı
type HitRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	URL          string    `json:"url"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int64     `json:"response_time_ms"`
	UserAgent    string    `json:"user_agent"`
	Error        string    `json:"error,omitempty"`
}

// Metrics toplam performans metrikleri
type Metrics struct {
	TotalHits       int     `json:"total_hits"`
	SuccessHits     int     `json:"success_hits"`
	FailedHits      int     `json:"failed_hits"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	MinResponseTime int64   `json:"min_response_time_ms"`
	MaxResponseTime int64   `json:"max_response_time_ms"`
	StatusCodes     map[int]int `json:"status_codes"`
	StartTime       time.Time   `json:"start_time"`
	EndTime         time.Time   `json:"end_time"`
}

type Reporter struct {
	mu            sync.RWMutex
	records       []HitRecord
	metrics       Metrics
	outputDir     string
	format        string
	logChan       chan string
	domain        string
	locale        string // "tr" veya "en"
	closed        bool   // kanal kapatıldı mı
	recordsFlushed int   // PERFORMANCE: Track flushed records count
}

func New(outputDir, format string, domain string) *Reporter {
	return NewWithLocale(outputDir, format, domain, "tr")
}

func NewWithLocale(outputDir, format string, domain, locale string) *Reporter {
	if locale != "en" {
		locale = "tr"
	}
	r := &Reporter{
		records:   make([]HitRecord, 0, 10000),
		outputDir: outputDir,
		format:    format,
		logChan:   make(chan string, 100),
		domain:    domain,
		locale:    locale,
		closed:    false,
	}
	r.metrics.StatusCodes = make(map[int]int)
	r.metrics.StartTime = time.Now()
	return r
}

func (r *Reporter) Record(h HitRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// PERFORMANCE FIX: Prevent unbounded memory growth
	// When records exceed maxRecords, keep only the last half
	if len(r.records) >= maxRecords {
		// Keep the last 50% of records
		keepFrom := len(r.records) / 2
		r.recordsFlushed += keepFrom
		r.records = r.records[keepFrom:]
	}

	r.records = append(r.records, h)
	r.metrics.TotalHits++

	if h.Error != "" {
		r.metrics.FailedHits++
	} else {
		r.metrics.SuccessHits++
		r.metrics.StatusCodes[h.StatusCode]++

		if r.metrics.MinResponseTime == 0 || h.ResponseTime < r.metrics.MinResponseTime {
			r.metrics.MinResponseTime = h.ResponseTime
		}
		if h.ResponseTime > r.metrics.MaxResponseTime {
			r.metrics.MaxResponseTime = h.ResponseTime
		}

		total := float64(r.metrics.SuccessHits-1)*r.metrics.AvgResponseTime + float64(h.ResponseTime)
		r.metrics.AvgResponseTime = total / float64(r.metrics.SuccessHits)
	}
}

func (r *Reporter) Log(msg string) {
	fmt.Println(msg)
	r.mu.RLock()
	closed := r.closed
	r.mu.RUnlock()
	if closed {
		return
	}
	select {
	case r.logChan <- msg:
	default:
		// Kanal doluysa mesajı atla (blocking önleme)
	}
}

// LogT locale'e göre çevrilmiş mesaj loglar
func (r *Reporter) LogT(key string, args ...interface{}) {
	msg := i18n.T(r.locale, key, args...)
	r.Log(msg)
}

func (r *Reporter) Finalize() {
	r.mu.Lock()
	r.metrics.EndTime = time.Now()
	r.mu.Unlock()
}

// Close log kanalını kapatır ve kaynakları temizler
func (r *Reporter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.closed = true
		close(r.logChan)
	}
}

func (r *Reporter) Export() error {
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return err
	}

	ts := time.Now().Format("20060102_150405")

	if r.format == "csv" || r.format == "both" {
		path := filepath.Join(r.outputDir, fmt.Sprintf("eroshit_hits_%s.csv", ts))
		if err := r.exportCSV(path); err != nil {
			return fmt.Errorf("CSV export: %w", err)
		}
		r.LogT(i18n.MsgReportCSV, path)
	}

	if r.format == "json" || r.format == "both" {
		path := filepath.Join(r.outputDir, fmt.Sprintf("eroshit_report_%s.json", ts))
		if err := r.exportJSON(path); err != nil {
			return fmt.Errorf("JSON export: %w", err)
		}
		r.LogT(i18n.MsgReportJSON, path)
	}

	if r.format == "html" || r.format == "both" {
		r.mu.RLock()
		m := r.metrics
		recs := make([]HitRecord, len(r.records))
		copy(recs, r.records)
		r.mu.RUnlock()
		htmlPath := filepath.Join(r.outputDir, fmt.Sprintf("eroshit_report_%s.html", ts))
		if err := NewHTMLReporter(m, recs, r.domain).GenerateReport(htmlPath); err != nil {
			return fmt.Errorf("HTML export: %w", err)
		}
		r.LogT(i18n.MsgReportHTML, htmlPath)
	}

	return nil
}

func (r *Reporter) exportCSV(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"timestamp", "url", "status_code", "response_time_ms", "user_agent", "error"})

	r.mu.RLock()
	for _, rec := range r.records {
		errStr := rec.Error
		if errStr == "" {
			errStr = "-"
		}
		_ = w.Write([]string{
			rec.Timestamp.Format(time.RFC3339),
			rec.URL,
			fmt.Sprintf("%d", rec.StatusCode),
			fmt.Sprintf("%d", rec.ResponseTime),
			rec.UserAgent,
			errStr,
		})
	}
	r.mu.RUnlock()

	w.Flush()
	return w.Error()
}

func (r *Reporter) exportJSON(path string) error {
	r.mu.RLock()
	out := struct {
		Records []HitRecord `json:"records"`
		Metrics Metrics     `json:"metrics"`
	}{
		Records: make([]HitRecord, len(r.records)),
		Metrics: r.metrics,
	}
	copy(out.Records, r.records)
	r.mu.RUnlock()

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (r *Reporter) GetMetrics() Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.metrics
}

// LogChan log mesajları için kanal (SSE/broadcast için)
func (r *Reporter) LogChan() <-chan string {
	return r.logChan
}
