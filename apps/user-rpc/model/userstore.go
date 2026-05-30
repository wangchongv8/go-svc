package model

import (
	"errors"
	"sync"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrUsernameEmpty   = errors.New("username must not be empty")
	ErrPasswordEmpty   = errors.New("password must not be empty")
	ErrPasswordShort   = errors.New("password must be at least 6 characters")
	ErrInvalidPassword = errors.New("invalid password")
)

type User struct {
	ID       int64
	Username string
	Password string
}

// UserStore is a thread-safe in-memory user store.
type UserStore struct {
	mu     sync.RWMutex
	users  map[int64]*User
	byName map[string]*User
	nextID int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[int64]*User),
		byName: make(map[string]*User),
		nextID: 1,
	}
}

func (s *UserStore) Create(username, password string) (*User, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}
	if password == "" {
		return nil, ErrPasswordEmpty
	}
	if len(password) < 6 {
		return nil, ErrPasswordShort
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[username]; exists {
		return nil, ErrUsernameExists
	}

	u := &User{
		ID:       s.nextID,
		Username: username,
		Password: password,
	}
	s.nextID++
	s.users[u.ID] = u
	s.byName[u.Username] = u
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

func (s *UserStore) FindByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byName[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}
