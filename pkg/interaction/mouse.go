package interaction

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// Point 2D koordinat
type Point struct {
	X, Y int
}

// MouseMovement insan benzeri mouse hareketi
type MouseMovement struct {
	StartX, StartY int
	EndX, EndY     int
	Duration       time.Duration
	Jitter         float64
}

var mu sync.Mutex
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func randInt(max int) int {
	mu.Lock()
	defer mu.Unlock()
	if max <= 0 {
		return 0
	}
	return rng.Intn(max)
}

func randFloat() float64 {
	mu.Lock()
	defer mu.Unlock()
	return rng.Float64()
}

// BezierCurve Bezier eğrisi ile smooth mouse yolu
func BezierCurve(start, end Point, duration time.Duration) []Point {
	steps := int(duration.Milliseconds() / 10)
	if steps < 2 {
		steps = 2
	}
	points := make([]Point, steps)

	cp1 := Point{
		X: start.X + randomOffset(end.X-start.X, 0.3),
		Y: start.Y + randomOffset(end.Y-start.Y, 0.3),
	}
	cp2 := Point{
		X: end.X + randomOffset(end.X-start.X, -0.3),
		Y: end.Y + randomOffset(end.Y-start.Y, -0.3),
	}

	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		points[i] = cubicBezier(start, cp1, cp2, end, t)
	}
	return points
}

func cubicBezier(p0, p1, p2, p3 Point, t float64) Point {
	mt := 1 - t
	mt2 := mt * mt
	mt3 := mt2 * mt
	t2 := t * t
	t3 := t2 * t
	return Point{
		X: int(mt3*float64(p0.X) + 3*mt2*t*float64(p1.X) + 3*mt*t2*float64(p2.X) + t3*float64(p3.X)),
		Y: int(mt3*float64(p0.Y) + 3*mt2*t*float64(p1.Y) + 3*mt*t2*float64(p2.Y) + t3*float64(p3.Y)),
	}
}

func randomOffset(diff int, factor float64) int {
	return int(float64(diff) * factor * (randFloat() - 0.5))
}

// MoveMouseToElement elemente insan benzeri mouse hareketi
func MoveMouseToElement(ctx context.Context, selector string) error {
	var rect struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}

	script := fmt.Sprintf(`(function(){
		var el = document.querySelector('%s');
		if (!el) return null;
		var r = el.getBoundingClientRect();
		return {x:r.left,y:r.top,width:r.width,height:r.height};
	})()`, selector)
	if err := chromedp.Evaluate(script, &rect).Do(ctx); err != nil {
		return err
	}
	if rect.Width == 0 && rect.Height == 0 {
		return fmt.Errorf("element not found: %s", selector)
	}

	targetX := int(rect.X + rect.Width*0.3 + rect.Width*0.4*randFloat())
	targetY := int(rect.Y + rect.Height*0.3 + rect.Height*0.4*randFloat())
	startX, startY := randInt(100), randInt(100)

	duration := time.Duration(300+randInt(500)) * time.Millisecond
	path := BezierCurve(
		Point{X: startX, Y: startY},
		Point{X: targetX, Y: targetY},
		duration,
	)

	for _, pt := range path {
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchMouseEvent(input.MouseMoved, float64(pt.X), float64(pt.Y)).
				Do(ctx)
		})); err != nil {
			return err
		}
		time.Sleep(8 * time.Millisecond)
	}
	return nil
}

// HumanClick insan benzeri tıklama
func HumanClick(ctx context.Context, selector string) error {
	if err := MoveMouseToElement(ctx, selector); err != nil {
		return err
	}
	time.Sleep(time.Duration(100+randInt(200)) * time.Millisecond)
	return chromedp.Run(ctx, chromedp.Click(selector, chromedp.ByQuery))
}

// RandomMouseMovement rastgele mouse hareketi
func RandomMouseMovement(ctx context.Context, count int) error {
	for i := 0; i < count; i++ {
		x := float64(randInt(800))
		y := float64(randInt(600))
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return input.DispatchMouseEvent(input.MouseMoved, x, y).Do(ctx)
		})); err != nil {
			return err
		}
		time.Sleep(time.Duration(500+randInt(1500)) * time.Millisecond)
	}
	return nil
}
