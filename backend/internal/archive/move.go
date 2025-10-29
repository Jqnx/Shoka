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
	oldHash := a.Hash
	// 1. Generate new hash based on updated file
	a.setHash()
	// 2. Update hash in database
	if err := a.app.Repo.UpdateHash(ctx, repository.UpdateHashParams{
		Hash:      a.Hash,
		ID:        a.ID,
		UpdatedAt: time.Now(),
	}); err != nil {
		return err
	}
	// 3. Concatonate new ThumbsPath based on newly generated hash
	d := fsutil.NewArchiveDir(a.app.Cfg.ThumbDir, a.Type, a.Hash)
	newPath, err := d.GenThumbDir()
	if err != nil {
		return err
	}
	// 4. Rename old ThumbsPath to new ThumbsPath
	if err := os.Rename(*a.ThumbsPath, newPath); err != nil {
		return err
	}
	// 5. Update thumbspath in database
	if err := a.app.Repo.UpdateThumbPath(ctx, repository.UpdateThumbPathParams{
		ThumbsPath: &newPath,
		ID:         a.ID,
		UpdatedAt:  time.Now(),
	}); err != nil {
		return err
	}
	// 6. Update cover name to newly generated hash
	oldCover := fmt.Sprintf("%s.webp", oldHash)
	newCover := fmt.Sprintf("%s.webp", a.Hash)
	oldCoverPath := filepath.Join(newPath, "cover", oldCover)
	newCoverPath := filepath.Join(newPath, "cover", newCover)
	if err := os.Rename(oldCoverPath, newCoverPath); err != nil {
		return err
	}

	return nil
}
