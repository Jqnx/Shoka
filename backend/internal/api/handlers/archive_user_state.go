package handlers

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"errors"
)

// bestEffortProgress loads the given user's reading progress for one archive,
// or nil when there is none. A genuine query failure is logged and swallowed
// rather than failing the request: the caller's response is still useful
// without the progress block, and this is a secondary read on a path whose
// primary work has already been committed.
//
// Note this is deliberately more forgiving than ArchiveHandler.archiveResponseFor,
// which propagates the same failure - the two paths' contracts differ.
func (h *ArchiveHandler) bestEffortProgress(ctx context.Context, archiveID, userID string) *sqlc.ReadingProgress {
	p, err := h.queries.GetProgressForArchive(ctx, sqlc.GetProgressForArchiveParams{
		ArchiveID: archiveID,
		UserID:    userID,
	})
	if err == nil {
		return &p
	}

	if !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get progress failed", "id", archiveID, "error", err)
	}

	return nil
}

// userArchiveRating loads the given user's rating (1-5) for one archive, or
// nil when they haven't rated it. A real query failure is returned for the
// caller to report - unlike bestEffortProgress it isn't safe to answer
// without knowing whether a rating exists, since the client would read the
// missing field as "unrated".
func (h *ArchiveHandler) userArchiveRating(ctx context.Context, archiveID, userID string) (*int, error) {
	v, err := h.queries.GetArchiveRating(ctx, sqlc.GetArchiveRatingParams{
		ArchiveID: archiveID,
		Uid:       userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	rating := int(v)

	return &rating, nil
}
