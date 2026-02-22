package metadata

import (
	"context"

	"Shoka/internal/repository"
	"Shoka/internal/util"
)

func RemoveArtists(ctx context.Context, qtx *repository.Queries, id string, artists []int32) error {
	if err := qtx.RemoveArtistFromArchive(ctx, repository.RemoveArtistFromArchiveParams{
		ArchiveID: id,
		Artists:   artists,
	}); err != nil {
		return err
	}

	if err := qtx.DecrementArtistCount(ctx, artists); err != nil {
		return err
	}

	return nil
}

func AddArtists(ctx context.Context, qtx *repository.Queries, id string, artists []int32) error {
	if err := qtx.BulkAddArchiveArtists(ctx, repository.BulkAddArchiveArtistsParams{
		ArchiveID: id,
		Artists:   artists,
	}); err != nil {
		return err
	}

	if err := qtx.IncrementArtistCount(ctx, artists); err != nil {
		return err
	}

	return nil
}

func UpdateArtists(ctx context.Context, qtx *repository.Queries, id string, oldArtistIDs []int32, newArtists []string) error {
	resolved, err := qtx.EnsureArtistExist(ctx, newArtists)
	if err != nil {
		return err
	}

	var newArtistIDs []int32
	for _, i := range resolved {
		newArtistIDs = append(newArtistIDs, i.ID)
	}

	insert, remove := util.CalculateDiff(oldArtistIDs, newArtistIDs)

	if len(insert) > 0 {
		if err := AddArtists(ctx, qtx, id, insert); err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := RemoveArtists(ctx, qtx, id, remove); err != nil {
			return err
		}
	}

	return nil
}
