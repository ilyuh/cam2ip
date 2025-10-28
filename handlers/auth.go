// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"fmt"
	"net/http"
)

// Auth handler.
type Auth struct {
}

// NewAuth returns new Auth handler.
func NewAuth() *Auth {
	return &Auth{}
}

// ServeHTTP handles requests on incoming connections.
func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Показываем форму авторизации
		a.showLoginForm(w, r)
	} else if r.Method == "POST" {
		// Обрабатываем данные авторизации
		a.handleLogin(w, r)
	} else {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

// showLoginForm отображает форму авторизации
func (a *Auth) showLoginForm(w http.ResponseWriter, r *http.Request) {
	loginHTML := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Авторизация - cam2ip</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .login-container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            width: 100%;
            max-width: 400px;
        }
        .login-header {
            text-align: center;
            margin-bottom: 2rem;
        }
        .login-header h1 {
            color: #333;
            margin: 0;
        }
        .form-group {
            margin-bottom: 1rem;
        }
        .form-group label {
            display: block;
            margin-bottom: 0.5rem;
            color: #555;
            font-weight: bold;
        }
        .form-group input {
            width: 100%;
            padding: 0.75rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 1rem;
            box-sizing: border-box;
        }
        .form-group input:focus {
            outline: none;
            border-color: #007bff;
            box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
        }
        .login-button {
            width: 100%;
            padding: 0.75rem;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            font-size: 1rem;
            cursor: pointer;
            transition: background-color 0.3s;
        }
        .login-button:hover {
            background-color: #0056b3;
        }
        .error-message {
            color: #dc3545;
            text-align: center;
            margin-top: 1rem;
            padding: 0.5rem;
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            border-radius: 4px;
            display: none;
        }
        .back-link {
            text-align: center;
            margin-top: 1rem;
        }
        .back-link a {
            color: #007bff;
            text-decoration: none;
        }
        .back-link a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-header">
            <h1>Авторизация</h1>
            <p>Войдите в систему cam2ip</p>
        </div>
        
        <form method="POST" action="/login">
            <div class="form-group">
                <label for="username">Имя пользователя:</label>
                <input type="text" id="username" name="username" required>
            </div>
            
            <div class="form-group">
                <label for="password">Пароль:</label>
                <input type="password" id="password" name="password" required>
            </div>
            
            <button type="submit" class="login-button">Войти</button>
        </form>
        
        <div id="error-message" class="error-message"></div>
        
        <div class="back-link">
            <a href="/dashboard">← Перейти к панели управления</a>
        </div>
    </div>

    <script>
        // Показываем ошибку если есть параметр error в URL
        const urlParams = new URLSearchParams(window.location.search);
        const error = urlParams.get('error');
        if (error) {
            document.getElementById('error-message').style.display = 'block';
            document.getElementById('error-message').textContent = decodeURIComponent(error);
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(loginHTML))
}

// handleLogin обрабатывает данные авторизации
func (a *Auth) handleLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Получаем экземпляры базы данных и логгера
	db := GetDatabase()
	logger := GetLogger()

	// Проверяем авторизацию через базу данных
	_, err = db.AuthenticateUser(username, password)
	success := err == nil

	// Логируем попытку авторизации
	logger.LogAuth(username, ipAddress, userAgent, success)

	// Также сохраняем в базу данных
	if dbErr := db.LogAuthAttempt(username, ipAddress, userAgent, success); dbErr != nil {
		logger.LogError("Failed to log auth attempt to database", dbErr)
	} else {
		logger.LogInfo(fmt.Sprintf("Auth attempt logged to database for user: %s", username))
	}

	if success {
		// Создаем сессию
		sessionID := sessionManager.CreateSession()
		sessionManager.SetSessionCookie(w, sessionID)

		// Логируем успешный доступ
		logger.LogAccess(username, ipAddress, "dashboard")

		// Успешная авторизация - переадресация на dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Неудачная авторизация - возвращаемся на страницу входа с ошибкой
	errorMsg := "Неверное имя пользователя или пароль"
	http.Redirect(w, r, "/?error="+errorMsg, http.StatusSeeOther)
}
