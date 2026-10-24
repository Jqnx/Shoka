package auth

import (
	"context"
	"slices"
)

// LibraryScope restricts a request to a set of libraries. A nil IDs means
// unrestricted (admin).
type LibraryScope struct{ IDs []string }

// Unrestricted reports whether the scope grants every library.
func (s LibraryScope) Unrestricted() bool { return s.IDs == nil }

// Allows reports whether the scope grants libraryID.
func (s LibraryScope) Allows(libraryID string) bool {
	return s.Unrestricted() || slices.Contains(s.IDs, libraryID)
}

// libraryScopeKey is a private context key so a scope can only be installed
// through WithLibraryScope.
type libraryScopeKey struct{}

// WithLibraryScope returns a copy of ctx carrying the caller's library scope.
func WithLibraryScope(ctx context.Context, s LibraryScope) context.Context {
	return context.WithValue(ctx, libraryScopeKey{}, s)
}

// LibraryScopeFromContext returns the library scope stored in ctx. A request
// without one (an admin, whose scope is bypassed) yields the zero value, which
// is unrestricted.
func LibraryScopeFromContext(ctx context.Context) LibraryScope {
	scope, _ := ctx.Value(libraryScopeKey{}).(LibraryScope)
	return scope
}
