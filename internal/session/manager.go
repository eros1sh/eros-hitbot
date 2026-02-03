package session

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"eroshit/pkg/fingerprint"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Session kullanıcı oturumu
type Session struct {
	ID             string
	Cookies        []CookieStore
	LocalStorage   map[string]string
	SessionStorage map[string]string
	CreatedAt      time.Time
	LastUsed       time.Time
	PageHistory    []string
	Fingerprint    *fingerprint.AdvancedFingerprint
}

// CookieStore basit cookie saklama
type CookieStore struct {
	Name   string
	Value  string
	Domain string
	Path   string
}

// SessionStats oturum istatistikleri
type SessionStats struct {
	TotalSessions   int
	ActiveSessions  int
	AvgCookieCount  float64
	AvgPageHistory  float64
}

// SessionManager oturum yöneticisi
type SessionManager struct {
	sessions       map[string]*Session
	mu             sync.RWMutex
	maxSessions    int
	newSessionRate float64
}

// NewSessionManager yeni oturum yöneticisi
func NewSessionManager(maxSessions int, newSessionRate float64) *SessionManager {
	if maxSessions <= 0 {
		maxSessions = 100
	}
	return &SessionManager{
		sessions:       make(map[string]*Session),
		maxSessions:    maxSessions,
		newSessionRate: newSessionRate,
	}
}

// GetOrCreateSession oturum al veya oluştur
func (s *SessionManager) GetOrCreateSession() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	useExisting := rand.Float64() > s.newSessionRate

	if useExisting && len(s.sessions) > 0 {
		return s.getRandomSessionUnlocked()
	}

	session := &Session{
		ID:             generateSessionID(),
		Cookies:        make([]CookieStore, 0),
		LocalStorage:   make(map[string]string),
		SessionStorage: make(map[string]string),
		CreatedAt:      time.Now(),
		LastUsed:       time.Now(),
		PageHistory:    make([]string, 0),
		Fingerprint:    fingerprint.GenerateAdvancedFingerprint(),
	}

	s.sessions[session.ID] = session

	if len(s.sessions) > s.maxSessions {
		s.removeOldestSessionUnlocked()
	}

	return session
}

// SaveSession oturumu kaydet
func (s *SessionManager) SaveSession(session *Session) {
	if session == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

// ApplySession Chromedp context'e oturum uygula
func (s *SessionManager) ApplySession(ctx context.Context, session *Session) error {
	if session == nil {
		return nil
	}

	if len(session.Cookies) > 0 {
		params := make([]*network.CookieParam, 0, len(session.Cookies))
		for _, c := range session.Cookies {
			params = append(params, &network.CookieParam{Name: c.Name, Value: c.Value})
		}
		if err := network.SetCookies(params).Do(ctx); err != nil {
			return err
		}
	}

	if len(session.LocalStorage) > 0 {
		data, _ := json.Marshal(session.LocalStorage)
		script := fmt.Sprintf(`
			(function(){
				var d = %s;
				for(var k in d){ try{ localStorage.setItem(k, d[k]); }catch(e){} }
			})();
		`, string(data))
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return err
		}
	}

	if len(session.SessionStorage) > 0 {
		data, _ := json.Marshal(session.SessionStorage)
		script := fmt.Sprintf(`
			(function(){
				var d = %s;
				for(var k in d){ try{ sessionStorage.setItem(k, d[k]); }catch(e){} }
			})();
		`, string(data))
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return err
		}
	}

	if session.Fingerprint != nil {
		script := session.Fingerprint.ToChromedpScript()
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return err
		}
	}

	return nil
}

// ExtractSession context'ten oturum bilgisini çıkar
func (s *SessionManager) ExtractSession(ctx context.Context, session *Session) error {
	if session == nil {
		return nil
	}

	cookies, err := network.GetCookies().Do(ctx)
	if err == nil && len(cookies) > 0 {
		session.Cookies = make([]CookieStore, 0, len(cookies))
		for _, c := range cookies {
			session.Cookies = append(session.Cookies, CookieStore{
				Name:   c.Name,
				Value:  c.Value,
				Domain: c.Domain,
				Path:   c.Path,
			})
		}
	}

	session.LastUsed = time.Now()
	return nil
}

// CleanOldSessions eski oturumları temizle
func (s *SessionManager) CleanOldSessions(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, sess := range s.sessions {
		if now.Sub(sess.LastUsed) > maxAge {
			delete(s.sessions, id)
		}
	}
}

// GetStats istatistikleri döner
func (s *SessionManager) GetStats() *SessionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SessionStats{TotalSessions: len(s.sessions)}
	cutoff := time.Now().Add(-10 * time.Minute)
	totalCookies, totalPages := 0, 0

	for _, sess := range s.sessions {
		if sess.LastUsed.After(cutoff) {
			stats.ActiveSessions++
		}
		totalCookies += len(sess.Cookies)
		totalPages += len(sess.PageHistory)
	}

	if len(s.sessions) > 0 {
		stats.AvgCookieCount = float64(totalCookies) / float64(len(s.sessions))
		stats.AvgPageHistory = float64(totalPages) / float64(len(s.sessions))
	}
	return stats
}

func (s *SessionManager) getRandomSessionUnlocked() *Session {
	if len(s.sessions) == 0 {
		return nil
	}
	keys := make([]string, 0, len(s.sessions))
	for k := range s.sessions {
		keys = append(keys, k)
	}
	return s.sessions[keys[rand.Intn(len(keys))]]
}

func (s *SessionManager) removeOldestSessionUnlocked() {
	var oldestID string
	var oldestTime time.Time

	for id, sess := range s.sessions {
		if oldestID == "" || sess.LastUsed.Before(oldestTime) {
			oldestID = id
			oldestTime = sess.LastUsed
		}
	}
	if oldestID != "" {
		delete(s.sessions, oldestID)
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = cryptorand.Read(b)
	return fmt.Sprintf("%x", b)
}
