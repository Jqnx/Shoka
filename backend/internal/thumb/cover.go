package thumb

import (
	"fmt"
	"path/filepath"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"

	"github.com/h2non/bimg"
)

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
	var fileByte []byte

	// Bind first image in Archive to temp file
	// TODO: Other types of media
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

	// Convert temp file to WEBP image
	img, err := ToWEBP(fileByte)
	if err != nil {
		c.App.Log.Error("error converting to webp image", "err", err.Error())
		return "", err
	}

	name := fmt.Sprintf("%s.webp", c.Archive.FileHash)
	fp := filepath.Join(c.Dir, name)
	full, err := filepath.Abs(fp)
	if err != nil {
		return "", err
	}

	// Save new image
	err = bimg.Write(full, img)
	if err != nil {
		return "", err
	}

	return fp, nil
}
