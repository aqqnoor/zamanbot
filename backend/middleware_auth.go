package main

import (
	"net/http"
	"os"
	"strings"
)

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(os.Getenv("API_TOKEN"))
		if token == "" {
			next.ServeHTTP(w, r); return
		}
		h := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(h, prefix) || strings.TrimSpace(h[len(prefix):]) != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized); return
		}
		next.ServeHTTP(w, r)
	})
}

// langFromRequest extracts language from ?lang or Accept-Language; defaults to "kk".
func langFromRequest(r *http.Request) string {
	lang := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang")))
	if lang != "" { return lang }
	al := strings.ToLower(r.Header.Get("Accept-Language"))
	if strings.HasPrefix(al, "kk") { return "kk" }
	if strings.HasPrefix(al, "ru") { return "ru" }
	if strings.HasPrefix(al, "en") { return "en" }
	return "kk"
}
