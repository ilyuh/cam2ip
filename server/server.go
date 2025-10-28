// Package server.
package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gen2brain/cam2ip/handlers"
)

// Server struct.
type Server struct {
	Name    string
	Version string

	Index int
	Delay int

	Width  float64
	Height float64

	Quality int
	Rotate  int
	Flip    string

	NoWebGL bool

	Timestamp  bool
	TimeFormat string

	Bind     string
	Htpasswd string

	Reader handlers.ImageReader
}

// NewServer returns new Server.
func NewServer() *Server {
	s := &Server{}

	return s
}

// ListenAndServe listens on the TCP address and serves requests.
func (s *Server) ListenAndServe() error {
	// Инициализируем базу данных
	if err := handlers.InitDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer handlers.GetDatabase().Close()

	// Инициализируем логгер
	if err := handlers.InitLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	// Note: Basic auth is disabled in favor of custom session-based authentication

	// Публичные маршруты (не требуют авторизации)
	http.Handle("/", handlers.NewAuth()) // Страница авторизации теперь на корневом маршруте

	// Отладочные маршруты (только для разработки)
	http.HandleFunc("/debug/headers", handlers.DebugHeaders)
	http.HandleFunc("/debug/ip", handlers.DebugIP)

	// Защищенные маршруты (требуют авторизации)
	http.Handle("/dashboard", handlers.AuthMiddleware(handlers.NewDashboard()))
	http.Handle("/logout", handlers.NewLogout())
	http.Handle("/html", handlers.AuthMiddleware(handlers.NewHTML(s.Width, s.Height, s.NoWebGL)))
	http.Handle("/jpeg", handlers.AuthMiddleware(handlers.NewJPEG(s.Reader, s.Quality)))
	http.Handle("/mjpeg", handlers.AuthMiddleware(handlers.NewMJPEG(s.Reader, s.Delay, s.Quality)))
	http.Handle("/socket", handlers.AuthMiddleware(handlers.NewSocket(s.Reader, s.Delay, s.Quality)))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", s.Bind)
	if err != nil {
		return err
	}

	return srv.Serve(listener)
}
