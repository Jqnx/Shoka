package database

import (
	"Shoka/internal/database/sqlc"
	"database/sql"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Ids the sort cases reuse; audioOne/audioTwo are shared with custom_test.go.
const (
	audioListUser = "user1"
	audioThree    = "au3"
)

const audioListSchema = `
CREATE TABLE library (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL
);

CREATE TABLE circle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE audio (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL REFERENCES library (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    folder_path TEXT UNIQUE NOT NULL,
    mod_time DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    release_date DATETIME,
    language TEXT,
    cover_index INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE audio_circle (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    circle_id INTEGER NOT NULL REFERENCES circle (id) ON DELETE CASCADE,
    PRIMARY KEY (audio_id, circle_id)
);

CREATE TABLE audio_played (
    audio_id TEXT NOT NULL REFERENCES audio (id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    played_at DATETIME NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (audio_id, user_id)
);
`

const audioListFixture = `
INSERT INTO library (id, name, path, type) VALUES ('lib1', 'Audio', '/audio', 'audio');

INSERT INTO audio (id, library_id, title, folder_path, mod_time, created_at, updated_at, release_date) VALUES
    ('au1', 'lib1', 'Beta', '/audio/au1', '2026-01-01', '2026-01-01', '2026-01-01', '2026-01-02'),
    ('au2', 'lib1', 'Alpha', '/audio/au2', '2026-01-01', '2026-01-03', '2026-01-01', '2026-01-01'),
    ('au3', 'lib1', 'Gamma', '/audio/au3', '2026-01-01', '2026-01-02', '2026-01-01', NULL);

INSERT INTO audio_played (audio_id, user_id, played_at, completed) VALUES
    ('au1', 'user1', '2026-02-01', 1),
    ('au3', 'user1', '2026-02-03', 1),
    ('au2', 'user2', '2026-02-04', 1);
`

func newAudioListDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	// :memory: is per-connection, so the fixture would vanish behind the pool.
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), audioListSchema+audioListFixture); err != nil {
		t.Fatalf("schema: %v", err)
	}

	return db
}

func audioIDs(t *testing.T, f AudioFilter) []string {
	t.Helper()

	rows, _, err := ListAudio(t.Context(), newAudioListDB(t), f)
	if err != nil {
		t.Fatalf("ListAudio: %v", err)
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	return ids
}

func TestListAudioSort(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filter AudioFilter
		want   []string
	}{
		"default title order": {
			filter: AudioFilter{Limit: 10, UserID: audioListUser},
			want:   []string{audioTwo, audioOne, audioThree},
		},
		"newest added first": {
			filter: AudioFilter{Sort: "created_at_desc", Limit: 10, UserID: audioListUser},
			want:   []string{audioTwo, audioThree, audioOne},
		},
		"newest released first, missing date last": {
			filter: AudioFilter{Sort: "release_date_desc", Limit: 10, UserID: audioListUser},
			want:   []string{audioOne, audioTwo, audioThree},
		},
		"recently played first, unplayed last": {
			filter: AudioFilter{Sort: "played_at_desc", Limit: 10, UserID: audioListUser},
			want:   []string{audioThree, audioOne, audioTwo},
		},
		"only played works": {
			filter: AudioFilter{
				Sort:       "played_at_desc",
				PlayedOnly: true,
				Limit:      10,
				UserID:     audioListUser,
			},
			want: []string{audioThree, audioOne},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := audioIDs(t, tt.filter); !slices.Equal(got, tt.want) {
				t.Errorf("ids = %v, want %v", got, tt.want)
			}
		})
	}
}

// Marking an already-played work must bump played_at, otherwise "Recently
// Played" would only ever reflect the first time each work was marked.
func TestMarkAudioPlayedBumpsTimestamp(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("single", func(t *testing.T) {
		t.Parallel()

		db := newAudioListDB(t)

		if _, err := db.ExecContext(ctx,
			"UPDATE audio_played SET played_at = '2020-01-01' WHERE audio_id = ? AND user_id = ?",
			audioOne, audioListUser,
		); err != nil {
			t.Fatalf("age played_at: %v", err)
		}

		err := sqlc.New(db).MarkAudioPlayed(ctx, sqlc.MarkAudioPlayedParams{
			AudioID: audioOne,
			UserID:  audioListUser,
		})
		if err != nil {
			t.Fatalf("MarkAudioPlayed: %v", err)
		}

		assertPlayedAtBumped(t, db, audioOne)
	})

	t.Run("started work becomes finished", func(t *testing.T) {
		t.Parallel()

		db := newAudioListDB(t)

		if _, err := db.ExecContext(ctx,
			"UPDATE audio_played SET played_at = '2020-01-01', completed = 0 WHERE audio_id = ? AND user_id = ?",
			audioOne, audioListUser,
		); err != nil {
			t.Fatalf("age played_at: %v", err)
		}

		err := sqlc.New(db).MarkAudioPlayed(ctx, sqlc.MarkAudioPlayedParams{
			AudioID: audioOne,
			UserID:  audioListUser,
		})
		if err != nil {
			t.Fatalf("MarkAudioPlayed: %v", err)
		}

		assertPlayedAtBumped(t, db, audioOne)
	})

	t.Run("bulk", func(t *testing.T) {
		t.Parallel()

		db := newAudioListDB(t)

		if _, err := db.ExecContext(ctx,
			"UPDATE audio_played SET played_at = '2020-01-01', completed = 0 WHERE audio_id = ? AND user_id = ?",
			audioOne, audioListUser,
		); err != nil {
			t.Fatalf("age played_at: %v", err)
		}

		err := sqlc.New(db).MarkAudioPlayedBulk(ctx, sqlc.MarkAudioPlayedBulkParams{
			Uid: audioListUser,
			Ids: []string{audioOne, audioTwo},
		})
		if err != nil {
			t.Fatalf("MarkAudioPlayedBulk: %v", err)
		}

		assertPlayedAtBumped(t, db, audioOne)
		assertPlayedAtBumped(t, db, audioTwo)
	})
}

func TestMarkAudioStartedPreservesCompletion(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("finished work stays finished", func(t *testing.T) {
		t.Parallel()

		db := newAudioListDB(t)

		if _, err := db.ExecContext(ctx,
			"UPDATE audio_played SET played_at = '2020-01-01' WHERE audio_id = ? AND user_id = ?",
			audioOne, audioListUser,
		); err != nil {
			t.Fatalf("age played_at: %v", err)
		}

		err := sqlc.New(db).MarkAudioStarted(ctx, sqlc.MarkAudioStartedParams{
			AudioID: audioOne,
			UserID:  audioListUser,
		})
		if err != nil {
			t.Fatalf("MarkAudioStarted: %v", err)
		}

		assertPlayedAtBumped(t, db, audioOne)
	})

	t.Run("new start is unfinished", func(t *testing.T) {
		t.Parallel()

		db := newAudioListDB(t)

		err := sqlc.New(db).MarkAudioStarted(ctx, sqlc.MarkAudioStartedParams{
			AudioID: audioTwo,
			UserID:  audioListUser,
		})
		if err != nil {
			t.Fatalf("MarkAudioStarted: %v", err)
		}

		var total, completed int
		if err := db.QueryRowContext(ctx,
			"SELECT count(*), coalesce(sum(completed = 1), 0) FROM audio_played WHERE audio_id = ? AND user_id = ?",
			audioTwo, audioListUser,
		).Scan(&total, &completed); err != nil {
			t.Fatalf("read play state: %v", err)
		}

		if total != 1 || completed != 0 {
			t.Errorf("new start state = rows:%d completed:%d, want rows:1 completed:0", total, completed)
		}
	})
}

func assertPlayedAtBumped(t *testing.T, db *sql.DB, audioID string) {
	t.Helper()

	var total, stale, completed int
	if err := db.QueryRowContext(t.Context(),
		`SELECT count(*),
			coalesce(sum(played_at = '2020-01-01'), 0),
			coalesce(sum(completed = 1), 0)
		FROM audio_played WHERE audio_id = ? AND user_id = ?`,
		audioID, audioListUser,
	).Scan(&total, &stale, &completed); err != nil {
		t.Fatalf("read played_at for %s: %v", audioID, err)
	}

	if total != 1 || stale != 0 || completed != 1 {
		t.Errorf(
			"%s played state wrong (rows=%d, stale=%d, completed=%d)",
			audioID, total, stale, completed,
		)
	}
}

// TestUpdateAudioCoverIndexKeepsMetadata locks the "presentation state, not
// metadata" contract: picking a cover must not resurface the work in
// "recently updated" or touch any metadata column.
func TestUpdateAudioCoverIndexKeepsMetadata(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	queries := sqlc.New(newAudioListDB(t))

	before, err := queries.GetAudioByID(ctx, audioOne)
	if err != nil {
		t.Fatalf("get audio before: %v", err)
	}

	if err := queries.UpdateAudioCoverIndex(ctx, sqlc.UpdateAudioCoverIndexParams{
		ID:         audioOne,
		CoverIndex: 2,
	}); err != nil {
		t.Fatalf("update cover index: %v", err)
	}

	after, err := queries.GetAudioByID(ctx, audioOne)
	if err != nil {
		t.Fatalf("get audio after: %v", err)
	}

	if after.CoverIndex != 2 {
		t.Errorf("cover_index = %d, want 2", after.CoverIndex)
	}

	if after.Title != before.Title {
		t.Errorf("title changed from %q to %q", before.Title, after.Title)
	}

	if !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Errorf("updated_at moved from %s to %s", before.UpdatedAt, after.UpdatedAt)
	}
}
