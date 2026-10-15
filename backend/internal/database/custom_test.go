package database

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// The DDL below mirrors the real migrations for exactly the tables the
// counter-aware deletes touch: library (00019, with the 00031/00032 column
// changes), archive (00001 + the library_id added in 00021), the five entity
// tables (00002, 00006, 00009, 00011, 00013) and their join tables with the
// cascading FKs (00005, 00010, 00012, 00014, 00018).
const taxonomySchema = `
CREATE TABLE library (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL default 'doujinshi'
                    CHECK (type IN ('doujinshi', 'audio')),
    created_at DATETIME NOT NULL default (datetime('now')),
    updated_at DATETIME NOT NULL default (datetime('now')),
    scan_interval_minutes INTEGER NOT NULL default 0,
    watch_enabled INTEGER NOT NULL default 1,
    last_scanned_at DATETIME
);

CREATE TABLE archive (
    id TEXT PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    language TEXT,
    category TEXT,
    page_count INTEGER NOT NULL default 0,
    file_path TEXT UNIQUE NOT NULL,
    file_size INTEGER NOT NULL,
    mod_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL default (datetime('now')),
    updated_at DATETIME NOT NULL default (datetime('now')),
    release_date DATETIME,
    library_id TEXT NOT NULL REFERENCES library (id) ON DELETE CASCADE
);

CREATE TABLE artist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL
);

CREATE TABLE circle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    count INTEGER NOT NULL
);

CREATE TABLE tag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    count INTEGER NOT NULL
);

CREATE TABLE parody (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL
);

CREATE TABLE character (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL
);

CREATE TABLE archive_artist (
    archive_id TEXT,
    artist_id INTEGER,
    PRIMARY KEY (archive_id, artist_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (artist_id) REFERENCES artist (id) ON DELETE CASCADE
);

CREATE TABLE archive_circle (
    archive_id TEXT,
    circle_id INTEGER,
    PRIMARY KEY (archive_id, circle_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (circle_id) REFERENCES circle (id) ON DELETE CASCADE
);

CREATE TABLE archive_tag (
    archive_id TEXT,
    tag_id INTEGER,
    PRIMARY KEY (archive_id, tag_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag (id) ON DELETE CASCADE
);

CREATE TABLE archive_parody (
    archive_id TEXT,
    parody_id INTEGER,
    PRIMARY KEY (archive_id, parody_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (parody_id) REFERENCES parody (id) ON DELETE CASCADE
);

CREATE TABLE archive_character (
    archive_id TEXT,
    character_id INTEGER,
    PRIMARY KEY (archive_id, character_id),
    FOREIGN KEY (archive_id) REFERENCES archive (id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES character (id) ON DELETE CASCADE
);
`

// The fixture stores counts equal to the true number of links: artist three
// (a1, a2, a4), circle two (a1, a2), tag two (a1, a2), parody one (a2 only),
// character two (a1, a2). a3 has no links at all, and both libraries hold at
// least one linked archive - so a library delete that ignored its own id
// would move the wrong counters and fail the assertions.
const taxonomyFixture = `
INSERT INTO library (id, name, path) VALUES
    ('lib1', 'Library One', '/library/one'),
    ('lib2', 'Library Two', '/library/two');

INSERT INTO archive (id, title, file_path, file_size, mod_time, library_id) VALUES
    ('a1', 'One', '/library/one/a1.zip', 1, '2026-01-01', 'lib1'),
    ('a2', 'Two', '/library/one/a2.zip', 1, '2026-01-01', 'lib1'),
    ('a3', 'Three', '/library/two/a3.zip', 1, '2026-01-01', 'lib2'),
    ('a4', 'Four', '/library/two/a4.zip', 1, '2026-01-01', 'lib2');

INSERT INTO artist (id, name, count) VALUES (1, 'artist-x', 3);
INSERT INTO circle (id, name, count) VALUES (1, 'circle-c', 2);
INSERT INTO tag (id, name, count) VALUES (1, 'tag-t', 2);
INSERT INTO parody (id, name, count) VALUES (1, 'parody-p', 1);
INSERT INTO character (id, name, count) VALUES (1, 'character-k', 2);

INSERT INTO archive_artist (archive_id, artist_id) VALUES ('a1', 1), ('a2', 1), ('a4', 1);
INSERT INTO archive_circle (archive_id, circle_id) VALUES ('a1', 1), ('a2', 1);
INSERT INTO archive_tag (archive_id, tag_id) VALUES ('a1', 1), ('a2', 1);
INSERT INTO archive_parody (archive_id, parody_id) VALUES ('a2', 1);
INSERT INTO archive_character (archive_id, character_id) VALUES ('a1', 1), ('a2', 1);
`

var (
	fixtureArchives  = []string{"a1", "a2", "a3", "a4"}
	fixtureLibraries = []string{libraryOne, libraryTwo}
)

// Ids the cases reuse: the two libraries seeded by the fixture, and an id no
// row ever has, proving a delete of a missing row is a no-op.
const (
	libraryOne  = "lib1"
	libraryTwo  = "lib2"
	neverExists = "never-existed"
)

func newTaxonomyDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	// :memory: is per-connection, so the fixture would vanish behind the pool.
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), "PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if _, err := db.ExecContext(t.Context(), taxonomySchema+taxonomyFixture); err != nil {
		t.Fatalf("schema: %v", err)
	}

	return db
}

// countRows runs a count query and returns its single value. Every query the
// test issues is a literal, so no SQL text is assembled from variables.
func countRows(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()

	var count int64
	if err := db.QueryRowContext(t.Context(), query, args...).Scan(&count); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}

	return count
}

// columnValues lists a text column in order; all ids in the fixture are text.
func columnValues(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), query)
	if err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	defer func() { _ = rows.Close() }()

	var values []string

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatalf("scan value: %v", err)
		}

		values = append(values, value)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %q: %v", query, err)
	}

	return values
}

// linkKey identifies one archive's rows within one relation, keyed by the
// link query itself so the snapshot read cannot drift from the assertion.
type linkKey struct {
	links     string
	archiveID string
}

// taxonomySnapshot reads the fixture archives' joined rows before a delete,
// so the archives that survive can be checked against their own rows.
func taxonomySnapshot(t *testing.T, db *sql.DB) map[linkKey]int64 {
	t.Helper()

	before := make(map[linkKey]int64)

	for _, rel := range taxonomyRelations {
		for _, id := range fixtureArchives {
			before[linkKey{rel.links, id}] = countRows(t, db, rel.links, id)
		}
	}

	return before
}

// taxonomyCounts is the expected counter of each taxonomy table.
type taxonomyCounts struct {
	artist, circle, tag, parody, character int64
}

// taxonomyRelation is one archive relation under test: its label plus the
// literal queries the assertions use — the archive's rows in the join table,
// the join table total, and the denormalized counter a delete must maintain.
// Nothing is read through the list/page queries, so the test cannot pass by
// agreeing with a query that shares the bug.
type taxonomyRelation struct {
	entity  string
	links   string
	total   string
	counter string
	// count pulls this relation's expected value out of a case's counts.
	count func(taxonomyCounts) int64
}

// taxonomyRelations covers all five counters a delete must maintain.
var taxonomyRelations = []taxonomyRelation{
	{
		"artist",
		"SELECT count(*) FROM archive_artist WHERE archive_id = ?",
		"SELECT count(*) FROM archive_artist",
		"SELECT count FROM artist WHERE id = 1",
		func(w taxonomyCounts) int64 { return w.artist },
	},
	{
		"circle",
		"SELECT count(*) FROM archive_circle WHERE archive_id = ?",
		"SELECT count(*) FROM archive_circle",
		"SELECT count FROM circle WHERE id = 1",
		func(w taxonomyCounts) int64 { return w.circle },
	},
	{
		"tag",
		"SELECT count(*) FROM archive_tag WHERE archive_id = ?",
		"SELECT count(*) FROM archive_tag",
		"SELECT count FROM tag WHERE id = 1",
		func(w taxonomyCounts) int64 { return w.tag },
	},
	{
		"parody",
		"SELECT count(*) FROM archive_parody WHERE archive_id = ?",
		"SELECT count(*) FROM archive_parody",
		"SELECT count FROM parody WHERE id = 1",
		func(w taxonomyCounts) int64 { return w.parody },
	},
	{
		"character",
		"SELECT count(*) FROM archive_character WHERE archive_id = ?",
		"SELECT count(*) FROM archive_character",
		"SELECT count FROM character WHERE id = 1",
		func(w taxonomyCounts) int64 { return w.character },
	},
}

// countsRemover is the shape both counter-aware deletes share.
type countsRemover func(ctx context.Context, q *sqlc.Queries, db *sql.DB, id string) error

// countsCase is one delete scenario: the ids to pass to the delete, the
// counters left behind, and the rows of archive/library that must survive.
type countsCase struct {
	name          string
	deletes       []string
	wantCounts    taxonomyCounts
	wantArchives  []string
	wantLibraries []string
}

// runCountsCases runs one table of scenarios against remove. Each scenario
// gets its own seeded database, so nothing depends on execution order.
func runCountsCases(t *testing.T, cases []countsCase, remove countsRemover) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := newTaxonomyDB(t)
			q := sqlc.New(db)
			ctx := t.Context()

			before := taxonomySnapshot(t, db)

			for _, id := range tc.deletes {
				if err := remove(ctx, q, db, id); err != nil {
					t.Fatalf("delete %q: %v", id, err)
				}
			}

			gotArchives := columnValues(t, db, "SELECT id FROM archive ORDER BY id")
			if !slices.Equal(gotArchives, tc.wantArchives) {
				t.Errorf("archives left = %v, want %v", gotArchives, tc.wantArchives)
			}

			gotLibraries := columnValues(t, db, "SELECT id FROM library ORDER BY id")
			if !slices.Equal(gotLibraries, tc.wantLibraries) {
				t.Errorf("libraries left = %v, want %v", gotLibraries, tc.wantLibraries)
			}

			for _, rel := range taxonomyRelations {
				want := rel.count(tc.wantCounts)

				if got := countRows(t, db, rel.total); got != want {
					t.Errorf("%s: %d join rows, want %d (the counter value)", rel.entity, got, want)
				}

				if got := countRows(t, db, rel.counter); got != want {
					t.Errorf("%s.count = %d, want %d", rel.entity, got, want)
				}

				for _, id := range fixtureArchives {
					wantRows := before[linkKey{rel.links, id}]
					if !slices.Contains(tc.wantArchives, id) {
						// The archive is gone, so its links cascaded with it.
						wantRows = 0
					}

					if got := countRows(t, db, rel.links, id); got != wantRows {
						t.Errorf("%s: rows for %s = %d, want %d", rel.entity, id, got, wantRows)
					}
				}
			}
		})
	}
}

func TestDeleteArchiveClearsRelationCounts(t *testing.T) {
	t.Parallel()

	cases := []countsCase{
		{
			name:          "decrements counts and removes only the deleted archive's links",
			deletes:       []string{"a1"},
			wantCounts:    taxonomyCounts{artist: 2, circle: 1, tag: 1, parody: 1, character: 1},
			wantArchives:  []string{"a2", "a3", "a4"},
			wantLibraries: fixtureLibraries,
		},
		{
			name:          "repeated and missing deletes are no-ops",
			deletes:       []string{"a1", "a1", neverExists},
			wantCounts:    taxonomyCounts{artist: 2, circle: 1, tag: 1, parody: 1, character: 1},
			wantArchives:  []string{"a2", "a3", "a4"},
			wantLibraries: fixtureLibraries,
		},
		{
			name:          "archive without relations leaves every count untouched",
			deletes:       []string{"a3"},
			wantCounts:    taxonomyCounts{artist: 3, circle: 2, tag: 2, parody: 1, character: 2},
			wantArchives:  []string{"a1", "a2", "a4"},
			wantLibraries: fixtureLibraries,
		},
	}

	runCountsCases(t, cases, DeleteArchive)
}

func TestDeleteLibraryClearsRelationCounts(t *testing.T) {
	t.Parallel()

	cases := []countsCase{
		{
			name:          "decrements counts for every archive of the deleted library",
			deletes:       []string{libraryOne},
			wantCounts:    taxonomyCounts{artist: 1},
			wantArchives:  []string{"a3", "a4"},
			wantLibraries: []string{libraryTwo},
		},
		{
			name:          "repeated and missing deletes are no-ops",
			deletes:       []string{libraryOne, libraryOne, neverExists},
			wantCounts:    taxonomyCounts{artist: 1},
			wantArchives:  []string{"a3", "a4"},
			wantLibraries: []string{libraryTwo},
		},
		{
			name:          "leaves the archives and counters of other libraries alone",
			deletes:       []string{libraryTwo},
			wantCounts:    taxonomyCounts{artist: 2, circle: 2, tag: 2, parody: 1, character: 2},
			wantArchives:  []string{"a1", "a2"},
			wantLibraries: []string{libraryOne},
		},
	}

	runCountsCases(t, cases, DeleteLibrary)
}
