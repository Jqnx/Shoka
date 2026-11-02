package image

import (
	"Shoka/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/webp"
)

// ThumbDir returns the directory that holds an archive's thumbnails.
func (p *Processor) ThumbDir(archiveID string) string {
	return filepath.Join(p.cacheDir, archiveID, "thumbs")
}

// AudioCacheKey is the cache directory key for an audio work. Thumbnail
// helpers are keyed by string, so audio works reuse them under their own
// namespace instead of colliding with an archive id.
func AudioCacheKey(audioID string) string {
	return "audio-" + audioID
}

// AudioCoverPath returns the path to an audio work's cover thumbnail, or ""
// when it has not been generated yet.
func (p *Processor) AudioCoverPath(audioID string) string {
	return p.ThumbPath(AudioCacheKey(audioID), 0)
}

// AudioCoverInfo is an audio work's cached cover. Path is empty when no cover
// has been rendered; Width/Height are zero when the file cannot be decoded.
type AudioCoverInfo struct {
	Path    string
	ModTime *time.Time
	Width   int
	Height  int
}

// CachedAudioCover returns the cached cover's path, mtime and pixel size. The
// size comes from the WebP header of the file being served, so list handlers
// can reserve the cover's space without a separate cache format.
func (p *Processor) CachedAudioCover(audioID string) AudioCoverInfo {
	path := p.thumbPath(AudioCacheKey(audioID), 0)

	info, err := os.Stat(path)
	if err != nil {
		return AudioCoverInfo{}
	}

	mod := info.ModTime().UTC()
	cover := AudioCoverInfo{Path: path, ModTime: &mod}

	//nolint:gosec // G304: path is built from an audio ID under the cache dir, not user input
	f, err := os.Open(path)
	if err != nil {
		return cover
	}
	defer func() { _ = f.Close() }()

	if cfg, err := webp.DecodeConfig(f); err == nil {
		cover.Width, cover.Height = cfg.Width, cfg.Height
	}

	return cover
}

// audioCoverMarker identifies the render settings an audio cover was produced
// with. Keep it derived from the constants so changing either invalidates
// every cached cover.
func audioCoverMarker() string {
	return fmt.Sprintf("%dx%d", audioCoverWidth, audioCoverQuality)
}

func audioCoverMarkerPath(cacheDir, audioID string) string {
	return filepath.Join(cacheDir, AudioCacheKey(audioID), "thumbs", "cover-version")
}

// AudioCoverStale reports whether an audio work's cached cover is missing or
// was rendered with settings other than the current constants. Scans use it to
// re-enqueue covers for works whose files did not otherwise change.
func AudioCoverStale(cacheDir, audioID string) bool {
	if _, err := os.Stat(thumbPathIn(cacheDir, AudioCacheKey(audioID), 0)); err != nil {
		return true
	}

	data, err := os.ReadFile(audioCoverMarkerPath(cacheDir, audioID))

	return err != nil || strings.TrimSpace(string(data)) != audioCoverMarker()
}

func thumbPathIn(cacheDir, archiveID string, index int) string {
	return filepath.Join(cacheDir, archiveID, "thumbs", fmt.Sprintf("%03d.webp", index+1))
}

func (p *Processor) thumbPath(archiveID string, index int) string {
	return thumbPathIn(p.cacheDir, archiveID, index)
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
