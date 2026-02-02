package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// Point 2D koordinat
type Point struct {
	X, Y int
}

// BehaviorConfig insan davranışı yapılandırması
type BehaviorConfig struct {
	MinPageDuration      time.Duration
	MaxPageDuration      time.Duration
	ScrollProbability    float64
	MouseMoveProbability float64
	ClickProbability     float64
	ReadingSpeed         int
}

// HumanBehavior insan benzeri davranış simülatörü
type HumanBehavior struct {
	config *BehaviorConfig
	rng    *rand.Rand
	mu     sync.Mutex
}

// NewHumanBehavior yeni davranış simülatörü oluşturur
func NewHumanBehavior(config *BehaviorConfig) *HumanBehavior {
	if config == nil {
		config = &BehaviorConfig{
			MinPageDuration:      5 * time.Second,
			MaxPageDuration:      45 * time.Second,
			ScrollProbability:    0.9,
			MouseMoveProbability: 0.7,
			ClickProbability:     0.2,
			ReadingSpeed:         250,
		}
	}
	return &HumanBehavior{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SimulatePageVisit sayfa ziyareti sırasında insan davranışı simüle eder
func (h *HumanBehavior) SimulatePageVisit(ctx context.Context, pageLength int) error {
	// Okuma süresi
	readTime := h.calculateReadingTime(pageLength)

	// Scroll (olasılıkla)
	if h.randFloat() < h.config.ScrollProbability {
		if err := h.randomScroll(ctx); err != nil {
			// Scroll hatası kritik değil, devam et
		}
	}

	// Mouse hareketleri (olasılıkla)
	if h.randFloat() < h.config.MouseMoveProbability {
		if err := h.randomMouseMovements(ctx); err != nil {
			// Devam et
		}
	}

	// Bekleme (okuma simülasyonu)
	select {
	case <-time.After(readTime):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Link tıklama (düşük olasılık)
	if h.randFloat() < h.config.ClickProbability {
		_ = h.maybeClickLink(ctx)
	}

	return nil
}

func (h *HumanBehavior) randFloat() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rng.Float64()
}

func (h *HumanBehavior) randIntn(n int) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if n <= 0 {
		return 0
	}
	return h.rng.Intn(n)
}

// randomScroll doğal scroll davranışı
func (h *HumanBehavior) randomScroll(ctx context.Context) error {
	var pageHeight int
	if err := chromedp.Evaluate(`document.documentElement.scrollHeight || document.body.scrollHeight || 0`, &pageHeight).Do(ctx); err != nil || pageHeight <= 100 {
		return nil
	}

	scrollCount := h.randIntn(5) + 3
	if scrollCount > 7 {
		scrollCount = 7
	}
	positions := h.generateScrollPositions(pageHeight, scrollCount)

	for _, pos := range positions {
		script := fmt.Sprintf(`window.scrollTo({top: %d, behavior: 'smooth'})`, pos)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return err
		}
		delay := 500 + h.randIntn(1500)
		select {
		case <-time.After(time.Duration(delay) * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (h *HumanBehavior) generateScrollPositions(maxHeight, count int) []int {
	if maxHeight <= 0 || count <= 0 {
		return nil
	}
	positions := make([]int, count)
	for i := 0; i < count; i++ {
		positions[i] = h.randIntn(maxHeight - 50)
		if positions[i] < 0 {
			positions[i] = 0
		}
	}
	return positions
}

// randomMouseMovements Bezier curve ile doğal mouse hareketi
func (h *HumanBehavior) randomMouseMovements(ctx context.Context) error {
	startX, startY := h.randIntn(800), h.randIntn(500)
	endX, endY := h.randIntn(800), h.randIntn(500)

	points := h.generateBezierCurve(startX, startY, endX, endY, 15)

	for _, p := range points {
		script := fmt.Sprintf(`
			(function(){
				var ev = new MouseEvent('mousemove', {clientX: %d, clientY: %d, bubbles: true});
				document.dispatchEvent(ev);
			})();
		`, p.X, p.Y)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return err
		}
		time.Sleep(8 * time.Millisecond)
	}

	return nil
}

// generateBezierCurve cubic Bezier curve noktaları
func (h *HumanBehavior) generateBezierCurve(x1, y1, x2, y2, steps int) []Point {
	cx1 := x1 + (x2-x1)/3 + h.randIntn(100) - 50
	cy1 := y1 + (y2-y1)/3 + h.randIntn(100) - 50
	cx2 := x1 + 2*(x2-x1)/3 + h.randIntn(100) - 50
	cy2 := y1 + 2*(y2-y1)/3 + h.randIntn(100) - 50

	points := make([]Point, steps)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		if steps == 1 {
			t = 1
		}
		points[i] = cubicBezier(t, x1, y1, cx1, cy1, cx2, cy2, x2, y2)
	}
	return points
}

func cubicBezier(t float64, x1, y1, cx1, cy1, cx2, cy2, x2, y2 int) Point {
	omt := 1 - t
	omt2 := omt * omt
	omt3 := omt2 * omt
	t2 := t * t
	t3 := t2 * t

	x := float64(x1)*omt3 + 3*float64(cx1)*omt2*t + 3*float64(cx2)*omt*t2 + float64(x2)*t3
	y := float64(y1)*omt3 + 3*float64(cy1)*omt2*t + 3*float64(cy2)*omt*t2 + float64(y2)*t3

	return Point{X: int(math.Round(x)), Y: int(math.Round(y))}
}

// calculateReadingTime sayfa uzunluğuna göre okuma süresi
func (h *HumanBehavior) calculateReadingTime(pageLength int) time.Duration {
	wordCount := pageLength / 5
	if wordCount <= 0 {
		wordCount = 50
	}

	speed := h.config.ReadingSpeed
	if speed <= 0 {
		speed = 250
	}
	seconds := (wordCount * 60) / speed

	if seconds < 5 {
		seconds = 5
	}
	if seconds > 45 {
		seconds = 45
	}

	return time.Duration(seconds) * time.Second
}

// maybeClickLink rastgele link tıklaması
func (h *HumanBehavior) maybeClickLink(ctx context.Context) error {
	var links []string
	if err := chromedp.Evaluate(`
		(function(){
			var a = Array.from(document.querySelectorAll('a[href]'));
			return a.map(function(el){ return el.href; }).filter(function(h){
				try {
					var u = new URL(h);
					return u.origin === window.location.origin;
				} catch(e){ return false; }
			});
		})()
	`, &links).Do(ctx); err != nil {
		return err
	}

	if len(links) == 0 {
		return nil
	}

	idx := h.randIntn(len(links))
	link := links[idx]
	linkJSON, _ := json.Marshal(link)

	script := fmt.Sprintf(`(function(){
		var target = %s;
		var as = document.querySelectorAll('a[href]');
		for(var i=0;i<as.length;i++){
			if(as[i].href === target){ as[i].click(); return true; }
		}
		return false;
	})();`, string(linkJSON))
	return chromedp.Evaluate(script, nil).Do(ctx)
}
