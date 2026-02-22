package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// clockSkewTolerance allows a small window for clock differences between servers.
// Tokens are valid starting 10 seconds before their IssuedAt, preventing
// spurious rejections in distributed deployments.
const clockSkewTolerance = 10 * time.Second

type JWTManager struct {
	Secret []byte
	TTL    time.Duration
	Issuer string
}

type Claims struct {
	UserID string   `json:"sub"`
	Name   string   `json:"name,omitempty"`
	Roles  []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

func (j JWTManager) Generate(u *User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: u.ID,
		Name:   u.DisplayName,
		Roles:  u.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-clockSkewTolerance)),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
			Subject:   u.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j JWTManager) Validate(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("missing token")
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return j.Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	if j.Issuer != "" && claims.Issuer != j.Issuer {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}
