package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/mux"
)

// dbFile is the path to the JSON "database" – same directory as the binary.
var dbFile string

func init() {
	// Resolve persons.json relative to this source file so it works whether the
	// binary is run from any working directory.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		filename = "."
	}
	dbFile = filepath.Join(filepath.Dir(filename), "persons.json")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	startTime := time.Now()

	r := mux.NewRouter()

	// ── Logging middleware ──────────────────────────────────────────────────
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Printf("[REQ]   %s %s | %s", req.Method, req.URL.Path, req.RemoteAddr)
			next.ServeHTTP(w, req)
		})
	})

	// ── CORS middleware ─────────────────────────────────────────────────────
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// ── Routes ──────────────────────────────────────────────────────────────
	r.HandleFunc("/signup", handleSignup).Methods(http.MethodPost)
	r.HandleFunc("/auth/signup", handleSignup).Methods(http.MethodPost)
	r.HandleFunc("/login", handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/users", handleUsers).Methods(http.MethodGet)
	r.HandleFunc("/health", handleHealth(startTime)).Methods(http.MethodGet)

	// ── 404 fallback ────────────────────────────────────────────────────────
	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
	r.MethodNotAllowedHandler = http.HandlerFunc(handleNotFound)

	addr := fmt.Sprintf(":%s", port)

	fmt.Printf("\n🚀 Pulse backend (Go) listening on http://localhost%s\n\n", addr)
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /signup      – Create new user")
	fmt.Println("  POST /auth/signup – Create new user (alias)")
	fmt.Println("  POST /login       – Login user")
	fmt.Println("  POST /auth/login  – Login user (alias)")
	fmt.Println("  GET  /users       – List all users (dev only)")
	fmt.Println("  GET  /health      – Health check")
	fmt.Println()

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v\n", err)
	}
}
