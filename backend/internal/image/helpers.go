package image

import (
	"Shoka/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ThumbDir returns the directory that holds an archive's thumbnails.
func (p *Processor) ThumbDir(archiveID string) string {
	return filepath.Join(p.cacheDir, archiveID, "thumbs")
}

func (p *Processor) thumbPath(archiveID string, index int) string {
	return filepath.Join(p.ThumbDir(archiveID), fmt.Sprintf("%03d.webp", index+1))
}

// ThumbPath returns the path to a page's thumbnail, or "" when it has not
// been generated yet.
func (p *Processor) ThumbPath(archiveID string, index int) string {
	path := p.thumbPath(archiveID, index)
	if _, err := os.Stat(path); err != nil {
		return ""
	}

	return path
}

// thumbReadyMarker returns the path of the marker the thumbnail job leaves once
// every page it could thumbnail has one on disk.
func (p *Processor) thumbReadyMarker(archiveID string) string {
	return filepath.Join(p.ThumbDir(archiveID), "ready")
}

// MarkThumbsReady records that a pageCount-page archive has been thumbnailed
// through its last page. Called by the thumbnail job once its loop completes, so
// the directory is usually there already - ensured rather than assumed, because
// a zero-page archive never generates a page to create it.
func (p *Processor) MarkThumbsReady(archiveID string, pageCount int) error {
	if err := util.EnsureDir(p.ThumbDir(archiveID)); err != nil {
		return err
	}

	return writeFileAtomic(p.thumbReadyMarker(archiveID), []byte(strconv.Itoa(pageCount)))
}

// ThumbsReady reports whether every page of a pageCount-page archive has a
// generated thumbnail on disk. It reads the job-written marker instead of
// stat-ing every page: this runs once per row of every archive list, so its cost
// has to be one read per archive, not one stat per page.
//
// The marker holds the number of pages the job thumbnailed, which counts the
// archive file and not necessarily archive.page_count - a metadata source can
// write its own count over the stored one (see metadata.ApplyMetadata) and a
// rescan never revises it. Asking whether the marker covers *at least* this
// archive's page count is what keeps both sides honest: a stored count below the
// thumbnailed one is ready, because every page it names has a thumbnail, and one
// above it stays unready until the pages it names do.
func (p *Processor) ThumbsReady(archiveID string, pageCount int) bool {
	// Nothing to thumbnail; no job will ever run to leave a marker.
	if pageCount <= 0 {
		return true
	}

	marked, err := os.ReadFile(p.thumbReadyMarker(archiveID))
	if err != nil {
		return false
	}

	thumbnailed, err := strconv.Atoi(strings.TrimSpace(string(marked)))

	return err == nil && thumbnailed >= pageCount
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
	defer func() { _ = os.Remove(tmpPath) }() // no-op once successfully renamed

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
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
