package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"Shoka/internal/database/sqlc"
	"Shoka/internal/metadata"
)

type ArchiveFilter struct {
	LibraryID  string
	Artists    []string
	Tags       []string
	Characters []string
	Parodies   []string
	Language   string
	Category   string
	Sort       string
	Limit      int64
	Offset     int64
	UserID     string
}

// SortOption is one valid value for ArchiveFilter.Sort. SortOptions is the
// single source of truth for what the "sort" query param on GET
// /api/archives accepts — both resolveSort and the /api/archives/sort-options
// endpoint (see handlers.GetArchiveSortOptions) read from it, so adding a
// new sort only requires a change here.
type SortOption struct {
	Value       string
	DisplayName string
	clause      string
}

var SortOptions = []SortOption{
	{"title_asc", "Title (A–Z)", "archive.title ASC"},
	{"title_desc", "Title (Z–A)", "archive.title DESC"},
	{"release_date_desc", "Release Date (Newest)", "archive.release_date DESC"},
	{"release_date_asc", "Release Date (Oldest)", "archive.release_date ASC"},
	{"created_at_desc", "Date Added (Newest)", "archive.created_at DESC"},
	{"created_at_asc", "Date Added (Oldest)", "archive.created_at ASC"},
	{"page_count_desc", "Page Count (High–Low)", "archive.page_count DESC"},
	{"page_count_asc", "Page Count (Low–High)", "archive.page_count ASC"},
}

const defaultSort = "created_at_desc"

func resolveSort(s string) string {
	for _, opt := range SortOptions {
		if opt.Value == s {
			return opt.clause
		}
	}
	for _, opt := range SortOptions {
		if opt.Value == defaultSort {
			return opt.clause
		}
	}
	return "archive.created_at DESC"
}

// addRelationFilterNames appends one "AND EXISTS (...)" clause per value in names,
// requiring the archive to be associated with every named row (AND semantics).
func addRelationFilterNames(sb *strings.Builder, args *[]any, joinTable, entityTable, joinCol string, names []string) {
	for _, name := range names {
		if name == "" {
			continue
		}
		sb.WriteString("AND EXISTS (SELECT 1 FROM " + joinTable + " JOIN " + entityTable + " ON " + joinTable + "." + joinCol + " = " + entityTable + ".id WHERE " + joinTable + ".archive_id = archive.id AND " + entityTable + ".name = ?) ")
		*args = append(*args, name)
	}
}

func buildWhere(f ArchiveFilter) (string, []any) {
	var sb strings.Builder
	var args []any

	sb.WriteString("WHERE 1=1 ")

	// library scoping is mandatory: the app has no unified cross-library view.
	sb.WriteString("AND archive.library_id = ? ")
	args = append(args, f.LibraryID)

	addRelationFilterNames(&sb, &args, "archive_artist", "artist", "artist_id", f.Artists)
	addRelationFilterNames(&sb, &args, "archive_tag", "tag", "tag_id", f.Tags)
	addRelationFilterNames(&sb, &args, "archive_character", "character", "character_id", f.Characters)
	addRelationFilterNames(&sb, &args, "archive_parody", "parody", "parody_id", f.Parodies)

	if f.Language != "" {
		sb.WriteString("AND archive.language = ? ")
		args = append(args, f.Language)
	}
	if f.Category != "" {
		sb.WriteString("AND archive.category = ? ")
		args = append(args, f.Category)
	}

	return sb.String(), args
}

func ListArchives(ctx context.Context, db *sql.DB, f ArchiveFilter) ([]sqlc.GetArchiveListRow, int64, error) {
	orderBy := resolveSort(f.Sort)
	whereSQL, whereArgs := buildWhere(f)

	var total int64
	countSQL := "SELECT COUNT(*) FROM archive " + whereSQL
	if err := db.QueryRowContext(ctx, countSQL, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count archives: %w", err)
	}

	selectSQL := `SELECT archive.id, archive.title, archive.summary, archive.language, archive.category,
		archive.page_count, archive.file_path, archive.file_size, archive.mod_time,
		archive.created_at, archive.updated_at, archive.release_date,
		reading_progress.page, reading_progress.last_read, reading_progress.completed
	FROM archive
	LEFT JOIN reading_progress ON archive.id = reading_progress.archive_id AND reading_progress.user_id = ?
	` + whereSQL + " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"

	selectArgs := append([]any{f.UserID}, whereArgs...)
	selectArgs = append(selectArgs, f.Limit, f.Offset)

	rows, err := db.QueryContext(ctx, selectSQL, selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list archives: %w", err)
	}
	defer rows.Close()

	var items []sqlc.GetArchiveListRow
	for rows.Next() {
		var row sqlc.GetArchiveListRow
		if err := rows.Scan(
			&row.ID, &row.Title, &row.Summary, &row.Language, &row.Category,
			&row.PageCount, &row.FilePath, &row.FileSize, &row.ModTime,
			&row.CreatedAt, &row.UpdatedAt, &row.ReleaseDate,
			&row.Page, &row.LastRead, &row.Completed,
		); err != nil {
			return nil, 0, fmt.Errorf("scan archive row: %w", err)
		}
		items = append(items, row)
	}

	return items, total, rows.Err()
}

// getRelationNames fetches, for each archive in archiveIDs, the names of related
// rows through a join table (e.g. archive_tag -> tag), grouped by archive ID.
func getRelationNames(ctx context.Context, db *sql.DB, joinTable, entityTable, joinCol string, archiveIDs []string) (map[string][]string, error) {
	result := make(map[string][]string)
	if len(archiveIDs) == 0 {
		return result, nil
	}

	placeholders := strings.Repeat(",?", len(archiveIDs))[1:]
	query := fmt.Sprintf(
		`SELECT %s.archive_id, %s.name
		FROM %s
		JOIN %s ON %s.%s = %s.id
		WHERE %s.archive_id IN (%s)
		ORDER BY %s.name`,
		joinTable, entityTable,
		joinTable,
		entityTable, joinTable, joinCol, entityTable,
		joinTable, placeholders,
		entityTable,
	)

	args := make([]any, len(archiveIDs))
	for i, id := range archiveIDs {
		args[i] = id
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var archiveID, name string
		if err := rows.Scan(&archiveID, &name); err != nil {
			return nil, err
		}
		result[archiveID] = append(result[archiveID], name)
	}

	return result, rows.Err()
}

// GetBulkArchiveMetadata fetches artists, tags, parodies, circles, and characters
// for a set of archives in a fixed number of queries (avoiding N+1).
func GetBulkArchiveMetadata(ctx context.Context, db *sql.DB, archiveIDs []string) (map[string]*metadata.Result, error) {
	result := make(map[string]*metadata.Result, len(archiveIDs))
	for _, id := range archiveIDs {
		result[id] = &metadata.Result{}
	}

	artists, err := getRelationNames(ctx, db, "archive_artist", "artist", "artist_id", archiveIDs)
	if err != nil {
		return nil, fmt.Errorf("get artists: %w", err)
	}
	tags, err := getRelationNames(ctx, db, "archive_tag", "tag", "tag_id", archiveIDs)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}
	parodies, err := getRelationNames(ctx, db, "archive_parody", "parody", "parody_id", archiveIDs)
	if err != nil {
		return nil, fmt.Errorf("get parodies: %w", err)
	}
	circles, err := getRelationNames(ctx, db, "archive_circle", "circle", "circle_id", archiveIDs)
	if err != nil {
		return nil, fmt.Errorf("get circles: %w", err)
	}
	characters, err := getRelationNames(ctx, db, "archive_character", "character", "character_id", archiveIDs)
	if err != nil {
		return nil, fmt.Errorf("get characters: %w", err)
	}

	for id, r := range result {
		r.Artists = artists[id]
		r.Tags = tags[id]
		r.Parodies = parodies[id]
		r.Circles = circles[id]
		r.Characters = characters[id]
	}

	return result, nil
}
