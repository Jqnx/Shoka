package metadata

import (
	"context"

	"Shoka/internal/repository"
	"Shoka/internal/util"
)

func RemoveParodies(ctx context.Context, qtx *repository.Queries, id string, parodies []int32) error {
	if err := qtx.RemoveParodyFromArchive(ctx, repository.RemoveParodyFromArchiveParams{
		ArchiveID: id,
		Parodies:  parodies,
	}); err != nil {
		return err
	}

	if err := qtx.DecrementParodyCount(ctx, parodies); err != nil {
		return err
	}

	return nil
}

func AddParodies(ctx context.Context, qtx *repository.Queries, id string, parodies []int32) error {
	if err := qtx.BulkAddArchiveParodies(ctx, repository.BulkAddArchiveParodiesParams{
		ArchiveID: id,
		Parodies:  parodies,
	}); err != nil {
		return err
	}

	if err := qtx.IncrementParodyCount(ctx, parodies); err != nil {
		return err
	}

	return nil
}

func UpdateParodies(ctx context.Context, qtx *repository.Queries, id string, oldParodyIDs []int32, newParodies []string) error {
	resolved, err := qtx.EnsureParodyExist(ctx, newParodies)
	if err != nil {
		return err
	}

	var newParodyIDs []int32
	for _, i := range resolved {
		newParodyIDs = append(newParodyIDs, i.ID)
	}

	insert, remove := util.CalculateDiff(oldParodyIDs, newParodyIDs)

	if len(insert) > 0 {
		if err := AddParodies(ctx, qtx, id, insert); err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := RemoveParodies(ctx, qtx, id, remove); err != nil {
			return err
		}
	}

	return nil
}
