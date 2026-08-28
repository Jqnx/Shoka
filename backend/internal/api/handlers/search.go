package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

// searchMinQueryRunes mirrors the trigram tokenizer's floor on the archive
// side (see database.buildWhere): a query shorter than this can't match
// anything meaningful, so every category is returned empty instead of
// running five-plus substring scans for nothing.
const searchMinQueryRunes = 3

const defaultSearchLimit = 5

const maxSearchLimit = 50

type SearchHandler struct {
	queries   *sqlc.Queries
	db        *sql.DB
	processor *image.Processor
	logger    *slog.Logger
}

func NewSearchHandler(queries *sqlc.Queries, db *sql.DB, processor *image.Processor, log *slog.Logger) *SearchHandler {
	return &SearchHandler{
		queries:   queries,
		db:        db,
		processor: processor,
		logger:    log.With("handler", "search"),
	}
}

// SearchEntityResult is one name match from a metadata entity category
// (artists, tags, parodies, circles, characters).
type SearchEntityResult struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// SearchEntityResponse is one category of non-archive search results. Total
// counts every matching row, even beyond the capped Items list, so a client
// can render a "N more" affordance.
type SearchEntityResponse struct {
	Items []SearchEntityResult `json:"items"`
	Total int64                `json:"total"`
}

// SearchArchiveResponse is the archive category of a search result - same
// shape as SearchEntityResponse but with full ArchiveResponse items.
type SearchArchiveResponse struct {
	Items []ArchiveResponse `json:"items"`
	Total int64             `json:"total"`
}

// SearchResponse groups results by category, each independently capped and
// counted - see SearchHandler.Search.
type SearchResponse struct {
	Archives   SearchArchiveResponse `json:"archives"`
	Artists    SearchEntityResponse  `json:"artists"`
	Circles    SearchEntityResponse  `json:"circles"`
	Tags       SearchEntityResponse  `json:"tags"`
	Parodies   SearchEntityResponse  `json:"parodies"`
	Characters SearchEntityResponse  `json:"characters"`
}

// emptySearchResponse is what Search returns for a query too short to
// match anything - every Items slice is non-nil so it serializes as `[]`
// rather than `null`.
func emptySearchResponse() SearchResponse {
	return SearchResponse{
		Archives:   SearchArchiveResponse{Items: []ArchiveResponse{}},
		Artists:    SearchEntityResponse{Items: []SearchEntityResult{}},
		Circles:    SearchEntityResponse{Items: []SearchEntityResult{}},
		Tags:       SearchEntityResponse{Items: []SearchEntityResult{}},
		Parodies:   SearchEntityResponse{Items: []SearchEntityResult{}},
		Characters: SearchEntityResponse{Items: []SearchEntityResult{}},
	}
}

func toSearchEntityResponse(items []database.EntitySearchResult, total int64) SearchEntityResponse {
	resp := SearchEntityResponse{Items: make([]SearchEntityResult, 0, len(items)), Total: total}
	for _, item := range items {
		resp.Items = append(resp.Items, SearchEntityResult{ID: item.ID, Name: item.Name, Count: item.Count})
	}

	return resp
}

// Search godoc
//
//	@Summary		Search across archives, artists, tags, parodies, circles and characters
//	@Description	Substring match (min 3 characters - shorter values return every category empty) over each category's name, scoped to a single library for the archives category. Artist/tag/parody/circle/character are global entities (not library-scoped), so those four categories search the whole install. Each category is independently capped at `limit` and separately counted via its `total`, for a "N more" affordance per category.
//	@Tags			search
//	@Produce		json
//	@Param			library_id	query		string	true	"Library ID to scope the archive category to"
//	@Param			q			query		string	true	"Search query (substring match, min 3 characters)"
//	@Param			limit		query		int		false	"Max results per category"	default(5)
//	@Success		200	{object}	SearchResponse
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/search [get]
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	q := r.URL.Query()

	libraryID := q.Get("library_id")
	if libraryID == "" {
		response.BadRequest(w, "library_id is required")
		return
	}

	limit := defaultSearchLimit
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= maxSearchLimit {
			limit = v
		}
	}

	query := strings.TrimSpace(q.Get("q"))
	if utf8.RuneCountInString(query) < searchMinQueryRunes {
		response.JSON(w, http.StatusOK, emptySearchResponse())
		return
	}

	ctx := r.Context()

	rows, archiveTotal, err := database.ListArchives(ctx, h.db, database.ArchiveFilter{
		LibraryID: libraryID,
		Query:     query,
		Limit:     int64(limit),
		Offset:    0,
		UserID:    userID,
	})
	if err != nil {
		h.logger.Error("search archives failed", "error", err)
		response.InternalError(w, "failed to search archives")

		return
	}

	archiveItems, err := buildArchiveListResponses(ctx, h.queries, h.db, h.processor, userID, rows)
	if err != nil {
		h.logger.Error("build archive search responses failed", "error", err)
		response.InternalError(w, "failed to search archives")

		return
	}

	artists, artistTotal, err := database.SearchEntities(ctx, h.db, "artist", query, int64(limit))
	if err != nil {
		h.logger.Error("search artists failed", "error", err)
		response.InternalError(w, "failed to search artists")

		return
	}

	circles, circleTotal, err := database.SearchEntities(ctx, h.db, "circle", query, int64(limit))
	if err != nil {
		h.logger.Error("search circles failed", "error", err)
		response.InternalError(w, "failed to search circles")

		return
	}

	tags, tagTotal, err := database.SearchEntities(ctx, h.db, "tag", query, int64(limit))
	if err != nil {
		h.logger.Error("search tags failed", "error", err)
		response.InternalError(w, "failed to search tags")

		return
	}

	parodies, parodyTotal, err := database.SearchEntities(ctx, h.db, "parody", query, int64(limit))
	if err != nil {
		h.logger.Error("search parodies failed", "error", err)
		response.InternalError(w, "failed to search parodies")

		return
	}

	characters, characterTotal, err := database.SearchEntities(ctx, h.db, "character", query, int64(limit))
	if err != nil {
		h.logger.Error("search characters failed", "error", err)
		response.InternalError(w, "failed to search characters")

		return
	}

	response.JSON(w, http.StatusOK, SearchResponse{
		Archives:   SearchArchiveResponse{Items: archiveItems, Total: archiveTotal},
		Artists:    toSearchEntityResponse(artists, artistTotal),
		Circles:    toSearchEntityResponse(circles, circleTotal),
		Tags:       toSearchEntityResponse(tags, tagTotal),
		Parodies:   toSearchEntityResponse(parodies, parodyTotal),
		Characters: toSearchEntityResponse(characters, characterTotal),
	})
}
