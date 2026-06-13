package kratos

import (
	"encoding/json"
	"testing"
)

func TestRegistrationRequest_IncludesMethodPassword(t *testing.T) {
	req := RegistrationRequest{
		Traits: struct {
			Username string `json:"username"`
		}{Username: "alice"},
		Password: "123456",
		Method:   "password",
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	// Verify method field is present and set to "password".
	if method, ok := decoded["method"]; !ok {
		t.Fatal("registration request is missing 'method' field")
	} else if method != "password" {
		t.Fatalf("expected method=password, got %v", method)
	}

	// Verify traits are present.
	traits, ok := decoded["traits"].(map[string]interface{})
	if !ok {
		t.Fatal("missing traits field")
	}
	if traits["username"] != "alice" {
		t.Fatalf("expected username=alice, got %v", traits["username"])
	}

	// Verify password is present (but not in logs).
	if password, ok := decoded["password"]; !ok {
		t.Fatal("registration request is missing 'password' field")
	} else if password != "123456" {
		t.Fatalf("expected correct password, got %v", password)
	}
}

func TestLoginRequest_IncludesMethodPassword(t *testing.T) {
	req := LoginRequest{
		Identifier: "alice",
		Password:   "123456",
		Method:     "password",
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if method, ok := decoded["method"]; !ok {
		t.Fatal("login request is missing 'method' field")
	} else if method != "password" {
		t.Fatalf("expected method=password, got %v", method)
	}
}

func TestSessionInfo_JSONFields(t *testing.T) {
	// Verify SessionInfo encodes/decodes correctly.
	body := `{"id":"sess-1","active":true,"identity":{"id":"ident-1","schema_id":"default","traits":{"username":"alice"},"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"},"expires_at":"2025-01-02T00:00:00Z","issued_at":"2025-01-01T00:00:00Z"}`

	var session SessionInfo
	if err := json.Unmarshal([]byte(body), &session); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if session.Identity.Traits.Username != "alice" {
		t.Fatalf("expected alice, got %s", session.Identity.Traits.Username)
	}
	if session.Identity.ID != "ident-1" {
		t.Fatalf("expected ident-1, got %s", session.Identity.ID)
	}
}

func TestClient_NewClient(t *testing.T) {
	client := NewClient("http://localhost:4433")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "http://localhost:4433" {
		t.Fatalf("expected baseURL, got %s", client.baseURL)
	}
}
