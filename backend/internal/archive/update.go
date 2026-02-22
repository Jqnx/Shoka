package archive

import (
	"context"

	"Shoka/internal/archive/metadata"
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"Shoka/internal/util"
)

// TODO: ArchivePayload to Archive converter

func (a *Archive) UpdateInDB(ctx context.Context, app *config.App) error {
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := app.Repo.WithTx(tx)

	archive, err := qtx.UpdateArchive(ctx, repository.UpdateArchiveParams{
		Title:       &a.Title,
		Summary:     a.Summary,
		Language:    a.Language,
		Category:    a.Category,
		UpdatedAt:   a.UpdatedAt,
		ReleaseDate: a.ReleaseDate,
		ID:          a.ID,
	})
	if err != nil {
		return err
	}

	old, err := metadata.GetCurrent(ctx, qtx, archive.ID)
	if err != nil {
		return err
	}
	newTags := util.ToSliceString(a.Tags)
	newURLs := util.ToSliceString(a.URL)
	newParodies := util.ToSliceString(a.Parody)
	newCharacters := util.ToSliceString(a.Character)
	newArtists := util.ToSliceString(a.Artist)

	if err := metadata.UpdateTags(ctx, qtx, archive.ID, old.Tags, newTags); err != nil {
		return err
	}

	if err := metadata.UpdateArtists(ctx, qtx, archive.ID, old.Artists, newArtists); err != nil {
		return err
	}

	if err := metadata.UpdateCharacters(ctx, qtx, archive.ID, old.Characters, newCharacters); err != nil {
		return err
	}

	if err := metadata.UpdateParodies(ctx, qtx, archive.ID, old.Parodies, newParodies); err != nil {
		return err
	}

	if err := metadata.UpdateURLs(ctx, qtx, archive.ID, old.URLs, newURLs); err != nil {
		return err
	}

	app.Log.Info("Successfully updated archive metadata", "archive", archive.ID)

	return tx.Commit(ctx)
}
