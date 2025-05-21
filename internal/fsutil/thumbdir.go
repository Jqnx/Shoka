package fsutil

import (
	"path/filepath"
)

type ThumbDir struct {
	ThumbsDir string
	MediaType string
	Hash      string
}

// CreateThumbDir creates the thumbnail directory for a specific archive
func (d *ThumbDir) CreateThumbDir() (string, error) {
	// Get relative path
	thumbdir, err := d.GetThumbDir()
	if err != nil {
		return "", err
	}
	if err := CreateDir(thumbdir); err != nil {
		return "", err
	}
	return thumbdir, nil
}

// GetThumbDir gets the thumbnail directory for a specific archive
func (d *ThumbDir) GetThumbDir() (string, error) {
	fp := filepath.Join(d.ThumbsDir, d.MediaType, d.Hash[0:2], d.Hash[2:4])
	return fp, nil
}

func NewThumbDir(cfgThumbDir, mediatype, hash string) *ThumbDir {
	return &ThumbDir{
		ThumbsDir: cfgThumbDir,
		MediaType: mediatype,
		Hash:      hash,
	}
}
