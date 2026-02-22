package auth

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

// mockUserRecord mirrors the structure of testdata/mock_users.json.
type mockUserRecord struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
	Disabled    bool     `json:"disabled"`
}

type mockDB struct {
	Users []mockUserRecord `json:"users"`
}

// TestMockProvider is a test-only Provider that loads users from a JSON file.
// It is intentionally in a _test.go file so it never enters the production binary.
//
// Scenario values:
//
//	"ok"            – normal flow, credentials validated against the JSON file.
//	"timeout"       – simulates a slow / hung provider; respects context cancellation.
//	"provider_down" – returns ErrProviderUnavailable immediately.
type TestMockProvider struct {
	Scenario string // "ok" | "timeout" | "provider_down"
	DBPath   string // path to testdata/mock_users.json relative to the test binary
}

func (m TestMockProvider) Authenticate(ctx context.Context, username, password string) (*User, error) {
	switch m.Scenario {
	case "timeout":
		select {
		case <-time.After(2 * time.Second):
			return nil, ErrProviderUnavailable
		case <-ctx.Done():
			return nil, ErrProviderUnavailable
		}
	case "provider_down":
		return nil, ErrProviderUnavailable
	}

	// Default ("ok"): validate against JSON fixture.
	db, err := loadMockDB(m.DBPath)
	if err != nil {
		return nil, ErrProviderUnavailable
	}

	for _, u := range db.Users {
		if u.Username != username {
			continue
		}
		if u.Disabled {
			return nil, ErrUserDisabled
		}
		if u.Password != password {
			return nil, ErrInvalidCredentials
		}
		return &User{
			ID:          firstNonEmpty(u.ID, u.Username),
			DisplayName: u.DisplayName,
			Roles:       u.Roles,
		}, nil
	}

	return nil, ErrUserNotFound
}

func loadMockDB(path string) (*mockDB, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var db mockDB
	if err := json.Unmarshal(b, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
