package engagement

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"

	"eroshit/pkg/interaction"
)

// ClickStrategy tıklama stratejisi
type ClickStrategy struct {
	TargetTypes   []string
	MaxClicks     int
	AvoidExternal bool
}

// RandomClicks sayfada rastgele tıklamalar
func RandomClicks(ctx context.Context, strategy ClickStrategy) error {
	selector := buildSelector(strategy.TargetTypes)
	if selector == "" {
		selector = "a, button"
	}

	var elements []map[string]interface{}
	script := fmt.Sprintf(`
		(function(){
			var els = Array.from(document.querySelectorAll('%s')).filter(function(el){return el.offsetParent!==null;});
			return els.slice(0,%d).map(function(el){
				var sel = el.tagName.toLowerCase();
				if(el.id) sel += '#'+el.id;
				else if(el.className&&typeof el.className==='string') sel += '.'+el.className.split(' ')[0];
				return {selector:sel,href:el.href||''};
			});
		})()
	`, selector, strategy.MaxClicks*3)

	if err := chromedp.Evaluate(script, &elements).Do(ctx); err != nil {
		return err
	}

	clickCount := 0
	baseURL := getCurrentOrigin(ctx)

	for _, elem := range elements {
		if clickCount >= strategy.MaxClicks {
			break
		}
		href, _ := elem["href"].(string)
		sel, _ := elem["selector"].(string)
		if sel == "" {
			continue
		}
		if strategy.AvoidExternal && isExternalLink(href, baseURL) {
			continue
		}
		if err := interaction.HumanClick(ctx, sel); err != nil {
			continue
		}
		clickCount++
	}
	return nil
}

func buildSelector(types []string) string {
	if len(types) == 0 {
		return "a, button"
	}
	return strings.Join(types, ", ")
}

func isExternalLink(href, baseHost string) bool {
	if href == "" || baseHost == "" {
		return false
	}
	u, err := url.Parse(href)
	if err != nil {
		return true
	}
	if u.Host == "" {
		return false
	}
	return u.Host != baseHost && !strings.HasSuffix(u.Host, "."+baseHost)
}

func getCurrentOrigin(ctx context.Context) string {
	var host string
	chromedp.Evaluate(`window.location.host||''`, &host).Do(ctx)
	return host
}
