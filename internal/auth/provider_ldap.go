package auth

import "context"

// LDAPConfig holds all parameters needed to connect to an Active Directory server.
// Loaded from environment variables via config.go.
type LDAPConfig struct {
	Addr         string // e.g. ldaps://dc.upb.edu.co:636
	BaseDN       string // e.g. DC=upb,DC=edu,DC=co
	BindDN       string // service account DN
	BindPassword string
	UserSearchDN string // base OU for user search, e.g. OU=Users,DC=upb,DC=edu,DC=co
	UseLDAPS     bool
	StartTLS     bool
	TimeoutSec   int
	MaxRetries   int
}

// LDAPProvider implements Provider against a real Active Directory server.
// Currently a stub — Authenticate returns ErrProviderUnavailable until the
// LDAP server is available and the implementation is completed.
type LDAPProvider struct {
	cfg LDAPConfig
}

func NewLDAPProvider(cfg LDAPConfig) *LDAPProvider {
	return &LDAPProvider{cfg: cfg}
}

// Authenticate validates user credentials against Active Directory.
//
// Full double-bind flow (to be implemented when LDAP is available):
//
//  1. Dial: connect using LDAPS or STARTTLS according to cfg.UseLDAPS / cfg.StartTLS.
//     Apply cfg.TimeoutSec as a connection + operation deadline via the context.
//
//  2. Service bind: bind with cfg.BindDN + cfg.BindPassword.
//     Required because AD blocks anonymous searches.
//
//  3. Search user DN: search cfg.UserSearchDN for the real DN of the user.
//     Filter: (&(objectClass=user)(sAMAccountName=<escaped_username>))
//     Always escape with ldap.EscapeFilter(username) to prevent LDAP injection.
//
//  4. User bind: bind as the found DN with the provided password.
//     A successful bind proves the password is correct.
//
//  5. Rebind as service: rebind with cfg.BindDN so we can read attributes.
//     Users often lack permission to read their own memberOf in AD.
//
//  6. Read attributes: fetch displayName and memberOf from the user entry.
//
//  7. Map memberOf → Roles: convert AD group DNs to internal role strings
//     (e.g. "CN=UPB-Admins,..." → "ROLE_ADMIN"). This mapping lives here
//     since User exposes Roles directly, not raw Groups.
//
//  8. Return *User with ID (sAMAccountName), DisplayName, Roles.
//
// Retry logic: on connection failure, retry up to cfg.MaxRetries times.
// On LDAP unavailable: return ErrProviderUnavailable (caller returns HTTP 503).
// New logins fail; active JWT tokens remain valid until expiry (coherence eventual).
func (p *LDAPProvider) Authenticate(_ context.Context, _, _ string) (*User, error) {
	// TODO: implement when LDAP server is available.
	return nil, ErrProviderUnavailable
}
