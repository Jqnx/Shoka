package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// AudioFilter describes the criteria, ordering, and paging applied when
// listing audio works. Zero values mean the corresponding condition is not
// applied. LibraryIDs is tri-state like ArchiveFilter.LibraryIDs: nil is
// unrestricted, a non-nil empty slice restricts to nothing.
type AudioFilter struct {
	LibraryID     string
	LibraryIDs    []string
	Query         string
	Artists       []string
	Tags          []string
	Circles       []string
	VoiceActors   []string
	FavoritesOnly bool
	Rating        int
	Limit         int64
	Offset        int64
	UserID        string
}

// AudioListRow is one row of the audio list.
type AudioListRow struct {
	ID              string
	Title           string
	LibraryID       string
	DurationSeconds int64
	ReleaseDate     *time.Time
	Language        *string
	Circle          *string
	UpdatedAt       time.Time
}

// addAudioRelationFilterNames appends one EXISTS clause per value, requiring
// the work to be linked to every named row (AND semantics). Table and column
// arguments are caller-supplied constants, never user input.
func addAudioRelationFilterNames(sb *strings.Builder, args *[]any, joinTable, entityTable, joinCol string, names []string) {
	for _, name := range names {
		fmt.Fprintf(
			sb,
			"AND EXISTS (SELECT 1 FROM %s JOIN %s ON %s.%s = %s.id WHERE %s.audio_id = audio.id AND %s.name = ?) ",
			joinTable, entityTable, joinTable, joinCol, entityTable, joinTable, entityTable,
		)

		*args = append(*args, name)
	}
}

// buildAudioWhere assembles the WHERE clause shared by the page and count
// queries.
func buildAudioWhere(f AudioFilter) (string, []any) {
	var (
		sb   strings.Builder
		args []any
	)

	sb.WriteString("WHERE library.type = 'audio' ")

	switch {
	case f.LibraryID != "":
		sb.WriteString("AND audio.library_id = ? ")

		args = append(args, f.LibraryID)
	case f.LibraryIDs != nil:
		if len(f.LibraryIDs) == 0 {
			sb.WriteString("AND 1=0 ")
		} else {
			sb.WriteString("AND audio.library_id IN (")
			sb.WriteString(strings.TrimSuffix(strings.Repeat("?,", len(f.LibraryIDs)), ","))
			sb.WriteString(") ")

			for _, id := range f.LibraryIDs {
				args = append(args, id)
			}
		}
	}

	if f.FavoritesOnly {
		sb.WriteString("AND EXISTS (SELECT 1 FROM audio_favorite af WHERE af.audio_id = audio.id AND af.user_id = ?) ")

		args = append(args, f.UserID)
	}

	if f.Rating > 0 {
		sb.WriteString("AND EXISTS (SELECT 1 FROM audio_rating ar WHERE ar.audio_id = audio.id AND ar.user_id = ? AND ar.rating = ?) ")

		args = append(args, f.UserID, f.Rating)
	}

	addAudioRelationFilterNames(&sb, &args, "audio_artist", "artist", "artist_id", f.Artists)
	addAudioRelationFilterNames(&sb, &args, "audio_tag", "tag", "tag_id", f.Tags)
	addAudioRelationFilterNames(&sb, &args, "audio_circle", "circle", "circle_id", f.Circles)
	addAudioRelationFilterNames(&sb, &args, "audio_voice_actor", "voice_actor", "voice_actor_id", f.VoiceActors)

	if query := strings.TrimSpace(f.Query); utf8.RuneCountInString(query) >= 3 {
		sb.WriteString("AND audio.id IN (SELECT audio_id FROM audio_fts WHERE audio_fts MATCH ?) ")

		args = append(args, ftsQuery(query))
	}

	return sb.String(), args
}

// ListAudio returns one page of audio works matching f, along with the total
// number of rows matching the filter regardless of limit and offset.
func ListAudio(ctx context.Context, db *sql.DB, f AudioFilter) ([]AudioListRow, int64, error) {
	whereSQL, whereArgs := buildAudioWhere(f)

	var total int64

	countSQL := "SELECT COUNT(*) FROM audio JOIN library ON library.id = audio.library_id " + whereSQL
	if err := db.QueryRowContext(ctx, countSQL, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audio: %w", err)
	}

	//nolint:gosec // G202: concatenation is of trusted constants; all values are bound parameters
	selectSQL := `SELECT audio.id, audio.title, audio.library_id, audio.duration_seconds,
		audio.release_date, audio.language,
		(SELECT c.name FROM audio_circle ac JOIN circle c ON c.id = ac.circle_id WHERE ac.audio_id = audio.id LIMIT 1) AS circle,
		audio.updated_at
	FROM audio
	JOIN library ON library.id = audio.library_id
	` + whereSQL + " ORDER BY audio.title ASC, audio.id ASC LIMIT ? OFFSET ?"

	selectArgs := append(append([]any{}, whereArgs...), f.Limit, f.Offset)

	rows, err := db.QueryContext(ctx, selectSQL, selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audio: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []AudioListRow

	for rows.Next() {
		var row AudioListRow
		if err := rows.Scan(
			&row.ID, &row.Title, &row.LibraryID, &row.DurationSeconds,
			&row.ReleaseDate, &row.Language, &row.Circle, &row.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audio row: %w", err)
		}

		items = append(items, row)
	}

	return items, total, rows.Err()
}
