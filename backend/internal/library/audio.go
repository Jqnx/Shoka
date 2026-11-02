package library

import (
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/media"
	"Shoka/internal/util"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var (
	audioExtensions = []string{".mp3", ".m4a", ".aac", ".flac", ".wav", ".ogg", ".opus"}
	imageExtensions = []string{".jpg", ".jpeg", ".png", ".webp", ".avif"}
)

// probeTimeout bounds one ffprobe call. A file that ffprobe cannot answer for
// quickly records a zero duration rather than stalling the whole scan.
const probeTimeout = 30 * time.Second

// audioIDAttempts is how many times a colliding generated audio id is retried
// before the add is reported as failed.
const audioIDAttempts = 5

// AudioScanner is the Scanner implementation for libraries of type "audio":
// every immediate child directory of the library root is one work, and a work
// is the audio/image files found recursively inside it.
type AudioScanner struct {
	cfg     *config.Config
	queries *sqlc.Queries
	db      *sql.DB
	queue   JobQueue
	log     *slog.Logger
	// probe is media.Duration in production; tests can substitute it.
	probe func(context.Context, string) (time.Duration, error)
}

// NewAudioScanner returns the Scanner for audio libraries.
func NewAudioScanner(cfg *config.Config, queries *sqlc.Queries, db *sql.DB, queue JobQueue, log *slog.Logger) *AudioScanner {
	return &AudioScanner{
		cfg:     cfg,
		queries: queries,
		db:      db,
		queue:   queue,
		log:     log.With("component", "audio_scanner"),
		probe:   media.Duration,
	}
}

// audioWork is one scanned work directory: its relative audio and image
// paths, sorted ascending, and the newest mtime among its files.
type audioWork struct {
	folderPath string
	audio      []string
	images     []string
	modTime    time.Time
}

// Scan reconciles lib's immediate child directories with the audio table.
// A work that cannot be read is skipped entirely so a partial listing can
// never truncate an intact work.
func (s *AudioScanner) Scan(ctx context.Context, lib sqlc.Library) error {
	s.log.Info("audio library scan started", "library_id", lib.ID, "library", lib.Name)

	start := time.Now()

	works, skipped, err := s.scanWorks(lib.Path)
	if err != nil {
		// Without a readable root there is no trustworthy "missing" set, so
		// nothing is removed - the next scan retries.
		s.log.Warn("audio library root unreadable, skipping scan", "path", lib.Path, "error", err)
		return nil
	}

	added, updated, removed, err := s.process(ctx, lib, works, skipped)
	if err != nil {
		return err
	}

	s.log.Info("audio library scan complete",
		"library_id", lib.ID,
		"duration", time.Since(start),
		"added", added,
		"updated", updated,
		"removed", removed,
	)

	return nil
}

// scanWorks reads every immediate child directory of root. A child that fails
// to scan is logged at warn and recorded as skipped: it contributes no work
// (so nothing is added or updated) and process must not treat it as missing
// either, or a partial listing would truncate an intact work.
func (s *AudioScanner) scanWorks(root string) (map[string]audioWork, map[string]struct{}, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}

	works := make(map[string]audioWork)
	skipped := make(map[string]struct{})

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(root, name)

		work, err := s.scanWork(path)
		if err != nil {
			s.log.Warn("audio work skipped", "path", path, "error", err)
			skipped[path] = struct{}{}

			continue
		}

		if len(work.audio) == 0 {
			continue
		}

		works[path] = work
	}

	return works, skipped, nil
}

// scanWork walks one work directory, collecting audio and image files and the
// newest mtime. Every filesystem error is returned: the caller skips the work
// rather than treating the partial result as a complete one.
func (s *AudioScanner) scanWork(root string) (audioWork, error) {
	work := audioWork{folderPath: root}

	var maxMod time.Time

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.ModTime().After(maxMod) {
			maxMod = info.ModTime()
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		switch {
		case slices.Contains(audioExtensions, ext):
			work.audio = append(work.audio, rel)
		case slices.Contains(imageExtensions, ext):
			work.images = append(work.images, rel)
		}

		return nil
	})
	if err != nil {
		return audioWork{}, err
	}

	slices.SortFunc(work.audio, naturalCompare)
	// Images keep lexicographic order: the first one is the work's cover, so
	// changing their order would silently change covers and gallery order.
	slices.Sort(work.images)
	work.modTime = maxMod

	return work, nil
}

// naturalCompare orders paths by numeric runs, so "第2章" sorts before
// "第10章" instead of lexicographically after it.
func naturalCompare(a, b string) int {
	switch {
	case util.NaturalLess(a, b):
		return -1
	case util.NaturalLess(b, a):
		return 1
	default:
		return 0
	}
}

// process adds, updates and removes audio works so the library matches disk.
// A path in skipped is left completely alone: it is neither missing nor
// scannable, so its stored rows stay as they were.
func (s *AudioScanner) process(ctx context.Context, lib sqlc.Library, works map[string]audioWork, skipped map[string]struct{}) (added, updated, removed int, err error) {
	existing, err := s.queries.GetAudioPathsByLibrary(ctx, lib.ID)
	if err != nil {
		return 0, 0, 0, err
	}

	inDB := make(map[string]sqlc.GetAudioPathsByLibraryRow, len(existing))
	for _, row := range existing {
		inDB[row.FolderPath] = row
	}

	for path, work := range works {
		record, known := inDB[path]

		addedHere, updatedHere, err := s.processWork(ctx, lib, record, known, work)
		if err != nil {
			s.log.Warn("audio work skipped", "path", path, "error", err)
			continue
		}

		added += addedHere
		updated += updatedHere
	}

	for path, record := range inDB {
		if _, exists := works[path]; exists {
			continue
		}

		if _, uncertain := skipped[path]; uncertain {
			continue
		}

		if err := s.removeAudio(ctx, record.ID); err != nil {
			s.log.Warn("failed to remove audio work", "path", path, "error", err)
			continue
		}

		removed++
	}

	return added, updated, removed, nil
}

// processWork reconciles one scanned work, reporting how many rows it added
// and updated. The caller logs failures per path.
func (s *AudioScanner) processWork(
	ctx context.Context,
	lib sqlc.Library,
	record sqlc.GetAudioPathsByLibraryRow,
	known bool,
	work audioWork,
) (added, updated int, err error) {
	change, err := s.audioChangeKind(ctx, record, known, work)
	if err != nil {
		return 0, 0, err
	}

	if !known {
		return 1, 0, s.addAudio(ctx, lib, work)
	}

	switch change {
	case audioChanged:
		if err := s.updateAudio(ctx, record.ID, work); err != nil {
			return 0, 0, err
		}

		return 0, 1, nil
	case audioReordered:
		if err := s.reorderAudio(ctx, record.ID, work.audio); err != nil {
			return 0, 0, err
		}

		return 0, 1, nil
	case audioUnchanged:
		// Unchanged files, but the cached cover may be missing or rendered
		// with older settings: re-render it without touching the work's rows.
		s.enqueueStaleCover(ctx, record.ID, work)

		return 0, 0, nil
	}

	return 0, 0, nil
}

// audioChange describes how a scanned work differs from its stored rows.
type audioChange int

const (
	audioUnchanged audioChange = iota
	// audioReordered has the same files but a different track order, which is
	// what a sort-rule change produces. Durations and the cover stay valid.
	audioReordered
	// audioChanged means content, files, images or the fingerprint differ, so
	// the work needs a full re-probe and cover refresh.
	audioChanged
)

// audioChangeKind compares a stored work with its scanned counterpart.
func (s *AudioScanner) audioChangeKind(ctx context.Context, record sqlc.GetAudioPathsByLibraryRow, known bool, work audioWork) (audioChange, error) {
	if !known {
		return audioChanged, nil
	}

	if !record.ModTime.Equal(work.modTime) {
		return audioChanged, nil
	}

	files, err := s.queries.GetAudioFiles(ctx, record.ID)
	if err != nil {
		return audioUnchanged, err
	}

	images, err := s.queries.GetAudioImages(ctx, record.ID)
	if err != nil {
		return audioUnchanged, err
	}

	if !slices.Equal(imagePaths(images), work.images) {
		return audioChanged, nil
	}

	return classifyFiles(filePaths(files), work.audio), nil
}

// classifyFiles distinguishes a pure track reorder from a real content change.
func classifyFiles(stored, scanned []string) audioChange {
	if slices.Equal(stored, scanned) {
		return audioUnchanged
	}

	if sameFileSet(stored, scanned) {
		return audioReordered
	}

	return audioChanged
}

func sameFileSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := slices.Clone(a)
	sortedB := slices.Clone(b)

	slices.Sort(sortedA)
	slices.Sort(sortedB)

	return slices.Equal(sortedA, sortedB)
}

// reorderAudio rewrites a work's track order without re-probing durations or
// invalidating its cover, which only the set of files would change.
func (s *AudioScanner) reorderAudio(ctx context.Context, audioID string, order []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	qtx := s.queries.WithTx(tx)

	for i, relativePath := range order {
		if err := qtx.UpdateAudioFileSortOrder(ctx, sqlc.UpdateAudioFileSortOrderParams{
			SortOrder:    int64(i),
			AudioID:      audioID,
			RelativePath: relativePath,
		}); err != nil {
			return fmt.Errorf("update sort order: %w", err)
		}
	}

	// Track URLs embed updated_at; changing the order remaps each index to a
	// different file, so the version has to change with it.
	if err := qtx.TouchAudio(ctx, audioID); err != nil {
		return fmt.Errorf("touch audio: %w", err)
	}

	return tx.Commit()
}

// enqueueStaleCover re-renders an unchanged work's cover when the cache is
// missing or predates the current render settings.
func (s *AudioScanner) enqueueStaleCover(ctx context.Context, audioID string, work audioWork) {
	if len(work.images) == 0 || !image.AudioCoverStale(s.cfg.Cache.Dir, audioID) {
		return
	}

	if err := s.queue.EnqueueOnce(ctx, jobs.JobTypeAudioCover, jobs.AudioCoverPayload{AudioID: audioID}); err != nil {
		s.log.Warn("failed to enqueue audio cover", "audio_id", audioID, "error", err)
	}
}

func filePaths(files []sqlc.AudioFile) []string {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.RelativePath
	}

	return paths
}

func imagePaths(images []sqlc.AudioImage) []string {
	paths := make([]string, len(images))
	for i, img := range images {
		paths[i] = img.RelativePath
	}

	return paths
}

// addAudio probes the work's files, inserts the row with its files and images,
// and queues cover generation when it has an image.
func (s *AudioScanner) addAudio(ctx context.Context, lib sqlc.Library, work audioWork) error {
	durations := s.probeFiles(ctx, work)

	audioID, err := s.createAudio(ctx, lib, work, durations)
	if err != nil {
		return err
	}

	if err := s.writeFiles(ctx, s.queries, audioID, work, durations); err != nil {
		return err
	}

	if len(work.images) > 0 {
		if err := s.queue.Enqueue(ctx, jobs.JobTypeAudioCover, jobs.AudioCoverPayload{AudioID: audioID}); err != nil {
			return err
		}
	}

	s.log.Info("audio work added", "path", work.folderPath, "id", audioID, "library_id", lib.ID)

	return nil
}

// createAudio inserts the work row under a fresh id, retrying generated-id
// collisions. Unlike ArchiveScanner.addArchive it returns the final error
// instead of silently continuing with a zero-value row.
func (s *AudioScanner) createAudio(ctx context.Context, lib sqlc.Library, work audioWork, durations []time.Duration) (string, error) {
	var total time.Duration
	for _, d := range durations {
		total += d
	}

	var lastErr error

	for range audioIDAttempts {
		id, err := util.GenerateID()
		if err != nil {
			return "", err
		}

		_, err = s.queries.CreateAudio(ctx, sqlc.CreateAudioParams{
			ID:              id,
			LibraryID:       lib.ID,
			Title:           filepath.Base(work.folderPath),
			FolderPath:      work.folderPath,
			ModTime:         work.modTime,
			DurationSeconds: int64(total / time.Second),
		})
		if err == nil {
			return id, nil
		}

		lastErr = err

		if !database.IsUniqueConstraintError(err) {
			return "", err
		}

		s.log.Warn("audio id collision, retrying", "path", work.folderPath)
	}

	return "", fmt.Errorf("create audio for %s: %w", work.folderPath, lastErr)
}

// probeFiles returns one duration per audio file, in order. A failure is
// logged and recorded as zero; a missing ffprobe therefore only costs the
// durations, not the scan.
func (s *AudioScanner) probeFiles(ctx context.Context, work audioWork) []time.Duration {
	durations := make([]time.Duration, len(work.audio))

	for i, rel := range work.audio {
		probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)

		d, err := s.probe(probeCtx, filepath.Join(work.folderPath, rel))
		if err != nil {
			s.log.Warn("probe audio duration failed", "path", rel, "work", work.folderPath, "error", err)
		} else {
			durations[i] = d
		}

		cancel()
	}

	return durations
}

// writeFiles replaces the work's file and image rows with the scanned lists,
// using the durations already probed for this scan. q is the caller's query
// set so an update can keep the file and work-level writes in one transaction.
func (s *AudioScanner) writeFiles(ctx context.Context, q *sqlc.Queries, audioID string, work audioWork, durations []time.Duration) error {
	if err := q.DeleteAudioFiles(ctx, audioID); err != nil {
		return err
	}

	if err := q.DeleteAudioImages(ctx, audioID); err != nil {
		return err
	}

	for i, rel := range work.audio {
		if err := q.CreateAudioFile(ctx, sqlc.CreateAudioFileParams{
			AudioID:         audioID,
			RelativePath:    rel,
			DurationSeconds: int64(durations[i] / time.Second),
			SortOrder:       int64(i),
		}); err != nil {
			return err
		}
	}

	for i, rel := range work.images {
		if err := q.CreateAudioImage(ctx, sqlc.CreateAudioImageParams{
			AudioID:      audioID,
			RelativePath: rel,
			SortOrder:    int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// updateAudio re-probes and replaces the files, updates the fingerprint and
// duration, evicts the old cover, and re-queues cover generation. Title and
// taxonomy relations are deliberately untouched.
func (s *AudioScanner) updateAudio(ctx context.Context, audioID string, work audioWork) error {
	durations := s.probeFiles(ctx, work)

	var total int64
	for _, d := range durations {
		total += int64(d / time.Second)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	qtx := s.queries.WithTx(tx)

	if err := s.writeFiles(ctx, qtx, audioID, work, durations); err != nil {
		return err
	}

	if err := qtx.UpdateAudioMeta(ctx, sqlc.UpdateAudioMetaParams{
		ID:              audioID,
		ModTime:         work.modTime,
		DurationSeconds: total,
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	_ = os.RemoveAll(filepath.Join(s.cfg.Cache.Dir, image.AudioCacheKey(audioID)))

	if len(work.images) > 0 {
		// Enqueue rather than EnqueueOnce: a cover job that was already
		// running read the pre-update image list, and EnqueueOnce would treat
		// it as pending and suppress the replacement.
		if err := s.queue.Enqueue(ctx, jobs.JobTypeAudioCover, jobs.AudioCoverPayload{AudioID: audioID}); err != nil {
			return err
		}
	}

	s.log.Info("audio work updated", "id", audioID)

	return nil
}

// removeAudio deletes a work's rows with the counter-maintaining helper, then
// its cache directory.
func (s *AudioScanner) removeAudio(ctx context.Context, audioID string) error {
	if err := database.DeleteAudio(ctx, s.queries, s.db, audioID); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(s.cfg.Cache.Dir, image.AudioCacheKey(audioID)))
}
