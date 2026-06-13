package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-svc/apps/gateway-api/internal/authctx"
	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/auth/kratos"

	"google.golang.org/grpc"
)

// stubSessionVerifier is a test double for authctx.SessionVerifier.
type stubSessionVerifier struct {
	session *kratos.SessionInfo
	err     error
}

func (s *stubSessionVerifier) WhoAmI(_ context.Context, _ string) (*kratos.SessionInfo, error) {
	return s.session, s.err
}

// stubUserRpc is a test double for userclient.UserRpc.
type stubUserRpc struct {
	getOrCreateResp *userclient.GetOrCreateByKratosIdentityResponse
	err             error
}

func (s *stubUserRpc) GetUser(_ context.Context, _ *userclient.GetUserRequest, _ ...grpc.CallOption) (*userclient.GetUserResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *stubUserRpc) GetOrCreateByKratosIdentity(_ context.Context, _ *userclient.GetOrCreateByKratosIdentityRequest, _ ...grpc.CallOption) (*userclient.GetOrCreateByKratosIdentityResponse, error) {
	return s.getOrCreateResp, s.err
}

func (s *stubUserRpc) GetByKratosIdentity(_ context.Context, _ *userclient.GetByKratosIdentityRequest, _ ...grpc.CallOption) (*userclient.GetByKratosIdentityResponse, error) {
	return nil, errors.New("not implemented")
}

func TestAuthMiddleware_MissingTokenReturns401AndDoesNotCallNext(t *testing.T) {
	mw := &AuthMiddleware{
		verifier: &stubSessionVerifier{},
		userRpc:  &stubUserRpc{},
	}

	var nextCalled bool
	handler := mw.Handle(func(w http.ResponseWriter, r *http.Request) {
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

func TestAuthMiddleware_InvalidTokenReturns401AndDoesNotCallNext(t *testing.T) {
	mw := &AuthMiddleware{
		verifier: &stubSessionVerifier{
			err: errors.New("token expired"),
		},
		userRpc: &stubUserRpc{},
	}

	var nextCalled bool
	handler := mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if nextCalled {
		t.Fatal("next should not be called when token is invalid")
	}
}

func TestAuthMiddleware_ValidTokenInjectsUserIDAndCallsNext(t *testing.T) {
	mw := &AuthMiddleware{
		verifier: &stubSessionVerifier{
			session: &kratos.SessionInfo{
				Active: true,
				Identity: kratos.IdentityInfo{
					ID: "kratos-id-1",
					Traits: kratos.IdentityTraits{
						Username: "alice",
					},
				},
			},
		},
		userRpc: &stubUserRpc{
			getOrCreateResp: &userclient.GetOrCreateByKratosIdentityResponse{
				Id:       int64(7),
				Username: "alice",
			},
		},
	}

	var capturedUserID int64
	handler := mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = authctx.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if capturedUserID != 7 {
		t.Fatalf("expected user_id 7, got %d", capturedUserID)
	}
}

func TestAuthMiddleware_IdentityMappingFailureReturns401(t *testing.T) {
	mw := &AuthMiddleware{
		verifier: &stubSessionVerifier{
			session: &kratos.SessionInfo{
				Active: true,
				Identity: kratos.IdentityInfo{
					ID: "kratos-id-2",
					Traits: kratos.IdentityTraits{
						Username: "bob",
					},
				},
			},
		},
		userRpc: &stubUserRpc{
			err: errors.New("db error"),
		},
	}

	var nextCalled bool
	handler := mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if nextCalled {
		t.Fatal("next should not be called when identity mapping fails")
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
