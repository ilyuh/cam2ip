// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"fmt"
	"net/http"
)

// DebugHeaders logs all request headers for debugging
func DebugHeaders(w http.ResponseWriter, r *http.Request) {
	logger := GetLogger()
	if logger == nil {
		return
	}

	// Логируем все заголовки для отладки
	logger.LogInfo("=== REQUEST HEADERS DEBUG ===")
	for name, values := range r.Header {
		for _, value := range values {
			logger.LogInfo(fmt.Sprintf("Header: %s = %s", name, value))
		}
	}
	logger.LogInfo(fmt.Sprintf("RemoteAddr: %s", r.RemoteAddr))
	logger.LogInfo(fmt.Sprintf("Host: %s", r.Host))
	logger.LogInfo("=== END HEADERS DEBUG ===")
}

// DebugIP shows detailed IP information
func DebugIP(w http.ResponseWriter, r *http.Request) {
	logger := GetLogger()
	if logger == nil {
		return
	}

	ip := GetClientIP(r)
	logger.LogInfo(fmt.Sprintf("=== IP DEBUG ==="))
	logger.LogInfo(fmt.Sprintf("Final IP: %s", ip))
	logger.LogInfo(fmt.Sprintf("RemoteAddr: %s", r.RemoteAddr))

	// Проверяем все заголовки с IP
	ipHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP",
		"True-Client-IP",
		"X-Cluster-Client-IP",
	}

	for _, header := range ipHeaders {
		if value := r.Header.Get(header); value != "" {
			logger.LogInfo(fmt.Sprintf("%s: %s", header, value))
		}
	}
	logger.LogInfo("=== END IP DEBUG ===")
}





