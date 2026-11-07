package handlers

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"log/slog"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// mergeAliasSchema mirrors the artist columns the merge planner reads.
const mergeAliasSchema = `
CREATE TABLE artist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL DEFAULT 0,
    cover_archive_id TEXT
);

CREATE TABLE artist_alias (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist_id INTEGER NOT NULL,
    alias TEXT UNIQUE NOT NULL
);

INSERT INTO artist (id, name) VALUES
    (1, 'Merge Source'),
    (2, 'Merge Target'),
    (3, 'Third Artist');
`

func TestMergeArtistAliasesConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		thirdAlias   string
		sourceAlias  string
		wantConflict string
	}{
		{"ASCII case variant", "CSHARED", "cshared", "cshared"},
		{"non-ASCII case variant", "ÉX", "éx", "éx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := newMergeAliasQueries(t, tt.thirdAlias, tt.sourceAlias)
			ctx := context.Background()

			source, err := q.GetArtistByID(ctx, 1)
			if err != nil {
				t.Fatalf("get source artist: %v", err)
			}

			target, err := q.GetArtistByID(ctx, 2)
			if err != nil {
				t.Fatalf("get target artist: %v", err)
			}

			handler := &ArtistHandler{logger: slog.New(slog.DiscardHandler)}

			aliases, conflict, err := handler.mergeArtistAliases(ctx, q, source, target)
			if err != nil {
				t.Fatalf("mergeArtistAliases: %v", err)
			}

			if conflict != tt.wantConflict {
				t.Errorf("conflict = %q, want %q (aliases %q)", conflict, tt.wantConflict, aliases)
			}
		})
	}
}

// newMergeAliasQueries seeds a source alias and a third artist's alias that
// differ only by case, then returns queries over the in-memory database.
func newMergeAliasQueries(t *testing.T, thirdAlias, sourceAlias string) *sqlc.Queries {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()

	if _, err := db.ExecContext(ctx, mergeAliasSchema); err != nil {
		t.Fatalf("seed schema: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO artist_alias (artist_id, alias) VALUES (3, ?), (1, ?)`,
		thirdAlias, sourceAlias,
	); err != nil {
		t.Fatalf("seed aliases: %v", err)
	}

	return sqlc.New(db)
}
