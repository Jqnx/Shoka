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

type Scanner struct {
	cfg        *config.Config
	queries    *sqlc.Queries
	queue      JobQueue
	log        *slog.Logger
	libraryDir string
}

func NewScanner(queries *sqlc.Queries, queue JobQueue, log *slog.Logger, libraryDir string) *Scanner {
	return &Scanner{
		queries:    queries,
		queue:      queue,
		log:        log.With("component", "scanner"),
		libraryDir: libraryDir,
	}
}

// Scan() scans the library directories and queues up new jobs for any new or updated files.
func (s *Scanner) Scan(ctx context.Context) error {
	s.log.Info("library scan started")
	start := time.Now()

	found := make(map[string]fs.FileInfo) // path → FileInfo

	if err := s.walk(ctx, s.libraryDir, found); err != nil {
		s.log.Error("walk failed", "dir", s.libraryDir, "error", err)
	}

	added, updated, removed, err := s.process(ctx, found)
	if err != nil {
		return err
	}

	s.log.Info("library scan complete",
		"duration", time.Since(start),
		"added", added,
		"updated", updated,
		"removed", removed,
	)
	return nil
}

// walk() recursively walks a directory and adds any files it finds to the found map
func (s *Scanner) walk(ctx context.Context, root string, found map[string]fs.FileInfo) error {
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

// process() processes the found files, adding, updating or removing them
func (s *Scanner) process(ctx context.Context, found map[string]fs.FileInfo) (added, updated, removed int, err error) {
	existing, err := s.queries.GetAllArchiveFilePaths(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	// build a lookup map: path → db record
	inDB := make(map[string]sqlc.GetAllArchiveFilePathsRow, len(existing))
	for _, a := range existing {
		inDB[a.FilePath] = a
	}

	// handle files found on disk
	for path, info := range found {
		record, known := inDB[path]

		if !known {
			if err := s.addArchive(ctx, path, info); err != nil {
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
func (s *Scanner) hasChanged(record sqlc.GetAllArchiveFilePathsRow, info fs.FileInfo) bool {
	return record.FileSize != info.Size() ||
		!record.ModTime.Equal(info.ModTime())
}

// addArchive() adds a new archive and queues up necessary jobs
func (s *Scanner) addArchive(ctx context.Context, path string, info fs.FileInfo) error {
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

	// TODO: Enabled Jobs
	//if err := s.queue.Enqueue(ctx, jobs.JobTypeCover, jobs.CoverPayload{
	//	ArchiveID: arch.ID,
	//	FilePath:  path,
	//}); err != nil {
	//	return err
	//}

	if err := s.queue.Enqueue(ctx, jobs.JobTypeMetadata, jobs.MetadataPayload{
		ArchiveID: arch.ID,
	}); err != nil {
		return err
	}

	//if err := s.queue.Enqueue(ctx, jobs.JobTypeIndex, jobs.IndexPayload{
	//	ArchiveID: archive.ID,
	//}); err != nil {
	//	return err
	//}

	s.log.Info("archive added", "path", path, "id", arch.ID)
	return nil
}

// updateArchive() updates an existing archive and queues up necessary jobs
func (s *Scanner) updateArchive(ctx context.Context, id string, info fs.FileInfo) error {
	if err := s.queries.UpdateArchiveMeta(ctx, sqlc.UpdateArchiveMetaParams{
		ID:       id,
		FileSize: info.Size(),
		ModTime:  info.ModTime(),
	}); err != nil {
		return err
	}

	// re-scrape metadata and re-index since the file changed
	// no need to regenerate thumbnails unless you want to

	// s.queue.Enqueue(ctx, jobs.JobTypeMetadata, jobs.MetadataPayload{ArchiveID: id})
	// s.queue.Enqueue(ctx, jobs.JobTypeIndex, jobs.IndexPayload{ArchiveID: id})

	s.log.Info("archive updated", "id", id)
	return nil
}

// removeArchive() removes an archive and its cache directory
func (s *Scanner) removeArchive(ctx context.Context, id string, path string) error {
	if err := s.queries.DeleteArchive(ctx, id); err != nil {
		return err
	}

	cacheDir := filepath.Join(s.cfg.Cache.Dir, fmt.Sprintf("%s", id))
	if err := os.RemoveAll(cacheDir); err != nil {
		s.log.Warn("failed to remove cache dir", "path", cacheDir, "error", err)
	}

	s.log.Info("archive removed", "path", path, "id", id)
	return nil
}
