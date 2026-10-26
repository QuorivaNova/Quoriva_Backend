package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Check func(context.Context) error

type statusResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func NewHandler(check Check, logger *slog.Logger, checkTimeout time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "alive"})
	})
	ready := func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), checkTimeout)
		defer cancel()
		if err := check(ctx); err != nil {
			logger.Warn("dependency unavailable", "request_id", w.Header().Get("X-Request-ID"))
			writeJSON(w, http.StatusServiceUnavailable, statusResponse{"quoriva-api", "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, statusResponse{"quoriva-api", "ready"})
	}
	mux.HandleFunc("GET /health/ready", ready)
	mux.HandleFunc("GET /api/status", ready)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := rand.Text()
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
		logger.Info("http request", "request_id", id, "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
