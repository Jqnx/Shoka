package metadata

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// aliasSchema mirrors the two tables GetArtistAliasTargets reads.
const aliasSchema = `
CREATE TABLE artist (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE artist_alias (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist_id INTEGER NOT NULL,
    alias TEXT UNIQUE NOT NULL
);

INSERT INTO artist (id, name) VALUES
    (1, 'Gagarin Kichi'),
    (2, 'shadow'),
    (3, 'Éx Artist'),
    (4, 'Bob'),
    (5, 'Robert X');
INSERT INTO artist_alias (artist_id, alias) VALUES
    (1, 'gagarinkichi'),
    (1, 'GKC'),
    (1, 'shadow'),
    (3, 'ÉX'),
    (5, 'bob');
`

func TestResolveArtistAliases(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), aliasSchema); err != nil {
		t.Fatalf("seed database: %v", err)
	}

	queries := sqlc.New(db)

	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{
			name:  "aliases resolve to the owning artist",
			names: []string{" Gagarinkichi ", "GKC", "shadow", "new name"},
			want:  []string{"Gagarin Kichi", "Gagarin Kichi", "shadow", "new name"},
		},
		{
			name:  "exact artist name beats a case-variant alias",
			names: []string{"Bob", "bob"},
			want:  []string{"Bob", "Robert X"},
		},
		{
			name:  "non-ASCII alias resolves case-insensitively",
			names: []string{"ÉX", "éx"},
			want:  []string{"Éx Artist", "Éx Artist"},
		},
		{
			name:  "nil stays nil",
			names: nil,
			want:  nil,
		},
		{
			name:  "empty stays non-nil",
			names: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveArtistAliases(context.Background(), queries, tt.names)
			if err != nil {
				t.Fatalf("resolveArtistAliases: %v", err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("resolveArtistAliases(%q) = %q, want %q", tt.names, got, tt.want)
			}

			if (got == nil) != (tt.want == nil) {
				t.Errorf("resolveArtistAliases(%q) nil = %v, want %v", tt.names, got == nil, tt.want == nil)
			}
		})
	}
}
