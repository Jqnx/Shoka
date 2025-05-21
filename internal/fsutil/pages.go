package fsutil

import "path/filepath"

func (d *ThumbDir) CreatePageDir(thumbdir string) (string, error) {
	p := filepath.Join(thumbdir, "pages")
	if err := CreateDir(p); err != nil {
		return "", err
	}
	return p, nil
}
