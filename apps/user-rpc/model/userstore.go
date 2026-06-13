package model

import (
	"errors"
	"sync"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameExists        = errors.New("username already exists")
	ErrUsernameEmpty         = errors.New("username must not be empty")
	ErrKratosIdentityIDEmpty = errors.New("kratos_identity_id must not be empty")
)

type User struct {
	ID               int64
	Username         string
	KratosIdentityID string // Phase 9: maps to Kratos identity.id
}

// UserStore is a thread-safe in-memory user store.
type UserStore struct {
	mu         sync.RWMutex
	users      map[int64]*User
	byName     map[string]*User
	byKratosID map[string]*User
	nextID     int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:      make(map[int64]*User),
		byName:     make(map[string]*User),
		byKratosID: make(map[string]*User),
		nextID:     1,
	}
}

// safeSuffix returns up to n characters from s, without panicking on short input.
func safeSuffix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// GetOrCreateByKratosIdentity finds or creates a local user mapped to a Kratos identity.
func (s *UserStore) GetOrCreateByKratosIdentity(kratosIdentityID, username string) (*User, error) {
	if kratosIdentityID == "" {
		return nil, ErrKratosIdentityIDEmpty
	}
	if username == "" {
		return nil, ErrUsernameEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Fast path: already mapped
	if u, ok := s.byKratosID[kratosIdentityID]; ok {
		return u, nil
	}

	// Slow path: create a new local profile
	if _, exists := s.byName[username]; exists {
		// Username taken by another identity — append suffix
		username = username + "-" + safeSuffix(kratosIdentityID, 8)
	}

	u := &User{
		ID:               s.nextID,
		Username:         username,
		KratosIdentityID: kratosIdentityID,
	}
	s.nextID++
	s.users[u.ID] = u
	s.byName[u.Username] = u
	s.byKratosID[kratosIdentityID] = u
	return u, nil
}

// FindByKratosIdentity returns the local user for a Kratos identity.
func (s *UserStore) FindByKratosIdentity(kratosIdentityID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byKratosID[kratosIdentityID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *UserStore) FindByID(id int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}
