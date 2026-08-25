package library

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"Shoka/internal/library/archive"
	"Shoka/internal/util"
)

var supportedExtensions = []string{".cbz", ".cbr", ".zip", ".rar", ".7z"}

type JobQueue interface {
	Enqueue(ctx context.Context, jobType string, payload any) error
	EnqueueOnce(ctx context.Context, jobType string, payload any) error
}

// Scanner scans a single library's root directory and reconciles it with
// the database. Different library types (doujinshi, audio, ...) get their
// own Scanner implementation, selected by the LibraryManager based on
// library.Type.
type Scanner interface {
	Scan(ctx context.Context, lib sqlc.Library) error
}

// ArchiveScanner is the Scanner implementation for libraries of type
// "doujinshi" — image archives (cbz/cbr/zip/rar/7z).
type ArchiveScanner struct {
	cfg     *config.Config
	queries *sqlc.Queries
	queue   JobQueue
	log     *slog.Logger
}

func NewArchiveScanner(cfg *config.Config, queries *sqlc.Queries, queue JobQueue, log *slog.Logger) *ArchiveScanner {
	return &ArchiveScanner{
		cfg:     cfg,
		queries: queries,
		queue:   queue,
		log:     log.With("component", "scanner"),
	}
}

// Scan scans lib's root directory and queues up new jobs for any new or
// updated files, scoping all comparisons to this library so other
// libraries' archives are never touched.
func (s *ArchiveScanner) Scan(ctx context.Context, lib sqlc.Library) error {
	s.log.Info("library scan started", "library_id", lib.ID, "library", lib.Name)
	start := time.Now()

	found := make(map[string]fs.FileInfo) // path → FileInfo

	if err := s.walk(ctx, lib.Path, found); err != nil {
		s.log.Error("walk failed", "dir", lib.Path, "error", err)
	}

	added, updated, removed, err := s.process(ctx, lib, found)
	if err != nil {
		return err
	}

	s.log.Info("library scan complete",
		"library_id", lib.ID,
		"duration", time.Since(start),
		"added", added,
		"updated", updated,
		"removed", removed,
	)
	return nil
}

// walk() recursively walks a directory and adds any files it finds to the found map
func (s *ArchiveScanner) walk(ctx context.Context, root string, found map[string]fs.FileInfo) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.log.Warn("could not access path", "path", path, "error", err)
			return nil
		}

		// skip hidden directories (e.g. .DS_Store folders, .git)
		if d.IsDir() && len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !slices.Contains(supportedExtensions, ext) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			s.log.Warn("could not stat file", "path", path, "error", err)
			return nil
		}

		found[path] = info
		return nil
	})
}

// process() processes the found files, adding, updating or removing them.
// The diff is scoped to lib's archives only — files belonging to other
// libraries are never considered "missing" here.
func (s *ArchiveScanner) process(ctx context.Context, lib sqlc.Library, found map[string]fs.FileInfo) (added, updated, removed int, err error) {
	existing, err := s.queries.GetArchiveFilePathsByLibrary(ctx, lib.ID)
	if err != nil {
		return 0, 0, 0, err
	}

	// build a lookup map: path → db record
	inDB := make(map[string]sqlc.GetArchiveFilePathsByLibraryRow, len(existing))
	for _, a := range existing {
		inDB[a.FilePath] = a
	}

	// handle files found on disk
	for path, info := range found {
		record, known := inDB[path]

		if !known {
			if err := s.addArchive(ctx, lib, path, info); err != nil {
				s.log.Error("failed to add archive", "path", path, "error", err)
				continue
			}
			added++
			continue
		}

		if s.hasChanged(record, info) {
			if err := s.updateArchive(ctx, record.ID, info); err != nil {
				s.log.Error("failed to update archive", "path", path, "error", err)
				continue
			}
			updated++
		}
	}

	for path, record := range inDB {
		if _, exists := found[path]; !exists {
			if err := s.removeArchive(ctx, record.ID, path); err != nil {
				s.log.Error("failed to remove archive", "path", path, "error", err)
				continue
			}
			removed++
		}
	}

	return added, updated, removed, nil
}

// hasChanged() checks if an archive has changed
func (s *ArchiveScanner) hasChanged(record sqlc.GetArchiveFilePathsByLibraryRow, info fs.FileInfo) bool {
	return record.FileSize != info.Size() ||
		!record.ModTime.Equal(info.ModTime())
}

// addArchive() adds a new archive and queues up necessary jobs
func (s *ArchiveScanner) addArchive(ctx context.Context, lib sqlc.Library, path string, info fs.FileInfo) error {
	var arch sqlc.Archive

	a, err := archive.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}

	pages, err := a.Pages()
	a.Close()
	if err != nil {
		return fmt.Errorf("list pages: %w", err)
	}

	for range 5 {
		id, err := util.GenerateID()
		if err != nil {
			return err
		}

		arch, err = s.queries.CreateArchive(ctx, sqlc.CreateArchiveParams{
			ID:        id,
			LibraryID: lib.ID,
			Title:     util.StripExtension(filepath.Base(path)),
			FilePath:  path,
			FileSize:  info.Size(),
			ModTime:   info.ModTime(),
			PageCount: int64(len(pages)),
		})

		if err == nil {
			break
		}

		if !database.IsUniqueConstraintError(err) {
			return err
		}

		s.log.Warn("archive id collision, retrying", "path", path)
	}

	if err := s.queue.Enqueue(ctx, jobs.JobTypeCover, jobs.CoverPayload{
		ArchiveID: arch.ID,
		FilePath:  path,
	}); err != nil {
		return err
	}

	if err := s.queue.Enqueue(ctx, jobs.JobTypePHash, jobs.PHashPayload{
		ArchiveID: arch.ID,
		FilePath:  path,
	}); err != nil {
		return err
	}

	// Per-page thumbnails are generated on demand (see the
	// POST /api/archives/{id}/thumbnails endpoint), not eagerly at scan
	// time — a library scan can touch thousands of archives at once and
	// most of them won't be opened for a while, if ever.

	if err := s.queue.Enqueue(ctx, jobs.JobTypeMetadata, jobs.MetadataPayload{
		ArchiveID: arch.ID,
	}); err != nil {
		return err
	}

	// Search indexing needs no job of its own - archive_fts is kept in sync
	// by SQL triggers on insert/update/delete (see migration 00028).

	s.log.Info("archive added", "path", path, "id", arch.ID, "library_id", lib.ID)
	return nil
}

// updateArchive() updates an existing archive and queues up necessary jobs
func (s *ArchiveScanner) updateArchive(ctx context.Context, id string, info fs.FileInfo) error {
	if err := s.queries.UpdateArchiveMeta(ctx, sqlc.UpdateArchiveMetaParams{
		ID:       id,
		FileSize: info.Size(),
		ModTime:  info.ModTime(),
	}); err != nil {
		return err
	}

	// re-scrape metadata since the file changed - no need to regenerate
	// thumbnails unless you want to. Search indexing needs no action here
	// either, same as addArchive above.

	// s.queue.Enqueue(ctx, jobs.JobTypeMetadata, jobs.MetadataPayload{ArchiveID: id})

	s.log.Info("archive updated", "id", id)
	return nil
}

// removeArchive() removes an archive and its cache directory
func (s *ArchiveScanner) removeArchive(ctx context.Context, id string, path string) error {
	if err := s.queries.DeleteArchive(ctx, id); err != nil {
		return err
	}

	cacheDir := filepath.Join(s.cfg.Cache.Dir, id)
	if err := os.RemoveAll(cacheDir); err != nil {
		s.log.Warn("failed to remove cache dir", "path", cacheDir, "error", err)
	}

	s.log.Info("archive removed", "path", path, "id", id)
	return nil
}
