package metadata

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
)

var (
	ErrUnknownSource  = errors.New("unknown metadata source")
	ErrDisabledSource = errors.New("metadata source is disabled")
	ErrNotSearchable  = errors.New("metadata source does not support search")
)

type Pipeline struct {
	sources []Source
	queries *sqlc.Queries
	logger  *slog.Logger
}

type SourceInfo struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

func NewPipeline(logger *slog.Logger, queries *sqlc.Queries, sources ...Source) *Pipeline {
	p := &Pipeline{
		sources: sources,
		queries: queries,
		logger:  logger.With("component", "metadata_pipeline"),
	}

	sort.Slice(p.sources, func(i, j int) bool {
		return p.sources[i].Priority() < p.sources[j].Priority()
	})

	// Explicit rather than init()-based: SourceFromURL's registry should
	// reflect the sources this pipeline was actually built with, not
	// whatever happened to be import-side-effect-linked into the binary.
	for _, source := range p.sources {
		if aware, ok := source.(HostAware); ok {
			RegisterSourceHost(source.Name(), aware.Hosts()...)
		}
	}

	return p
}

// resolveSettings loads a source's per-library configuration from the
// library_source table. A missing row (source never configured for this
// library) is treated as present-but-disabled, not an error.
func (p *Pipeline) resolveSettings(ctx context.Context, libraryID, name string) (SourceSettings, error) {
	row, err := p.queries.GetLibrarySource(ctx, sqlc.GetLibrarySourceParams{
		LibraryID: libraryID,
		Source:    name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SourceSettings{}, nil
		}

		return SourceSettings{}, err
	}

	settings := SourceSettings{
		Enabled:           row.Enabled != 0,
		MagazineBlocklist: unmarshalStringSlice(row.MagazineBlocklist),
		MiscBlocklist:     unmarshalStringSlice(row.MiscBlocklist),
	}
	if row.Cookies != nil {
		settings.Cookies = *row.Cookies
	}

	if row.ApiKey != nil {
		settings.APIKey = *row.ApiKey
	}

	return settings, nil
}

// IsEnabled checks if a given source is enabled for a library.
func (p *Pipeline) IsEnabled(ctx context.Context, libraryID, name string) bool {
	settings, err := p.resolveSettings(ctx, libraryID, name)
	if err != nil {
		p.logger.Warn("failed to resolve source settings", "library_id", libraryID, "source", name, "error", err)
		return false
	}

	return settings.Enabled
}

// Run tries each source in priority order and merges their results.
// The first source to provide a field wins, later sources only fill gaps.
func (p *Pipeline) Run(ctx context.Context, input Input) (*Result, error) {
	return p.run(ctx, input, false)
}

// RunLocal is like Run but skips remote (network/rate-limited) sources.
// Used so local sources can be queried quickly without waiting for network sources.
func (p *Pipeline) RunLocal(ctx context.Context, input Input) (*Result, error) {
	return p.run(ctx, input, true)
}

func (p *Pipeline) run(ctx context.Context, input Input, localOnly bool) (*Result, error) {
	final := &Result{}

	for _, source := range p.sources {
		settings, err := p.resolveSettings(ctx, input.LibraryID, source.Name())
		if err != nil {
			p.logger.Warn("failed to resolve source settings", "source", source.Name(), "archive_id", input.ArchiveID, "error", err)
			continue
		}

		if !settings.Enabled {
			continue
		}

		if localOnly && !source.IsLocal() {
			continue
		}

		if isComplete(final) {
			break
		}

		p.logger.Debug("trying metadata source",
			"source", source.Name(),
			"archive_id", input.ArchiveID,
		)

		sourceInput := input
		sourceInput.SourceConfig = settings

		result, err := source.Fetch(ctx, sourceInput)
		if err != nil {
			p.logger.Warn("metadata source failed",
				"source", source.Name(),
				"archive_id", input.ArchiveID,
				"error", err,
			)

			continue
		}

		if result == nil {
			p.logger.Debug("metadata source returned nothing",
				"source", source.Name(),
				"archive_id", input.ArchiveID,
			)

			continue
		}

		merge(final, result)
		p.logger.Debug("metadata source contributed",
			"source", source.Name(),
			"archive_id", input.ArchiveID,
		)
	}

	return final, nil
}

// merge copies fields from src into dst only if dst's field is not yet set.
func merge(dst, src *Result) {
	if dst.Title == nil && src.Title != nil {
		dst.Title = src.Title
	}

	if dst.Summary == nil && src.Summary != nil {
		dst.Summary = src.Summary
	}

	if dst.Language == nil && src.Language != nil {
		dst.Language = src.Language
	}

	if dst.Category == nil && src.Category != nil {
		dst.Category = src.Category
	}

	if dst.ReleaseDate == nil && src.ReleaseDate != nil {
		dst.ReleaseDate = src.ReleaseDate
	}

	if dst.PageCount == nil && src.PageCount != nil {
		dst.PageCount = src.PageCount
	}
	// for slices, only copy if dst is empty
	if len(dst.Artists) == 0 && len(src.Artists) > 0 {
		dst.Artists = src.Artists
	}

	if len(dst.Tags) == 0 && len(src.Tags) > 0 {
		dst.Tags = src.Tags
	}

	if len(dst.Parodies) == 0 && len(src.Parodies) > 0 {
		dst.Parodies = src.Parodies
	}

	if len(dst.Circles) == 0 && len(src.Circles) > 0 {
		dst.Circles = src.Circles
	}

	if len(dst.Characters) == 0 && len(src.Characters) > 0 {
		dst.Characters = src.Characters
	}
	// URLs are the exception: union across sources rather than first-wins, so
	// a ComicInfo e-hentai link and the nhentai source's own link both
	// survive. NormalizeURLs dedupes by URL, keeping priority order.
	if len(src.URLs) > 0 {
		dst.URLs = NormalizeURLs(append(dst.URLs, src.URLs...))
	}
}

// isComplete returns true if all fields in the result are populated,
// including at least one source URL - otherwise the loop would break on a
// fully-tagged ComicInfo result before ever reaching the lower-priority
// sources (nhentai, e-hentai) that the URL union in merge() relies on.
func isComplete(r *Result) bool {
	return r.Title != nil &&
		r.Summary != nil &&
		r.Language != nil &&
		r.Category != nil &&
		r.ReleaseDate != nil &&
		r.PageCount != nil &&
		len(r.Artists) > 0 &&
		len(r.Tags) > 0 &&
		len(r.URLs) > 0
}

// FetchWithSource runs a single named source.
func (p *Pipeline) FetchWithSource(ctx context.Context, name string, input Input) (*Result, error) {
	var found Source

	for _, s := range p.sources {
		if s.Name() == name {
			found = s
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownSource, name)
	}

	settings, err := p.resolveSettings(ctx, input.LibraryID, name)
	if err != nil {
		return nil, fmt.Errorf("resolve source settings: %w", err)
	}

	if !settings.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrDisabledSource, name)
	}

	input.SourceConfig = settings

	result, err := found.Fetch(ctx, input)
	if err != nil || result == nil {
		return result, err
	}

	result.URLs = NormalizeURLs(result.URLs)

	return result, nil
}

// SearchWithSource runs the Search method of a named remote source.
// Returns ErrNotSearchable if the source doesn't support manual search.
func (p *Pipeline) SearchWithSource(ctx context.Context, name string, input Input) ([]*SearchResult, error) {
	var found Source

	for _, s := range p.sources {
		if s.Name() == name {
			found = s
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownSource, name)
	}

	settings, err := p.resolveSettings(ctx, input.LibraryID, name)
	if err != nil {
		return nil, fmt.Errorf("resolve source settings: %w", err)
	}

	if !settings.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrDisabledSource, name)
	}

	input.SourceConfig = settings

	searchable, ok := found.(SearchableSource)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotSearchable, name)
	}

	return searchable.Search(ctx, input)
}

// FetchFromSourceByID fetches full metadata from a named remote source
// using a source-specific ID chosen by the user from search results.
// input identifies the archive/library the result will be applied to, so
// the correct per-library source settings (cookies/API key) can be used.
func (p *Pipeline) FetchFromSourceByID(ctx context.Context, name string, input Input, id string) (*Result, error) {
	var found Source

	for _, s := range p.sources {
		if s.Name() == name {
			found = s
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownSource, name)
	}

	settings, err := p.resolveSettings(ctx, input.LibraryID, name)
	if err != nil {
		return nil, fmt.Errorf("resolve source settings: %w", err)
	}

	if !settings.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrDisabledSource, name)
	}

	input.SourceConfig = settings

	remote, ok := found.(RemoteSource)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotSearchable, name)
	}

	result, err := remote.FetchByID(ctx, input, id)
	if err != nil || result == nil {
		return result, err
	}

	result.URLs = NormalizeURLs(result.URLs)

	return result, nil
}

// SourceIsLocal reports whether name is a registered source (known) and,
// if so, whether it reads from the local filesystem rather than the network
// (isLocal). Callers use this to validate a caller-supplied source name and
// to decide which job queue to run it on - remote sources are deliberately
// processed at low concurrency so they aren't hammered.
func (p *Pipeline) SourceIsLocal(name string) (isLocal, known bool) {
	for _, s := range p.sources {
		if s.Name() == name {
			return s.IsLocal(), true
		}
	}

	return false, false
}

// Sources returns info about all registered sources and their enabled
// state for the given library.
func (p *Pipeline) Sources(ctx context.Context, libraryID string) []SourceInfo {
	info := make([]SourceInfo, len(p.sources))
	for i, s := range p.sources {
		info[i] = SourceInfo{
			Name:     s.Name(),
			Priority: s.Priority(),
			Enabled:  p.IsEnabled(ctx, libraryID, s.Name()),
		}
	}

	return info
}

func unmarshalStringSlice(raw string) []string {
	if raw == "" {
		return nil
	}

	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}

	return out
}
