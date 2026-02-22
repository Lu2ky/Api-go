package auth

import "context"

// Provider is the contract for any authentication backend.
// The API never speaks to LDAP directly — only through this interface.
type Provider interface {
	Authenticate(ctx context.Context, username, password string) (*User, error)
}

// Service coordinates authentication. It delegates to the configured Provider
// and is the single entry point for login logic in the rest of the API.
// Future additions (role enrichment, audit logging) go here, not in the provider.
type Service struct {
	provider Provider
}

func NewService(p Provider) *Service {
	return &Service{provider: p}
}

func (s *Service) Login(ctx context.Context, username, password string) (*User, error) {
	return s.provider.Authenticate(ctx, username, password)
}
