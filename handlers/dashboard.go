// Package handlers provides HTTP handlers for the cam2ip application.
package handlers

import (
	"net/http"
)

// Dashboard handler.
type Dashboard struct {
}

// NewDashboard returns new Dashboard handler.
func NewDashboard() *Dashboard {
	return &Dashboard{}
}

// ServeHTTP handles requests on incoming connections.
func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	dashboardHTML := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Панель управления - cam2ip</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
        }
        .header {
            background-color: #007bff;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            margin: 0;
            display: inline-block;
        }
        .logout-btn {
            float: right;
            background-color: #dc3545;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-top: 0.5rem;
        }
        .logout-btn:hover {
            background-color: #c82333;
        }
        .container {
            max-width: 1200px;
            margin: 2rem auto;
            padding: 0 2rem;
        }
        .welcome {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 2rem;
        }
        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 2rem;
        }
        .service-card {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
            transition: transform 0.3s, box-shadow 0.3s;
        }
        .service-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 4px 20px rgba(0,0,0,0.15);
        }
        .service-card h3 {
            color: #333;
            margin-top: 0;
        }
        .service-card p {
            color: #666;
            margin-bottom: 1.5rem;
        }
        .service-link {
            display: inline-block;
            background-color: #007bff;
            color: white;
            padding: 0.75rem 1.5rem;
            text-decoration: none;
            border-radius: 4px;
            transition: background-color 0.3s;
        }
        .service-link:hover {
            background-color: #0056b3;
        }
        .status {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.875rem;
            font-weight: bold;
            margin-left: 1rem;
        }
        .status.online {
            background-color: #d4edda;
            color: #155724;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>cam2ip - Панель управления</h1>
        <a href="/logout" class="logout-btn">Выйти</a>
    </div>
    
    <div class="container">
        <div class="welcome">
            <h2>Добро пожаловать в панель управления cam2ip!</h2>
            <p>Вы успешно авторизованы. Выберите один из доступных сервисов:</p>
        </div>
        
        <div class="services-grid">
            <div class="service-card">
                <h3>HTML Видеопоток</h3>
                <p>Просмотр видео с камеры в браузере с использованием WebSocket</p>
                <a href="/html" class="service-link">Открыть HTML</a>
            </div>
            
            <div class="service-card">
                <h3>JPEG Изображение</h3>
                <p>Получение статического изображения с камеры в формате JPEG</p>
                <a href="/jpeg" class="service-link">Получить JPEG</a>
            </div>
            
            <div class="service-card">
                <h3>MJPEG Поток</h3>
                <p>Motion JPEG поток для просмотра в медиаплеерах</p>
                <a href="/mjpeg" class="service-link">Открыть MJPEG</a>
            </div>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dashboardHTML))
}

