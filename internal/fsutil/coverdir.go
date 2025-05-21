package fsutil

import (
	"path/filepath"
)

func (d *ThumbDir) CreateCoverDir(thumbdir string) (string, error) {
	c := filepath.Join(thumbdir, "cover")
	if err := CreateDir(c); err != nil {
		return "", err
	}
	return c, nil
}
