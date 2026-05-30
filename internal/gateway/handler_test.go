package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	t.Run("GET returns 200 with status ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		HealthzHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
		}

		expectedBody := `{"status":"ok"}`
		if rec.Body.String() != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
		}
	})

	t.Run("non-GET returns 405 Method Not Allowed", func(t *testing.T) {
		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

		for _, m := range methods {
			req := httptest.NewRequest(m, "/healthz", nil)
			rec := httptest.NewRecorder()

			HealthzHandler(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s: expected status %d, got %d", m, http.StatusMethodNotAllowed, rec.Code)
			}
		}
	})
}
