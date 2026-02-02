package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// HTMLReporter HTML rapor üretici
type HTMLReporter struct {
	metrics   *Metrics
	records   []HitRecord
	domain    string
	timestamp time.Time
}

// NewHTMLReporter yeni HTML rapor üretici
func NewHTMLReporter(metrics Metrics, records []HitRecord, domain string) *HTMLReporter {
	return &HTMLReporter{
		metrics:   &metrics,
		records:   records,
		domain:    domain,
		timestamp: time.Now(),
	}
}

// GenerateReport HTML raporu oluşturur
func (h *HTMLReporter) GenerateReport(filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	data := h.prepareTemplateData()
	tmpl, err := template.New("report").Parse(htmlReportTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func (h *HTMLReporter) prepareTemplateData() map[string]interface{} {
	m := h.metrics
	duration := m.EndTime.Sub(m.StartTime)
	successRate := 0.0
	if m.TotalHits > 0 {
		successRate = float64(m.SuccessHits) / float64(m.TotalHits) * 100
	}
	rpm := 0.0
	if duration.Minutes() > 0 {
		rpm = float64(m.TotalHits) / duration.Minutes()
	}

	uniquePages := make(map[string]bool)
	for _, r := range h.records {
		uniquePages[r.URL] = true
	}

	recent := h.records
	if len(recent) > 100 {
		recent = recent[len(recent)-100:]
	}

	type recView struct {
		TimeStr     string
		URL         string
		StatusCode  int
		ResponseTime int64
	}
	recentViews := make([]recView, len(recent))
	for i, r := range recent {
		recentViews[i] = recView{
			TimeStr:      r.Timestamp.Format("15:04:05"),
			URL:          r.URL,
			StatusCode:   r.StatusCode,
			ResponseTime: r.ResponseTime,
		}
	}

	timelineData := h.buildTimelineData()
	statusData := h.buildStatusData()
	responseData := h.buildResponseTimeData()

	return map[string]interface{}{
		"Timestamp":          h.timestamp.Format("2006-01-02 15:04:05"),
		"Domain":             h.domain,
		"TotalHits":          m.TotalHits,
		"SuccessRate":        fmt.Sprintf("%.1f", successRate),
		"AvgResponseTime":    fmt.Sprintf("%.0f", m.AvgResponseTime),
		"Duration":           formatDuration(duration),
		"RequestsPerMinute":  fmt.Sprintf("%.1f", rpm),
		"UniquePages":        len(uniquePages),
		"TimelineData":       timelineData,
		"StatusData":         statusData,
		"ResponseTimeData":   responseData,
		"RecentRequests":     recentViews,
	}
}

func (h *HTMLReporter) buildTimelineData() string {
	if len(h.records) == 0 {
		return "{}"
	}
	var ts []string
	var counts []int
	for i := 0; i < 20 && i < len(h.records); i++ {
		r := h.records[i]
		ts = append(ts, r.Timestamp.Format("15:04:05"))
		counts = append(counts, 1)
	}
	data := map[string]interface{}{
		"timestamps":   ts,
		"requests":     counts,
		"successRate":  []float64{100},
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func (h *HTMLReporter) buildStatusData() string {
	labels := []string{}
	values := []int{}
	for code, count := range h.metrics.StatusCodes {
		labels = append(labels, fmt.Sprintf("%d", code))
		values = append(values, count)
	}
	if len(labels) == 0 {
		labels = []string{"N/A"}
		values = []int{0}
	}
	data := map[string]interface{}{"labels": labels, "values": values}
	b, _ := json.Marshal(data)
	return string(b)
}

func (h *HTMLReporter) buildResponseTimeData() string {
	bins := []string{"0-200", "200-500", "500-1000", "1000-2000", "2000+"}
	counts := make([]int, len(bins))
	for _, r := range h.records {
		rt := int(r.ResponseTime)
		switch {
		case rt < 200:
			counts[0]++
		case rt < 500:
			counts[1]++
		case rt < 1000:
			counts[2]++
		case rt < 2000:
			counts[3]++
		default:
			counts[4]++
		}
	}
	data := map[string]interface{}{"bins": bins, "counts": counts}
	b, _ := json.Marshal(data)
	return string(b)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

const htmlReportTemplate = `<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Eros Hit Bot - Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: system-ui, sans-serif; background: #0f172a; color: #e2e8f0; padding: 24px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { text-align: center; padding: 32px; background: linear-gradient(135deg,#1e3a5f,#0f172a); border-radius: 12px; margin-bottom: 24px; }
        .header h1 { font-size: 2rem; margin-bottom: 8px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; margin-bottom: 24px; }
        .stat-card { background: #1e293b; padding: 20px; border-radius: 12px; }
        .stat-card .value { font-size: 2rem; font-weight: bold; color: #38bdf8; }
        .stat-card .label { font-size: 0.85rem; color: #94a3b8; margin-top: 4px; }
        .chart-box { background: #1e293b; padding: 20px; border-radius: 12px; margin-bottom: 24px; height: 300px; }
        .chart-box h2 { font-size: 1.1rem; margin-bottom: 16px; }
        table { width: 100%; border-collapse: collapse; background: #1e293b; border-radius: 12px; overflow: hidden; }
        th, td { padding: 12px; text-align: left; }
        th { background: #334155; }
        tr:nth-child(even) { background: #0f172a33; }
        .footer { text-align: center; padding: 16px; color: #64748b; font-size: 0.9rem; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Eros Hit Bot Report</h1>
            <p>Generated: {{.Timestamp}} | Target: {{.Domain}}</p>
        </div>
        <div class="stats">
            <div class="stat-card"><div class="value">{{.TotalHits}}</div><div class="label">Total Requests</div></div>
            <div class="stat-card"><div class="value">{{.SuccessRate}}%</div><div class="label">Success Rate</div></div>
            <div class="stat-card"><div class="value">{{.AvgResponseTime}}ms</div><div class="label">Avg Response</div></div>
            <div class="stat-card"><div class="value">{{.Duration}}</div><div class="label">Duration</div></div>
            <div class="stat-card"><div class="value">{{.RequestsPerMinute}}</div><div class="label">Req/Min</div></div>
            <div class="stat-card"><div class="value">{{.UniquePages}}</div><div class="label">Unique Pages</div></div>
        </div>
        <div class="chart-box">
            <h2>Status Code Distribution</h2>
            <canvas id="statusChart"></canvas>
        </div>
        <div class="chart-box">
            <h2>Response Time Distribution</h2>
            <canvas id="responseChart"></canvas>
        </div>
        <div style="margin-bottom: 24px;">
            <h2 style="margin-bottom: 12px;">Recent Requests</h2>
            <table>
                <thead><tr><th>Time</th><th>URL</th><th>Status</th><th>Response (ms)</th></tr></thead>
                <tbody>
                {{range .RecentRequests}}
                <tr><td>{{.TimeStr}}</td><td style="max-width:400px;overflow:hidden;text-overflow:ellipsis;">{{.URL}}</td><td>{{.StatusCode}}</td><td>{{.ResponseTime}}</td></tr>
                {{end}}
                </tbody>
            </table>
        </div>
        <div class="footer">Eros Hit Bot | Report generated automatically</div>
    </div>
    <script>
        const statusData = {{.StatusData}};
        new Chart(document.getElementById('statusChart'), {
            type: 'doughnut',
            data: { labels: statusData.labels, datasets: [{ data: statusData.values, backgroundColor: ['#22c55e','#eab308','#ef4444'] }] }
        });
        const respData = {{.ResponseTimeData}};
        new Chart(document.getElementById('responseChart'), {
            type: 'bar',
            data: { labels: respData.bins, datasets: [{ label: 'Count', data: respData.counts, backgroundColor: '#38bdf8' }] },
            options: { scales: { y: { beginAtZero: true } } }
        });
    </script>
</body>
</html>
`
