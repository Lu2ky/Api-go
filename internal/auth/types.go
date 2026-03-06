package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrProviderUnavailable = errors.New("auth provider unavailable")
	ErrUserDisabled        = errors.New("user disabled")
)

type User struct {
	ID          string
	DisplayName string
	Email       string
	Groups      []string
	Roles       []string
}