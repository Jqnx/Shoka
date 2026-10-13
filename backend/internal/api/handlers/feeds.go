package handlers

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/language"
	"Shoka/internal/metadata"
)

// recentlyReadResponses renders the continue-reading feed's rows, attaching
// the favorite flags, ratings, and relational metadata the caller already
// fetched in bulk for this page. The rows carry their progress inline (the
// feed only lists archives that have some).
func (h *ArchiveHandler) recentlyReadResponses(
	rows []sqlc.GetRecentlyReadArchivesRow,
	favorited map[string]bool,
	ratings map[string]int,
	metaByID map[string]*metadata.Result,
) []ArchiveResponse {
	lc := language.NewConverter()

	items := make([]ArchiveResponse, 0, len(rows))

	for _, row := range rows {
		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    lc.DisplayName(row.Language),
			Category:    row.Category,
			ReleaseDate: row.ReleaseDate,
			PageCount:   int(row.PageCount),
			FilePath:    row.FilePath,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			ThumbsReady: h.processor.ThumbsReady(row.ID, int(row.PageCount)),
			IsFavorited: favorited[row.ID],
			Progress: &ProgressResponse{
				CurrentPage: int(row.Page),
				LastRead:    row.LastRead,
				Completed:   row.Completed,
			},
		}

		if rating, ok := ratings[row.ID]; ok {
			resp.Rating = &rating
		}

		if meta, ok := metaByID[row.ID]; ok {
			resp.Artists = meta.Artists
			resp.Tags = meta.Tags
			resp.Parodies = meta.Parodies
			resp.Circles = meta.Circles
			resp.Characters = meta.Characters
		}

		items = append(items, resp)
	}

	return items
}

// favoriteFeedResponses renders the favorites feed's rows. Every archive here
// is favorited by definition, and progress is optional — unlike the
// recently-read feed, this one also lists archives that were never opened,
// which is why the three progress columns arrive nullable and the field is
// left nil when there is no progress row.
func (h *ArchiveHandler) favoriteFeedResponses(rows []sqlc.GetUserFavoriteArchiveListRow, ratings map[string]int) []ArchiveResponse {
	lc := language.NewConverter()

	items := make([]ArchiveResponse, 0, len(rows))

	for _, row := range rows {
		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    lc.DisplayName(row.Language),
			Category:    row.Category,
			ReleaseDate: row.ReleaseDate,
			PageCount:   int(row.PageCount),
			FilePath:    row.FilePath,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			ThumbsReady: h.processor.ThumbsReady(row.ID, int(row.PageCount)),
			IsFavorited: true,
		}

		if rating, ok := ratings[row.ID]; ok {
			resp.Rating = &rating
		}

		if row.Page != nil {
			resp.Progress = &ProgressResponse{
				CurrentPage: int(*row.Page),
				Completed:   row.Completed != nil && *row.Completed,
			}

			if row.LastRead != nil {
				resp.Progress.LastRead = *row.LastRead
			}
		}

		items = append(items, resp)
	}

	return items
}
