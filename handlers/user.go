// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Не включаем пароль в JSON
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthLog represents an authentication log entry
type AuthLog struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

// HashPassword creates a SHA256 hash of the password
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

