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

func TestGetUserLogic_Success(t *testing.T) {
	svcCtx := svc.NewServiceContext(config.Config{})

	reg, _ := NewRegisterLogic(context.Background(), svcCtx).Register(&user.RegisterRequest{
		Username: "alice",
		Password: "123456",
	})

	resp, err := NewGetUserLogic(context.Background(), svcCtx).GetUser(&user.GetUserRequest{
		Id: reg.Id,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Id != reg.Id {
		t.Errorf("expected id %d, got %d", reg.Id, resp.Id)
	}
	if resp.Username != "alice" {
		t.Errorf("expected username %q, got %q", "alice", resp.Username)
	}
}

func TestGetUserLogic_NotFound(t *testing.T) {
	svcCtx := svc.NewServiceContext(config.Config{})

	_, err := NewGetUserLogic(context.Background(), svcCtx).GetUser(&user.GetUserRequest{
		Id: 999,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}
	if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", err)
	}
}
