package archive

import (
	"context"
	"os"
	"path/filepath"

	"Shoka/internal/archive/metadata"
	"Shoka/internal/config"
)

// Delete fully deletes an archive with all its metadata
func (a *Archive) Delete(ctx context.Context, app *config.App, removeFile bool) error {
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	qtx := app.Repo.WithTx(tx)

	// Recount metadata
	old, err := metadata.GetCurrent(ctx, qtx, a.ID)
	if err := metadata.RemoveTags(ctx, qtx, a.ID, old.Tags); err != nil {
		return err
	}
	if err := metadata.RemoveArtists(ctx, qtx, a.ID, old.Artists); err != nil {
		return err
	}
	if err := metadata.RemoveCharacters(ctx, qtx, a.ID, old.Characters); err != nil {
		return err
	}
	if err := metadata.RemoveParodies(ctx, qtx, a.ID, old.Parodies); err != nil {
		return err
	}
	if err := metadata.RemoveURLs(ctx, qtx, a.ID, old.URLs); err != nil {
		return err
	}

	// Remove from database
	if err := qtx.DeleteArchive(ctx, a.ID); err != nil {
		return err
	}

	if removeFile {
		// Soft Delete
		tempFilePath := filepath.Join(app.Cfg.TempDir, filepath.Base(a.FilePath)+".deleted")
		err = os.Rename(a.FilePath, tempFilePath)
		if err != nil {
			return err
		}

		// Commit
		err = tx.Commit(ctx)
		if err != nil {
			revertErr := os.Rename(tempFilePath, a.FilePath)
			if revertErr != nil {
				return err
			}
			return err
		}

		err = os.Remove(tempFilePath)
		if err != nil {
			app.Log.Warn("Archive deleted from database, but failed to clean up tempfile", "id", a.ID, "tempfile", tempFilePath)
		} else {
			app.Log.Info("Archive removed", "id", a.ID, "file", a.FilePath)
		}
	} else {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		app.Log.Info("Archive removed", "id", a.ID)
	}

	return nil
}
