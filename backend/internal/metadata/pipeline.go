package metadata

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/spf13/viper"
)

var (
	ErrUnknownSource  = errors.New("unknown metadata source")
	ErrDisabledSource = errors.New("metadata source is disabled")
)

type Pipeline struct {
	sources []Source
	logger  *slog.Logger
}

type SourceInfo struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

func NewPipeline(logger *slog.Logger, sources ...Source) *Pipeline {
	p := &Pipeline{
		sources: sources,
		logger:  logger.With("component", "metadata_pipeline"),
	}

	sort.Slice(p.sources, func(i, j int) bool {
		return p.sources[i].Priority() < p.sources[j].Priority()
	})

	return p
}

// IsEnabled checks if a given source is enabled in the config.
func (p *Pipeline) IsEnabled(name string) bool {
	key := fmt.Sprintf("metadata.sources.%s.enabled", name)
	return viper.GetBool(key)
}

// Run tries each source in priority order and merges their results.
// The first source to provide a field wins — later sources only fill gaps.
func (p *Pipeline) Run(ctx context.Context, input Input) (*Result, error) {
	final := &Result{}

	for _, source := range p.sources {
		if !p.IsEnabled(source.Name()) {
			continue
		}

		if isComplete(final) {
			break
		}

		p.logger.Debug("trying metadata source",
			"source", source.Name(),
			"archive_id", input.ArchiveID,
		)

		result, err := source.Fetch(ctx, input)
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
}

// isComplete returns true if all fields in the result are populated.
func isComplete(r *Result) bool {
	return r.Title != nil &&
		r.Summary != nil &&
		r.Language != nil &&
		r.Category != nil &&
		r.ReleaseDate != nil &&
		r.PageCount != nil &&
		len(r.Artists) > 0 &&
		len(r.Tags) > 0
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

	if !p.IsEnabled(name) {
		return nil, fmt.Errorf("%w: %s", ErrDisabledSource, name)
	}

	return found.Fetch(ctx, input)
}

// Sources returns info about all registered sources and their enabled state.
func (p *Pipeline) Sources() []SourceInfo {
	info := make([]SourceInfo, len(p.sources))
	for i, s := range p.sources {
		info[i] = SourceInfo{
			Name:     s.Name(),
			Priority: s.Priority(),
			Enabled:  p.IsEnabled(s.Name()),
		}
	}

	return info
}
