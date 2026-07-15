package image

import (
	"fmt"
	"os"
	"path/filepath"
)

func (p *Processor) ThumbDir(archiveID string) string {
	return filepath.Join(p.cacheDir, archiveID, "thumbs")
}

func (p *Processor) thumbPath(archiveID string, index int) string {
	return filepath.Join(p.ThumbDir(archiveID), fmt.Sprintf("%03d.webp", index+1))
}

func (p *Processor) ThumbPath(archiveID string, index int) string {
	path := p.thumbPath(archiveID, index)
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

// writeFileAtomic writes data to a uniquely-named temp file in the same
// directory as destPath, then renames it into place. Different jobs can
// legitimately race to produce the same output file (e.g. the cover job
// and an on-demand thumbnail job both wanting page 0) — writing directly
// with os.WriteFile would let concurrent writers interleave and corrupt
// the file; a rename is atomic, so the file at destPath is always either
// absent or fully written.
func writeFileAtomic(destPath string, data []byte) error {
	dir := filepath.Dir(destPath)

	tmp, err := os.CreateTemp(dir, ".tmp-*"+filepath.Ext(destPath))
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once successfully renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
