package database

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// languageFacetSchema is the minimum GetAllLanguage reads: both language-bearing
// tables plus the access table used for the restricted-user branch.
const languageFacetSchema = `
CREATE TABLE archive (library_id TEXT NOT NULL, language TEXT);
CREATE TABLE audio (library_id TEXT NOT NULL, language TEXT);
CREATE TABLE library_access (user_id TEXT NOT NULL, library_id TEXT NOT NULL);

INSERT INTO archive (library_id, language) VALUES
    ('lib1', 'english'),
    ('lib1', ''),
    ('lib1', '   '),
    ('lib1', NULL);
INSERT INTO audio (library_id, language) VALUES
    ('lib2', 'japanese'),
    ('lib2', '');
`

// TestGetAllLanguageExcludesBlanks covers the facet contract: a blank language
// is "unset", so it must not surface as an option. An empty option previously
// collided with the editors' own "None" entry and broke their keyed lists.
func TestGetAllLanguageExcludesBlanks(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "language.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), languageFacetSchema); err != nil {
		t.Fatalf("seed schema: %v", err)
	}

	languages, err := sqlc.New(db).GetAllLanguage(context.Background(), sqlc.GetAllLanguageParams{
		AllLibraries: int64(1),
	})
	if err != nil {
		t.Fatalf("get languages: %v", err)
	}

	want := []string{"english", "japanese"}
	if len(languages) != len(want) {
		t.Fatalf("languages = %v, want %v", languages, want)
	}

	for i, language := range languages {
		if language == nil {
			t.Fatalf("languages[%d] is nil, want %q", i, want[i])
		}

		if *language != want[i] {
			t.Errorf("languages[%d] = %q, want %q", i, *language, want[i])
		}
	}
}
