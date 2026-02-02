package analytics

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
)

// EcommerceItem e-ticaret ürünü
type EcommerceItem struct {
	ItemID    string
	ItemName  string
	Price     float64
	Quantity  int
	Category  string
}

// EcommerceEvent e-ticaret eventi
type EcommerceEvent struct {
	EventName     string
	TransactionID string
	Currency      string
	Value         float64
	Items         []EcommerceItem
}

// SendEcommerceEvent GA4'e e-ticaret eventi gönderir
func (m *Manager) SendEcommerceEvent(ctx context.Context, event EcommerceEvent) error {
	var itemsParts []string
	for _, it := range event.Items {
		itemsParts = append(itemsParts, fmt.Sprintf(`{
			item_id:'%s',item_name:'%s',price:%.2f,quantity:%d,item_category:'%s'
		}`,
			escapeJS(it.ItemID),
			escapeJS(it.ItemName),
			it.Price,
			it.Quantity,
			escapeJS(it.Category),
		))
	}
	itemsJSON := "[" + strings.Join(itemsParts, ",") + "]"

	script := fmt.Sprintf(`(function(){
		if(typeof gtag==='function'){
			gtag('event','%s',{
				transaction_id:'%s',currency:'%s',value:%.2f,items:%s
			});
		}
	})();`,
		escapeJS(event.EventName),
		escapeJS(event.TransactionID),
		escapeJS(event.Currency),
		event.Value,
		itemsJSON,
	)
	return chromedp.Evaluate(script, nil).Do(ctx)
}
