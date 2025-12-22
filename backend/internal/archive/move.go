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
	// 1. Generate new hash based on updated file
	a.setHash()
	// 2. Update hash in database
	if err := a.app.Repo.UpdateFileHash(ctx, repository.UpdateFileHashParams{
		FileHash:  a.FileHash,
		ID:        a.ID,
		UpdatedAt: time.Now(),
	}); err != nil {
		return err
	}
	// 3. Concatenate new ThumbPath based on newly generated hash
	d := fsutil.NewArchiveDir(a.app.Cfg.ThumbDir, a.FileHash)
	newPath, err := d.GenThumbDir()
	if err != nil {
		return err
	}
	// 4. Rename old ThumbPath to new ThumbPath
	if err := os.Rename(*a.ThumbPath, newPath); err != nil {
		return err
	}
	// 5. Update ThumbPath in database
	if err := a.app.Repo.UpdateThumbPath(ctx, repository.UpdateThumbPathParams{
		ThumbPath: &newPath,
		ID:        a.ID,
		UpdatedAt: time.Now(),
	}); err != nil {
		return err
	}
	// 6. Update cover name to newly generated hash
	oldCover := fmt.Sprintf("%s.webp", oldHash)
	newCover := fmt.Sprintf("%s.webp", a.FileHash)
	oldCoverPath := filepath.Join(newPath, "cover", oldCover)
	newCoverPath := filepath.Join(newPath, "cover", newCover)
	if err := os.Rename(oldCoverPath, newCoverPath); err != nil {
		return err
	}

	return nil
}
