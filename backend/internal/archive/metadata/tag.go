package metadata

import (
	"context"

	"Shoka/internal/repository"
	"Shoka/internal/util"
)

func RemoveTags(ctx context.Context, qtx *repository.Queries, id string, tags []int32) error {
	if err := qtx.RemoveTagFromArchive(ctx, repository.RemoveTagFromArchiveParams{
		ArchiveID: id,
		Tags:      tags,
	}); err != nil {
		return err
	}

	if err := qtx.DecrementTagCount(ctx, tags); err != nil {
		return err
	}

	return nil
}

func AddTags(ctx context.Context, qtx *repository.Queries, id string, tags []int32) error {
	if err := qtx.BulkAddArchiveTags(ctx, repository.BulkAddArchiveTagsParams{
		ArchiveID: id,
		Tags:      tags,
	}); err != nil {
		return err
	}

	if err := qtx.IncrementTagCount(ctx, tags); err != nil {
		return err
	}

	return nil
}

func UpdateTags(ctx context.Context, qtx *repository.Queries, id string, oldTagIDs []int32, newTags []string) error {
	resolved, err := qtx.EnsureTagExist(ctx, newTags)
	if err != nil {
		return err
	}

	var newTagIDs []int32
	for _, i := range resolved {
		newTagIDs = append(newTagIDs, i.ID)
	}

	insert, remove := util.CalculateDiff(oldTagIDs, newTagIDs)

	if len(insert) > 0 {
		if err := AddTags(ctx, qtx, id, insert); err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := RemoveTags(ctx, qtx, id, remove); err != nil {
			return err
		}
	}

	return nil
}
