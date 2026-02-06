// Package session provides advanced session management with persistence capabilities.
// Supports cookie, localStorage, sessionStorage and IndexedDB persistence with encryption.
package session

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"eroshit/pkg/canvas"
)

// Session represents a browser session with all persistence data
type Session struct {
	ID                string            `json:"id"`
	Cookies           []*http.Cookie    `json:"cookies"`
	LocalStorage      map[string]string `json:"local_storage"`
	SessionStorage    map[string]string `json:"session_storage"`
	IndexedDB         map[string]any    `json:"indexed_db"`
	CanvasFingerprint string            `json:"canvas_fingerprint"`
	UserAgent         string            `json:"user_agent"`
	ScreenResolution  string            `json:"screen_resolution"`
	Timezone          string            `json:"timezone"`
	Language          string            `json:"language"`
	CreatedAt         time.Time         `json:"created_at"`
	LastUsedAt        time.Time         `json:"last_used_at"`
	VisitCount        int               `json:"visit_count"`
	IsReturning       bool              `json:"is_returning"`
}

// IsExpired checks if session has exceeded TTL
func (s *Session) IsExpired(ttl time.Duration) bool {
	return time.Since(s.LastUsedAt) > ttl
}

// UpdateLastUsed updates the last used timestamp and increments visit count
func (s *Session) UpdateLastUsed() {
	s.LastUsedAt = time.Now()
	s.VisitCount++
}

// GenerateFingerprint creates a unique fingerprint for the session
func (s *Session) GenerateFingerprint() {
	cf := canvas.GenerateFingerprint()
	s.CanvasFingerprint = fmt.Sprintf("%s|%s|%f", cf.WebGLVendor, cf.WebGLRenderer, cf.CanvasNoise)
}

// SessionStore defines the interface for session persistence
type SessionStore interface {
	Save(session *Session) error
	Load(id string) (*Session, error)
	Delete(id string) error
	List() ([]string, error)
	Close() error
}

// FileStore implements SessionStore with file-based persistence
type FileStore struct {
	basePath  string
	mu        sync.RWMutex
	encrypt   bool
	secretKey []byte
}

// NewFileStore creates a new file-based session store
func NewFileStore(basePath string, encrypt bool, secretKey string) (*FileStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	store := &FileStore{
		basePath: basePath,
		encrypt:  encrypt,
	}

	if encrypt {
		if secretKey == "" {
			// Generate a random key if not provided
			store.secretKey = generateKey()
		} else {
			// Derive key from provided secret
			hash := sha256.Sum256([]byte(secretKey))
			store.secretKey = hash[:]
		}
	}

	return store, nil
}

// Save persists a session to disk
func (fs *FileStore) Save(session *Session) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if fs.encrypt {
		data, err = fs.encryptData(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt session: %w", err)
		}
	}

	filename := filepath.Join(fs.basePath, session.ID+".json")
	if fs.encrypt {
		filename = filepath.Join(fs.basePath, session.ID+".enc")
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load retrieves a session from disk
func (fs *FileStore) Load(id string) (*Session, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	filename := filepath.Join(fs.basePath, id+".json")
	if fs.encrypt {
		filename = filepath.Join(fs.basePath, id+".enc")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	if fs.encrypt {
		data, err = fs.decryptData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt session: %w", err)
		}
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Delete removes a session from disk
func (fs *FileStore) Delete(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	filename := filepath.Join(fs.basePath, id+".json")
	if fs.encrypt {
		filename = filepath.Join(fs.basePath, id+".enc")
	}

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// List returns all session IDs in the store
func (fs *FileStore) List() ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	var ids []string
	ext := ".json"
	if fs.encrypt {
		ext = ".enc"
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ext {
			id := entry.Name()[:len(entry.Name())-len(ext)]
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// Close implements SessionStore
func (fs *FileStore) Close() error {
	return nil
}

func (fs *FileStore) encryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(fs.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(cryptorand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (fs *FileStore) decryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(fs.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// SessionManager manages browser sessions with persistence
type SessionManager struct {
	store              SessionStore
	ttl                time.Duration
	mu                 sync.RWMutex
	sessions           map[string]*Session
	returningVisitorRate int
	encrypt            bool
}

// SessionManagerConfig configuration for session manager
type SessionManagerConfig struct {
	StoragePath          string
	TTL                  time.Duration
	Encrypt              bool
	EncryptionKey        string
	ReturningVisitorRate int // 0-100 percentage
}

// NewSessionManager creates a new session manager
func NewSessionManager(cfg SessionManagerConfig) (*SessionManager, error) {
	if cfg.TTL <= 0 {
		cfg.TTL = 168 * time.Hour // 7 days default
	}
	if cfg.ReturningVisitorRate < 0 || cfg.ReturningVisitorRate > 100 {
		cfg.ReturningVisitorRate = 30 // 30% default
	}

	store, err := NewFileStore(cfg.StoragePath, cfg.Encrypt, cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}

	sm := &SessionManager{
		store:                store,
		ttl:                  cfg.TTL,
		sessions:             make(map[string]*Session),
		returningVisitorRate: cfg.ReturningVisitorRate,
		encrypt:              cfg.Encrypt,
	}

	// Load existing sessions from disk
	if err := sm.loadAllSessions(); err != nil {
		return nil, fmt.Errorf("failed to load existing sessions: %w", err)
	}

	return sm, nil
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession() *Session {
	session := &Session{
		ID:             generateSessionID(),
		Cookies:        make([]*http.Cookie, 0),
		LocalStorage:   make(map[string]string),
		SessionStorage: make(map[string]string),
		IndexedDB:      make(map[string]any),
		CreatedAt:      time.Now(),
		LastUsedAt:     time.Now(),
		VisitCount:     0,
		IsReturning:    false,
	}

	session.GenerateFingerprint()

	sm.mu.Lock()
	sm.sessions[session.ID] = session
	sm.mu.Unlock()

	// Persist to disk
	_ = sm.store.Save(session)

	return session
}

// CreateReturningSession creates a session marked as returning visitor
func (sm *SessionManager) CreateReturningSession() *Session {
	session := sm.CreateSession()
	session.IsReturning = true
	session.VisitCount = 1 + rand.IntN(5) // Previous visits
	_ = sm.store.Save(session)
	return session
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session := sm.sessions[id]
	if session != nil {
		session.UpdateLastUsed()
	}

	return session
}

// LoadSession loads a session from persistent store
func (sm *SessionManager) LoadSession(id string) (*Session, error) {
	// First check in-memory cache
	sm.mu.RLock()
	session, exists := sm.sessions[id]
	sm.mu.RUnlock()

	if exists {
		return session, nil
	}

	// Load from disk
	session, err := sm.store.Load(id)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, nil
	}

	// Check TTL
	if session.IsExpired(sm.ttl) {
		_ = sm.store.Delete(id)
		return nil, nil
	}

	// Add to cache
	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()

	return session, nil
}

// GetRandomExistingSession returns a random existing session for returning visitor simulation
func (sm *SessionManager) GetRandomExistingSession() *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.sessions) == 0 {
		return nil
	}

	// Filter non-expired sessions
	validSessions := make([]*Session, 0)
	for _, s := range sm.sessions {
		if !s.IsExpired(sm.ttl) {
			validSessions = append(validSessions, s)
		}
	}

	if len(validSessions) == 0 {
		return nil
	}

	return validSessions[rand.IntN(len(validSessions))]
}

// GetOrCreateSession returns an existing session or creates a new one
// based on returningVisitorRate probability
func (sm *SessionManager) GetOrCreateSession() *Session {
	// Check if we should use a returning visitor
	if rand.IntN(100) < sm.returningVisitorRate {
		if session := sm.GetRandomExistingSession(); session != nil {
			session.IsReturning = true
			session.UpdateLastUsed()
			_ = sm.store.Save(session)
			return session
		}
	}

	return sm.CreateSession()
}

// SaveCookies saves cookies for a session
func (sm *SessionManager) SaveCookies(sessionID string, cookies []*http.Cookie) error {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Cookies = cookies
	session.UpdateLastUsed()

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return sm.store.Save(session)
}

// GetCookies retrieves cookies for a session
func (sm *SessionManager) GetCookies(sessionID string) ([]*http.Cookie, error) {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.Cookies, nil
}

// SaveLocalStorage saves localStorage data for a session
func (sm *SessionManager) SaveLocalStorage(sessionID string, data map[string]string) error {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.LocalStorage = data
	session.UpdateLastUsed()

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return sm.store.Save(session)
}

// GetLocalStorage retrieves localStorage data for a session
func (sm *SessionManager) GetLocalStorage(sessionID string) (map[string]string, error) {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.LocalStorage, nil
}

// SaveSessionStorage saves sessionStorage data for a session
func (sm *SessionManager) SaveSessionStorage(sessionID string, data map[string]string) error {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.SessionStorage = data
	session.UpdateLastUsed()

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return sm.store.Save(session)
}

// GetSessionStorage retrieves sessionStorage data for a session
func (sm *SessionManager) GetSessionStorage(sessionID string) (map[string]string, error) {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.SessionStorage, nil
}

// SaveIndexedDB saves IndexedDB data for a session
func (sm *SessionManager) SaveIndexedDB(sessionID string, data map[string]any) error {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.IndexedDB = data
	session.UpdateLastUsed()

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return sm.store.Save(session)
}

// GetIndexedDB retrieves IndexedDB data for a session
func (sm *SessionManager) GetIndexedDB(sessionID string) (map[string]any, error) {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.IndexedDB, nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(id string) error {
	sm.mu.Lock()
	delete(sm.sessions, id)
	sm.mu.Unlock()

	return sm.store.Delete(id)
}

// CleanupExpired removes all expired sessions
func (sm *SessionManager) CleanupExpired() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var expired []string
	for id, session := range sm.sessions {
		if session.IsExpired(sm.ttl) {
			expired = append(expired, id)
		}
	}

	for _, id := range expired {
		delete(sm.sessions, id)
		_ = sm.store.Delete(id)
	}

	return nil
}

// GetSessionCount returns the number of active sessions
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, session := range sm.sessions {
		if !session.IsExpired(sm.ttl) {
			count++
		}
	}

	return count
}

// GetAllSessions returns all non-expired sessions
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, session := range sm.sessions {
		if !session.IsExpired(sm.ttl) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// Close cleans up resources
func (sm *SessionManager) Close() error {
	return sm.store.Close()
}

// loadAllSessions loads all sessions from disk into memory
func (sm *SessionManager) loadAllSessions() error {
	ids, err := sm.store.List()
	if err != nil {
		return err
	}

	for _, id := range ids {
		session, err := sm.store.Load(id)
		if err != nil {
			continue // Skip corrupted sessions
		}
		if session != nil && !session.IsExpired(sm.ttl) {
			sm.sessions[id] = session
		} else if session != nil {
			// Clean up expired session
			_ = sm.store.Delete(id)
		}
	}

	return nil
}

// generateSessionID creates a cryptographically secure session ID
func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = cryptorand.Read(b)
	return fmt.Sprintf("sess_%x", b)
}

// generateKey creates a random encryption key
func generateKey() []byte {
	key := make([]byte, 32)
	_, _ = cryptorand.Read(key)
	return key
}

// GetFingerprintJS returns JavaScript to generate canvas fingerprint
func GetFingerprintJS() string {
	return `(function() {
		var canvas = document.createElement('canvas');
		var ctx = canvas.getContext('2d');
		canvas.width = 200;
		canvas.height = 50;
		
		// Text with gradient
		var gradient = ctx.createLinearGradient(0, 0, canvas.width, 0);
		gradient.addColorStop(0, 'rgb(255, 0, 0)');
		gradient.addColorStop(0.5, 'rgb(0, 255, 0)');
		gradient.addColorStop(1, 'rgb(0, 0, 255)');
		ctx.fillStyle = gradient;
		ctx.font = '20px Arial';
		ctx.fillText('ErosHit FP v2.0', 10, 35);
		
		// Geometric shapes
		ctx.strokeStyle = 'rgb(128, 128, 128)';
		ctx.beginPath();
		ctx.arc(150, 25, 15, 0, 2 * Math.PI);
		ctx.stroke();
		
		return canvas.toDataURL();
	})();`
}

// GetLocalStorageJS returns JavaScript to extract localStorage
func GetLocalStorageJS() string {
	return `(function() {
		var data = {};
		for (var i = 0; i < localStorage.length; i++) {
			var key = localStorage.key(i);
			data[key] = localStorage.getItem(key);
		}
		return JSON.stringify(data);
	})();`
}

// GetSessionStorageJS returns JavaScript to extract sessionStorage
func GetSessionStorageJS() string {
	return `(function() {
		var data = {};
		for (var i = 0; i < sessionStorage.length; i++) {
			var key = sessionStorage.key(i);
			data[key] = sessionStorage.getItem(key);
		}
		return JSON.stringify(data);
	})();`
}

// GetIndexedDBJS returns JavaScript to extract IndexedDB (basic implementation)
func GetIndexedDBJS() string {
	return `(async function() {
		var data = {};
		try {
			var databases = await indexedDB.databases();
			for (var db of databases) {
				data[db.name] = { version: db.version };
			}
		} catch(e) {}
		return JSON.stringify(data);
	})();`
}

// SetLocalStorageJS returns JavaScript to restore localStorage
func SetLocalStorageJS(data map[string]string) string {
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf(`(function() {
		var data = %s;
		for (var key in data) {
			try { localStorage.setItem(key, data[key]); } catch(e) {}
		}
	})();`, string(jsonData))
}

// SetSessionStorageJS returns JavaScript to restore sessionStorage
func SetSessionStorageJS(data map[string]string) string {
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf(`(function() {
		var data = %s;
		for (var key in data) {
			try { sessionStorage.setItem(key, data[key]); } catch(e) {}
		}
	})();`, string(jsonData))
}

// SessionStats provides statistics about sessions
type SessionStats struct {
	TotalSessions    int            `json:"total_sessions"`
	ReturningCount   int            `json:"returning_count"`
	NewCount         int            `json:"new_count"`
	AvgVisitCount    float64        `json:"avg_visit_count"`
	AvgSessionAge    time.Duration  `json:"avg_session_age"`
}

// GetStats returns statistics about all sessions
func (sm *SessionManager) GetStats() *SessionStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := &SessionStats{}
	if len(sm.sessions) == 0 {
		return stats
	}

	var totalVisits int
	var totalAge time.Duration
	now := time.Now()

	for _, session := range sm.sessions {
		if session.IsExpired(sm.ttl) {
			continue
		}

		stats.TotalSessions++
		totalVisits += session.VisitCount
		totalAge += now.Sub(session.CreatedAt)

		if session.IsReturning {
			stats.ReturningCount++
		} else {
			stats.NewCount++
		}
	}

	if stats.TotalSessions > 0 {
		stats.AvgVisitCount = float64(totalVisits) / float64(stats.TotalSessions)
		stats.AvgSessionAge = totalAge / time.Duration(stats.TotalSessions)
	}

	return stats
}

// ExportSession exports a session to a portable format (base64 encoded JSON)
func (sm *SessionManager) ExportSession(sessionID string) (string, error) {
	session, err := sm.LoadSession(sessionID)
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// ImportSession imports a session from a portable format
func (sm *SessionManager) ImportSession(encoded string) (*Session, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	// Generate new ID to avoid conflicts
	session.ID = generateSessionID()
	session.LastUsedAt = time.Now()

	sm.mu.Lock()
	sm.sessions[session.ID] = &session
	sm.mu.Unlock()

	_ = sm.store.Save(&session)

	return &session, nil
}
