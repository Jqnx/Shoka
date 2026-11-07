package api

import (
	"Shoka/internal/auth"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// testToken is the session token the auth fixtures store and send back.
const (
	testToken    = "tok"
	testLibraryA = "lib-a"
	testLibraryB = "lib-b"
)

// Mirrors the frontend-owned Better Auth columns the middleware reads.
func newAuthDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	// :memory: is per-connection, so the fixture would vanish behind the pool.
	db.SetMaxOpenConns(1)

	_, err = db.ExecContext(t.Context(), `
		CREATE TABLE user (id TEXT PRIMARY KEY, role TEXT, banned INTEGER);
		CREATE TABLE session (token TEXT PRIMARY KEY, user_id TEXT NOT NULL, expires_at INTEGER NOT NULL);
		CREATE TABLE library_access (
			user_id TEXT NOT NULL,
			library_id TEXT NOT NULL,
			PRIMARY KEY (user_id, library_id)
		);
	`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}

	return db
}

func addSession(t *testing.T, db *sql.DB, token, userID string, role any, banned bool, expires time.Time) {
	t.Helper()

	if _, err := db.ExecContext(t.Context(), `INSERT INTO user (id, role, banned) VALUES (?, ?, ?)`, userID, role, banned); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err := db.ExecContext(t.Context(),
		`INSERT INTO session (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expires.UnixMilli(),
	)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
}

// Sentinels so a test can assert "reached the handler" without a status it
// would otherwise have to share with a real response.
func teapotHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

// Wraps the sentinel in the auth middleware alone.
func authOnlyRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(authMiddleware(db))
	r.Get("/", teapotHandler)

	return r
}

// Wraps the sentinel in the same middleware pair routes.go mounts.
func guardedRouter(db *sql.DB, permission auth.Permission) http.Handler {
	r := chi.NewRouter()
	r.Use(authMiddleware(db))
	r.With(requirePermission(permission)).Get("/", teapotHandler)

	return r
}

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		token    string
		role     any
		banned   bool
		expires  time.Time
		sendAuth bool
		want     int
	}{
		{name: "no token", want: http.StatusUnauthorized},
		{name: "unknown token", token: "nope", sendAuth: true, want: http.StatusUnauthorized},
		{
			name: "expired session", token: testToken, role: auth.RoleAdmin, sendAuth: true,
			expires: time.Now().Add(-time.Hour), want: http.StatusUnauthorized,
		},
		{
			name: "banned user", token: testToken, role: auth.RoleAdmin, banned: true, sendAuth: true,
			expires: future, want: http.StatusForbidden,
		},
		{
			name: "valid session", token: testToken, role: auth.RoleUser, sendAuth: true,
			expires: future, want: http.StatusTeapot,
		},
		{
			name: "null role is accepted", token: testToken, role: nil, sendAuth: true,
			expires: future, want: http.StatusTeapot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := newAuthDB(t)
			if tt.token != "" && !tt.expires.IsZero() {
				addSession(t, db, tt.token, "u1", tt.role, tt.banned, tt.expires)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			if tt.sendAuth {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rec := httptest.NewRecorder()
			authOnlyRouter(db).ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       any
		permission auth.Permission
		want       int
	}{
		{"admin may update archives", auth.RoleAdmin, auth.ArchiveUpdate, http.StatusTeapot},
		{"admin may create libraries", auth.RoleAdmin, auth.LibraryCreate, http.StatusTeapot},
		{"user may not update archives", auth.RoleUser, auth.ArchiveUpdate, http.StatusForbidden},
		{"user may not create libraries", auth.RoleUser, auth.LibraryCreate, http.StatusForbidden},
		{"null role may not update archives", nil, auth.ArchiveUpdate, http.StatusForbidden},
		{"unknown role may not update archives", "librarian", auth.ArchiveUpdate, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := newAuthDB(t)
			addSession(t, db, testToken, "u1", tt.role, false, time.Now().Add(time.Hour))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+testToken)

			rec := httptest.NewRecorder()
			guardedRouter(db, tt.permission).ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}

			if tt.want == http.StatusForbidden && rec.Body.String() == "" {
				t.Error("forbidden response should carry an error body")
			}
		})
	}
}

func TestAuthMiddlewareResolvesLibraryScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		role         string
		grants       []string
		wantScope    auth.LibraryScope
		wantAllows   []string
		wantDeniedID string
	}{
		{
			name:         "admin is unrestricted",
			role:         auth.RoleAdmin,
			grants:       nil,
			wantScope:    auth.LibraryScope{},
			wantAllows:   []string{testLibraryA, testLibraryB},
			wantDeniedID: "",
		},
		{
			name:         "comma-separated admin role is unrestricted",
			role:         auth.RoleUser + "," + auth.RoleAdmin,
			grants:       nil,
			wantScope:    auth.LibraryScope{},
			wantAllows:   []string{testLibraryA, testLibraryB},
			wantDeniedID: "",
		},
		{
			name:         "user gets only granted ids",
			role:         auth.RoleUser,
			grants:       []string{testLibraryA},
			wantScope:    auth.LibraryScope{IDs: []string{testLibraryA}},
			wantAllows:   []string{testLibraryA},
			wantDeniedID: testLibraryB,
		},
		{
			name:         "user with no grants is denied everything",
			role:         auth.RoleUser,
			grants:       nil,
			wantScope:    auth.LibraryScope{IDs: []string{}},
			wantAllows:   nil,
			wantDeniedID: testLibraryA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := newAuthDB(t)
			addSession(t, db, testToken, "u1", tt.role, false, time.Now().Add(time.Hour))

			for _, libraryID := range tt.grants {
				if _, err := db.ExecContext(t.Context(),
					`INSERT INTO library_access (user_id, library_id) VALUES (?, ?)`, "u1", libraryID,
				); err != nil {
					t.Fatalf("insert grant: %v", err)
				}
			}

			r := chi.NewRouter()
			r.Use(authMiddleware(db))
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				scope := auth.LibraryScopeFromContext(req.Context())
				if scope.Unrestricted() != tt.wantScope.Unrestricted() {
					t.Errorf("unrestricted = %v, want %v", scope.Unrestricted(), tt.wantScope.Unrestricted())
				}

				for _, id := range tt.wantAllows {
					if !scope.Allows(id) {
						t.Errorf("scope should allow %q", id)
					}
				}

				if tt.wantDeniedID != "" && scope.Allows(tt.wantDeniedID) {
					t.Errorf("scope should deny %q", tt.wantDeniedID)
				}

				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+testToken)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
		})
	}
}

// A failure while resolving the allowlist must deny the request, never fall
// through to an unrestricted scope.
func TestAuthMiddlewareFailsClosedOnAccessLookupError(t *testing.T) {
	t.Parallel()

	db := newAuthDB(t)
	addSession(t, db, testToken, "u1", auth.RoleUser, false, time.Now().Add(time.Hour))

	if _, err := db.ExecContext(t.Context(), `DROP TABLE library_access`); err != nil {
		t.Fatalf("drop library_access: %v", err)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	rec := httptest.NewRecorder()
	authOnlyRouter(db).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
