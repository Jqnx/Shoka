package metadata

import (
	"context"

	"Shoka/internal/repository"
	"Shoka/internal/util"
)

func RemoveURLs(ctx context.Context, qtx *repository.Queries, id string, urls []string) error {
	if err := qtx.RemoveArchiveUrl(ctx, repository.RemoveArchiveUrlParams{
		ArchiveID: id,
		Urls:      urls,
	}); err != nil {
		return err
	}

	return nil
}

func AddURLs(ctx context.Context, qtx *repository.Queries, id string, urls []string) error {
	if err := qtx.BulkAddArchiveURLs(ctx, repository.BulkAddArchiveURLsParams{
		ArchiveID: id,
		Urls:      urls,
	}); err != nil {
		return err
	}

	return nil
}

func UpdateURLs(ctx context.Context, qtx *repository.Queries, id string, oldURLs []string, newURLs []string) error {
	insert, remove := util.CalculateDiff(oldURLs, newURLs)

	if len(insert) > 0 {
		if err := AddURLs(ctx, qtx, id, insert); err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := RemoveURLs(ctx, qtx, id, remove); err != nil {
			return err
		}
	}

	return nil
}
