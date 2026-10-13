package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/language"
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// The taxonomy endpoints - characters, parodies, tags and artists - repeat
// four request shapes: a paginated list, an "all" list, the archives linked
// to one row, and delete-by-id. Those algorithms live here; the per-entity
// handlers stay in their own files as thin call sites.
//
// The sqlc parameter and row structs of the equivalent queries are
// structurally identical (they differ only in struct tags, which Go's
// conversion rules ignore), so the canonical types below stand in for them
// and let one helper serve every entity.

const (
	// defaultListLimit is the page size the taxonomy list endpoints use when
	// the request does not ask for another one.
	defaultListLimit = 50
	// defaultArchiveLimit is the page size the archives-by-taxonomy endpoints
	// use when the request does not ask for another one.
	defaultArchiveLimit = 24
	// maxListLimit caps the page size a request may ask for.
	maxListLimit = 100
)

// parsePagination reads the page/limit query parameters, falling back to
// defaultLimit and ignoring unparsable or out-of-range values.
func parsePagination(r *http.Request, defaultLimit int) (page, limit int) {
	page, limit = 1, defaultLimit

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= maxListLimit {
			limit = v
		}
	}

	return page, limit
}

// pathID parses the "id" path parameter, writing a 400 response naming
// resource when it is not a valid integer.
func pathID(w http.ResponseWriter, r *http.Request, resource string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid "+resource+" id")

		return 0, false
	}

	return id, true
}

// taxonomyPage is the paginated payload shared by the list endpoints. The
// response structs the handlers return are conversions of it.
type taxonomyPage[T any] struct {
	Items []T
	Total int64
	Page  int
	Limit int
}

// taxonomyListParams is the page window of a taxonomy list query, and
// taxonomyListParamsType the sqlc structs it converts to.
type taxonomyListParams sqlc.GetCharacterListParams

// taxonomyListParamsType lists the sqlc parameters of the taxonomy list
// queries.
type taxonomyListParamsType interface {
	sqlc.GetCharacterListParams | sqlc.GetParodyListParams | sqlc.GetTagListParams
}

// listTaxonomy counts and fetches one page of kind ("characters", "tags", ...),
// mapping every row with item. It writes the failure response and returns
// false when either query fails.
func listTaxonomy[P taxonomyListParamsType, R, T any](
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	kind string,
	count func(context.Context) (int64, error),
	fetch func(context.Context, P) ([]R, error),
	item func(R) T,
) (taxonomyPage[T], bool) {
	page, limit := parsePagination(r, defaultListLimit)
	offset := int64((page - 1) * limit)

	total, err := count(r.Context())
	if err != nil {
		logger.Error("count "+kind+" failed", "error", err)
		response.InternalError(w, "failed to count "+kind)

		return taxonomyPage[T]{}, false
	}

	rows, err := fetch(r.Context(), P(taxonomyListParams{Limit: int64(limit), Offset: offset}))
	if err != nil {
		logger.Error("list "+kind+" failed", "error", err)
		response.InternalError(w, "failed to list "+kind)

		return taxonomyPage[T]{}, false
	}

	items := make([]T, 0, len(rows))
	for _, row := range rows {
		items = append(items, item(row))
	}

	return taxonomyPage[T]{Items: items, Total: total, Page: page, Limit: limit}, true
}

// allTaxonomy fetches every row of kind and maps it with item. It writes the
// failure response and returns false when the query fails.
func allTaxonomy[R, T any](
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	kind string,
	fetch func(context.Context) ([]R, error),
	item func(R) T,
) ([]T, bool) {
	rows, err := fetch(r.Context())
	if err != nil {
		logger.Error("get all "+kind+" failed", "error", err)
		response.InternalError(w, "failed to get "+kind)

		return nil, false
	}

	items := make([]T, 0, len(rows))
	for _, row := range rows {
		items = append(items, item(row))
	}

	return items, true
}

// archiveRow mirrors the columns selected by every archives-by-taxonomy
// query, and archiveRowType the sqlc row types it converts from.
type archiveRow sqlc.GetArchivesByCharacterNameRow

// archiveRowType lists the sqlc rows of the archives-by-taxonomy queries.
type archiveRowType interface {
	sqlc.GetArchivesByCharacterNameRow | sqlc.GetArchivesByParodyNameRow |
		sqlc.GetArchivesByTagNameRow | sqlc.GetArchivesByArtistIDRow
}

// archiveNameParams is the parameter set of a name-keyed archives query, and
// archiveNameParamsType the sqlc structs it converts to.
type archiveNameParams sqlc.GetArchivesByCharacterNameParams

// archiveNameParamsType lists the sqlc parameters of the name-keyed
// archives-by-taxonomy queries.
type archiveNameParamsType interface {
	sqlc.GetArchivesByCharacterNameParams | sqlc.GetArchivesByParodyNameParams |
		sqlc.GetArchivesByTagNameParams
}

// archiveResponses maps a page of archive rows to the shared list item.
func archiveResponses[R archiveRowType](processor *image.Processor, rows []R) []ArchiveResponse {
	lc := language.NewConverter()

	items := make([]ArchiveResponse, 0, len(rows))
	for _, raw := range rows {
		row := archiveRow(raw)

		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    lc.DisplayName(row.Language),
			Category:    row.Category,
			ReleaseDate: row.ReleaseDate,
			PageCount:   int(row.PageCount),
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			ThumbsReady: processor.ThumbsReady(row.ID, int(row.PageCount)),
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

// archivesByName counts and fetches one page of the archives linked to a
// taxonomy row selected by name (character, parody, tag) and writes the list
// response. kind names the relation in logs and messages.
func archivesByName[P archiveNameParamsType, R archiveRowType](
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	kind, name string,
	count func(context.Context, string) (int64, error),
	fetch func(context.Context, P) ([]R, error),
	processor *image.Processor,
) {
	page, limit := parsePagination(r, defaultArchiveLimit)
	offset := int64((page - 1) * limit)
	uid := auth.UserIDFromContext(r.Context())

	total, err := count(r.Context(), name)
	if err != nil {
		logger.Error("count archives by "+kind+" failed", "error", err)
		response.InternalError(w, "failed to count archives")

		return
	}

	rows, err := fetch(r.Context(), P(archiveNameParams{Uid: uid, Name: name, Offset: offset, Limit: int64(limit)}))
	if err != nil {
		logger.Error("list archives by "+kind+" failed", "error", err)
		response.InternalError(w, "failed to list archives")

		return
	}

	response.JSON(w, http.StatusOK, ArchiveListResponse{
		Items: archiveResponses(processor, rows),
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// archivesByID is archivesByName for the artist, whose archives are selected
// by id and whose failure log lines name the artist id.
func archivesByID[R archiveRowType](
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	id int64,
	count func(context.Context, int64) (int64, error),
	fetch func(context.Context, sqlc.GetArchivesByArtistIDParams) ([]R, error),
	processor *image.Processor,
) {
	page, limit := parsePagination(r, defaultArchiveLimit)
	offset := int64((page - 1) * limit)
	uid := auth.UserIDFromContext(r.Context())

	total, err := count(r.Context(), id)
	if err != nil {
		logger.Error("count archives by artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to count archives")

		return
	}

	rows, err := fetch(r.Context(), sqlc.GetArchivesByArtistIDParams{
		ID:     id,
		Uid:    uid,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		logger.Error("list archives by artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to list archives")

		return
	}

	response.JSON(w, http.StatusOK, ArchiveListResponse{
		Items: archiveResponses(processor, rows),
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
