package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phillezi/server-room-temperature/internal/middleware"
)

func TestCORS_PreflightOptions(t *testing.T) {
	dummyCalled := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dummyCalled = true
	})

	handler := middleware.CORS(dummyHandler)

	req := httptest.NewRequest(http.MethodOptions, "/api/history", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", rec.Code)
	}

	if dummyCalled {
		t.Fatalf("expected preflight not to call underlying handler")
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin: *, got %q", got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Methods to be set")
	}

	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Headers: *, got %q", got)
	}
}

func TestCORS_NormalRequest(t *testing.T) {
	dummyCalled := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dummyCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := middleware.CORS(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	req.Header.Set("Origin", "http://another-host.local:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	if !dummyCalled {
		t.Fatalf("expected underlying handler to be called")
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin: *, got %q", got)
	}
}
