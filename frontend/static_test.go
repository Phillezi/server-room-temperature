package frontend_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phillezi/server-room-temperature/frontend"
)

func TestHandler_ServeIndex(t *testing.T) {
	handler := frontend.Handler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("expected text/html, got %q", contentType)
	}

	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header")
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Server Room Temperature") {
		t.Fatalf("expected html body to contain 'Server Room Temperature'")
	}
}

func TestHandler_SSRParameterInjection(t *testing.T) {
	cfg := frontend.Config{
		NatsWSURL:    "wss://nats.custom-domain.com:443",
		NatsUser:     "custom-reader",
		NatsPassword: "custom-secret-password",
		Subject:      "serverroom.temperature.rack9.sensor5",
	}

	handler := frontend.Handler(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"nats_url":"wss://nats.custom-domain.com:443"`) {
		t.Fatalf("expected injected nats_url in html body, got: %s", body)
	}
	if !strings.Contains(body, `"nats_user":"custom-reader"`) {
		t.Fatalf("expected injected nats_user in html body")
	}
	if !strings.Contains(body, `"nats_password":"custom-secret-password"`) {
		t.Fatalf("expected injected nats_password in html body")
	}
	if !strings.Contains(body, `"subject":"serverroom.temperature.rack9.sensor5"`) {
		t.Fatalf("expected injected subject in html body")
	}
}

func TestHandler_SPAFallback(t *testing.T) {
	handler := frontend.Handler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/custom-view", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on SPA fallback, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Server Room Temperature") {
		t.Fatalf("expected SPA fallback to serve index.html")
	}
}

func TestHandler_ETagConditionalRequest(t *testing.T) {
	handler := frontend.Handler()

	// Initial request to get ETag
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("expected ETag header")
	}

	// Conditional request with If-None-Match
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("If-None-Match", etag)
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Fatalf("expected 304 Not Modified, got %d", rec2.Code)
	}
}

func TestHandler_GzipCompression(t *testing.T) {
	handler := frontend.Handler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding: gzip")
	}

	gzReader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to read decompressed body: %v", err)
	}

	if !strings.Contains(string(decompressed), "Server Room Temperature") {
		t.Fatalf("expected decompressed body to contain 'Server Room Temperature'")
	}
}
