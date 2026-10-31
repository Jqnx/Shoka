package metadata

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/util"
	"context"
	"encoding/json"
	"fmt"
)

// queries aliases the sqlc queries type so the relation descriptors below can
// reference its methods as compact method expressions.
type queries = *sqlc.Queries

// relationSpec describes one archive relation (artists, circles, tags, …) so
// the shared applyRelation reconciler can drive it. The sqlc calls differ only
// in their parameter structs and row types, so newRelation adapts the raw sqlc
// functions into the closures below, keeping applyRelation free of sqlc row
// and parameter types.
type relationSpec struct {
	// name is the plural label used in upsert/get/link error messages.
	name string
	// countName is the singular noun used in the counter error messages.
	countName  string
	upsert     func(ctx context.Context, q *sqlc.Queries, namesJSON []byte) error
	desiredIDs func(ctx context.Context, q *sqlc.Queries, names []string) ([]int64, error)
	currentIDs func(ctx context.Context, q *sqlc.Queries, archiveID string) ([]int64, error)
	link       func(ctx context.Context, q *sqlc.Queries, archiveID string, idsJSON []byte) error
	unlink     func(ctx context.Context, q *sqlc.Queries, archiveID string, idsJSON []byte) error
	increment  func(ctx context.Context, q *sqlc.Queries, ids []int64) error
	decrement  func(ctx context.Context, q *sqlc.Queries, ids []int64) error
}

// newRelation assembles a relationSpec from a relation's raw sqlc functions
// (passed as method expressions) plus the two small builders that turn a set
// of IDs into the relation's BulkAdd*/BulkRemove* parameter struct. Row is the
// bulk-get row type; AddP/RemoveP are the link/unlink parameter structs.
func newRelation[Row, AddP, RemoveP any](
	name string,
	countName string,
	upsert func(q queries, ctx context.Context, namesJSON any) error,
	getRows func(q queries, ctx context.Context, names []string) ([]Row, error),
	current func(q queries, ctx context.Context, archiveID string) ([]int64, error),
	add func(q queries, ctx context.Context, params AddP) error,
	addParams func(archiveID string, idsJSON []byte) AddP,
	remove func(q queries, ctx context.Context, params RemoveP) error,
	removeParams func(archiveID string, idsJSON []byte) RemoveP,
	increment func(q queries, ctx context.Context, ids []int64) error,
	decrement func(q queries, ctx context.Context, ids []int64) error,
	rowID func(Row) int64,
) relationSpec {
	return relationSpec{
		name:      name,
		countName: countName,
		upsert: func(ctx context.Context, q *sqlc.Queries, namesJSON []byte) error {
			return upsert(q, ctx, namesJSON)
		},
		desiredIDs: func(ctx context.Context, q *sqlc.Queries, names []string) ([]int64, error) {
			rows, err := getRows(q, ctx, names)
			if err != nil {
				return nil, err
			}

			return idsOf(rows, rowID), nil
		},
		currentIDs: func(ctx context.Context, q *sqlc.Queries, archiveID string) ([]int64, error) {
			return current(q, ctx, archiveID)
		},
		link: func(ctx context.Context, q *sqlc.Queries, archiveID string, idsJSON []byte) error {
			return add(q, ctx, addParams(archiveID, idsJSON))
		},
		unlink: func(ctx context.Context, q *sqlc.Queries, archiveID string, idsJSON []byte) error {
			return remove(q, ctx, removeParams(archiveID, idsJSON))
		},
		increment: func(ctx context.Context, q *sqlc.Queries, ids []int64) error {
			return increment(q, ctx, ids)
		},
		decrement: func(ctx context.Context, q *sqlc.Queries, ids []int64) error {
			return decrement(q, ctx, ids)
		},
	}
}

// idsOf extracts the primary key from each row produced by a BulkGet* query.
func idsOf[T any](rows []T, id func(T) int64) []int64 {
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = id(row)
	}

	return ids
}

var (
	artistsRelation = newRelation(
		"artists", "artist",
		queries.BulkAddArtists,
		queries.BulkGetArtists,
		queries.GetArchiveArtistIDs,
		queries.BulkAddArchiveArtists,
		func(archiveID string, idsJSON []byte) sqlc.BulkAddArchiveArtistsParams {
			return sqlc.BulkAddArchiveArtistsParams{ArchiveID: &archiveID, Artists: idsJSON}
		},
		queries.BulkRemoveArtistFromArchive,
		func(archiveID string, idsJSON []byte) sqlc.BulkRemoveArtistFromArchiveParams {
			return sqlc.BulkRemoveArtistFromArchiveParams{ArchiveID: &archiveID, Artists: idsJSON}
		},
		queries.IncrementArtistCount,
		queries.DecrementArtistCount,
		func(r sqlc.BulkGetArtistsRow) int64 { return r.ID },
	)

	circlesRelation = newRelation(
		"circles", "circle",
		queries.BulkAddCircle,
		queries.BulkGetCircle,
		queries.GetArchiveCircleIDs,
		queries.BulkAddArchiveCircle,
		func(archiveID string, idsJSON []byte) sqlc.BulkAddArchiveCircleParams {
			return sqlc.BulkAddArchiveCircleParams{ArchiveID: &archiveID, Circle: idsJSON}
		},
		queries.BulkRemoveCircleFromArchive,
		func(archiveID string, idsJSON []byte) sqlc.BulkRemoveCircleFromArchiveParams {
			return sqlc.BulkRemoveCircleFromArchiveParams{ArchiveID: &archiveID, Circle: idsJSON}
		},
		queries.IncrementCircleCount,
		queries.DecrementCircleCount,
		func(r sqlc.BulkGetCircleRow) int64 { return r.ID },
	)

	tagsRelation = newRelation(
		"tags", "tag",
		queries.BulkAddTags,
		queries.BulkGetTags,
		queries.GetArchiveTagIDs,
		queries.BulkAddArchiveTags,
		func(archiveID string, idsJSON []byte) sqlc.BulkAddArchiveTagsParams {
			return sqlc.BulkAddArchiveTagsParams{ArchiveID: &archiveID, Tags: idsJSON}
		},
		queries.BulkRemoveTagFromArchive,
		func(archiveID string, idsJSON []byte) sqlc.BulkRemoveTagFromArchiveParams {
			return sqlc.BulkRemoveTagFromArchiveParams{ArchiveID: &archiveID, Tags: idsJSON}
		},
		queries.IncrementTagCount,
		queries.DecrementTagCount,
		func(r sqlc.BulkGetTagsRow) int64 { return r.ID },
	)

	parodiesRelation = newRelation(
		"parodies", "parody",
		queries.BulkAddParodies,
		queries.BulkGetParodies,
		queries.GetArchiveParodyIDs,
		queries.BulkAddArchiveParodies,
		func(archiveID string, idsJSON []byte) sqlc.BulkAddArchiveParodiesParams {
			return sqlc.BulkAddArchiveParodiesParams{ArchiveID: &archiveID, Parodies: idsJSON}
		},
		queries.BulkRemoveParodiesFromArchive,
		func(archiveID string, idsJSON []byte) sqlc.BulkRemoveParodiesFromArchiveParams {
			return sqlc.BulkRemoveParodiesFromArchiveParams{ArchiveID: &archiveID, Parodies: idsJSON}
		},
		queries.IncrementParodyCount,
		queries.DecrementParodyCount,
		func(r sqlc.BulkGetParodiesRow) int64 { return r.ID },
	)

	charactersRelation = newRelation(
		"characters", "character",
		queries.BulkAddCharacters,
		queries.BulkGetCharacters,
		queries.GetArchiveCharacterIDs,
		queries.BulkAddArchiveCharacters,
		func(archiveID string, idsJSON []byte) sqlc.BulkAddArchiveCharactersParams {
			return sqlc.BulkAddArchiveCharactersParams{ArchiveID: &archiveID, Characters: idsJSON}
		},
		queries.BulkRemoveCharactersFromArchive,
		func(archiveID string, idsJSON []byte) sqlc.BulkRemoveCharactersFromArchiveParams {
			return sqlc.BulkRemoveCharactersFromArchiveParams{ArchiveID: &archiveID, Characters: idsJSON}
		},
		queries.IncrementCharacterCount,
		queries.DecrementCharacterCount,
		func(r sqlc.BulkGetCharactersRow) int64 { return r.ID },
	)
)

// applyRelation reconciles one archive relation to the desired set: it upserts
// the names, resolves them to IDs, diffs against the archive's current links,
// then links/unlinks and adjusts the relation's counter table. Behaviour is
// identical for every relation; each supplies its sqlc calls via rel.
func applyRelation(ctx context.Context, q *sqlc.Queries, archiveID string, names []string, rel relationSpec) error {
	namesJSON, _ := json.Marshal(names)

	if err := rel.upsert(ctx, q, namesJSON); err != nil {
		return fmt.Errorf("upsert %s: %w", rel.name, err)
	}

	desiredIDs, err := rel.desiredIDs(ctx, q, names)
	if err != nil {
		return fmt.Errorf("get %s: %w", rel.name, err)
	}

	currentIDs, err := rel.currentIDs(ctx, q, archiveID)
	if err != nil {
		return fmt.Errorf("get current %s: %w", rel.name, err)
	}

	toAdd, toRemove := util.Diff(currentIDs, desiredIDs)

	if len(toAdd) > 0 {
		toAddJSON, _ := json.Marshal(toAdd)
		if err := rel.link(ctx, q, archiveID, toAddJSON); err != nil {
			return fmt.Errorf("link %s: %w", rel.name, err)
		}
	}

	if len(toRemove) > 0 {
		toRemoveJSON, _ := json.Marshal(toRemove)
		if err := rel.unlink(ctx, q, archiveID, toRemoveJSON); err != nil {
			return fmt.Errorf("unlink %s: %w", rel.name, err)
		}
	}

	if len(toAdd) > 0 {
		if err := rel.increment(ctx, q, toAdd); err != nil {
			return fmt.Errorf("increment %s counts: %w", rel.countName, err)
		}
	}

	if len(toRemove) > 0 {
		if err := rel.decrement(ctx, q, toRemove); err != nil {
			return fmt.Errorf("decrement %s counts: %w", rel.countName, err)
		}
	}

	return nil
}

// linkArchiveArtistsToCircles records every (artist, circle) pair the
// archive currently has in artist_circle. Insert-only: a pair is never
// removed here, because another archive may still justify it.
func linkArchiveArtistsToCircles(ctx context.Context, q *sqlc.Queries, archiveID string) error {
	artists, err := q.GetArchiveArtistIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get archive artists: %w", err)
	}

	circles, err := q.GetArchiveCircleIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get archive circles: %w", err)
	}

	if len(artists) == 0 || len(circles) == 0 {
		return nil
	}

	artistsJSON, _ := json.Marshal(artists)
	circlesJSON, _ := json.Marshal(circles)

	if err := q.BulkAddArtistCircle(ctx, sqlc.BulkAddArtistCircleParams{
		Artists: artistsJSON,
		Circles: circlesJSON,
	}); err != nil {
		return fmt.Errorf("add artist circle pairs: %w", err)
	}

	return nil
}

// ClearRelationCounts subtracts the archive's contribution from every relation
// counter. Run it in the same transaction that deletes the archive, and before
// it: the archive_* join rows cascade away with the archive row, so its links
// are only readable beforehand.
func ClearRelationCounts(ctx context.Context, q *sqlc.Queries, archiveID string) error {
	for _, rel := range []relationSpec{artistsRelation, circlesRelation, tagsRelation, parodiesRelation, charactersRelation} {
		ids, err := rel.currentIDs(ctx, q, archiveID)
		if err != nil {
			return fmt.Errorf("get current %s: %w", rel.name, err)
		}

		if len(ids) == 0 {
			continue
		}

		if err := rel.decrement(ctx, q, ids); err != nil {
			return fmt.Errorf("decrement %s counts: %w", rel.countName, err)
		}
	}

	return nil
}

// applyURLs reconciles an archive's source links to the desired set. It
// mirrors applyRelation, diffing on the URL string via util.Diff (generic
// over comparable element types) rather than integer IDs. There are no
// counter tables for URLs, so there's no increment/decrement step. A
// non-nil empty slice clears every link.
func applyURLs(ctx context.Context, q *sqlc.Queries, archiveID string, urls []string) error {
	desired := NormalizeURLs(urls)

	current, err := q.GetArchiveUrls(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get current urls: %w", err)
	}

	toAdd, toRemove := util.Diff(current, desired)

	if len(toAdd) > 0 {
		toAddJSON, _ := json.Marshal(toAdd)
		if err := q.BulkAddArchiveURLs(ctx, sqlc.BulkAddArchiveURLsParams{
			ArchiveID: archiveID,
			Urls:      toAddJSON,
		}); err != nil {
			return fmt.Errorf("add urls: %w", err)
		}
	}

	if len(toRemove) > 0 {
		if err := q.RemoveArchiveUrl(ctx, sqlc.RemoveArchiveUrlParams{
			ArchiveID: archiveID,
			Urls:      toRemove,
		}); err != nil {
			return fmt.Errorf("remove urls: %w", err)
		}
	}

	return nil
}
