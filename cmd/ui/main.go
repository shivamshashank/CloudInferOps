package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/ui"
)

func main() {
	service := ui.NewService()
	handler := ui.NewHandler(service)
	assets := envOr("CLOUDINFEROPS_WEB_ROOT", "./web/dist")
	files := http.FileServer(http.Dir(assets))
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			handler.ServeHTTP(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/cloudinferops")
		if path == "" {
			http.Redirect(w, r, "/cloudinferops/", http.StatusTemporaryRedirect)
			return
		}
		if strings.HasPrefix(path, "/api/") {
			r.URL.Path = path
			handler.ServeHTTP(w, r)
			return
		}
		clean := strings.TrimPrefix(path, "/")
		if clean == "" {
			serveIndex(w, assets)
			return
		}
		if _, err := os.Stat(filepath.Join(assets, filepath.Clean(clean))); err != nil {
			serveIndex(w, assets)
			return
		} else {
			r.URL.Path = "/" + clean
		}
		files.ServeHTTP(w, r)
	})
	port := envOr("PORT", "8080")
	server := &http.Server{Addr: ":" + port, Handler: securityHeaders(root), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 20 * time.Minute, IdleTimeout: 60 * time.Second}
	log.Printf("CloudInferOps portal listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func serveIndex(w http.ResponseWriter, assets string) {
	data, err := os.ReadFile(filepath.Join(assets, "index.html"))
	if err != nil {
		http.Error(w, "portal assets unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}
