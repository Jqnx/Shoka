package library

import (
	"Shoka/internal/config"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// AllLibraryTypes lists every library type known to the schema (must stay
// in sync with the CHECK constraint on library.type in the migrations),
// regardless of whether a Scanner is registered for it yet. Used to tell
// clients which types exist versus which are actually usable today.
var AllLibraryTypes = []string{"doujinshi", "audio"}

// Manager owns the set of active libraries: for each one it runs a
// filesystem Watcher, and it dispatches scan jobs (see jobs.Scannable) to
// the Scanner implementation registered for that library's type. It lets
// libraries be added/removed at runtime via the admin API without
// restarting the server.
type Manager struct {
	cfg      *config.Config
	queries  *sqlc.Queries
	queue    JobQueue
	log      *slog.Logger
	scanners map[string]Scanner // keyed by library.type

	mu sync.Mutex
	// baseCtx is the application-lifetime context every watcher derives from,
	// captured in Start. Never derive a watcher from a request context —
	// OnLibraryChanged runs in HTTP handlers, so the watcher would die as
	// soon as the response is written.
	baseCtx context.Context
	cancels map[string]context.CancelFunc // library ID -> stop func for its watcher
}

func NewManager(cfg *config.Config, queries *sqlc.Queries, queue JobQueue, log *slog.Logger) *Manager {
	logger := log.With("component", "library_manager")

	m := &Manager{
		cfg:     cfg,
		queries: queries,
		queue:   queue,
		log:     logger,
		baseCtx: context.Background(),
		cancels: make(map[string]context.CancelFunc),
	}

	// Registry of Scanner implementations by library type. "audio" has no
	// implementation yet — libraries of that type are accepted by the
	// schema but rejected at the admin API layer until that scanner exists.
	m.scanners = map[string]Scanner{
		"doujinshi": NewArchiveScanner(cfg, queries, queue, logger),
	}

	return m
}

// SupportsType reports whether a Scanner is registered for the given
// library type — i.e. whether it can actually be scanned/watched today.
// Used by the admin API to validate library creation and to report
// per-type support alongside AllLibraryTypes.
func (m *Manager) SupportsType(libraryType string) bool {
	_, ok := m.scanners[libraryType]
	return ok
}

// Start loads every library and begins watching it. Call once at
// application startup, after config/DB are ready.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	m.baseCtx = ctx
	m.mu.Unlock()

	libs, err := m.queries.ListLibraries(ctx)
	if err != nil {
		return fmt.Errorf("list libraries: %w", err)
	}

	for _, lib := range libs {
		m.startWatching(lib)
	}

	// Periodic scans are a safety net for changes the watcher never sees —
	// network shares and bind mounts don't propagate inotify events.
	go m.runScheduler(ctx)

	return nil
}

// EnqueueInitialScans enqueues one scan job per library. Intended to be
// called once at application startup; the worker pool parallelizes the
// resulting jobs across libraries.
func (m *Manager) EnqueueInitialScans(ctx context.Context) error {
	libs, err := m.queries.ListLibraries(ctx)
	if err != nil {
		return fmt.Errorf("list libraries: %w", err)
	}

	for _, lib := range libs {
		if _, ok := m.scanners[lib.Type]; !ok {
			continue
		}

		if err := m.queue.EnqueueOnce(ctx, jobs.JobTypeScan, jobs.ScanPayload{LibraryID: lib.ID}); err != nil {
			m.log.Error("failed to enqueue initial scan", "library_id", lib.ID, "error", err)
		}
	}

	return nil
}

// Shutdown stops every active watcher.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, cancel := range m.cancels {
		cancel()
		delete(m.cancels, id)
	}
}

// OnLibraryChanged reconciles lib's watcher with its current settings, so a
// create or update takes effect without a restart. Must handle both
// directions: startWatching is a no-op when already watching, so turning
// watch_enabled off has to explicitly stop the running watcher.
//
// Takes no context by design — see Manager.baseCtx.
func (m *Manager) OnLibraryChanged(lib sqlc.Library) {
	if lib.WatchEnabled == 0 {
		m.stopWatching(lib.ID)
		return
	}

	m.startWatching(lib)
}

// OnLibraryDeleted stops watching a deleted library. DB cleanup (archives,
// library_source rows) is handled by ON DELETE CASCADE.
func (m *Manager) OnLibraryDeleted(libraryID string) {
	m.stopWatching(libraryID)
}

// ScanLibrary implements jobs.Scannable. It resolves the library by ID and
// dispatches to the Scanner registered for its type.
func (m *Manager) ScanLibrary(ctx context.Context, libraryID string) error {
	lib, err := m.queries.GetLibraryByID(ctx, libraryID)
	if err != nil {
		return fmt.Errorf("get library: %w", err)
	}

	scanner, ok := m.scanners[lib.Type]
	if !ok {
		return fmt.Errorf("no scanner available for library type %q", lib.Type)
	}

	// Record the attempt even on failure. Otherwise a library that always
	// fails to scan exhausts its retries, leaves nothing pending for
	// EnqueueOnce to suppress, and gets re-enqueued every tick.
	defer func() {
		if err := m.queries.MarkLibraryScanned(ctx, libraryID); err != nil {
			m.log.Error("failed to record scan time", "library_id", libraryID, "error", err)
		}
	}()

	return scanner.Scan(ctx, lib)
}

// startWatching begins watching a library's directory for changes. Safe to
// call more than once — a no-op if already watching, if the library has the
// filesystem watcher turned off, or if the library's type has no registered
// Scanner (nothing would ever act on the scan jobs it would trigger).
func (m *Manager) startWatching(lib sqlc.Library) {
	if _, ok := m.scanners[lib.Type]; !ok {
		m.log.Warn("no scanner registered for library type, not watching", "library_id", lib.ID, "type", lib.Type)
		return
	}

	if lib.WatchEnabled == 0 {
		m.log.Info("filesystem watcher disabled for library", "library_id", lib.ID)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.cancels[lib.ID]; exists {
		return
	}

	watchCtx, cancel := context.WithCancel(m.baseCtx)

	watcher := NewWatcher(m.queue, lib.ID, lib.Path, m.log)
	if err := watcher.Start(watchCtx); err != nil {
		m.log.Error("failed to start watcher", "library_id", lib.ID, "path", lib.Path, "error", err)
		cancel()

		return
	}

	m.cancels[lib.ID] = cancel
}

// stopWatching stops watching a library's directory, if currently watched.
func (m *Manager) stopWatching(libraryID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.cancels[libraryID]; ok {
		cancel()
		delete(m.cancels, libraryID)
	}
}
