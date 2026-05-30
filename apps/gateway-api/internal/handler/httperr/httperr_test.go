package httperr

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWrite_GRPCError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"NotFound", status.Error(codes.NotFound, "not found"), http.StatusNotFound},
		{"InvalidArgument", status.Error(codes.InvalidArgument, "bad request"), http.StatusBadRequest},
		{"AlreadyExists", status.Error(codes.AlreadyExists, "exists"), http.StatusConflict},
		{"Unauthenticated", status.Error(codes.Unauthenticated, "no auth"), http.StatusUnauthorized},
		{"FailedPrecondition", status.Error(codes.FailedPrecondition, "stock low"), http.StatusBadRequest},
		{"Internal", status.Error(codes.Internal, "boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Write(rec, tt.err)
			if rec.Code != tt.wantStatus {
				t.Errorf("expected HTTP %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestWrite_NonGRPCError(t *testing.T) {
	rec := httptest.NewRecorder()
	Write(rec, errors.New("parse error"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected HTTP 400 for non-gRPC error, got %d", rec.Code)
	}
}
