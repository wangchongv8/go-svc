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

func TestLoginLogic_Success(t *testing.T) {
	svcCtx := svc.NewServiceContext(config.Config{})

	reg, _ := NewRegisterLogic(context.Background(), svcCtx).Register(&user.RegisterRequest{
		Username: "alice",
		Password: "123456",
	})

	resp, err := NewLoginLogic(context.Background(), svcCtx).Login(&user.LoginRequest{
		Username: "alice",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Id != reg.Id {
		t.Errorf("expected id %d, got %d", reg.Id, resp.Id)
	}
}

func TestLoginLogic_WrongPassword(t *testing.T) {
	svcCtx := svc.NewServiceContext(config.Config{})

	NewRegisterLogic(context.Background(), svcCtx).Register(&user.RegisterRequest{
		Username: "alice",
		Password: "123456",
	})

	_, err := NewLoginLogic(context.Background(), svcCtx).Login(&user.LoginRequest{
		Username: "alice",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if st, ok := status.FromError(err); !ok || st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestLoginLogic_UserNotFound(t *testing.T) {
	svcCtx := svc.NewServiceContext(config.Config{})

	_, err := NewLoginLogic(context.Background(), svcCtx).Login(&user.LoginRequest{
		Username: "nonexistent",
		Password: "123456",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}
	if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", err)
	}
}
