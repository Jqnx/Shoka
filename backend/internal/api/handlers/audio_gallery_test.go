package handlers

import (
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	stdimage "image"
	"image/color"
	"image/png"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// audioGallerySchema is the part of migrations 00017 and 00044 the gallery
// handlers read and write: the work and image tables, the relation tables
// SetCover's AudioResponse reads, and the job queue.
const audioGallerySchema = `
CREATE TABLE audio (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL,
    title TEXT NOT NULL,
    folder_path TEXT UNIQUE NOT NULL,
    mod_time DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    release_date DATETIME,
    language TEXT,
    cover_index INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE audio_image (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_id TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    UNIQUE (audio_id, relative_path)
);

CREATE TABLE audio_file (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_id TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL
);

CREATE TABLE artist (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE tag (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE circle (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE voice_actor (id INTEGER PRIMARY KEY, name TEXT NOT NULL);

CREATE TABLE audio_artist (audio_id TEXT NOT NULL, artist_id INTEGER NOT NULL);
CREATE TABLE audio_tag (audio_id TEXT NOT NULL, tag_id INTEGER NOT NULL);
CREATE TABLE audio_circle (audio_id TEXT NOT NULL, circle_id INTEGER NOT NULL);
CREATE TABLE audio_voice_actor (audio_id TEXT NOT NULL, voice_actor_id INTEGER NOT NULL);
CREATE TABLE audio_favorite (audio_id TEXT NOT NULL, user_id TEXT NOT NULL, favorited_at DATETIME NOT NULL);
CREATE TABLE audio_played (audio_id TEXT NOT NULL, user_id TEXT NOT NULL, played_at DATETIME NOT NULL, completed INTEGER NOT NULL DEFAULT 0);
CREATE TABLE audio_rating (audio_id TEXT NOT NULL, user_id TEXT NOT NULL, rating INTEGER NOT NULL, rated_at DATETIME NOT NULL);
CREATE TABLE audio_image_favorite (
    audio_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    image_index INTEGER NOT NULL CHECK (image_index >= 0),
    favorited_at DATETIME NOT NULL,
    PRIMARY KEY (audio_id, user_id, image_index)
);

CREATE TABLE job (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    payload TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending', 'running', 'done', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    error TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    run_after DATETIME NOT NULL DEFAULT (datetime('now'))
);
`

// audioGalleryFixture is one work with one image plus the job queue and
// broadcaster the gallery endpoints depend on.
type audioGalleryFixture struct {
	handler     *AudioHandler
	db          *sql.DB
	processor   *image.Processor
	folder      string
	broadcaster *events.ThumbnailBroadcaster
	queue       *jobs.Queue
}

// newAudioGalleryHandler seeds work "work1" with one 64x48 PNG.
func newAudioGalleryHandler(t *testing.T) audioGalleryFixture {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), audioGallerySchema); err != nil {
		t.Fatalf("schema: %v", err)
	}

	folder := t.TempDir()

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO audio (id, library_id, title, folder_path, mod_time) VALUES (?, ?, ?, ?, ?)`,
		"work1", "lib1", "work", folder, "2026-01-01 00:00:00"); err != nil {
		t.Fatalf("seed audio: %v", err)
	}

	img := stdimage.NewRGBA(stdimage.Rect(0, 0, 64, 48))

	for x := range 64 {
		for y := range 48 {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 0x40, A: 0xff})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode image: %v", err)
	}

	if err := os.WriteFile(filepath.Join(folder, "001.png"), buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write image: %v", err)
	}

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES (?, ?, ?)`,
		"work1", "001.png", 0); err != nil {
		t.Fatalf("seed image: %v", err)
	}

	logger := slog.New(slog.DiscardHandler)
	processor := image.NewProcessor(t.TempDir(), logger)

	cache, err := image.NewCache(processor, logger)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}

	queries := sqlc.New(db)
	broadcaster := events.NewThumbnailBroadcaster()
	queue := jobs.NewQueue(queries, logger)

	return audioGalleryFixture{
		handler:     NewAudioHandler(queries, db, processor, cache, queue, broadcaster, logger),
		db:          db,
		processor:   processor,
		folder:      folder,
		broadcaster: broadcaster,
		queue:       queue,
	}
}

// seedAudioGalleryThumb renders work1's only image and marks the gallery
// current, as the thumbnail job would.
func seedAudioGalleryThumb(t *testing.T, fx audioGalleryFixture) {
	t.Helper()

	src, err := os.Open(filepath.Join(fx.folder, "001.png"))
	if err != nil {
		t.Fatalf("open source image: %v", err)
	}

	err = fx.processor.GenerateAudioThumb(context.Background(), "work1", 0, src)
	_ = src.Close()

	if err != nil {
		t.Fatalf("generate thumbnail: %v", err)
	}

	if err := fx.processor.WriteAudioGalleryVersion("work1"); err != nil {
		t.Fatalf("write gallery version: %v", err)
	}
}

// audioGalleryRequest builds a request carrying chi's id and optional index
// path params.
func audioGalleryRequest(t *testing.T, method, path, id, index string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	if index != "" {
		rctx.URLParams.Add("index", index)
	}

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func audioGalleryGet(t *testing.T, id, index string) *http.Request {
	t.Helper()

	return audioGalleryRequest(t, http.MethodGet, "/api/audio/"+id+"/images/"+index, id, index)
}

func TestAudioGetImageThumbnail(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	before := httptest.NewRecorder()
	fx.handler.GetImageThumbnail(before, audioGalleryGet(t, "work1", "0"))

	if before.Code != http.StatusNotFound {
		t.Errorf("status before generation = %d, want 404", before.Code)
	}

	if got := before.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control before generation = %q, want no-store", got)
	}

	seedAudioGalleryThumb(t, fx)

	after := httptest.NewRecorder()
	fx.handler.GetImageThumbnail(after, audioGalleryGet(t, "work1", "0"))

	if after.Code != http.StatusOK {
		t.Fatalf("status after generation = %d, want 200", after.Code)
	}

	if got := after.Header().Get("Content-Type"); got != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", got)
	}

	if got := after.Header().Get("Cache-Control"); got != "private, max-age=86400" {
		t.Errorf("Cache-Control = %q, want private, max-age=86400", got)
	}

	if after.Body.Len() == 0 {
		t.Error("thumbnail body is empty")
	}

	outOfRange := httptest.NewRecorder()
	fx.handler.GetImageThumbnail(outOfRange, audioGalleryGet(t, "work1", "99"))

	if outOfRange.Code != http.StatusNotFound {
		t.Errorf("status for index 99 = %d, want 404", outOfRange.Code)
	}
}

func TestAudioGetImagePreview(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	first := httptest.NewRecorder()
	fx.handler.GetImagePreview(first, audioGalleryGet(t, "work1", "0"))

	if first.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200", first.Code)
	}

	if got := first.Header().Get("Content-Type"); got != "image/webp" {
		t.Errorf("preview Content-Type = %q, want image/webp", got)
	}

	if first.Body.Len() == 0 {
		t.Fatal("preview body is empty")
	}

	// The second request is served from the in-memory LRU: removing the source
	// must not turn it into a 404.
	if err := os.Remove(filepath.Join(fx.folder, "001.png")); err != nil {
		t.Fatalf("remove source image: %v", err)
	}

	second := httptest.NewRecorder()
	fx.handler.GetImagePreview(second, audioGalleryGet(t, "work1", "0"))

	if second.Code != http.StatusOK {
		t.Fatalf("cached preview status = %d, want 200", second.Code)
	}

	if !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
		t.Error("cached preview bytes differ from the rendered preview")
	}
}

func TestAudioGenerateImageThumbnails(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	started := httptest.NewRecorder()
	fx.handler.GenerateImageThumbnails(started,
		audioGalleryRequest(t, http.MethodPost, "/api/audio/work1/thumbnails", "work1", ""))

	if started.Code != http.StatusAccepted {
		t.Errorf("status for a gallery without thumbs = %d, want 202", started.Code)
	}

	pending, err := fx.queue.HasPendingJob(context.Background(), jobs.JobTypeAudioThumbnail,
		jobs.AudioThumbnailPayload{AudioID: "work1"})
	if err != nil {
		t.Fatalf("check pending job: %v", err)
	}

	if !pending {
		t.Error("no audio thumbnail job was enqueued")
	}

	seedAudioGalleryThumb(t, fx)

	ready := httptest.NewRecorder()
	fx.handler.GenerateImageThumbnails(ready,
		audioGalleryRequest(t, http.MethodPost, "/api/audio/work1/thumbnails", "work1", ""))

	if ready.Code != http.StatusNoContent {
		t.Errorf("status for a complete gallery = %d, want 204", ready.Code)
	}
}

// runAudioStream runs the events handler on its own goroutine and fails the
// test instead of hanging when the stream never terminates. The deadline
// matters for the early-return cases: a regression that removes the terminal
// event would otherwise block the whole package until go test's timeout.
func runAudioStream(t *testing.T, handler *AudioHandler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	req = req.WithContext(ctx)

	finished := make(chan struct{})

	go func() {
		defer close(finished)

		handler.StreamImageThumbnailEvents(rec, req)
	}()

	select {
	case <-finished:
		return rec
	case <-time.After(2 * time.Second):
		cancel()
		<-finished

		t.Fatal("stream did not finish within 2s")

		return nil
	}
}

func TestAudioStreamImageThumbnailEventsReady(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)
	seedAudioGalleryThumb(t, fx)

	rec := runAudioStream(t, fx.handler,
		audioGalleryRequest(t, http.MethodGet, "/api/audio/work1/thumbnails/events", "work1", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()

	for _, want := range []string{"event: ready", `"index":0`, "event: done", `"done":true`} {
		if !strings.Contains(body, want) {
			t.Errorf("stream body missing %q:\n%s", want, body)
		}
	}

	if strings.Contains(body, `"error"`) {
		t.Errorf("ready stream reports an error:\n%s", body)
	}
}

func TestAudioStreamImageThumbnailEventsEmptyGallery(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	// A work with no images has nothing to generate; no marker exists because
	// no job was ever enqueued.
	if _, err := fx.db.ExecContext(t.Context(), "DELETE FROM audio_image"); err != nil {
		t.Fatalf("clear images: %v", err)
	}

	rec := runAudioStream(t, fx.handler,
		audioGalleryRequest(t, http.MethodGet, "/api/audio/work1/thumbnails/events", "work1", ""))

	body := rec.Body.String()

	if !strings.Contains(body, `"done":true`) {
		t.Errorf("stream for a work without images is not done:\n%s", body)
	}

	if strings.Contains(body, `"error"`) {
		t.Errorf("stream for a work without images reports an error:\n%s", body)
	}
}

func TestAudioStreamImageThumbnailEventsReportsMissingJob(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	rec := runAudioStream(t, fx.handler,
		audioGalleryRequest(t, http.MethodGet, "/api/audio/work1/thumbnails/events", "work1", ""))

	body := rec.Body.String()

	if !strings.Contains(body, "event: done") {
		t.Fatalf("stream for a work with no generation running has no done event:\n%s", body)
	}

	if !strings.Contains(body, "thumbnail generation is not running") {
		t.Errorf("done event does not explain the missing job:\n%s", body)
	}
}

func TestAudioStreamImageThumbnailEventsForwardsTerminalEvent(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	// A pending job is what keeps the stream waiting instead of reporting that
	// generation is not running.
	if err := fx.queue.Enqueue(context.Background(), jobs.JobTypeAudioThumbnail,
		jobs.AudioThumbnailPayload{AudioID: "work1"}); err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	rec := httptest.NewRecorder()
	finished := make(chan struct{})

	req := audioGalleryRequest(t, http.MethodGet, "/api/audio/work1/thumbnails/events", "work1", "")

	ctx, cancel := context.WithCancel(req.Context())
	t.Cleanup(cancel)

	req = req.WithContext(ctx)

	go func() {
		defer close(finished)

		fx.handler.StreamImageThumbnailEvents(rec, req)
	}()

	key := image.AudioCacheKey("work1")
	deadline := time.After(2 * time.Second)

	for {
		select {
		case <-finished:
			body := rec.Body.String()

			if !strings.Contains(body, "event: done") || !strings.Contains(body, `"error":"boom"`) {
				t.Errorf("stream did not forward the terminal event:\n%s", body)
			}

			return
		case <-deadline:
			cancel()
			<-finished

			t.Fatal("stream did not finish after a terminal event was published")
		case <-time.After(20 * time.Millisecond):
			// The subscription is installed before the snapshot, but publishing
			// repeatedly tolerates the gap between starting the goroutine and
			// reaching Subscribe.
			fx.broadcaster.Publish(key, events.ThumbnailEvent{Done: true, Error: "boom"})
		}
	}
}

// seedAudioImage writes an extra PNG of the given size under the work and
// appends its image row, in the scanner's sort order.
func seedAudioImage(t *testing.T, fx audioGalleryFixture, name string, sortOrder, width, height int) {
	t.Helper()

	img := stdimage.NewRGBA(stdimage.Rect(0, 0, width, height))

	for x := range width {
		for y := range height {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 0x80, A: 0xff})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode image: %v", err)
	}

	if err := os.WriteFile(filepath.Join(fx.folder, name), buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write image: %v", err)
	}

	if _, err := fx.db.ExecContext(t.Context(),
		`INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES (?, ?, ?)`,
		"work1", name, sortOrder); err != nil {
		t.Fatalf("seed image: %v", err)
	}
}

// audioSetCoverRequest builds a PUT with the work id and a JSON body.
func audioSetCoverRequest(t *testing.T, id, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut,
		"/api/audio/"+id+"/cover", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestAudioSetCover(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)
	// The second image is wider than the first, so the rendered cover's size
	// proves the selected image was used rather than image 0.
	seedAudioImage(t, fx, "002.png", 1, 120, 80)

	rec := httptest.NewRecorder()
	fx.handler.SetCover(rec, audioSetCoverRequest(t, "work1", `{"index":1}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp AudioResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.CoverIndex != 1 {
		t.Errorf("response cover_index = %d, want 1", resp.CoverIndex)
	}

	if !resp.CoverReady || resp.CoverWidth != 120 || resp.CoverHeight != 80 {
		t.Errorf("response cover = ready:%t %dx%d, want ready 120x80",
			resp.CoverReady, resp.CoverWidth, resp.CoverHeight)
	}

	var stored int64
	if err := fx.db.QueryRowContext(t.Context(),
		`SELECT cover_index FROM audio WHERE id = 'work1'`).Scan(&stored); err != nil {
		t.Fatalf("read cover_index: %v", err)
	}

	if stored != 1 {
		t.Errorf("stored cover_index = %d, want 1", stored)
	}

	if cover := fx.processor.CachedAudioCover("work1"); cover.Width != 120 || cover.Height != 80 {
		t.Errorf("rendered cover = %dx%d, want 120x80", cover.Width, cover.Height)
	}
}

func TestAudioSetCoverRejectsIndexOutOfRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "negative", body: `{"index":-1}`},
		{name: "past last image", body: `{"index":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newAudioGalleryHandler(t)

			rec := httptest.NewRecorder()
			fx.handler.SetCover(rec, audioSetCoverRequest(t, "work1", tt.body))

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}

			var stored int64
			if err := fx.db.QueryRowContext(t.Context(),
				`SELECT cover_index FROM audio WHERE id = 'work1'`).Scan(&stored); err != nil {
				t.Fatalf("read cover_index: %v", err)
			}

			if stored != 0 {
				t.Errorf("stored cover_index = %d, want 0 after a rejected request", stored)
			}

			if cover := fx.processor.CachedAudioCover("work1"); cover.Path != "" {
				t.Errorf("rejected request rendered a cover at %s", cover.Path)
			}
		})
	}
}

func TestAudioSetCoverMissingImage(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	// Start from a rendered cover so the test can prove a failed render leaves
	// the existing one untouched instead of evicting it.
	src, err := os.Open(filepath.Join(fx.folder, "001.png"))
	if err != nil {
		t.Fatalf("open source image: %v", err)
	}

	err = fx.processor.GenerateAudioCover(context.Background(), "work1", src)
	_ = src.Close()

	if err != nil {
		t.Fatalf("generate cover: %v", err)
	}

	before := fx.processor.CachedAudioCover("work1")
	if before.Path == "" || before.ModTime == nil {
		t.Fatalf("cover was not rendered: %+v", before)
	}

	// A row whose file is gone (deleted on disk before the next scan).
	if _, err := fx.db.ExecContext(t.Context(),
		`INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES (?, ?, ?)`,
		"work1", "missing.png", 1); err != nil {
		t.Fatalf("seed image row: %v", err)
	}

	rec := httptest.NewRecorder()
	fx.handler.SetCover(rec, audioSetCoverRequest(t, "work1", `{"index":1}`))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}

	var stored int64
	if err := fx.db.QueryRowContext(t.Context(),
		`SELECT cover_index FROM audio WHERE id = 'work1'`).Scan(&stored); err != nil {
		t.Fatalf("read cover_index: %v", err)
	}

	if stored != 0 {
		t.Errorf("stored cover_index = %d, want 0 after a missing image", stored)
	}

	after := fx.processor.CachedAudioCover("work1")
	if after.Path != before.Path || after.ModTime == nil || !after.ModTime.Equal(*before.ModTime) {
		t.Errorf("existing cover changed after a failed request: before=%+v after=%+v", before, after)
	}
}

// favoriteImagesRequest adds the caller identity the favorite handlers read
// from the request context.
func favoriteImagesRequest(t *testing.T, method, id, index, uid string) *http.Request {
	t.Helper()

	req := audioGalleryRequest(t, method, "/api/audio/"+id, id, index)

	return req.WithContext(auth.WithUserID(req.Context(), uid))
}

func favoriteImageIndices(t *testing.T, fx audioGalleryFixture, uid string) []int32 {
	t.Helper()

	rec := httptest.NewRecorder()
	fx.handler.GetFavoriteImageIndices(rec, favoriteImagesRequest(t, http.MethodGet, "work1", "", uid))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp AudioImageFavoriteIndicesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return resp.Indices
}

func addFavoriteImage(t *testing.T, fx audioGalleryFixture, index, uid string) {
	t.Helper()

	rec := httptest.NewRecorder()
	fx.handler.AddFavoriteImage(rec, favoriteImagesRequest(t, http.MethodPut, "work1", index, uid))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("add status = %d, want 204: %s", rec.Code, rec.Body.String())
	}
}

func removeFavoriteImage(t *testing.T, fx audioGalleryFixture, index, uid string) {
	t.Helper()

	rec := httptest.NewRecorder()
	fx.handler.RemoveFavoriteImage(rec, favoriteImagesRequest(t, http.MethodDelete, "work1", index, uid))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("remove status = %d, want 204: %s", rec.Code, rec.Body.String())
	}
}

func TestAudioFavoriteImages(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)
	seedAudioImage(t, fx, "002.png", 1, 120, 80)

	const (
		user  = "user1"
		other = "user2"
	)

	if got := favoriteImageIndices(t, fx, user); len(got) != 0 {
		t.Fatalf("initial favorites = %v, want none", got)
	}

	// Idempotent add.
	for range 2 {
		addFavoriteImage(t, fx, "1", user)
	}

	if got := favoriteImageIndices(t, fx, user); len(got) != 1 || got[0] != 1 {
		t.Errorf("favorites for the caller = %v, want [1]", got)
	}

	// Favorites are per-user.
	if got := favoriteImageIndices(t, fx, other); len(got) != 0 {
		t.Errorf("favorites for another user = %v, want none", got)
	}

	// Another user's favorite is invisible and cannot be cleared by the caller.
	addFavoriteImage(t, fx, "0", other)

	if got := favoriteImageIndices(t, fx, other); len(got) != 1 || got[0] != 0 {
		t.Errorf("favorites for another user = %v, want [0]", got)
	}

	removeFavoriteImage(t, fx, "0", user)

	if got := favoriteImageIndices(t, fx, other); len(got) != 1 || got[0] != 0 {
		t.Errorf("another user's favorite was removed: %v, want [0]", got)
	}

	removeFavoriteImage(t, fx, "1", user)

	if got := favoriteImageIndices(t, fx, user); len(got) != 0 {
		t.Errorf("favorites after removal = %v, want none", got)
	}

	if got := favoriteImageIndices(t, fx, other); len(got) != 1 || got[0] != 0 {
		t.Errorf("another user's favorite changed during removal: %v, want [0]", got)
	}
}

func TestAudioFavoriteImageRejectsInvalidIndex(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	tests := []struct {
		name  string
		index string
		want  int
	}{
		{name: "negative", index: "-1", want: http.StatusBadRequest},
		{name: "malformed", index: "nope", want: http.StatusBadRequest},
		{name: "past the last image", index: "1", want: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			fx.handler.AddFavoriteImage(rec,
				favoriteImagesRequest(t, http.MethodPut, "work1", tt.index, "user1"))

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d: %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAudioRemoveFavoriteImageAllowsStaleIndex(t *testing.T) {
	t.Parallel()

	fx := newAudioGalleryHandler(t)

	// Removing an index the work no longer has must stay possible so a stale
	// favorite can be cleared.
	rec := httptest.NewRecorder()
	fx.handler.RemoveFavoriteImage(rec,
		favoriteImagesRequest(t, http.MethodDelete, "work1", "99", "user1"))

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204: %s", rec.Code, rec.Body.String())
	}
}
