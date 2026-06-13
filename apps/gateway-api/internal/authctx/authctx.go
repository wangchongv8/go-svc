// Package authctx provides context keys and helpers for authentication.
// It is a neutral package that both handler/middleware and logic layers can import.
package authctx

import (
	"context"
	"errors"
	"net/http"
	"strings"

	userclient "go-svc/apps/user-rpc/userrpc"
	"go-svc/pkg/auth/kratos"
	"go-svc/pkg/observability/traceid"

	"github.com/zeromicro/go-zero/core/logx"
)

type contextKey string

const (
	// UserIDKey is the context key for the authenticated user id.
	UserIDKey contextKey = "user_id"
)

// SessionVerifier validates a session token and returns session information.
// *kratos.Client implements this interface.
type SessionVerifier interface {
	WhoAmI(ctx context.Context, sessionToken string) (*kratos.SessionInfo, error)
}

// BearerToken extracts the Bearer token from the Authorization header.
func BearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok {
		return ""
	}
	return token
}

// WithUserID adds the user id to the context.
func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

// UserIDFromContext returns the authenticated user id.
func UserIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(UserIDKey).(int64)
	return id
}

// AuthenticateRequest validates the Bearer token, verifies the session with
// Kratos /sessions/whoami, and maps the Kratos identity to a local user via
// user-rpc. Returns the local user_id on success, or an error.
func AuthenticateRequest(r *http.Request, kratosClient SessionVerifier, userRpc userclient.UserRpc) (int64, error) {
	sessionToken := BearerToken(r)
	if sessionToken == "" {
		return 0, errors.New("missing Authorization header")
	}

	session, err := kratosClient.WhoAmI(r.Context(), sessionToken)
	if err != nil {
		logx.WithContext(r.Context()).Errorw("auth: whoami failed",
			logx.Field("err", err))
		return 0, errors.New("invalid or expired token")
	}

	kratosID := session.Identity.ID
	username := session.Identity.Traits.Username

	regResp, err := userRpc.GetOrCreateByKratosIdentity(traceid.WithOutgoingMetadata(r.Context()), &userclient.GetOrCreateByKratosIdentityRequest{
		KratosIdentityId: kratosID,
		Username:         username,
	})
	if err != nil {
		logx.WithContext(r.Context()).Errorw("auth: identity mapping failed",
			logx.Field("kratos_identity_id", kratosID),
			logx.Field("username", username),
			logx.Field("err", err))
		return 0, errors.New("identity mapping failed")
	}

	logx.WithContext(r.Context()).Infow("auth: authenticated",
		logx.Field("kratos_identity_id", kratosID),
		logx.Field("user_id", regResp.Id))

	return regResp.Id, nil
}
