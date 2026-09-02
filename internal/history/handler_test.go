package history_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phillezi/server-room-temperature/internal/history"
)

func TestHistoryHandler_InvalidParams(t *testing.T) {
	handler := history.HistoryHandler{Service: nil}

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "invalid from format",
			query:      "?from=not-a-date",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid to format",
			query:      "?to=not-a-date",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "to before from",
			query: "?from=" + time.Now().UTC().Format(time.RFC3339) +
				"&to=" + time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/history"+tt.query, nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
