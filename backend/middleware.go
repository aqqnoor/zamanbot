package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rid := requestID()
		w.Header().Set("X-Request-Id", rid)

		log.Printf("[REQ] %s %s rid=%s", r.Method, r.URL.Path, rid)
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		dur := time.Since(start)
		log.Printf("[RES] %d %s %s rid=%s dur=%s", ww.status, r.Method, r.URL.Path, rid, dur)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func requestID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func withCORS(next http.Handler) http.Handler {
	origins := os.Getenv("CORS_ORIGINS")
	allowAll := false
	allowed := map[string]struct{}{}
	if origins == "" || origins == "*" {
		allowAll = true
	} else {
		for _, o := range strings.Split(origins, ",") {
			allowed[strings.TrimSpace(o)] = struct{}{}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Language")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
