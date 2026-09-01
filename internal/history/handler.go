package history

import (
	"encoding/json"
	"net/http"
	"time"
)

type HistoryHandler struct {
	Service *Service
}

func (h HistoryHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	subject := r.URL.Query().Get("subject")

	now := time.Now().UTC()

	from := startOfHour(now)
	to := now

	if value := r.URL.Query().Get("from"); value != "" {
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			http.Error(
				w,
				"invalid from",
				http.StatusBadRequest,
			)
			return
		}

		from = t
	}

	if value := r.URL.Query().Get("to"); value != "" {
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			http.Error(
				w,
				"invalid to",
				http.StatusBadRequest,
			)
			return
		}

		to = t
	}

	if !to.After(from) {
		http.Error(
			w,
			"to must be after from",
			http.StatusBadRequest,
		)
		return
	}

	readings, err := h.Service.History(
		r.Context(),
		subject,
		from,
		to,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(readings)
}

func startOfHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}
