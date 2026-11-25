package handlers

import (
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// favoritesSchema covers the two favorite tables and the columns the mixed
// feed joins on.
const favoritesSchema = `
CREATE TABLE archive (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL,
    title TEXT NOT NULL,
    page_count INTEGER NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE favorite_page (
    archive_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    page_index INTEGER NOT NULL,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (archive_id, user_id, page_index)
);

CREATE TABLE audio (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL,
    title TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE audio_image (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_id TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL
);

CREATE TABLE audio_image_favorite (
    audio_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    image_index INTEGER NOT NULL,
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (audio_id, user_id, image_index)
);

CREATE TABLE library_access (
    user_id TEXT NOT NULL,
    library_id TEXT NOT NULL
);
`

const favoritesFixture = `
INSERT INTO archive (id, library_id, title, page_count) VALUES
    ('a1', 'lib1', 'Archive One', 5),
    ('a2', 'lib1', 'Archive Two', 5),
    ('a3', 'lib2', 'Archive Hidden', 3);

INSERT INTO favorite_page (archive_id, user_id, page_index, favorited_at) VALUES
    ('a1', 'user1', 1, '2026-01-01 10:00:00'),
    ('a2', 'user1', 0, '2026-01-01 12:00:00'),
    ('a1', 'user1', 9, '2026-01-01 15:00:00'),
    ('a3', 'user1', 0, '2026-01-01 13:00:00'),
    ('a1', 'user2', 2, '2026-01-01 17:00:00');

INSERT INTO audio (id, library_id, title) VALUES
    ('au1', 'lib1', 'Audio One'),
    ('au2', 'lib2', 'Audio Hidden');

INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES
    ('au1', 'a.png', 0),
    ('au1', 'b.png', 1),
    ('au2', 'c.png', 0);

INSERT INTO audio_image_favorite (audio_id, user_id, image_index, favorited_at) VALUES
    ('au1', 'user1', 1, '2026-01-01 12:00:00'),
    ('au1', 'user1', 5, '2026-01-01 16:00:00'),
    ('au2', 'user1', 0, '2026-01-01 14:00:00'),
    ('au1', 'user2', 0, '2026-01-01 18:00:00');

INSERT INTO library_access (user_id, library_id) VALUES ('user1', 'lib1');
`

func newFavoritesHandler(t *testing.T) *FavoriteHandler {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), favoritesSchema+favoritesFixture); err != nil {
		t.Fatalf("schema: %v", err)
	}

	return NewFavoriteHandler(sqlc.New(db), slog.New(slog.DiscardHandler))
}

func getFavoritesPage(t *testing.T, h *FavoriteHandler, page, limit string) FavoriteImageListResponse {
	t.Helper()

	ctx := auth.WithUserID(t.Context(), "user1")
	ctx = auth.WithLibraryScope(ctx, auth.LibraryScope{IDs: []string{"lib1"}})

	req := httptest.NewRequestWithContext(ctx, http.MethodGet,
		"/api/favorites/pages?page="+page+"&limit="+limit, nil)
	rec := httptest.NewRecorder()

	h.GetFavoritePages(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp FavoriteImageListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return resp
}

type favoriteWant struct {
	kind      string
	archiveID string
	audioID   string
	index     int
}

func assertFavoriteItem(t *testing.T, item FavoriteImageItem, want favoriteWant) {
	t.Helper()

	if item.Kind != want.kind || item.ArchiveID != want.archiveID ||
		item.AudioID != want.audioID || item.PageIndex != want.index {
		t.Errorf("item = %+v, want %s %s%s index %d",
			item, want.kind, want.archiveID, want.audioID, want.index)
	}
}

func assertFavoritePage(t *testing.T, h *FavoriteHandler, page string, total int, want favoriteWant) {
	t.Helper()

	resp := getFavoritesPage(t, h, page, "1")
	if resp.Total != int64(total) {
		t.Fatalf("page %s total = %d, want %d", page, resp.Total, total)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("page %s items = %d, want 1", page, len(resp.Items))
	}

	assertFavoriteItem(t, resp.Items[0], want)
}

func TestFavoritePagesMixedFeed(t *testing.T) {
	t.Parallel()

	h := newFavoritesHandler(t)

	first := getFavoritesPage(t, h, "1", "2")
	if first.Total != 3 {
		t.Fatalf("total = %d, want 3 (stale, out-of-scope and other users' rows excluded)", first.Total)
	}

	if len(first.Items) != 2 {
		t.Fatalf("page 1 items = %d, want 2", len(first.Items))
	}

	// Same-second tie: the archive page sorts before the audio image, and the
	// id/index tail makes the order stable across separate page requests.
	assertFavoriteItem(t, first.Items[0], favoriteWant{kind: "archive", archiveID: "a2"})
	assertFavoriteItem(t, first.Items[1], favoriteWant{kind: "audio", audioID: "au1", index: 1})

	if first.Items[1].UpdatedAt == "" {
		t.Error("audio feed item has no updated_at for cache-busting")
	}

	want := []favoriteWant{
		{kind: "archive", archiveID: "a2", index: 0},
		{kind: "audio", audioID: "au1", index: 1},
		{kind: "archive", archiveID: "a1", index: 1},
	}

	for i, w := range want {
		assertFavoritePage(t, h, strconv.Itoa(i+1), 3, w)
	}
}
