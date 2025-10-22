package thumb

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"

	"github.com/h2non/bimg"
)

type Thumb struct {
	PageDir  string
	ThumbDir string
	App      *config.App
	Archive  *repository.GetArchiveByIDRow
}

type tempThumb struct {
	tempdir  string
	pagesdir string
	app      *config.App
	file     os.DirEntry
}

var (
	wg    sync.WaitGroup
	count int
)

// TODO: REDO

func NewThumb(a *repository.GetArchiveByIDRow, app *config.App, pagedir, thumbdir string) *Thumb {
	return &Thumb{
		PageDir:  pagedir,
		ThumbDir: thumbdir,
		App:      app,
		Archive:  a,
	}
}

func (t *Thumb) CreatePageDir(thumbdir string) {
	p := filepath.Join(thumbdir, "pages")
	if err := fsutil.CreateDir(p); err != nil {
		return
	}
	t.PageDir = p
}

func (t *Thumb) GetThumbDir() {
	fp := filepath.Join(t.App.Cfg.ThumbDir, t.Archive.Type, t.Archive.Hash[0:2], t.Archive.Hash[2:4])
	t.ThumbDir = fp
}

func (t *Thumb) Generate() error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", t.Archive.ID)
	if err != nil {
		return err
	}

	// Defer removal of temp folder and all contents
	defer os.RemoveAll(tempDir)

	// Unzip archive
	if err := fsutil.Unzip(t.Archive.FilePath, tempDir); err != nil {
		return err
	}

	// List extracted files
	list, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	// Convert every image to WEBP
	// Create goroutine for every image
	for _, i := range list {
		if fsutil.MatchExtension(i.Name(), config.ImageExtensions) {
			temp := tempThumb{
				tempdir:  tempDir,
				pagesdir: t.PageDir,
				app:      t.App,
				file:     i,
			}
			// Add to the waitgroup
			wg.Add(1)
			// Run conversion of thumbnails in goroutine
			go func() {
				defer wg.Done()
				convertThumb(temp)
			}()
		}
	}

	// Wait for goroutines to finish
	wg.Wait()

	// t.App.Log.Info("extract thumbs")

	return nil
}

func convertThumb(a tempThumb) {
	src := filepath.Join(a.tempdir, a.file.Name())

	img, err := ToWEBP(src)
	if err != nil {
		return
	}

	strippedFn := fsutil.StripExtension(a.file.Name())
	fn := fmt.Sprintf("%s.webp", strippedFn)
	fp := filepath.Join(a.pagesdir, fn)

	err = bimg.Write(fp, img)
	if err != nil {
		return
	}
}
