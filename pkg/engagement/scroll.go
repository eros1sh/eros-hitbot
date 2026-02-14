package engagement

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// ScrollBehavior scroll davranışı
type ScrollBehavior struct {
	Strategy    string // "gradual", "fast", "reader"
	PausePoints []int  // 0-100% pozisyonlar
	ReadSpeed   int    // kelime/dakika
}

var engRng = rand.New(rand.NewSource(time.Now().UnixNano()))
var engMu sync.Mutex

func engRandInt(max int) int {
	engMu.Lock()
	defer engMu.Unlock()
	if max <= 0 {
		return 0
	}
	return engRng.Intn(max)
}

// HumanScroll insan benzeri scroll
func HumanScroll(ctx context.Context, behavior ScrollBehavior) error {
	var pageHeight int
	if err := chromedp.Evaluate(`document.documentElement.scrollHeight || document.body.scrollHeight || 0`, &pageHeight).Do(ctx); err != nil || pageHeight <= 100 {
		return nil
	}

	switch behavior.Strategy {
	case "reader":
		return scrollAsReader(ctx, pageHeight, behavior.ReadSpeed)
	case "fast":
		return scrollFast(ctx, pageHeight)
	default:
		return scrollGradual(ctx, pageHeight, behavior.PausePoints)
	}
}

func scrollAsReader(ctx context.Context, pageHeight, wpm int) error {
	if wpm <= 0 {
		wpm = 200
	}
	var wordCount int
	chromedp.Evaluate(`(document.body&&document.body.innerText)?document.body.innerText.split(/\s+/).length:100`, &wordCount).Do(ctx)
	if wordCount <= 0 {
		wordCount = 100
	}

	readingTime := time.Duration(wordCount*60/wpm) * time.Second
	if readingTime > 60*time.Second {
		readingTime = 60 * time.Second
	}
	scrollSteps := int(readingTime.Seconds()/2) + 1
	if scrollSteps > 20 {
		scrollSteps = 20
	}

	for i := 0; i < scrollSteps; i++ {
		scrollPercent := float64(i) / float64(scrollSteps)
		scrollTo := int(float64(pageHeight) * scrollPercent * 0.95)
		chromedp.Evaluate(fmt.Sprintf(`window.scrollTo({top:%d,behavior:'smooth'})`, scrollTo), nil).Do(ctx)
		time.Sleep(time.Duration(500+engRandInt(500)) * time.Millisecond)
	}
	return nil
}

func scrollGradual(ctx context.Context, pageHeight int, pausePoints []int) error {
	if len(pausePoints) == 0 {
		pausePoints = []int{25, 50, 75}
	}
	// Batch scroll: single JS call per pause point with smooth behavior, then Go-side sleep
	for _, pct := range pausePoints {
		targetScroll := pageHeight * pct / 100
		chromedp.Evaluate(fmt.Sprintf(`window.scrollTo({top:%d,behavior:'smooth'})`, targetScroll), nil).Do(ctx)
		// Smooth scroll animation time (~300ms) + reading pause
		time.Sleep(time.Duration(400+engRandInt(600)) * time.Millisecond)
	}
	return nil
}

func scrollFast(ctx context.Context, pageHeight int) error {
	steps := 5 + engRandInt(10)
	scrollAmount := pageHeight / steps
	if scrollAmount <= 0 {
		scrollAmount = 200
	}
	for i := 0; i < steps; i++ {
		chromedp.Evaluate(fmt.Sprintf(`window.scrollBy(0,%d)`, scrollAmount), nil).Do(ctx)
		time.Sleep(time.Duration(200+engRandInt(300)) * time.Millisecond)
	}
	return nil
}
