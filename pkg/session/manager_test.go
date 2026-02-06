package session

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewSessionManager(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath:          tmpDir,
		TTL:                  1 * time.Hour,
		Encrypt:              false,
		ReturningVisitorRate: 30,
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	if sm == nil {
		t.Fatal("Session manager is nil")
	}

	if sm.ttl != 1*time.Hour {
		t.Errorf("Expected TTL 1h, got %v", sm.ttl)
	}

	if sm.returningVisitorRate != 30 {
		t.Errorf("Expected returning visitor rate 30, got %d", sm.returningVisitorRate)
	}
}

func TestCreateSession(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	session := sm.CreateSession()
	if session == nil {
		t.Fatal("Session is nil")
	}

	if session.ID == "" {
		t.Error("Session ID is empty")
	}

	if !strings.HasPrefix(session.ID, "sess_") {
		t.Errorf("Session ID should start with 'sess_', got: %s", session.ID)
	}

	if session.CanvasFingerprint == "" {
		t.Error("Canvas fingerprint is empty")
	}

	if session.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}

	if session.LastUsedAt.IsZero() {
		t.Error("LastUsedAt is zero")
	}

	if session.VisitCount != 0 {
		t.Errorf("Expected VisitCount 0, got %d", session.VisitCount)
	}
}

func TestSessionPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm1, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Create a session
	session := sm1.CreateSession()
	sessionID := session.ID

	// Add some data
	session.UserAgent = "Mozilla/5.0 Test"
	session.ScreenResolution = "1920x1080"
	session.Timezone = "Europe/Istanbul"
	session.Language = "tr-TR"

	// Save cookies
	cookies := []*http.Cookie{
		{Name: "test_cookie", Value: "test_value", Domain: "example.com"},
		{Name: "session_id", Value: "abc123", Domain: "example.com"},
	}
	if err := sm1.SaveCookies(sessionID, cookies); err != nil {
		t.Fatalf("Failed to save cookies: %v", err)
	}

	// Save localStorage
	localStorage := map[string]string{
		"user_pref":    "dark_mode",
		"last_visit":   time.Now().Format(time.RFC3339),
		"visitor_type": "premium",
	}
	if err := sm1.SaveLocalStorage(sessionID, localStorage); err != nil {
		t.Fatalf("Failed to save localStorage: %v", err)
	}

	// Save sessionStorage
	sessionStorage := map[string]string{
		"temp_data":   "some_value",
		"form_values": `{"name":"test"}`,
	}
	if err := sm1.SaveSessionStorage(sessionID, sessionStorage); err != nil {
		t.Fatalf("Failed to save sessionStorage: %v", err)
	}

	// Save IndexedDB
	indexedDB := map[string]any{
		"app_data": map[string]string{
			"version": "1.0.0",
			"theme":   "dark",
		},
	}
	if err := sm1.SaveIndexedDB(sessionID, indexedDB); err != nil {
		t.Fatalf("Failed to save IndexedDB: %v", err)
	}

	sm1.Close()

	// Create a new session manager to test persistence
	sm2, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create second session manager: %v", err)
	}
	defer sm2.Close()

	// Load the session
	loadedSession, err := sm2.LoadSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession == nil {
		t.Fatal("Loaded session is nil")
	}

	if loadedSession.ID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, loadedSession.ID)
	}

	if loadedSession.UserAgent != "Mozilla/5.0 Test" {
		t.Errorf("Expected UserAgent 'Mozilla/5.0 Test', got '%s'", loadedSession.UserAgent)
	}

	// Check cookies
	if len(loadedSession.Cookies) != 2 {
		t.Errorf("Expected 2 cookies, got %d", len(loadedSession.Cookies))
	}

	// Check localStorage
	if len(loadedSession.LocalStorage) != 3 {
		t.Errorf("Expected 3 localStorage items, got %d", len(loadedSession.LocalStorage))
	}

	if loadedSession.LocalStorage["user_pref"] != "dark_mode" {
		t.Errorf("Expected localStorage['user_pref'] = 'dark_mode', got '%s'", loadedSession.LocalStorage["user_pref"])
	}

	// Check sessionStorage
	if len(loadedSession.SessionStorage) != 2 {
		t.Errorf("Expected 2 sessionStorage items, got %d", len(loadedSession.SessionStorage))
	}

	// Check IndexedDB
	if len(loadedSession.IndexedDB) != 1 {
		t.Errorf("Expected 1 IndexedDB entry, got %d", len(loadedSession.IndexedDB))
	}
}

func TestEncryptedPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath:   tmpDir,
		TTL:           1 * time.Hour,
		Encrypt:       true,
		EncryptionKey: "test-secret-key-for-encryption",
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	session := sm.CreateSession()
	sessionID := session.ID

	// Add sensitive data
	session.Cookies = []*http.Cookie{
		{Name: "auth_token", Value: "super_secret_token_12345", Domain: "example.com"},
	}
	_ = sm.SaveCookies(sessionID, session.Cookies)

	sm.Close()

	// Verify file is encrypted (can't read as plain JSON)
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read tmp dir: %v", err)
	}

	var found bool
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".enc") {
			found = true
			content, err := os.ReadFile(filepath.Join(tmpDir, f.Name()))
			if err != nil {
				t.Fatalf("Failed to read encrypted file: %v", err)
			}

			// Try to parse as JSON - should fail if properly encrypted
			var data map[string]interface{}
			if err := json.Unmarshal(content, &data); err == nil {
				t.Error("File should be encrypted, but was readable as JSON")
			}
			break
		}
	}

	if !found {
		t.Error("No encrypted file found")
	}

	// Create new session manager with same key - should decrypt successfully
	sm2, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create second session manager: %v", err)
	}
	defer sm2.Close()

	loadedSession, err := sm2.LoadSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to load encrypted session: %v", err)
	}

	if loadedSession == nil {
		t.Fatal("Loaded session is nil")
	}

	if len(loadedSession.Cookies) != 1 || loadedSession.Cookies[0].Value != "super_secret_token_12345" {
		t.Error("Decrypted session data is incorrect")
	}
}

func TestSessionTTL(t *testing.T) {
	// Test IsExpired method
	session := &Session{
		ID:         generateSessionID(),
		LastUsedAt: time.Now(),
	}

	// Should not be expired with 1 hour TTL
	if session.IsExpired(1 * time.Hour) {
		t.Error("Fresh session should not be expired")
	}

	// Set LastUsedAt to 2 hours ago
	session.LastUsedAt = time.Now().Add(-2 * time.Hour)

	// Should be expired with 1 hour TTL
	if !session.IsExpired(1 * time.Hour) {
		t.Error("Old session should be expired")
	}
}

func TestCleanupExpired(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         100 * time.Millisecond,
		Encrypt:     false,
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	// Create multiple sessions
	for i := 0; i < 5; i++ {
		sm.CreateSession()
	}

	if count := sm.GetSessionCount(); count != 5 {
		t.Errorf("Expected 5 sessions, got %d", count)
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Cleanup expired sessions
	if err := sm.CleanupExpired(); err != nil {
		t.Fatalf("CleanupExpired failed: %v", err)
	}

	if count := sm.GetSessionCount(); count != 0 {
		t.Errorf("Expected 0 sessions after cleanup, got %d", count)
	}
}

func TestReturningVisitorSimulation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath:          tmpDir,
		TTL:                  1 * time.Hour,
		Encrypt:              false,
		ReturningVisitorRate: 100, // Always return existing session
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	// Create initial sessions
	for i := 0; i < 3; i++ {
		sm.CreateSession()
	}

	// With 100% returning rate, should always get existing session
	session := sm.GetOrCreateSession()
	if session == nil {
		t.Fatal("Expected a session, got nil")
	}

	if !session.IsReturning {
		t.Error("Expected returning session with 100% rate")
	}

	if session.VisitCount < 1 {
		t.Errorf("Expected VisitCount >= 1 for returning visitor, got %d", session.VisitCount)
	}
}

func TestGetRandomExistingSession(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	// No sessions yet
	if s := sm.GetRandomExistingSession(); s != nil {
		t.Error("Expected nil with no sessions")
	}

	// Create sessions
	createdIDs := make(map[string]bool)
	for i := 0; i < 5; i++ {
		s := sm.CreateSession()
		createdIDs[s.ID] = true
	}

	// Get random session multiple times
	for i := 0; i < 10; i++ {
		s := sm.GetRandomExistingSession()
		if s == nil {
			t.Error("Expected a session, got nil")
			continue
		}
		if !createdIDs[s.ID] {
			t.Errorf("Got unexpected session ID: %s", s.ID)
		}
	}
}

func TestSessionStats(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	// Initial stats should be zero
	stats := sm.GetStats()
	if stats.TotalSessions != 0 {
		t.Errorf("Expected 0 total sessions, got %d", stats.TotalSessions)
	}

	// Create sessions
	for i := 0; i < 5; i++ {
		s := sm.CreateSession()
		if i%2 == 0 {
			s.IsReturning = true
		}
		for j := 0; j < i; j++ {
			s.UpdateLastUsed()
		}
	}

	stats = sm.GetStats()
	if stats.TotalSessions != 5 {
		t.Errorf("Expected 5 total sessions, got %d", stats.TotalSessions)
	}

	if stats.ReturningCount != 3 {
		t.Errorf("Expected 3 returning sessions, got %d", stats.ReturningCount)
	}

	if stats.NewCount != 2 {
		t.Errorf("Expected 2 new sessions, got %d", stats.NewCount)
	}

	if stats.AvgVisitCount == 0 {
		t.Error("Expected non-zero average visit count")
	}
}

func TestExportImportSession(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm1, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Create and populate session
	session := sm1.CreateSession()
	session.UserAgent = "TestAgent/1.0"
	session.Cookies = []*http.Cookie{
		{Name: "test", Value: "value"},
	}
	session.LocalStorage = map[string]string{"key": "value"}
	_ = sm1.SaveCookies(session.ID, session.Cookies)

	// Export
	encoded, err := sm1.ExportSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to export session: %v", err)
	}
	if encoded == "" {
		t.Error("Exported session is empty")
	}

	sm1.Close()

	// Create new manager and import
	sm2, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create second session manager: %v", err)
	}
	defer sm2.Close()

	imported, err := sm2.ImportSession(encoded)
	if err != nil {
		t.Fatalf("Failed to import session: %v", err)
	}

	if imported == nil {
		t.Fatal("Imported session is nil")
	}

	// New ID should be generated
	if imported.ID == session.ID {
		t.Error("Imported session should have a new ID")
	}

	// Data should be preserved
	if imported.UserAgent != "TestAgent/1.0" {
		t.Errorf("Expected UserAgent 'TestAgent/1.0', got '%s'", imported.UserAgent)
	}

	if len(imported.Cookies) != 1 || imported.Cookies[0].Name != "test" {
		t.Error("Cookies not preserved correctly")
	}

	if imported.LocalStorage["key"] != "value" {
		t.Error("LocalStorage not preserved correctly")
	}
}

func TestDeleteSession(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm, err := NewSessionManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sm.Close()

	session := sm.CreateSession()
	sessionID := session.ID

	if sm.GetSessionCount() != 1 {
		t.Error("Expected 1 session")
	}

	if err := sm.DeleteSession(sessionID); err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	if sm.GetSessionCount() != 0 {
		t.Error("Expected 0 sessions after deletion")
	}

	// Load should return nil
	loaded, _ := sm.LoadSession(sessionID)
	if loaded != nil {
		t.Error("Expected nil for deleted session")
	}
}

func TestFileStoreList(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewFileStore(tmpDir, false, "")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Create some sessions
	for i := 0; i < 3; i++ {
		session := &Session{
			ID:         generateSessionID(),
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
		}
		if err := store.Save(session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}
	}

	ids, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(ids) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(ids))
	}
}

func TestGetFingerprintJS(t *testing.T) {
	js := GetFingerprintJS()
	if js == "" {
		t.Error("GetFingerprintJS returned empty string")
	}
	if !strings.Contains(js, "canvas") {
		t.Error("Fingerprint JS should contain 'canvas'")
	}
	if !strings.Contains(js, "toDataURL") {
		t.Error("Fingerprint JS should contain 'toDataURL'")
	}
}

func TestGetLocalStorageJS(t *testing.T) {
	js := GetLocalStorageJS()
	if js == "" {
		t.Error("GetLocalStorageJS returned empty string")
	}
	if !strings.Contains(js, "localStorage") {
		t.Error("LocalStorage JS should contain 'localStorage'")
	}
}

func TestGetSessionStorageJS(t *testing.T) {
	js := GetSessionStorageJS()
	if js == "" {
		t.Error("GetSessionStorageJS returned empty string")
	}
	if !strings.Contains(js, "sessionStorage") {
		t.Error("SessionStorage JS should contain 'sessionStorage'")
	}
}

func TestGetIndexedDBJS(t *testing.T) {
	js := GetIndexedDBJS()
	if js == "" {
		t.Error("GetIndexedDBJS returned empty string")
	}
	if !strings.Contains(js, "indexedDB") {
		t.Error("IndexedDB JS should contain 'indexedDB'")
	}
}

func TestSetLocalStorageJS(t *testing.T) {
	data := map[string]string{"key1": "value1", "key2": "value2"}
	js := SetLocalStorageJS(data)
	if js == "" {
		t.Error("SetLocalStorageJS returned empty string")
	}
	if !strings.Contains(js, "localStorage.setItem") {
		t.Error("SetLocalStorage JS should contain 'localStorage.setItem'")
	}
	if !strings.Contains(js, "key1") || !strings.Contains(js, "value1") {
		t.Error("SetLocalStorage JS should contain the data")
	}
}

func TestSetSessionStorageJS(t *testing.T) {
	data := map[string]string{"temp": "data"}
	js := SetSessionStorageJS(data)
	if js == "" {
		t.Error("SetSessionStorageJS returned empty string")
	}
	if !strings.Contains(js, "sessionStorage.setItem") {
		t.Error("SetSessionStorage JS should contain 'sessionStorage.setItem'")
	}
}

func TestSessionIsExpired(t *testing.T) {
	session := &Session{
		ID:         generateSessionID(),
		LastUsedAt: time.Now().Add(-2 * time.Hour),
	}

	if !session.IsExpired(1 * time.Hour) {
		t.Error("Session should be expired")
	}

	if session.IsExpired(3 * time.Hour) {
		t.Error("Session should not be expired with 3h TTL")
	}
}

func TestSessionUpdateLastUsed(t *testing.T) {
	session := &Session{
		ID:         generateSessionID(),
		VisitCount: 5,
	}

	oldVisitCount := session.VisitCount
	session.UpdateLastUsed()

	if session.VisitCount != oldVisitCount+1 {
		t.Errorf("Expected VisitCount %d, got %d", oldVisitCount+1, session.VisitCount)
	}

	if time.Since(session.LastUsedAt) > time.Second {
		t.Error("LastUsedAt should be updated to current time")
	}
}

func TestGenerateFingerprint(t *testing.T) {
	session := &Session{ID: generateSessionID()}
	session.GenerateFingerprint()

	if session.CanvasFingerprint == "" {
		t.Error("CanvasFingerprint should not be empty")
	}

	// Should contain WebGL vendor/renderer info
	if !strings.Contains(session.CanvasFingerprint, "Google") &&
		!strings.Contains(session.CanvasFingerprint, "Apple") {
		t.Error("Fingerprint should contain vendor info")
	}
}

func BenchmarkCreateSession(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm, _ := NewSessionManager(cfg)
	defer sm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CreateSession()
	}
}

func BenchmarkSaveLoadSession(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := SessionManagerConfig{
		StoragePath: tmpDir,
		TTL:         1 * time.Hour,
		Encrypt:     false,
	}

	sm, _ := NewSessionManager(cfg)
	defer sm.Close()

	session := sm.CreateSession()
	sessionID := session.ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.SaveCookies(sessionID, []*http.Cookie{
			{Name: "test", Value: "value"},
		})
		sm.LoadSession(sessionID)
	}
}


