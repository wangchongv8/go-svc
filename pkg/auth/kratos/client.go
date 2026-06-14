// Package kratos provides a thin client for Ory Kratos public API.
package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps the Kratos public API for registration, login, and session.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Kratos client targeting the public API base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegistrationRequest is sent to POST /self-service/registration/api.
type RegistrationRequest struct {
	Traits struct {
		Username string `json:"username"`
	} `json:"traits"`
	Password string `json:"password"`
	Method   string `json:"method"`
}

// LoginRequest is sent to POST /self-service/login/api.
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	Method     string `json:"method"`
}

// SessionInfo is returned by /sessions/whoami.
type SessionInfo struct {
	ID        string        `json:"id"`
	Active    bool          `json:"active"`
	Identity  IdentityInfo  `json:"identity"`
	ExpiresAt time.Time     `json:"expires_at"`
	IssuedAt  time.Time     `json:"issued_at"`
	Devices   []interface{} `json:"devices"`
}

// IdentityInfo holds the identity fields from a session.
type IdentityInfo struct {
	ID        string         `json:"id"`
	SchemaID  string         `json:"schema_id"`
	Traits    IdentityTraits `json:"traits"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// IdentityTraits represents the identity traits schema.
type IdentityTraits struct {
	Username string `json:"username"`
}

// RegisterFlowResponse is the Kratos API registration flow.
type RegisterFlowResponse struct {
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	Session      *SessionInfo `json:"session"`
	SessionToken string       `json:"session_token"`
	Identity     IdentityInfo `json:"identity"`
}

// LoginFlowResponse is the Kratos API login flow.
type LoginFlowResponse struct {
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	Session      *SessionInfo `json:"session"`
	SessionToken string       `json:"session_token"`
	Identity     IdentityInfo `json:"identity"`
}

// InitializeRegistrationFlow creates a new Kratos registration flow.
func (c *Client) InitializeRegistrationFlow(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/self-service/registration/api", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var flow RegisterFlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return "", err
	}
	return flow.ID, nil
}

// CompleteRegistrationFlow submits the registration data to Kratos.
func (c *Client) CompleteRegistrationFlow(ctx context.Context, flowID string, req RegistrationRequest) (*RegisterFlowResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/self-service/registration?flow=%s", c.baseURL, flowID),
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kratos registration failed (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var flow RegisterFlowResponse
	if err := json.Unmarshal(bodyBytes, &flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

// InitializeLoginFlow creates a new Kratos login flow.
func (c *Client) InitializeLoginFlow(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/self-service/login/api", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var flow LoginFlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return "", err
	}
	return flow.ID, nil
}

// CompleteLoginFlow submits login credentials to Kratos.
func (c *Client) CompleteLoginFlow(ctx context.Context, flowID string, req LoginRequest) (*LoginFlowResponse, error) {
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/self-service/login?flow=%s", c.baseURL, flowID),
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kratos login failed (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var flow LoginFlowResponse
	if err := json.Unmarshal(bodyBytes, &flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

// WhoAmI validates a session token and returns session info.
func (c *Client) WhoAmI(ctx context.Context, sessionToken string) (*SessionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/sessions/whoami", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kratos whoami failed (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var session SessionInfo
	if err := json.Unmarshal(bodyBytes, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
