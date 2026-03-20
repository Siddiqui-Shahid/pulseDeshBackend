package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"pulse-backend/internal/db"
	"pulse-backend/internal/handlers"
	"github.com/gorilla/mux"
)

func init() {
	// Determine database file path
	exePath, err := os.Executable()
	if err != nil {
		exePath = "."
	}
	baseDir := filepath.Dir(exePath)
	dbPath := filepath.Join(baseDir, "persons.json")
	db.Init(dbPath)
	log.Printf("Database initialized: %s", dbPath)
}

func main() {
	r := mux.NewRouter()

	// Middleware
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)

	// Routes
	r.HandleFunc("/auth/signup", handlers.SignupHandler).Methods("POST")
	r.HandleFunc("/auth/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/api/users", handlers.UsersHandler).Methods("GET")
	r.HandleFunc("/api/health", handlers.HealthHandler).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(handlers.NotFoundHandler)

	// Start server
	port := "3000"
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📨 [%s] %s | Host: %s | Origin: %s", r.Method, r.URL.Path, r.Host, r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔐 CORS: Setting headers for Origin: %s", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			log.Printf("✅ Handling OPTIONS request")
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
