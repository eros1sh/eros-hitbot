package seo

import (
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

// Keyword anahtar kelime
type Keyword struct {
	Term         string
	LongTail     []string
	SearchEngine string
	Position     int
}

// KeywordManager anahtar kelime yöneticisi
type KeywordManager struct {
	Keywords []Keyword
	mu       sync.Mutex
	rng      *rand.Rand
}

// NewKeywordManager yeni keyword manager
func NewKeywordManager(keywords []Keyword) *KeywordManager {
	return &KeywordManager{
		Keywords: keywords,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetRandomKeyword rastgele keyword
func (km *KeywordManager) GetRandomKeyword() Keyword {
	km.mu.Lock()
	defer km.mu.Unlock()
	if len(km.Keywords) == 0 {
		return Keyword{Term: "default search"}
	}
	return km.Keywords[km.rng.Intn(len(km.Keywords))]
}

// GenerateSearchQuery gerçekçi arama sorgusu
func (k *Keyword) GenerateSearchQuery() string {
	if rand.Float64() < 0.7 {
		return k.Term
	}
	if len(k.LongTail) > 0 {
		return k.LongTail[rand.Intn(len(k.LongTail))]
	}
	return k.Term
}

// AddQuestionWords soru kelimeleri ekler
func (k *Keyword) AddQuestionWords() string {
	questions := []string{
		"nasıl", "ne", "nedir", "ne zaman", "nerede",
		"kim", "hangi", "kaç", "how", "what", "when", "where",
	}
	if rand.Float64() < 0.3 {
		q := questions[rand.Intn(len(questions))]
		return fmt.Sprintf("%s %s", q, k.Term)
	}
	return k.Term
}

// GetSearchEngineURL arama motoru URL'i
func (k *Keyword) GetSearchEngineURL() string {
	query := k.GenerateSearchQuery()
	encoded := url.QueryEscape(query)
	switch k.SearchEngine {
	case "bing":
		return fmt.Sprintf("https://www.bing.com/search?q=%s", encoded)
	case "yandex":
		return fmt.Sprintf("https://yandex.com/search/?text=%s", encoded)
	default:
		return fmt.Sprintf("https://www.google.com/search?q=%s", encoded)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
