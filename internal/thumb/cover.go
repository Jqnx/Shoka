package thumb

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"fmt"
	"os"
	"path/filepath"

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

func (c *Cover) CreateDir(thumbdir string) {
	cd := filepath.Join(thumbdir, "cover")
	err := fsutil.CreateDir(cd)
	if err != nil {
		return
	}
	c.Dir = cd
}

func (c *Cover) Generate() (string, error) {
	// Create temporary file
	temp, err := os.CreateTemp("", "tempCover-*")
	if err != nil {
		return "", err
	}
	// Defer removal of temp file
	defer os.Remove(temp.Name())

	// Bind first image in Archive to temp file
	// TODO: Other types of media
	switch c.Archive.Type {
	case "archive":
		fsutil.ExtractFirstPage(*c.Archive.FilePath, temp)
	}

	// Convert temp file to WEBP image
	img, err := ToWEBP(temp.Name())
	if err != nil {
		return "", err
	}

	// Save new image
	name := fmt.Sprintf("%v.webp", c.Archive.Hash)
	fp := filepath.Join(c.Dir, name)
	full, err := filepath.Abs(fp)
	if err != nil {
		return "", err
	}
	err = bimg.Write(full, img)
	if err != nil {
		return "", err
	}

	// Close temp file
	if err := temp.Close(); err != nil {
		return "", err
	}

	return fp, nil
}

//func GenerateBigCover(a *archive.Archive, thumbdir string, app *config.App) (string, error) {
//	// Create temporary file
//	temp, err := os.CreateTemp("", "tempCover-*")
//	if err != nil {
//		return "", err
//	}
//	// Defer removal of temp file
//	defer os.Remove(temp.Name())
//
//	// Bind first image in Archive to temp file
//	// TODO: Other types of media
//	switch a.Type {
//	case "archive":
//		fsutil.ExtractFirstPage(*a.FilePath, temp)
//	}
//
//	// Convert temp file to WEBP image
//	img, err := ToWEBP(temp.Name())
//	if err != nil {
//		return "", err
//	}
//
//	// Save new image
//	name := fmt.Sprintf("%v.webp", *a.Hash)
//	fp := filepath.Join(thumbdir, name)
//	err = bimg.Write(fp, img)
//	if err != nil {
//		return "", err
//	}
//
//	// Close temp file
//	if err := temp.Close(); err != nil {
//		return "", err
//	}
//
//	return name, nil
//}
