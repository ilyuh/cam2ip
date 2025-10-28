// Package main provides a user management utility for cam2ip.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/gen2brain/cam2ip/handlers"
)

func main() {
	// Инициализируем базу данных
	if err := handlers.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer handlers.GetDatabase().Close()

	fmt.Println("=== cam2ip User Manager ===")
	fmt.Println("1. Create new user")
	fmt.Println("2. List users")
	fmt.Println("3. Exit")
	fmt.Print("Choose an option: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		createUser()
	case "2":
		listUsers()
	case "3":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid option")
		os.Exit(1)
	}
}

func createUser() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Println("Username cannot be empty")
		return
	}

	// Проверяем, существует ли пользователь
	db := handlers.GetDatabase()
	_, err := db.GetUserByUsername(username)
	if err == nil {
		fmt.Printf("User '%s' already exists\n", username)
		return
	}

	fmt.Print("Enter email (optional): ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}
	fmt.Println() // New line after password input

	if len(password) == 0 {
		fmt.Println("Password cannot be empty")
		return
	}

	// Создаем пользователя
	err = db.CreateUser(username, string(password), email)
	if err != nil {
		fmt.Printf("Failed to create user: %v\n", err)
		return
	}

	fmt.Printf("User '%s' created successfully!\n", username)
}

func listUsers() {
	// Получаем список пользователей (простая реализация)
	fmt.Println("Users in the system:")
	fmt.Println("===================")

	// Для простоты, выводим только информацию о том, что пользователи есть
	// В реальном приложении здесь был бы запрос к базе данных для получения списка
	fmt.Println("Use the database directly to view all users:")
	fmt.Println("sqlite3 data/cam2ip.db \"SELECT id, username, email, is_active, created_at FROM users;\"")
}
