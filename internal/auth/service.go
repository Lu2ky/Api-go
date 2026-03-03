package auth

import "context"

type Provider interface {
	Authenticate(ctx context.Context, username, password string) (*User, error)
}

type Service struct {
	provider Provider
}

func NewService(p Provider) *Service {
	return &Service{provider: p}
}

func (s *Service) Login(ctx context.Context, username, password string) (*User, error) {
	return s.provider.Authenticate(ctx, username, password)
}