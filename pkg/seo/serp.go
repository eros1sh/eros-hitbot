package seo

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// SERPSimulator SERP etkileşim simülatörü
type SERPSimulator struct {
	TargetDomain string
}

// NewSERPSimulator yeni SERP simülatörü
func NewSERPSimulator(targetDomain string) *SERPSimulator {
	return &SERPSimulator{TargetDomain: targetDomain}
}

// SimulateOrganicClick organik tıklama simüle eder - arama sayfasına gidip hedefe tıklar
func (s *SERPSimulator) SimulateOrganicClick(ctx context.Context, keyword Keyword) (string, error) {
	searchURL := keyword.GetSearchEngineURL()
	if err := chromedp.Run(ctx, chromedp.Navigate(searchURL)); err != nil {
		return "", err
	}
	time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)

	if err := s.scrollSERP(ctx); err != nil {
		return "", err
	}

	targetURL, err := s.findTargetInResults(ctx)
	if err != nil {
		return "", err
	}

	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
	return targetURL, nil
}

func (s *SERPSimulator) scrollSERP(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		chromedp.Evaluate(`window.scrollBy(0, 200)`, nil).Do(ctx)
		time.Sleep(time.Duration(300+rand.Intn(500)) * time.Millisecond)
	}
	return nil
}

func (s *SERPSimulator) findTargetInResults(ctx context.Context) (string, error) {
	domain := strings.ReplaceAll(s.TargetDomain, ".", `\.`)
	var links []string
	script := fmt.Sprintf(`
		(function(){
			var as = document.querySelectorAll('div#search a[href], a[href]');
			for(var i=0;i<as.length;i++){
				if(as[i].href && as[i].href.indexOf('%s')>=0) return [as[i].href];
			}
			return [];
		})()
	`, s.TargetDomain)
	if err := chromedp.Evaluate(script, &links).Do(ctx); err != nil {
		return "", err
	}
	_ = domain
	if len(links) == 0 {
		return "", fmt.Errorf("target domain not found in results")
	}
	return links[0], nil
}
