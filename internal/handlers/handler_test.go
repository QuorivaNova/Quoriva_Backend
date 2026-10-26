package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLivenessDoesNotDependOnDatabase(t *testing.T) {
	h := NewHandler(func(context.Context) error { t.Fatal("liveness queried dependency"); return nil }, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Second)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if w.Code != 200 || w.Body.String() != "{\"status\":\"alive\"}\n" {
		t.Fatalf("unexpected liveness response: %d %s", w.Code, w.Body.String())
	}
}

func TestReadinessAndStatusContract(t *testing.T) {
	for _, path := range []string{"/health/ready", "/api/status"} {
		for _, unavailable := range []bool{false, true} {
			t.Run(path+"/"+map[bool]string{false: "ready", true: "unavailable"}[unavailable], func(t *testing.T) {
				h := NewHandler(func(context.Context) error {
					if unavailable {
						return errors.New("postgres://user:secret@db/private")
					}
					return nil
				}, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Second)
				w := httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				wantCode, wantStatus := 200, "ready"
				if unavailable {
					wantCode, wantStatus = 503, "unavailable"
				}
				if w.Code != wantCode || w.Body.String() != "{\"service\":\"quoriva-api\",\"status\":\""+wantStatus+"\"}\n" {
					t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
				}
				if w.Header().Get("X-Request-ID") == "" || w.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("missing correlation/cache header")
				}
				if strings.Contains(w.Body.String(), "secret") {
					t.Fatal("leaked internal error")
				}
			})
		}
	}
}

func TestCheckReceivesDeadlineAndRequestCancellation(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(map[bool]string{false: "timeout", true: "request cancellation"}[canceled], func(t *testing.T) {
			h := NewHandler(func(ctx context.Context) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("missing dependency deadline")
				}
				<-ctx.Done()
				return ctx.Err()
			}, slog.New(slog.NewTextHandler(io.Discard, nil)), 20*time.Millisecond)
			r := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			if canceled {
				ctx, cancel := context.WithCancel(r.Context())
				cancel()
				r = r.WithContext(ctx)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 503 {
				t.Fatalf("expected unavailable, got %d", w.Code)
			}
		})
	}
}
