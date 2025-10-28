// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"net/http"
)

// AuthMiddleware проверяет авторизацию пользователя
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем session ID из cookie
		sessionID := sessionManager.GetSessionFromRequest(r)

		// Проверяем валидность сессии
		if !sessionManager.IsValidSession(sessionID) {
			// Если сессия недействительна, переадресовываем на страницу авторизации
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Логируем доступ к защищенному ресурсу
		logger := GetLogger()
		if logger != nil {
			ipAddress := GetClientIP(r)
			logger.LogAccess("authenticated_user", ipAddress, r.URL.Path)
		}

		// Если сессия действительна, передаем управление следующему handler
		next.ServeHTTP(w, r)
	})
}
