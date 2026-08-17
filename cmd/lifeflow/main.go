package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mitja6889/HTTPUtils/internal/api"
	"github.com/mitja6889/HTTPUtils/internal/store"
)

//go:embed all:web
var webFS embed.FS

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dataPath := flag.String("data", defaultDataPath(), "Path to JSON data file")
	seed := flag.Bool("seed", false, "Load demo data (plans for 1 week, goals, habits)")
	flag.Parse()

	s, err := store.New(*dataPath)
	if err != nil {
		log.Fatalf("store init: %v", err)
	}

	if *seed {
		if err := s.SeedDemo(); err != nil {
			log.Fatalf("seed demo: %v", err)
		}
		log.Println("Demo data loaded")
	}

	mux := http.NewServeMux()
	api.New(s).Register(mux)

	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("web fs: %v", err)
	}

	fileServer := http.FileServer(http.FS(webRoot))
	mux.Handle("/", spaHandler(webRoot, fileServer))

	server := &http.Server{
		Addr:              *addr,
		Handler:           withMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("LifeFlow running at http://localhost%s", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

func defaultDataPath() string {
	if dir := os.Getenv("LIFEFLOW_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "lifeflow.json")
	}

	return filepath.Join("data", "lifeflow.json")
}

func spaHandler(webRoot fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.WriteNotFound(w)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(webRoot, path); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
