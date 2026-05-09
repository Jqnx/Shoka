package api

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"database/sql"
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

			userID, err := database.GetSession(r.Context(), db, token)
			if err != nil || userID == "" {
				response.Unauthorized(w)
				return
			}

			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
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

	// Better Auth sets a cookie named "better-auth.session_token" by default
	cookie, err := r.Cookie("better-auth.session_token")
	if err != nil {
		return "", nil
	}

	return cookie.Value, nil
}
