package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const (
	CtxUserID ctxKey = "user_id"
	CtxRoles  ctxKey = "roles"
)

// RequireAuth returns middleware that validates a Bearer JWT and injects
// user_id and roles into the request context.
// Routes behind this middleware can read identity with UserIDFromCtx / RolesFromCtx.
func RequireAuth(jwtm JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			claims, err := jwtm.Validate(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxRoles, claims.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromCtx extracts the authenticated user ID from the request context.
func UserIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(CtxUserID).(string)
	return v
}

// RolesFromCtx extracts the authenticated user roles from the request context.
func RolesFromCtx(ctx context.Context) []string {
	v, _ := ctx.Value(CtxRoles).([]string)
	return v
}

func bearerToken(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
