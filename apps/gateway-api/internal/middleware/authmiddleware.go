package middleware

import (
	"encoding/json"
	"net/http"

	"go-svc/apps/gateway-api/internal/authctx"
	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/auth/kratos"
)

// AuthMiddleware validates the Kratos session token and maps the Kratos
// identity to a local user, injecting the local user_id into the request
// context. It returns 401 on any authentication failure.
type AuthMiddleware struct {
	kratosClient *kratos.Client // concrete client for KratosClient() accessor
	verifier     authctx.SessionVerifier
	userRpc      userclient.UserRpc
}

// NewAuthMiddleware creates an AuthMiddleware.
func NewAuthMiddleware(kratosClient *kratos.Client, userRpc userclient.UserRpc) *AuthMiddleware {
	return &AuthMiddleware{
		kratosClient: kratosClient,
		verifier:     kratosClient, // *kratos.Client implements SessionVerifier
		userRpc:      userRpc,
	}
}

// KratosClient returns the underlying Kratos client (for registration/login flows).
func (m *AuthMiddleware) KratosClient() *kratos.Client {
	return m.kratosClient
}

// UserRpcClient returns the underlying UserRpc client.
func (m *AuthMiddleware) UserRpcClient() userclient.UserRpc {
	return m.userRpc
}

// Handle is the go-zero middleware handler. It validates the Bearer token,
// verifies the session with Kratos /sessions/whoami, maps the Kratos
// identity to a local user, and injects user_id into the context.
// On failure it writes exactly one 401 response and does NOT call next.
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authctx.AuthenticateRequest(r, m.verifier, m.userRpc)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ctx := authctx.WithUserID(r.Context(), userID)
		next(w, r.WithContext(ctx))
	}
}
