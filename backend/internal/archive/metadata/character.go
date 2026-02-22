package metadata

import (
	"context"

	"Shoka/internal/repository"
	"Shoka/internal/util"
)

func RemoveCharacters(ctx context.Context, qtx *repository.Queries, id string, characters []int32) error {
	if err := qtx.RemoveCharacterFromArchive(ctx, repository.RemoveCharacterFromArchiveParams{
		ArchiveID:  id,
		Characters: characters,
	}); err != nil {
		return err
	}

	if err := qtx.DecrementCharacterCount(ctx, characters); err != nil {
		return err
	}

	return nil
}

func AddCharacters(ctx context.Context, qtx *repository.Queries, id string, characters []int32) error {
	if err := qtx.BulkAddArchiveCharacters(ctx, repository.BulkAddArchiveCharactersParams{
		ArchiveID:  id,
		Characters: characters,
	}); err != nil {
		return err
	}

	if err := qtx.IncrementCharacterCount(ctx, characters); err != nil {
		return err
	}

	return nil
}

func UpdateCharacters(ctx context.Context, qtx *repository.Queries, id string, oldCharacterIDs []int32, newCharacters []string) error {
	resolved, err := qtx.EnsureCharacterExist(ctx, newCharacters)
	if err != nil {
		return err
	}

	var newCharacterIDs []int32
	for _, i := range resolved {
		newCharacterIDs = append(newCharacterIDs, i.ID)
	}

	insert, remove := util.CalculateDiff(oldCharacterIDs, newCharacterIDs)

	if len(insert) > 0 {
		if err := AddCharacters(ctx, qtx, id, insert); err != nil {
			return err
		}
	}

	if len(remove) > 0 {
		if err := RemoveCharacters(ctx, qtx, id, remove); err != nil {
			return err
		}
	}

	return nil
}
