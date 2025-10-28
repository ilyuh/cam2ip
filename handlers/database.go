// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection and operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase() (*Database, error) {
	// Создаем директорию для базы данных если её нет
	dbDir := "data"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Подключаемся к SQLite базе данных
	dbPath := filepath.Join(dbDir, "cam2ip.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	database := &Database{db: db}

	// Инициализируем таблицы
	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// initTables creates the necessary tables if they don't exist
func (d *Database) initTables() error {
	// Создаем таблицу пользователей
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Создаем таблицу логов авторизации
	createAuthLogsTable := `
	CREATE TABLE IF NOT EXISTS auth_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		ip_address TEXT NOT NULL,
		user_agent TEXT,
		success BOOLEAN NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Создаем индекс для быстрого поиска по username в auth_logs
	createIndex := `
	CREATE INDEX IF NOT EXISTS idx_auth_logs_username ON auth_logs(username);
	CREATE INDEX IF NOT EXISTS idx_auth_logs_created_at ON auth_logs(created_at);`

	if _, err := d.db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if _, err := d.db.Exec(createAuthLogsTable); err != nil {
		return fmt.Errorf("failed to create auth_logs table: %v", err)
	}

	if _, err := d.db.Exec(createIndex); err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	// Создаем пользователя по умолчанию если его нет
	return d.createDefaultUser()
}

// createDefaultUser creates a default admin user if no users exist
func (d *Database) createDefaultUser() error {
	// Проверяем, есть ли уже пользователи
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %v", err)
	}

	// Если пользователей нет, создаем админа по умолчанию
	if count == 0 {
		hashedPassword := HashPassword("admin")
		_, err := d.db.Exec(`
			INSERT INTO users (username, password, email, is_active) 
			VALUES (?, ?, ?, ?)`,
			"admin", hashedPassword, "admin@cam2ip.local", true)

		if err != nil {
			return fmt.Errorf("failed to create default user: %v", err)
		}

		log.Println("Created default user: admin/admin")
	}

	return nil
}

// AuthenticateUser authenticates a user with username and password
func (d *Database) AuthenticateUser(username, password string) (*User, error) {
	var user User
	var hashedPassword string

	query := `
		SELECT id, username, password, email, is_active, created_at, updated_at 
		FROM users 
		WHERE username = ? AND is_active = 1`

	err := d.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &hashedPassword, &user.Email,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Проверяем пароль
	if !VerifyPassword(password, hashedPassword) {
		return nil, fmt.Errorf("invalid password")
	}

	// Не возвращаем хеш пароля
	user.Password = ""

	return &user, nil
}

// LogAuthAttempt logs an authentication attempt
func (d *Database) LogAuthAttempt(username, ipAddress, userAgent string, success bool) error {
	_, err := d.db.Exec(`
		INSERT INTO auth_logs (username, ip_address, user_agent, success) 
		VALUES (?, ?, ?, ?)`,
		username, ipAddress, userAgent, success)

	if err != nil {
		return fmt.Errorf("failed to log auth attempt: %v", err)
	}

	return nil
}

// GetUserByUsername retrieves a user by username
func (d *Database) GetUserByUsername(username string) (*User, error) {
	var user User

	query := `
		SELECT id, username, email, is_active, created_at, updated_at 
		FROM users 
		WHERE username = ?`

	err := d.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (d *Database) CreateUser(username, password, email string) error {
	hashedPassword := HashPassword(password)

	_, err := d.db.Exec(`
		INSERT INTO users (username, password, email, is_active) 
		VALUES (?, ?, ?, ?)`,
		username, hashedPassword, email, true)

	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

// GetRecentAuthLogs retrieves recent authentication logs
func (d *Database) GetRecentAuthLogs(limit int) ([]AuthLog, error) {
	query := `
		SELECT id, username, ip_address, user_agent, success, created_at 
		FROM auth_logs 
		ORDER BY created_at DESC 
		LIMIT ?`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query auth logs: %v", err)
	}
	defer rows.Close()

	var logs []AuthLog
	for rows.Next() {
		var log AuthLog
		err := rows.Scan(
			&log.ID, &log.Username, &log.IPAddress, &log.UserAgent,
			&log.Success, &log.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auth log: %v", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// Global database instance
var globalDB *Database

// InitDatabase initializes the global database connection
func InitDatabase() error {
	db, err := NewDatabase()
	if err != nil {
		return err
	}
	globalDB = db
	return nil
}

// GetDatabase returns the global database instance
func GetDatabase() *Database {
	return globalDB
}

