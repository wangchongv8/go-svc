package authctx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeAuthHandler exercises the authctx helpers (BearerToken + WithUserID)
// in a http.Handler chain. It tests the context injection contract without
// depending on real Kratos or user-rpc.
// For tests of the production middleware, see internal/middleware/authmiddleware_test.go.
func fakeAuthHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := BearerToken(r)
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"missing Authorization header"}`))
			return
		}
		// In real middleware, token would be validated against Kratos.
		// Here we inject a test user_id.
		ctx := WithUserID(r.Context(), int64(42))
		next(w, r.WithContext(ctx))
	}
}

func TestAuthMiddleware_MissingTokenReturns401(t *testing.T) {
	var nextCalled bool
	handler := fakeAuthHandler(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if nextCalled {
		t.Fatal("next should not be called when auth fails")
	}
	if !contains(w.Body.String(), "missing Authorization header") {
		t.Fatalf("expected error message in body, got %q", w.Body.String())
	}
}

func TestAuthMiddleware_ValidTokenCallsNextWithUserID(t *testing.T) {
	var capturedUserID int64
	handler := fakeAuthHandler(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if capturedUserID != 42 {
		t.Fatalf("expected user_id 42, got %d", capturedUserID)
	}
}

func TestAuthMiddleware_NonBearerTokenReturns401(t *testing.T) {
	var nextCalled bool
	handler := fakeAuthHandler(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "NotBearer test-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if nextCalled {
		t.Fatal("next should not be called for non-Bearer token")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
