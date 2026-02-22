package auth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AuthMode string

const (
	AuthModeMock AuthMode = "mock"
	AuthModeLDAP AuthMode = "ldap"
)

type Config struct {
	Mode AuthMode

	JWTSecret string
	JWTTTLMin int
	JWTIssuer string

	LDAP LDAPConfig
}

// LoadConfigFromEnv reads all configuration from environment variables.
// It calls Validate() before returning — if validation fails the application
// should not start (fail-fast).
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{}

	mode := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_MODE")))
	if mode == "" {
		mode = string(AuthModeMock)
	}
	cfg.Mode = AuthMode(mode)

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	cfg.JWTIssuer = os.Getenv("JWT_ISSUER")
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "api"
	}
	cfg.JWTTTLMin = envInt("JWT_TTL_MIN", 60)

	cfg.LDAP = LDAPConfig{
		Addr:         os.Getenv("LDAP_ADDR"),
		BaseDN:       os.Getenv("LDAP_BASE_DN"),
		BindDN:       os.Getenv("LDAP_BIND_DN"),
		BindPassword: os.Getenv("LDAP_BIND_PASS"),
		UserSearchDN: os.Getenv("LDAP_USER_SEARCH_BASE"),
		UseLDAPS:     envBool("LDAP_USE_LDAPS", true),
		StartTLS:     envBool("LDAP_STARTTLS", false),
		TimeoutSec:   envInt("LDAP_TIMEOUT_SEC", 8),
		MaxRetries:   envInt("LDAP_MAX_RETRIES", 1),
	}

	return cfg, cfg.Validate()
}

// Validate checks for missing required configuration and returns an error
// that should be treated as fatal at startup.
func (c Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("missing env: JWT_SECRET")
	}

	switch c.Mode {
	case AuthModeMock:
		// No additional env vars required for mock mode.
		return nil
	case AuthModeLDAP:
		var missing []string
		if c.LDAP.Addr == "" {
			missing = append(missing, "LDAP_ADDR")
		}
		if c.LDAP.BaseDN == "" {
			missing = append(missing, "LDAP_BASE_DN")
		}
		if c.LDAP.BindDN == "" {
			missing = append(missing, "LDAP_BIND_DN")
		}
		if c.LDAP.BindPassword == "" {
			missing = append(missing, "LDAP_BIND_PASS")
		}
		if len(missing) > 0 {
			return fmt.Errorf("AUTH_MODE=ldap but missing env vars: %s", strings.Join(missing, ", "))
		}
		return nil
	default:
		return fmt.Errorf("invalid AUTH_MODE: %q (valid: mock, ldap)", c.Mode)
	}
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envBool(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "yes" || v == "y"
}
