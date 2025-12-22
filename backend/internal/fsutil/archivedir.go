package fsutil

import (
	"path/filepath"
)

type ArchiveDir struct {
	GlobalThumbDir string
	FileHash       string
}

func NewArchiveDir(globalThumbDir, filehash string) *ArchiveDir {
	return &ArchiveDir{
		GlobalThumbDir: globalThumbDir,
		FileHash:       filehash,
	}
}

// GenThumbDir generates the thumbnail directory for a specific archive
func (d *ArchiveDir) GenThumbDir() (string, error) {
	fp := filepath.Join(d.GlobalThumbDir, d.FileHash[0:2], d.FileHash[2:4], d.FileHash[4:6])
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
