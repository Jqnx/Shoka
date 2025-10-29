package thumb

import (
	"fmt"
	"path/filepath"
	"sync"

	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"

	"github.com/h2non/bimg"
)

type Thumb struct {
	App     *config.App
	Archive *archive.Archive
}

var (
	wg    sync.WaitGroup
	count int
)

// TODO: REDO

func NewThumb(a *archive.Archive, app *config.App) *Thumb {
	return &Thumb{
		App:     app,
		Archive: a,
	}
}

func (t *Thumb) Generate() error {
	zip, err := fsutil.OpenArchive(t.Archive.FilePath)
	if err != nil {
		return err
	}
	list, err := zip.GetFileNames(true)
	if err != nil {
		return err
	}

	// Convert every image to WEBP
	// Create goroutine for every image
	for _, i := range list {
		file, err := zip.ReadFile(i)
		if err != nil {
			return err
		}
		// Add to the waitgroup
		wg.Add(1)
		// Run conversion of thumbnails in goroutine
		go func() {
			defer wg.Done()
			t.convertThumb(filepath.Base(i), file)
		}()
	}

	// Wait for goroutines to finish
	wg.Wait()

	// t.App.Log.Info("extract thumbs")

	return nil
}

func (t *Thumb) convertThumb(dest string, content []byte) {
	img, err := ToWEBP(content)
	if err != nil {
		t.App.Log.Error("error converting thumbnail", "err", err.Error())
		return
	}

	strippedFn := fsutil.StripExtension(dest)
	fn := fmt.Sprintf("%s.webp", strippedFn)
	fp := filepath.Join(filepath.Join(*t.Archive.ThumbsPath, "pages"), fn)

	err = bimg.Write(fp, img)
	if err != nil {
		t.App.Log.Error("error writing thumbnail", "err", err.Error())
		return
	}
}
