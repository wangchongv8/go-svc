package authctx

import (
	"net/http"
	"testing"
)

func TestBearerToken_Valid(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer abc123")
	if got := BearerToken(r); got != "abc123" {
		t.Fatalf("expected abc123, got %q", got)
	}
}

func TestBearerToken_Missing(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if got := BearerToken(r); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestBearerToken_NoPrefix(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "abc123")
	if got := BearerToken(r); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestBearerToken_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "")
	if got := BearerToken(r); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestBearerToken_WithExtraSpaces(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer token-value")
	if got := BearerToken(r); got != "token-value" {
		t.Fatalf("expected token-value, got %q", got)
	}
}

func TestWithUserID_Roundtrip(t *testing.T) {
	ctx := WithUserID(t.Context(), int64(42))
	if got := UserIDFromContext(ctx); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestUserIDFromContext_NoValue(t *testing.T) {
	if got := UserIDFromContext(t.Context()); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}
