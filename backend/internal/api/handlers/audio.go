package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/metadata"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// audioIDPattern matches the IDs produced by util.GenerateID. The id arrives
// as an untrusted path parameter and is used to build cache paths, so the
// alphabet is checked before it can escape the cache root.
var audioIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func validAudioID(id string) bool {
	return audioIDPattern.MatchString(id)
}

// AudioHandler serves the audio work list, detail, media, and per-user state
// endpoints.
type AudioHandler struct {
	queries   *sqlc.Queries
	db        *sql.DB
	processor *image.Processor
	logger    *slog.Logger
}

// NewAudioHandler returns an AudioHandler backed by queries, db, and the
// shared image processor.
func NewAudioHandler(queries *sqlc.Queries, db *sql.DB, processor *image.Processor, log *slog.Logger) *AudioHandler {
	return &AudioHandler{
		queries:   queries,
		db:        db,
		processor: processor,
		logger:    log.With("handler", "audio"),
	}
}

// AudioTrackResponse is one playable file of a work.
type AudioTrackResponse struct {
	Index           int    `json:"index"`
	Name            string `json:"name"`
	DurationSeconds int64  `json:"duration_seconds"`
}

// AudioImageResponse is one viewable image of a work.
type AudioImageResponse struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// AudioListItem is one row of the audio list.
type AudioListItem struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	LibraryID       string     `json:"library_id"`
	DurationSeconds int64      `json:"duration_seconds"`
	ReleaseDate     *time.Time `json:"release_date"`
	Language        *string    `json:"language"`
	Circle          *string    `json:"circle"`
	CoverReady      bool       `json:"cover_ready"`
	IsFavorited     bool       `json:"is_favorited"`
	Rating          *int       `json:"rating"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AudioResponse is the full representation of one audio work.
type AudioResponse struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	LibraryID       string               `json:"library_id"`
	DurationSeconds int64                `json:"duration_seconds"`
	ReleaseDate     *time.Time           `json:"release_date"`
	Language        *string              `json:"language"`
	Circle          *string              `json:"circle"`
	Artists         []string             `json:"artists"`
	Tags            []string             `json:"tags"`
	VoiceActors     []string             `json:"voice_actors"`
	Tracks          []AudioTrackResponse `json:"tracks"`
	Images          []AudioImageResponse `json:"images"`
	CoverReady      bool                 `json:"cover_ready"`
	IsFavorited     bool                 `json:"is_favorited"`
	Rating          *int                 `json:"rating"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// AudioListResponse is a single page of audio works plus the total count.
type AudioListResponse struct {
	Items []AudioListItem `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

// audioRelations holds the names of the four taxonomies an audio work can be
// linked to.
type audioRelations struct {
	artists     []string
	tags        []string
	circles     []string
	voiceActors []string
}

// loadAudioRelations reads every relation name in a fixed number of queries,
// normalizing nil to empty slices.
func (h *AudioHandler) loadAudioRelations(ctx context.Context, audioID string) (audioRelations, error) {
	var rel audioRelations

	queries := []struct {
		load func(context.Context, string) ([]string, error)
		into *[]string
	}{
		{h.queries.GetAudioArtistNames, &rel.artists},
		{h.queries.GetAudioTagNames, &rel.tags},
		{h.queries.GetAudioCircleNames, &rel.circles},
		{h.queries.GetAudioVoiceActorNames, &rel.voiceActors},
	}

	for _, q := range queries {
		names, err := q.load(ctx, audioID)
		if err != nil {
			return audioRelations{}, err
		}

		*q.into = append(make([]string, 0, len(names)), names...)
	}

	return rel, nil
}

// loadScopedAudio resolves an audio work under the caller's library scope,
// writing the failure response itself.
func (h *AudioHandler) loadScopedAudio(w http.ResponseWriter, r *http.Request, id string) (sqlc.Audio, bool) {
	if !validAudioID(id) {
		response.NotFound(w, "audio not found")
		return sqlc.Audio{}, false
	}

	audio, err := h.queries.GetAudioByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "audio not found")
			return sqlc.Audio{}, false
		}

		h.logger.Error("get audio failed", "id", id, "error", err)
		response.InternalError(w, "failed to get audio")

		return sqlc.Audio{}, false
	}

	if !auth.LibraryScopeFromContext(r.Context()).Allows(audio.LibraryID) {
		response.NotFound(w, "audio not found")
		return sqlc.Audio{}, false
	}

	return audio, true
}

// audioResponseFor assembles the full AudioResponse for one work: relations,
// ordered tracks/images, and this user's favorite/rating state.
func (h *AudioHandler) audioResponseFor(ctx context.Context, audio sqlc.Audio, userID string) (AudioResponse, error) {
	rel, err := h.loadAudioRelations(ctx, audio.ID)
	if err != nil {
		return AudioResponse{}, err
	}

	files, err := h.queries.GetAudioFiles(ctx, audio.ID)
	if err != nil {
		return AudioResponse{}, err
	}

	images, err := h.queries.GetAudioImages(ctx, audio.ID)
	if err != nil {
		return AudioResponse{}, err
	}

	isFavorited, err := h.queries.AudioIsFavorited(ctx, sqlc.AudioIsFavoritedParams{
		AudioID: audio.ID,
		Uid:     userID,
	})
	if err != nil {
		return AudioResponse{}, err
	}

	var rating *int

	if v, err := h.queries.GetAudioRating(ctx, sqlc.GetAudioRatingParams{
		AudioID: audio.ID,
		Uid:     userID,
	}); err == nil {
		iv := int(v)
		rating = &iv
	} else if !errors.Is(err, sql.ErrNoRows) {
		return AudioResponse{}, err
	}

	resp := AudioResponse{
		ID:              audio.ID,
		Title:           audio.Title,
		LibraryID:       audio.LibraryID,
		DurationSeconds: audio.DurationSeconds,
		ReleaseDate:     audio.ReleaseDate,
		Language:        audio.Language,
		Artists:         rel.artists,
		Tags:            rel.tags,
		VoiceActors:     rel.voiceActors,
		Tracks:          make([]AudioTrackResponse, 0, len(files)),
		Images:          make([]AudioImageResponse, 0, len(images)),
		CoverReady:      h.processor.AudioCoverPath(audio.ID) != "",
		IsFavorited:     isFavorited,
		Rating:          rating,
		UpdatedAt:       audio.UpdatedAt,
	}

	if len(rel.circles) > 0 {
		resp.Circle = &rel.circles[0]
	}

	for i, f := range files {
		resp.Tracks = append(resp.Tracks, AudioTrackResponse{
			Index:           i,
			Name:            filepath.Base(f.RelativePath),
			DurationSeconds: f.DurationSeconds,
		})
	}

	for i, img := range images {
		resp.Images = append(resp.Images, AudioImageResponse{
			Index: i,
			Name:  filepath.Base(img.RelativePath),
		})
	}

	return resp, nil
}

// GetAudios godoc
//
//	@Summary		List audio works with pagination
//	@Tags			audio
//	@Produce		json
//	@Param			library_id	query		string	false	"Only this audio library"
//	@Param			q			query		string	false	"Title/artist/tag/circle/voice actor search (3+ characters)"
//	@Param			favorites	query		bool	false	"Only works the current user has favorited"
//	@Param			rating		query		int		false	"Only works the current user rated this many stars (1-5)"
//	@Param			tag			query		string	false	"Filter by tag name (repeatable)"
//	@Param			artist		query		string	false	"Filter by artist name (repeatable)"
//	@Param			circle		query		string	false	"Filter by circle name (repeatable)"
//	@Param			voice_actor	query		string	false	"Filter by voice actor name (repeatable)"
//	@Param			page		query		int		false	"Page number (1-based)"	default(1)
//	@Param			limit		query		int		false	"Items per page"		default(50)
//	@Success		200	{object}	AudioListResponse
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio [get]
func (h *AudioHandler) GetAudios(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := parsePagination(r, defaultListLimit)

	f := audioFilterFromRequest(r, page, limit)

	if !h.applyAudioScope(w, r, &f) {
		return
	}

	rows, total, err := database.ListAudio(ctx, h.db, f)
	if err != nil {
		h.logger.Error("list audio failed", "error", err)
		response.InternalError(w, "failed to list audio")

		return
	}

	items, err := h.audioListItems(ctx, rows, f.UserID)
	if err != nil {
		h.logger.Error("enrich audio list failed", "error", err)
		response.InternalError(w, "failed to list audio")

		return
	}

	response.JSON(w, http.StatusOK, AudioListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// audioFilterFromRequest reads the list query parameters into an unfiltered
// (scope-wise) AudioFilter.
func audioFilterFromRequest(r *http.Request, page, limit int) database.AudioFilter {
	query := r.URL.Query()

	return database.AudioFilter{
		LibraryID:     query.Get("library_id"),
		Query:         query.Get("q"),
		Artists:       query["artist"],
		Tags:          query["tag"],
		Circles:       query["circle"],
		VoiceActors:   query["voice_actor"],
		FavoritesOnly: query.Get("favorites") == "true",
		Rating:        int(ratingFilterParam(query.Get("rating"))),
		Limit:         int64(limit),
		Offset:        int64((page - 1) * limit),
		UserID:        auth.UserIDFromContext(r.Context()),
	}
}

// applyAudioScope narrows f to the caller's library access, writing the
// failure response and returning false when it cannot.
func (h *AudioHandler) applyAudioScope(w http.ResponseWriter, r *http.Request, f *database.AudioFilter) bool {
	ctx := r.Context()
	scope := auth.LibraryScopeFromContext(ctx)

	if f.LibraryID == "" {
		if scope.Unrestricted() {
			return true
		}

		f.LibraryIDs = scope.IDs
		if f.LibraryIDs == nil {
			f.LibraryIDs = []string{}
		}

		return true
	}

	lib, err := h.queries.GetLibraryByID(ctx, f.LibraryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "library not found")
			return false
		}

		h.logger.Error("get library failed", "id", f.LibraryID, "error", err)
		response.InternalError(w, "failed to get library")

		return false
	}

	if lib.Type != "audio" || !scope.Allows(lib.ID) {
		response.NotFound(w, "library not found")
		return false
	}

	return true
}

// audioListItems attaches this user's favorite/rating state and cover
// readiness to a page of list rows.
func (h *AudioHandler) audioListItems(ctx context.Context, rows []database.AudioListRow, userID string) ([]AudioListItem, error) {
	favorited := make(map[string]bool)
	ratings := make(map[string]int)

	if len(rows) > 0 {
		ids := make([]string, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}

		favoritedIDs, err := h.queries.GetFavoritedAudioIDs(ctx, sqlc.GetFavoritedAudioIDsParams{
			Uid: userID,
			Ids: ids,
		})
		if err != nil {
			return nil, err
		}

		for _, id := range favoritedIDs {
			favorited[id] = true
		}

		ratingRows, err := h.queries.GetAudioRatingsForIDs(ctx, sqlc.GetAudioRatingsForIDsParams{
			Uid: userID,
			Ids: ids,
		})
		if err != nil {
			return nil, err
		}

		for _, row := range ratingRows {
			ratings[row.AudioID] = int(row.Rating)
		}
	}

	items := make([]AudioListItem, 0, len(rows))
	for _, row := range rows {
		item := AudioListItem{
			ID:              row.ID,
			Title:           row.Title,
			LibraryID:       row.LibraryID,
			DurationSeconds: row.DurationSeconds,
			ReleaseDate:     row.ReleaseDate,
			Language:        row.Language,
			Circle:          row.Circle,
			CoverReady:      h.processor.AudioCoverPath(row.ID) != "",
			IsFavorited:     favorited[row.ID],
			UpdatedAt:       row.UpdatedAt,
		}

		if rating, ok := ratings[row.ID]; ok {
			item.Rating = &rating
		}

		items = append(items, item)
	}

	return items, nil
}

// GetAudio godoc
//
//	@Summary		Get a single audio work
//	@Tags			audio
//	@Produce		json
//	@Param			id	path		string	true	"Audio ID"
//	@Success		200	{object}	AudioResponse
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/{id} [get]
func (h *AudioHandler) GetAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	audio, ok := h.loadScopedAudio(w, r, id)
	if !ok {
		return
	}

	resp, err := h.audioResponseFor(r.Context(), audio, auth.UserIDFromContext(r.Context()))
	if err != nil {
		h.logger.Error("build audio response failed", "id", id, "error", err)
		response.InternalError(w, "failed to get audio")

		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// UpdateAudioRequest is the body of a full metadata edit. Every array is a
// full replacement; a missing or null array clears that relation.
type UpdateAudioRequest struct {
	Title       string   `json:"title"`
	ReleaseDate *string  `json:"release_date"` // "YYYY-MM-DD" or null
	Language    *string  `json:"language"`
	Circle      *string  `json:"circle"`
	Artists     []string `json:"artists"`
	Tags        []string `json:"tags"`
	VoiceActors []string `json:"voice_actors"`
}

// UpdateAudio godoc
//
//	@Summary		Update an audio work's metadata
//	@Description	Full replacement of the scalar metadata and every relation. This is a manual edit: there is no remote metadata source for audio.
//	@Tags			audio
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Audio ID"
//	@Param			body	body		UpdateAudioRequest	true	"Fields to set"
//	@Success		200	{object}	AudioResponse
//	@Failure		400	{object}	response.Error
//	@Failure		403	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/{id} [patch]
func (h *AudioHandler) UpdateAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	audio, ok := h.loadScopedAudio(w, r, id)
	if !ok {
		return
	}

	var body UpdateAudioRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		response.BadRequest(w, "title is required")
		return
	}

	releaseDate, ok := parseAudioReleaseDate(w, body.ReleaseDate)
	if !ok {
		return
	}

	result := &metadata.AudioResult{
		Title:       title,
		ReleaseDate: releaseDate,
		Language:    optionalName(body.Language),
		Artists:     body.Artists,
		Tags:        body.Tags,
		VoiceActors: body.VoiceActors,
	}

	if circle := optionalName(body.Circle); circle != nil {
		result.Circles = []string{*circle}
	}

	if err := metadata.ApplyAudioMetadata(r.Context(), h.queries, h.db, audio.ID, result); err != nil {
		h.logger.Error("apply audio metadata failed", "id", id, "error", err)
		response.InternalError(w, "failed to update audio")

		return
	}

	updated, err := h.queries.GetAudioByID(r.Context(), audio.ID)
	if err != nil {
		h.logger.Error("reload audio failed", "id", id, "error", err)
		response.InternalError(w, "failed to update audio")

		return
	}

	resp, err := h.audioResponseFor(r.Context(), updated, auth.UserIDFromContext(r.Context()))
	if err != nil {
		h.logger.Error("build audio response failed", "id", id, "error", err)
		response.InternalError(w, "failed to update audio")

		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// parseAudioReleaseDate parses the optional "YYYY-MM-DD" release date,
// writing the 400 response and returning false when it is malformed.
func parseAudioReleaseDate(w http.ResponseWriter, raw *string) (*time.Time, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}

	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*raw))
	if err != nil {
		response.BadRequest(w, "invalid release date")

		return nil, false
	}

	return &parsed, true
}

// optionalName trims an optional name, turning a blank into nil.
func optionalName(raw *string) *string {
	if raw == nil {
		return nil
	}

	name := strings.TrimSpace(*raw)
	if name == "" {
		return nil
	}

	return &name
}

// GetCover godoc
//
//	@Summary		Get an audio work's cover image
//	@Tags			audio
//	@Produce		image/webp
//	@Param			id	path		string	true	"Audio ID"
//	@Success		200	{file}		binary
//	@Failure		404	{object}	response.Error
//	@Router			/api/audio/{id}/cover [get]
func (h *AudioHandler) GetCover(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, ok := h.loadScopedAudio(w, r, id); !ok {
		return
	}

	path := h.processor.AudioCoverPath(id)
	if path == "" {
		// A 404 is cacheable, and this one stops being true as soon as the
		// cover job finishes.
		w.Header().Set("Cache-Control", "no-store")
		response.NotFound(w, "cover not ready")

		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	//nolint:gosec // G703: id is constrained to [A-Za-z0-9]+ by validAudioID
	http.ServeFile(w, r, path)
}

// GetImage godoc
//
//	@Summary		Get one image of an audio work
//	@Tags			audio
//	@Produce		image/*
//	@Param			id		path		string	true	"Audio ID"
//	@Param			index	path		int		true	"0-based image index"
//	@Success		200		{file}		binary
//	@Failure		404		{object}	response.Error
//	@Router			/api/audio/{id}/images/{index} [get]
func (h *AudioHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	audio, ok := h.loadScopedAudio(w, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}

	images, err := h.queries.GetAudioImages(r.Context(), audio.ID)
	if err != nil {
		h.logger.Error("get audio images failed", "id", audio.ID, "error", err)
		response.InternalError(w, "failed to get image")

		return
	}

	index, ok := mediaIndex(w, r, len(images))
	if !ok {
		return
	}

	path, ok := audioMediaPath(w, audio.FolderPath, images[index].RelativePath)
	if !ok {
		return
	}

	w.Header().Set("Cache-Control", "private, max-age=86400")

	http.ServeFile(w, r, path)
}

// GetTrack godoc
//
//	@Summary		Get one audio file of a work
//	@Description	Supports HTTP Range requests, so the browser can seek.
//	@Tags			audio
//	@Produce		audio/*
//	@Param			id		path		string	true	"Audio ID"
//	@Param			index	path		int		true	"0-based track index"
//	@Success		200		{file}		binary
//	@Success		206		{file}		binary
//	@Failure		404		{object}	response.Error
//	@Router			/api/audio/{id}/tracks/{index} [get]
func (h *AudioHandler) GetTrack(w http.ResponseWriter, r *http.Request) {
	audio, ok := h.loadScopedAudio(w, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}

	files, err := h.queries.GetAudioFiles(r.Context(), audio.ID)
	if err != nil {
		h.logger.Error("get audio files failed", "id", audio.ID, "error", err)
		response.InternalError(w, "failed to get track")

		return
	}

	index, ok := mediaIndex(w, r, len(files))
	if !ok {
		return
	}

	path, ok := audioMediaPath(w, audio.FolderPath, files[index].RelativePath)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", audioContentType(path))

	http.ServeFile(w, r, path)
}

// mediaIndex parses the index path parameter, answering 404 when it is not a
// valid position in a list of length n.
func mediaIndex(w http.ResponseWriter, r *http.Request, n int) (int, bool) {
	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil || index < 0 || index >= n {
		response.NotFound(w, "not found")
		return 0, false
	}

	return index, true
}

// audioMediaPath joins a stored relative path onto the work folder, refusing
// anything that could escape it.
func audioMediaPath(w http.ResponseWriter, folderPath, relativePath string) (string, bool) {
	if !filepath.IsLocal(relativePath) {
		response.NotFound(w, "not found")
		return "", false
	}

	return filepath.Join(folderPath, relativePath), true
}

// audioContentType maps a track extension to the type the browser needs to
// play it.
func audioContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		return "audio/mpeg"
	case ".m4a", ".aac":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	case ".wav":
		return "audio/wav"
	case ".ogg", ".opus":
		return "audio/ogg"
	default:
		return "application/octet-stream"
	}
}

// mutateScopedAudio resolves the work under the caller's scope, runs mutate,
// then answers 204.
func (h *AudioHandler) mutateScopedAudio(
	w http.ResponseWriter,
	r *http.Request,
	logMsg, failMsg string,
	mutate func(id, userID string) error,
) {
	mutateScoped(w, r, h.logger, logMsg, failMsg, func(id string) bool {
		_, ok := h.loadScopedAudio(w, r, id)
		return ok
	}, mutate)
}

// AddFavorite godoc
//
//	@Summary		Favorite an audio work
//	@Tags			audio
//	@Param			id	path	string	true	"Audio ID"
//	@Success		204
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/{id}/favorite [put]
func (h *AudioHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	h.mutateScopedAudio(w, r, "add favorite audio failed", "failed to favorite audio",
		func(id, userID string) error {
			return h.queries.AddFavoriteAudio(r.Context(), sqlc.AddFavoriteAudioParams{
				AudioID: id,
				UserID:  userID,
			})
		})
}

// RemoveFavorite godoc
//
//	@Summary		Unfavorite an audio work
//	@Tags			audio
//	@Param			id	path	string	true	"Audio ID"
//	@Success		204
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/{id}/favorite [delete]
func (h *AudioHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	h.mutateScopedAudio(w, r, "remove favorite audio failed", "failed to unfavorite audio",
		func(id, userID string) error {
			return h.queries.RemoveFavoriteAudio(r.Context(), sqlc.RemoveFavoriteAudioParams{
				AudioID: id,
				Uid:     userID,
			})
		})
}

// SetRating godoc
//
//	@Summary		Rate an audio work
//	@Tags			audio
//	@Accept			json
//	@Param			id		path	string				true	"Audio ID"
//	@Param			body	body	SetRatingRequest	true	"Rating (1-5)"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/{id}/rating [put]
func (h *AudioHandler) SetRating(w http.ResponseWriter, r *http.Request) {
	var body SetRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if body.Rating < 1 || body.Rating > 5 {
		response.BadRequest(w, "rating must be between 1 and 5")
		return
	}

	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	if _, ok := h.loadScopedAudio(w, r, id); !ok {
		return
	}

	if _, err := h.queries.UpsertAudioRating(r.Context(), sqlc.UpsertAudioRatingParams{
		AudioID: id,
		UserID:  userID,
		Rating:  int64(body.Rating),
	}); err != nil {
		h.logger.Error("upsert audio rating failed", "id", id, "error", err)
		response.InternalError(w, "failed to rate audio")

		return
	}

	response.NoContent(w)
}

// RemoveRating godoc
//
//	@Summary		Clear an audio work's rating
//	@Tags			audio
//	@Param			id	path	string	true	"Audio ID"
//	@Success		204
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/{id}/rating [delete]
func (h *AudioHandler) RemoveRating(w http.ResponseWriter, r *http.Request) {
	h.mutateScopedAudio(w, r, "remove audio rating failed", "failed to clear rating",
		func(id, userID string) error {
			return h.queries.RemoveAudioRating(r.Context(), sqlc.RemoveAudioRatingParams{
				AudioID: id,
				Uid:     userID,
			})
		})
}
