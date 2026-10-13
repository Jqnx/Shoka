// Package api exposes the HTTP API: it wires the chi router, the auth and
// permission middleware, and delegates each route to a handler in
// internal/api/handlers.
package api

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

func authMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractToken(r)
			if err != nil || token == "" {
				response.Unauthorized(w)
				return
			}

			session, err := database.GetSession(r.Context(), db, token)
			if err != nil || session == nil {
				response.Unauthorized(w)
				return
			}

			// Banning deletes the user's sessions, so this is normally
			// unreachable - it's here so a stale token can't outlive a ban.
			if session.Banned {
				response.Forbidden(w, "account is banned")
				return
			}

			ctx := auth.WithRole(auth.WithUserID(r.Context(), session.UserID), session.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requirePermission gates a route behind the role matrix in internal/auth.
// Must be mounted inside the authMiddleware group - it reads the role that
// puts on the request context.
func requirePermission(permission auth.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !auth.Can(auth.RoleFromContext(r.Context()), permission) {
				response.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractToken checks the Authorization header first, then falls back
// to the Better Auth session cookie.
func extractToken(r *http.Request) (string, error) {
	// Authorization: Bearer <token>
	header := r.Header.Get("Authorization")

	token, ok := strings.CutPrefix(header, "Bearer ")
	if ok {
		return token, nil
	}

	// Better Auth sets a cookie named "better-auth.session_token" by default.
	// A missing cookie just means "no token" — any other error is real.
	cookie, err := r.Cookie("better-auth.session_token")
	if errors.Is(err, http.ErrNoCookie) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	cookieToken, _, _ := strings.Cut(cookie.Value, ".")

	return cookieToken, nil
}
