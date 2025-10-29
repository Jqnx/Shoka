package archive

import (
	"context"

	"Shoka/internal/config"
	"Shoka/internal/repository"
)

// Insert inserts an Archive struct into the db
func (a *Archive) Insert(c context.Context, app *config.App) error {
	_, err := app.Repo.CreateArchive(c, repository.CreateArchiveParams{
		ID:         a.ID,
		Title:      a.Title,
		PageCount:  a.PageCount,
		FilePath:   a.FilePath,
		FileName:   a.FileName,
		Hash:       a.Hash,
		ThumbsPath: a.ThumbsPath,
		Type:       a.Type,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	})
	if err != nil {
		app.Log.Error("failed to insert archive in db", "error", err)
		return err
	}
	return nil
}
