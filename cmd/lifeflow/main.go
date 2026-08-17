package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitja6889/HTTPUtils/internal/api"
	"github.com/mitja6889/HTTPUtils/internal/store"
)

//go:embed all:web
var webFS embed.FS

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dataPath := flag.String("data", defaultDataPath(), "Path to JSON data file")
	flag.Parse()

	s, err := store.New(*dataPath)
	if err != nil {
		log.Fatalf("store init: %v", err)
	}

	mux := http.NewServeMux()
	api.New(s).Register(mux)

	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("web fs: %v", err)
	}

	fileServer := http.FileServer(http.FS(webRoot))
	mux.Handle("/", spaHandler(webRoot, fileServer))

	log.Printf("LifeFlow running at http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		log.Fatalf("server: %v", err)
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
			http.NotFound(w, r)
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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
