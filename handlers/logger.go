// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Logger represents a logger instance
type Logger struct {
	fileLogger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() (*Logger, error) {
	// Создаем директорию для логов если её нет
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Создаем файл лога с текущей датой
	logFile := filepath.Join(logDir, "cam2ip-"+time.Now().Format("2006-01-02")+".log")

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Создаем логгер с префиксом времени
	fileLogger := log.New(file, "", log.LstdFlags)

	return &Logger{
		fileLogger: fileLogger,
	}, nil
}

// LogAuth logs an authentication attempt
func (l *Logger) LogAuth(username, ipAddress, userAgent string, success bool) {
	status := "FAILED"
	if success {
		status = "SUCCESS"
	}

	// Обрезаем UserAgent если он слишком длинный
	if len(userAgent) > 100 {
		userAgent = userAgent[:100] + "..."
	}

	// Логируем в файл
	l.fileLogger.Printf("AUTH %s - User: %s, IP: %s, UserAgent: %s",
		status, username, ipAddress, userAgent)

	// Также логируем в консоль с дополнительной информацией
	location := getLocationFromIP(ipAddress)
	log.Printf("AUTH %s - User: %s, IP: %s%s", status, username, ipAddress, location)
}

// getLocationFromIP returns additional info about the IP
func getLocationFromIP(ip string) string {
	// Убираем все метки для проверки
	cleanIP := strings.TrimSuffix(ip, " (localhost)")
	cleanIP = strings.TrimSuffix(cleanIP, " (private)")
	cleanIP = strings.TrimSuffix(cleanIP, " (LAN)")

	if strings.HasPrefix(cleanIP, "127.") || cleanIP == "::1" {
		return " (localhost)"
	} else if strings.HasPrefix(cleanIP, "192.168.") {
		return " (LAN)"
	} else if strings.HasPrefix(cleanIP, "10.") {
		return " (LAN)"
	} else if strings.HasPrefix(cleanIP, "172.") {
		return " (LAN)"
	} else if strings.Contains(cleanIP, "(private)") {
		return " (private network)"
	}
	return ""
}

// LogAccess logs general access to protected resources
func (l *Logger) LogAccess(username, ipAddress, resource string) {
	l.fileLogger.Printf("ACCESS - User: %s, IP: %s, Resource: %s",
		username, ipAddress, resource)
}

// LogError logs an error
func (l *Logger) LogError(message string, err error) {
	if err != nil {
		l.fileLogger.Printf("ERROR - %s: %v", message, err)
		log.Printf("ERROR - %s: %v", message, err)
	} else {
		l.fileLogger.Printf("ERROR - %s", message)
		log.Printf("ERROR - %s", message)
	}
}

// LogInfo logs an info message
func (l *Logger) LogInfo(message string) {
	l.fileLogger.Printf("INFO - %s", message)
	log.Printf("INFO - %s", message)
}

// GetClientIP extracts the real IP address from the request
func GetClientIP(r *http.Request) string {
	// Проверяем заголовки прокси в порядке приоритета
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
		"True-Client-IP",   // Cloudflare Enterprise
		"X-Cluster-Client-IP",
	}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			// X-Forwarded-For может содержать несколько IP через запятую
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					ip = strings.TrimSpace(ips[0])
				}
			}

			// Проверяем, что это не локальный адрес
			if !isLocalIP(ip) {
				return ip
			}
		}
	}

	// Если заголовки прокси не найдены или содержат локальные адреса, используем RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	// Добавляем метки для лучшей идентификации
	if isLocalIP(ip) {
		return ip + " (localhost)"
	} else if isPrivateIP(ip) {
		return ip + " (private)"
	}

	return ip
}

// isLocalIP checks if the IP address is a loopback/localhost address
func isLocalIP(ip string) bool {
	// Парсим IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // Если не можем распарсить, считаем локальным
	}

	// Проверяем только на loopback адреса (127.0.0.1, ::1)
	return parsedIP.IsLoopback() ||
		parsedIP.IsUnspecified() ||
		parsedIP.Equal(net.IPv4zero) ||
		parsedIP.Equal(net.IPv6zero)
}

// isPrivateIP checks if the IP address is a private network address
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsPrivate()
}

// Global logger instance
var globalLogger *Logger

// InitLogger initializes the global logger
func InitLogger() error {
	logger, err := NewLogger()
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	return globalLogger
}
