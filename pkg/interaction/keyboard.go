package interaction

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// TypeHumanLike insan benzeri yazı
func TypeHumanLike(ctx context.Context, selector, text string) error {
	if err := chromedp.Run(ctx, chromedp.Focus(selector, chromedp.ByQuery)); err != nil {
		return err
	}
	for i, char := range text {
		if err := chromedp.Run(ctx, chromedp.SendKeys(selector, string(char), chromedp.ByQuery)); err != nil {
			return err
		}
		delay := calculateKeystrokeDelay(i, len(text))
		time.Sleep(delay)
	}
	return nil
}

func calculateKeystrokeDelay(index, totalLength int) time.Duration {
	baseDelay := 80 + randInt(120)
	variation := randInt(51) - 25 // -25 to 25
	ms := baseDelay + variation
	if ms < 50 {
		ms = 50
	}
	return time.Duration(ms) * time.Millisecond
}
