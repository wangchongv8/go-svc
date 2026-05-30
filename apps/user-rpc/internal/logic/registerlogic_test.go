package logic

import (
	"context"
	"testing"

	"go-svc/apps/user-rpc/internal/config"
	"go-svc/apps/user-rpc/internal/svc"
	"go-svc/apps/user-rpc/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return svc.NewServiceContext(config.Config{})
}

func TestRegisterLogic_Success(t *testing.T) {
	svcCtx := setupServiceContext(t)
	ctx := context.Background()

	resp, err := NewRegisterLogic(ctx, svcCtx).Register(&user.RegisterRequest{
		Username: "alice",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Id == 0 {
		t.Error("expected non-zero id")
	}
	if resp.Username != "alice" {
		t.Errorf("expected username %q, got %q", "alice", resp.Username)
	}

	// verify user persisted
	saved, err := svcCtx.UserStore.FindByID(resp.Id)
	if err != nil {
		t.Fatalf("expected user in store, got error: %v", err)
	}
	if saved.Username != "alice" {
		t.Errorf("expected stored username %q, got %q", "alice", saved.Username)
	}
}

func TestRegisterLogic_DuplicateUsername(t *testing.T) {
	svcCtx := setupServiceContext(t)
	ctx := context.Background()

	_, err := NewRegisterLogic(ctx, svcCtx).Register(&user.RegisterRequest{
		Username: "bob",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("first register should succeed, got %v", err)
	}

	_, err = NewRegisterLogic(ctx, svcCtx).Register(&user.RegisterRequest{
		Username: "bob",
		Password: "654321",
	})
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
	if st, ok := status.FromError(err); !ok || st.Code() != codes.AlreadyExists {
		t.Errorf("expected AlreadyExists, got %v", err)
	}
}

func TestRegisterLogic_EmptyUsername(t *testing.T) {
	svcCtx := setupServiceContext(t)
	ctx := context.Background()

	_, err := NewRegisterLogic(ctx, svcCtx).Register(&user.RegisterRequest{
		Username: "",
		Password: "123456",
	})
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
	if st, ok := status.FromError(err); !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}
