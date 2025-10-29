package archive

import (
	"context"

	"Shoka/internal/config"
	"Shoka/internal/repository"
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

	m := NewMetadata(qtx, a, archive.ID)

	if err := m.Artist(ctx); err != nil {
		return err
	}

	if err := m.Tag(ctx); err != nil {
		return err
	}

	if err := m.Character(ctx); err != nil {
		return err
	}

	if err := m.Parody(ctx); err != nil {
		return err
	}

	if err := m.URL(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
