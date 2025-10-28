// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"net/http"
)

// Logout handler.
type Logout struct {
}

// NewLogout returns new Logout handler.
func NewLogout() *Logout {
	return &Logout{}
}

// ServeHTTP handles requests on incoming connections.
func (l *Logout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем session ID из cookie
	sessionID := sessionManager.GetSessionFromRequest(r)
	if sessionID != "" {
		// Удаляем сессию
		sessionManager.DeleteSession(sessionID)
	}

	// Очищаем cookie
	sessionManager.ClearSessionCookie(w)

	// Переадресовываем на страницу авторизации
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
