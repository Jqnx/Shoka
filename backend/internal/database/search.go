package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// EntitySearchResult is one name-matched row from a metadata entity table
// (tag, artist, circle, parody, character) - every one of those tables
// shares this id/name/count shape.
type EntitySearchResult struct {
	ID    int64
	Name  string
	Count int64
}

// entityLikePattern turns raw user input into a SQL LIKE pattern that
// matches it as a literal substring: '%', '_' and the escape character
// itself are escaped so they can't be smuggled in as LIKE wildcards,
// mirroring how ftsQuery (above) treats the archive FTS query as a literal
// phrase rather than a query language.
func entityLikePattern(raw string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(raw)
	return "%" + escaped + "%"
}

// SearchEntities does a substring search over the `name` column of a
// metadata entity table - the same tables GetBulkArchiveMetadata reads
// from. These tables are global (not scoped to a library, see their
// migrations), so unlike archive search there's no library_id filter here.
// Results are ordered by count (most-used first) then name, and capped at
// limit; total reports how many rows matched regardless of that cap, so
// callers can render a "N more" affordance.
//
// table is always a caller-supplied constant, never user input - same
// contract as getRelationNames/CanonicalNames below.
func SearchEntities(ctx context.Context, db *sql.DB, table, query string, limit int64) ([]EntitySearchResult, int64, error) {
	pattern := entityLikePattern(query)

	var total int64

	//nolint:gosec // G202: table is a caller-supplied constant (see doc comment); query is bound via a "?" placeholder
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE name LIKE ? ESCAPE '\\'", table)
	if err := db.QueryRowContext(ctx, countSQL, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count %s: %w", table, err)
	}

	//nolint:gosec // G202: table is a caller-supplied constant (see doc comment); query/limit are bound via "?" placeholders
	selectSQL := fmt.Sprintf(
		"SELECT id, name, count FROM %s WHERE name LIKE ? ESCAPE '\\' ORDER BY count DESC, name ASC LIMIT ?",
		table,
	)

	rows, err := db.QueryContext(ctx, selectSQL, pattern, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("search %s: %w", table, err)
	}
	defer rows.Close()

	var items []EntitySearchResult

	for rows.Next() {
		var item EntitySearchResult
		if err := rows.Scan(&item.ID, &item.Name, &item.Count); err != nil {
			return nil, 0, fmt.Errorf("scan %s row: %w", table, err)
		}

		items = append(items, item)
	}

	return items, total, rows.Err()
}
