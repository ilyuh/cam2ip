// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// SessionManager manages user sessions
type SessionManager struct {
	sessions map[string]time.Time
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]time.Time),
	}

	// Запускаем горутину для очистки истекших сессий
	go sm.cleanupExpiredSessions()

	return sm
}

// CreateSession creates a new session and returns session ID
func (sm *SessionManager) CreateSession() string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Генерируем случайный session ID
	sessionID := make([]byte, 16)
	rand.Read(sessionID)
	sessionIDStr := hex.EncodeToString(sessionID)

	// Сохраняем сессию с временем истечения (24 часа)
	sm.sessions[sessionIDStr] = time.Now().Add(24 * time.Hour)

	return sessionIDStr
}

// IsValidSession checks if session is valid
func (sm *SessionManager) IsValidSession(sessionID string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	expiry, exists := sm.sessions[sessionID]
	if !exists {
		return false
	}

	return time.Now().Before(expiry)
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.sessions, sessionID)
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.Lock()
		now := time.Now()
		for sessionID, expiry := range sm.sessions {
			if now.After(expiry) {
				delete(sm.sessions, sessionID)
			}
		}
		sm.mutex.Unlock()
	}
}

// GetSessionFromRequest extracts session ID from request cookies
func (sm *SessionManager) GetSessionFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SetSessionCookie sets session cookie in response
func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie removes session cookie
func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// Global session manager instance
var sessionManager = NewSessionManager()

