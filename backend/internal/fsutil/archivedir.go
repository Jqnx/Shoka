package fsutil

import (
	"path/filepath"
)

type ArchiveDir struct {
	ThumbsDir string
	MediaType string
	Hash      string
}

func NewArchiveDir(cfgThumbDir, mediatype, hash string) *ArchiveDir {
	return &ArchiveDir{
		ThumbsDir: cfgThumbDir,
		MediaType: mediatype,
		Hash:      hash,
	}
}

// GenThumbDir generates the thumbnail directory for a specific archive
func (d *ArchiveDir) GenThumbDir() (string, error) {
	fp := filepath.Join(d.ThumbsDir, d.MediaType, d.Hash[0:2], d.Hash[2:4])
	return fp, nil
}

// CreateDirs creates the thumbnail directory containing a cover and
// a pages directory for a specific archive, returns only the top thumbnail directory
func (d *ArchiveDir) CreateDirs() (string, error) {
	// Get relative path
	thumbdir, err := d.GenThumbDir()
	if err != nil {
		return "", err
	}
	if err := CreateDir(thumbdir); err != nil {
		return "", err
	}
	if err := d.CreateCoverDir(thumbdir); err != nil {
		return "", err
	}
	if err := d.CreatePagesDir(thumbdir); err != nil {
		return "", err
	}

	return thumbdir, nil
}

// CreateCoverDir creates a new cover directory on the filesystem
// for a single archive.
func (d *ArchiveDir) CreateCoverDir(thumbdir string) error {
	if err := CreateDir(filepath.Join(thumbdir, "cover")); err != nil {
		return err
	}
	return nil
}

// CreatePagesDir creates a new pages directory on the filesystem
// for a single archive.
func (d *ArchiveDir) CreatePagesDir(thumbdir string) error {
	if err := CreateDir(filepath.Join(thumbdir, "pages")); err != nil {
		return err
	}
	return nil
}
