package metadata

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// AudioResult is the metadata of one audio work: the scalar columns plus the
// names of every related artist, tag, circle and voice actor. A full edit
// replaces everything, so nil and empty slices both clear a relation.
type AudioResult struct {
	Title       string
	ReleaseDate *time.Time
	Language    *string
	Artists     []string
	Tags        []string
	Circles     []string
	VoiceActors []string
}

var (
	audioArtistsRelation = newRelation(
		"artists", "artist",
		queries.BulkAddArtists,
		queries.BulkGetAudioArtists,
		queries.GetAudioArtistIDs,
		queries.BulkAddAudioArtists,
		func(audioID string, idsJSON []byte) sqlc.BulkAddAudioArtistsParams {
			return sqlc.BulkAddAudioArtistsParams{AudioID: audioID, Artists: idsJSON}
		},
		queries.BulkRemoveAudioArtists,
		func(audioID string, idsJSON []byte) sqlc.BulkRemoveAudioArtistsParams {
			return sqlc.BulkRemoveAudioArtistsParams{AudioID: audioID, Artists: idsJSON}
		},
		queries.IncrementArtistCount,
		queries.DecrementArtistCount,
		func(r sqlc.BulkGetAudioArtistsRow) int64 { return r.ID },
	)

	audioTagsRelation = newRelation(
		"tags", "tag",
		queries.BulkAddTags,
		queries.BulkGetAudioTags,
		queries.GetAudioTagIDs,
		queries.BulkAddAudioTags,
		func(audioID string, idsJSON []byte) sqlc.BulkAddAudioTagsParams {
			return sqlc.BulkAddAudioTagsParams{AudioID: audioID, Tags: idsJSON}
		},
		queries.BulkRemoveAudioTags,
		func(audioID string, idsJSON []byte) sqlc.BulkRemoveAudioTagsParams {
			return sqlc.BulkRemoveAudioTagsParams{AudioID: audioID, Tags: idsJSON}
		},
		queries.IncrementTagCount,
		queries.DecrementTagCount,
		func(r sqlc.BulkGetAudioTagsRow) int64 { return r.ID },
	)

	audioCirclesRelation = newRelation(
		"circles", "circle",
		queries.BulkAddCircle,
		queries.BulkGetAudioCircles,
		queries.GetAudioCircleIDs,
		queries.BulkAddAudioCircles,
		func(audioID string, idsJSON []byte) sqlc.BulkAddAudioCirclesParams {
			return sqlc.BulkAddAudioCirclesParams{AudioID: audioID, Circles: idsJSON}
		},
		queries.BulkRemoveAudioCircles,
		func(audioID string, idsJSON []byte) sqlc.BulkRemoveAudioCirclesParams {
			return sqlc.BulkRemoveAudioCirclesParams{AudioID: audioID, Circles: idsJSON}
		},
		queries.IncrementCircleCount,
		queries.DecrementCircleCount,
		func(r sqlc.BulkGetAudioCirclesRow) int64 { return r.ID },
	)

	audioVoiceActorsRelation = newRelation(
		"voice actors", "voice actor",
		queries.BulkAddVoiceActors,
		queries.BulkGetAudioVoiceActors,
		queries.GetAudioVoiceActorIDs,
		queries.BulkAddAudioVoiceActors,
		func(audioID string, idsJSON []byte) sqlc.BulkAddAudioVoiceActorsParams {
			return sqlc.BulkAddAudioVoiceActorsParams{AudioID: audioID, VoiceActors: idsJSON}
		},
		queries.BulkRemoveAudioVoiceActors,
		func(audioID string, idsJSON []byte) sqlc.BulkRemoveAudioVoiceActorsParams {
			return sqlc.BulkRemoveAudioVoiceActorsParams{AudioID: audioID, VoiceActors: idsJSON}
		},
		queries.IncrementVoiceActorCount,
		queries.DecrementVoiceActorCount,
		func(r sqlc.BulkGetAudioVoiceActorsRow) int64 { return r.ID },
	)
)

// audioRelations maps each AudioResult list onto its relation descriptor, in
// the order they are applied.
func audioRelations(result *AudioResult) []struct {
	names []string
	rel   relationSpec
} {
	return []struct {
		names []string
		rel   relationSpec
	}{
		{result.Artists, audioArtistsRelation},
		{result.Tags, audioTagsRelation},
		{result.Circles, audioCirclesRelation},
		{result.VoiceActors, audioVoiceActorsRelation},
	}
}

// normalizeNames trims every name, drops blanks and drops case-insensitive
// duplicates while keeping the first spelling seen.
func normalizeNames(names []string) []string {
	out := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		out = append(out, name)
	}

	return out
}

// ApplyAudioMetadata replaces an audio work's scalar metadata and its full
// relation set in one transaction. Circles holds at most one name; extra
// values are ignored. artist_circle is insert-only here, same as archives: a
// pair is never removed because another work may still justify it.
func ApplyAudioMetadata(ctx context.Context, queries *sqlc.Queries, db *sql.DB, audioID string, result *AudioResult) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	qtx := queries.WithTx(tx)

	if err := qtx.UpdateAudio(ctx, sqlc.UpdateAudioParams{
		Title:       strings.TrimSpace(result.Title),
		ReleaseDate: result.ReleaseDate,
		Language:    result.Language,
		ID:          audioID,
	}); err != nil {
		return fmt.Errorf("update audio metadata: %w", err)
	}

	artists, err := resolveArtistAliases(ctx, qtx, normalizeNames(result.Artists))
	if err != nil {
		return fmt.Errorf("resolve artist aliases: %w", err)
	}

	result.Artists = artists
	result.Tags = normalizeNames(result.Tags)
	result.Circles = normalizeNames(result.Circles)
	result.VoiceActors = normalizeNames(result.VoiceActors)

	if len(result.Circles) > 1 {
		result.Circles = result.Circles[:1]
	}

	for _, r := range audioRelations(result) {
		if err := applyRelation(ctx, qtx, audioID, r.names, r.rel); err != nil {
			return err
		}
	}

	if err := linkAudioArtistsToCircles(ctx, qtx, audioID); err != nil {
		return fmt.Errorf("link audio artists to circles: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// linkAudioArtistsToCircles records the audio work's current artist×circle
// pairs in artist_circle.
func linkAudioArtistsToCircles(ctx context.Context, q *sqlc.Queries, audioID string) error {
	return linkArtistsToCircles(ctx, q, audioID, "audio",
		q.GetAudioArtistIDs, q.GetAudioCircleIDs)
}

// ClearAudioRelationCounts subtracts an audio work's contribution from every
// relation counter. Run it in the same transaction that deletes the work, and
// before it: the audio_* join rows cascade away with the audio row.
func ClearAudioRelationCounts(ctx context.Context, q *sqlc.Queries, audioID string) error {
	specs := []relationSpec{audioArtistsRelation, audioCirclesRelation, audioTagsRelation, audioVoiceActorsRelation}

	for _, rel := range specs {
		ids, err := rel.currentIDs(ctx, q, audioID)
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
