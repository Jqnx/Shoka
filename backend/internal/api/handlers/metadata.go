package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/library/archive"
	"Shoka/internal/metadata"
	"Shoka/internal/metadata/sources"

	"github.com/go-chi/chi/v5"
)

type MetadataHandler struct {
	queries   *sqlc.Queries
	pipeline  *metadata.Pipeline
	logger    *slog.Logger
	db        *sql.DB
	processor *image.Processor
}

func NewMetadataHandler(queries *sqlc.Queries, db *sql.DB, pipeline *metadata.Pipeline, processor *image.Processor, logger *slog.Logger) *MetadataHandler {
	return &MetadataHandler{
		queries:   queries,
		db:        db,
		pipeline:  pipeline,
		processor: processor,
		logger:    logger.With("handler", "metadata"),
	}
}

// GetSources godoc
//
//	@Summary		List metadata sources
//	@Description	Returns all registered metadata sources and their enabled state for a library
//	@Tags			metadata
//	@Produce		json
//	@Param			library_id	query		string	true	"Library ID"
//	@Success		200	{array}		metadata.SourceInfo
//	@Failure		400	{object}	response.Error
//	@Router			/metadata/sources [get]
func (h *MetadataHandler) GetSources(w http.ResponseWriter, r *http.Request) {
	libraryID := r.URL.Query().Get("library_id")
	if libraryID == "" {
		response.BadRequest(w, "library_id is required")
		return
	}

	response.JSON(w, http.StatusOK, h.pipeline.Sources(r.Context(), libraryID))
}

// FetchMetadata godoc
//
//	@Summary		Fetch metadata for an archive
//	@Description	Runs the full metadata pipeline for the given archive and applies results
//	@Tags			metadata
//	@Produce		json
//	@Param			id	path		string	true	"Archive ID"
//	@Success		200	{object}	metadata.Result
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/archives/{id}/metadata [post]
func (h *MetadataHandler) FetchMetadata(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "archive not found")
		return
	}

	result, err := h.pipeline.Run(r.Context(), metadata.Input{
		ArchiveID: archive.ID,
		LibraryID: archive.LibraryID,
		FilePath:  archive.FilePath,
		Title:     archive.Title,
	})
	if err != nil {
		h.logger.Error("metadata pipeline failed", "archive_id", id, "error", err)
		response.InternalError(w, "failed to fetch metadata")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// FetchMetadataFromSource godoc
//
//	@Summary		Fetch metadata from a specific source
//	@Description	Runs a single named metadata source for the given archive
//	@Tags			metadata
//	@Produce		json
//	@Param			id		path		string	true	"Archive ID"
//	@Param			source	path		string	true	"Source name (e.g. e-hentai, nhentai, comicinfo.xml)"
//	@Success		200		{object}	metadata.Result
//	@Failure		404		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Router			/archives/{id}/metadata/{source} [post]
func (h *MetadataHandler) FetchMetadataFromSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sourceName := chi.URLParam(r, "source")

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "archive not found")
		return
	}

	result, err := h.pipeline.FetchWithSource(r.Context(), sourceName, metadata.Input{
		ArchiveID: archive.ID,
		LibraryID: archive.LibraryID,
		FilePath:  archive.FilePath,
		Title:     archive.Title,
	})
	if err != nil {
		switch {
		case errors.Is(err, metadata.ErrUnknownSource):
			response.NotFound(w, err.Error())
		case errors.Is(err, metadata.ErrDisabledSource):
			response.Forbidden(w, err.Error())
		default:
			response.InternalError(w, err.Error())
		}
		return
	}

	if result == nil {
		response.NotFound(w, fmt.Sprintf("source %q found no metadata for this archive", sourceName))
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// SaveMetadataToFile godoc
//
//	@Summary		Save metadata to archive
//	@Description	Writes current metadata as ComicInfo.xml into the archive file
//	@Tags			metadata
//	@Param			id	path	string	true	"Archive ID"
//	@Success		204
//	@Failure		404	{object}	response.Error
//	@Failure		422	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/archives/{id}/metadata/save [post]
func (h *MetadataHandler) SaveMetadataToFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	arch, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "archive not found")
		return
	}

	result, err := database.GetArchiveMetadata(r.Context(), h.queries, id)
	if err != nil {
		response.InternalError(w, "failed to load metadata")
		return
	}

	data, err := sources.MarshalComicInfo(result)
	if err != nil {
		response.InternalError(w, "failed to build ComicInfo.xml")
		return
	}

	a, err := archive.Open(arch.FilePath)
	if err != nil {
		response.InternalError(w, "failed to open archive")
		return
	}
	defer a.Close()

	if err := a.WriteFile("ComicInfo.xml", data); err != nil {
		response.UnprocessableEntity(w, err.Error())
		return
	}

	// if the archive path changed (RAR → CBZ conversion) update the DB
	if updatedPath := a.Path(); updatedPath != arch.FilePath {
		info, _ := os.Stat(updatedPath)
		h.queries.UpdateFilePath(r.Context(), sqlc.UpdateFilePathParams{
			ID:       id,
			FilePath: updatedPath,
			FileSize: info.Size(),
			ModTime:  info.ModTime(),
		})
	}

	response.NoContent(w)
}

// SearchMetadataSource godoc
//
//	@Summary		Search a metadata source
//	@Description	Returns a list of candidates from a remote source for manual selection
//	@Tags			metadata
//	@Produce		json
//	@Param			id		path		string	true	"Archive ID"
//	@Param			source	path		string	true	"Source name"
//	@Param			q		query		string	false	"Override search query (defaults to archive title)"
//	@Success		200		{array}		metadata.SearchResult
//	@Failure		400		{object}	response.Error
//	@Failure		404		{object}	response.Error
//	@Failure		422		{object}	response.Error
//	@Router			/archives/{id}/metadata/{source}/search [get]
func (h *MetadataHandler) SearchMetadataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sourceName := chi.URLParam(r, "source")

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "archive not found")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		query = archive.Title
	}

	results, err := h.pipeline.SearchWithSource(r.Context(), sourceName, metadata.Input{
		ArchiveID: archive.ID,
		LibraryID: archive.LibraryID,
		FilePath:  archive.FilePath,
		Title:     query,
	})
	if err != nil {
		switch {
		case errors.Is(err, metadata.ErrUnknownSource):
			response.NotFound(w, err.Error())
		case errors.Is(err, metadata.ErrDisabledSource):
			response.Forbidden(w, err.Error())
		case errors.Is(err, metadata.ErrNotSearchable):
			response.BadRequest(w, err.Error())
		default:
			response.InternalError(w, err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, results)
}

// ApplyMetadataFromSource godoc
//
//	@Summary		Preview metadata from a specific source result
//	@Description	Fetches full metadata by source-specific ID and returns a preview of the archive as it would look with that metadata applied — it does NOT save anything. Use PATCH /api/archives/{id} to actually persist the (possibly user-edited) values.
//	@Tags			metadata
//	@Produce		json
//	@Param			id		path	string	true	"Archive ID"
//	@Param			source	path	string	true	"Source name"
//	@Param			source_id	path	string	true	"Source-specific result ID"
//	@Success		200	{object}	ArchiveResponse
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/archives/{id}/metadata/{source}/{source_id} [post]
func (h *MetadataHandler) ApplyMetadataFromSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sourceName := chi.URLParam(r, "source")
	sourceID := chi.URLParam(r, "source_id")

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "archive not found")
		return
	}

	result, err := h.pipeline.FetchFromSourceByID(r.Context(), sourceName, metadata.Input{
		ArchiveID: archive.ID,
		LibraryID: archive.LibraryID,
		FilePath:  archive.FilePath,
		Title:     archive.Title,
	}, sourceID)
	if err != nil {
		switch {
		case errors.Is(err, metadata.ErrUnknownSource):
			response.NotFound(w, err.Error())
		case errors.Is(err, metadata.ErrDisabledSource):
			response.Forbidden(w, err.Error())
		case errors.Is(err, metadata.ErrNotSearchable):
			response.BadRequest(w, err.Error())
		default:
			response.InternalError(w, err.Error())
		}
		return
	}

	if result == nil {
		response.NotFound(w, fmt.Sprintf("source %q found no result for id %q", sourceName, sourceID))
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	var progress *sqlc.ReadingProgress
	if p, err := h.queries.GetProgressForArchive(r.Context(), sqlc.GetProgressForArchiveParams{
		ArchiveID: id,
		UserID:    userID,
	}); err == nil {
		progress = &p
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get progress failed", "id", id, "error", err)
	}

	isFavorited, err := h.queries.ArchiveIsFavorited(r.Context(), sqlc.ArchiveIsFavoritedParams{
		ArchiveID: id,
		Uid:       userID,
	})
	if err != nil {
		h.logger.Error("get favorite status failed", "id", id, "error", err)
	}

	var rating *int
	if v, err := h.queries.GetArchiveRating(r.Context(), sqlc.GetArchiveRatingParams{
		ArchiveID: id,
		Uid:       userID,
	}); err == nil {
		iv := int(v)
		rating = &iv
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get rating failed", "id", id, "error", err)
	}

	preview := overlayResult(archive, result)

	response.JSON(w, http.StatusOK, buildArchiveResponse(h.processor, preview, result, progress, isFavorited, rating))
}
