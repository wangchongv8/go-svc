package model

import (
	"testing"
)

func TestGetOrCreateByKratosIdentity_EmptyKratosID(t *testing.T) {
	s := NewUserStore()
	_, err := s.GetOrCreateByKratosIdentity("", "alice")
	if err != ErrKratosIdentityIDEmpty {
		t.Fatalf("expected ErrKratosIdentityIDEmpty, got %v", err)
	}
}

func TestGetOrCreateByKratosIdentity_EmptyUsername(t *testing.T) {
	s := NewUserStore()
	_, err := s.GetOrCreateByKratosIdentity("kratos-1", "")
	if err != ErrUsernameEmpty {
		t.Fatalf("expected ErrUsernameEmpty, got %v", err)
	}
}

func TestGetOrCreateByKratosIdentity_Idempotent(t *testing.T) {
	s := NewUserStore()

	u1, err := s.GetOrCreateByKratosIdentity("kratos-1", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.ID != 1 || u1.Username != "alice" {
		t.Fatalf("unexpected user: %+v", u1)
	}

	// Same call returns the same user.
	u2, err := s.GetOrCreateByKratosIdentity("kratos-1", "alice")
	if err != nil {
		t.Fatalf("unexpected error on idempotent call: %v", err)
	}
	if u2.ID != u1.ID {
		t.Fatalf("idempotent call returned different user: %d vs %d", u2.ID, u1.ID)
	}
}

func TestGetOrCreateByKratosIdentity_DuplicateUsername(t *testing.T) {
	s := NewUserStore()

	// Create a user with a taken username but different Kratos ID.
	u1, err := s.GetOrCreateByKratosIdentity("kratos-1", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.Username != "alice" {
		t.Fatalf("expected alice, got %s", u1.Username)
	}

	u2, err := s.GetOrCreateByKratosIdentity("kratos-2", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should append a suffix because username is taken.
	if u2.Username == "alice" {
		t.Fatalf("expected suffixed username, got %s", u2.Username)
	}
	if u2.ID == u1.ID {
		t.Fatalf("expected different user IDs")
	}
}

func TestGetOrCreateByKratosIdentity_ShortKratosID(t *testing.T) {
	s := NewUserStore()

	u1, err := s.GetOrCreateByKratosIdentity("k-1", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.Username != "alice" {
		t.Fatalf("expected alice, got %s", u1.Username)
	}

	// Short Kratos ID should not panic: safeSuffix returns full string if < 8 chars.
	u2, err := s.GetOrCreateByKratosIdentity("k-2", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u2.Username == "alice" {
		t.Fatalf("expected suffixed username, got %s", u2.Username)
	}
}

func TestFindByKratosIdentity_NotFound(t *testing.T) {
	s := NewUserStore()
	_, err := s.FindByKratosIdentity("nonexistent")
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFindByKratosIdentity_Found(t *testing.T) {
	s := NewUserStore()
	u, _ := s.GetOrCreateByKratosIdentity("kratos-1", "alice")

	found, err := s.FindByKratosIdentity("kratos-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != u.ID {
		t.Fatalf("expected user ID %d, got %d", u.ID, found.ID)
	}
}
