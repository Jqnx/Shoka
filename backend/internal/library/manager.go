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

	mu      sync.Mutex
	cancels map[string]context.CancelFunc // library ID -> stop func for its watcher
}

func NewManager(cfg *config.Config, queries *sqlc.Queries, queue JobQueue, log *slog.Logger) *Manager {
	logger := log.With("component", "library_manager")

	m := &Manager{
		cfg:     cfg,
		queries: queries,
		queue:   queue,
		log:     logger,
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
	libs, err := m.queries.ListLibraries(ctx)
	if err != nil {
		return fmt.Errorf("list libraries: %w", err)
	}

	for _, lib := range libs {
		m.startWatching(ctx, lib)
	}

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

// OnLibraryChanged begins watching lib. Call after creating or updating a
// library so the change takes effect immediately, without a restart.
// startWatching is idempotent, so calling this for an already-watched
// library is a no-op.
func (m *Manager) OnLibraryChanged(ctx context.Context, lib sqlc.Library) {
	m.startWatching(ctx, lib)
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

	return scanner.Scan(ctx, lib)
}

// startWatching begins watching a library's directory for changes. Safe to
// call more than once — a no-op if already watching, and a no-op if the
// library's type has no registered Scanner (nothing would ever act on the
// scan jobs it would trigger).
func (m *Manager) startWatching(ctx context.Context, lib sqlc.Library) {
	if _, ok := m.scanners[lib.Type]; !ok {
		m.log.Warn("no scanner registered for library type, not watching", "library_id", lib.ID, "type", lib.Type)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.cancels[lib.ID]; exists {
		return
	}

	watchCtx, cancel := context.WithCancel(ctx)

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
