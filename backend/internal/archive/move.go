package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
)

func (a *Archive) MoveOnFileUpdate() error {
	ctx := context.Background()
	oldHash := a.FileHash

	a.setHash()

	d := fsutil.NewArchiveDir(a.app.Cfg.ThumbDir, a.FileHash)
	newPath, err := d.GenThumbDir()
	if err != nil {
		return err
	}

	newPathParent := filepath.Dir(newPath)
	if err := os.MkdirAll(newPathParent, 0o755); err != nil {
		return err
	}

	if err := os.Rename(*a.ThumbPath, newPath); err != nil {
		return err
	}

	if err := a.app.Repo.UpdateThumbPath(ctx, repository.UpdateThumbPathParams{
		ThumbPath: &newPath,
		ID:        a.ID,
		UpdatedAt: time.Now(),
	}); err != nil {
		return err
	}

	oldCover := fmt.Sprintf("%s.webp", oldHash)
	newCover := fmt.Sprintf("%s.webp", a.FileHash)
	oldCoverPath := filepath.Join(newPath, "cover", oldCover)
	newCoverPath := filepath.Join(newPath, "cover", newCover)
	if err := os.Rename(oldCoverPath, newCoverPath); err != nil {
		return err
	}

	if err := a.app.Repo.UpdateFileHash(ctx, repository.UpdateFileHashParams{
		FileHash:  a.FileHash,
		ID:        a.ID,
		UpdatedAt: time.Now(),
	}); err != nil {
		return err
	}

	return nil
}
