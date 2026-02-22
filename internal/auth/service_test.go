package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

const testDBPath = "../../testdata/mock_users.json"

func TestServiceLogin_OK(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "ok", DBPath: testDBPath})

	user, err := svc.Login(context.Background(), "student", "student123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID != "student" {
		t.Errorf("expected ID=student, got %q", user.ID)
	}
	if len(user.Roles) == 0 {
		t.Error("expected at least one role")
	}
}

func TestServiceLogin_AdminOK(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "ok", DBPath: testDBPath})

	user, err := svc.Login(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID != "admin" {
		t.Errorf("expected ID=admin, got %q", user.ID)
	}
}

func TestServiceLogin_InvalidCredentials(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "ok", DBPath: testDBPath})

	_, err := svc.Login(context.Background(), "student", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestServiceLogin_UserNotFound(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "ok", DBPath: testDBPath})

	_, err := svc.Login(context.Background(), "ghost", "whatever")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestServiceLogin_UserDisabled(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "ok", DBPath: testDBPath})

	_, err := svc.Login(context.Background(), "disabled_user", "pass123")
	if !errors.Is(err, ErrUserDisabled) {
		t.Errorf("expected ErrUserDisabled, got: %v", err)
	}
}

func TestServiceLogin_ProviderDown(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "provider_down", DBPath: testDBPath})

	_, err := svc.Login(context.Background(), "student", "student123")
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable, got: %v", err)
	}
}

func TestServiceLogin_Timeout_RespectsContext(t *testing.T) {
	svc := NewService(TestMockProvider{Scenario: "timeout", DBPath: testDBPath})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := svc.Login(ctx, "student", "student123")
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable on timeout, got: %v", err)
	}
}
