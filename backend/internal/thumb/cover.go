package thumb

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"

	"github.com/h2non/bimg"
)

// TODO:
// Update transaction to save thumbhash and thumb_path to db
// GenerateCover, GenerateThumbnail, GeneratePreview
//
// Generate hash function in fsutil
// https://github.com/photoprism/photoprism/blob/develop/pkg/fs/hash.go#L13
//
//	Generate path for thumb
// https://github.com/photoprism/photoprism/blob/7d56a454dc64a88fc76194b84ca2e030632afca9/internal/thumb/create.go#L45

type Cover struct {
	Dir     string
	App     *config.App
	Archive *repository.GetArchiveByIDRow
}

func NewCover(a *repository.GetArchiveByIDRow, app *config.App, coverdir string) *Cover {
	return &Cover{
		Dir:     coverdir,
		App:     app,
		Archive: a,
	}
}

func (c *Cover) Generate() (string, error) {
	ctx := context.Background()
	var fileByte []byte
	var fileName string

	// Bind first image in Archive to temp file
	// TODO: Other types of media
	switch c.Archive.Type {
	case "archive":
		arch, err := fsutil.OpenArchive(c.Archive.FilePath)
		if err != nil {
			return "", err
		}
		first, err := arch.GetFirstFileName(true)
		if err != nil {
			return "", err
		}
		readFile, err := arch.ReadFile(first)
		if err != nil {
			return "", err
		}
		fileByte = readFile
		fileName = filepath.Base(first)
	}

	// Convert temp file to WEBP image
	img, err := ToWEBP(fileByte)
	if err != nil {
		c.App.Log.Error("error converting to webp image", "err", err.Error())
		return "", err
	}

	name := fmt.Sprintf("%s.webp", c.Archive.Hash)
	fp := filepath.Join(c.Dir, name)
	full, err := filepath.Abs(fp)
	if err != nil {
		return "", err
	}

	// Update new image in database
	coverPath := fmt.Sprintf("/cover/%s", name)
	if err := c.App.Repo.UpdateCoverInfo(ctx, repository.UpdateCoverInfoParams{
		ID:        c.Archive.ID,
		CoverPath: &coverPath,
		CoverImg:  &fileName,
		UpdatedAt: time.Now(),
	}); err != nil {
		return "", err
	}

	// Save new image
	err = bimg.Write(full, img)
	if err != nil {
		return "", err
	}

	return fp, nil
}
