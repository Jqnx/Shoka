package image

import (
	"fmt"
	"os"
	"path/filepath"
)

func (p *Processor) ThumbDir(archiveID string) string {
	return filepath.Join(p.cacheDir, archiveID, "thumbs")
}

func (p *Processor) ThumbPath(archiveID string, index int) string {
	path := filepath.Join(p.ThumbDir(archiveID), fmt.Sprintf("%03d.webp", index+1))
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

func (p *Processor) ThumbsReady(archiveID string, pageCount int) bool {
	for i := range pageCount {
		if p.ThumbPath(archiveID, i) == "" {
			return false
		}
	}
	return true
}
