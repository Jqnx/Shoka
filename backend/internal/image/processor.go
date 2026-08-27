package image

import (
	"Shoka/internal/util"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/corona10/goimagehash"
	"github.com/davidbyttow/govips/v2/vips"
)

const (
	thumbWidth   = 360
	maxPageWidth = 2400
	thumbQuality = 75
	pageQuality  = 85
)

type Processor struct {
	cacheDir string
	log      *slog.Logger
}

func NewProcessor(cacheDir string, log *slog.Logger) *Processor {
	return &Processor{
		cacheDir: cacheDir,
		log:      log.With("component", "image_processor"),
	}
}

// processImage() resizes an image and exports it as a webp.
func (p *Processor) processImage(r io.Reader, destPath string, maxWidth, quality int) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read image data: %w", err)
	}

	img, err := vips.NewImageFromBuffer(data)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}
	defer img.Close()

	if img.Width() > maxWidth {
		scale := float64(maxWidth) / float64(img.Width())
		if err := img.Resize(scale, vips.KernelLanczos3); err != nil {
			return fmt.Errorf("resize: %w", err)
		}
	}

	img.RemoveMetadata()

	ep := vips.NewWebpExportParams()
	ep.Quality = quality
	ep.Lossless = false
	ep.StripMetadata = true

	bytes, _, err := img.ExportWebp(ep)
	if err != nil {
		return fmt.Errorf("export webp: %w", err)
	}

	return writeFileAtomic(destPath, bytes)
}

// ProcessToBytes processes an image from a reader and returns WebP bytes
// without writing to disk. Used for serving full resolution pages into the LRU.
func (p *Processor) ProcessToBytes(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read image data: %w", err)
	}

	img, err := vips.NewImageFromBuffer(data)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	defer img.Close()

	if img.Width() > maxPageWidth {
		scale := float64(maxPageWidth) / float64(img.Width())
		if err := img.Resize(scale, vips.KernelLanczos3); err != nil {
			return nil, fmt.Errorf("resize: %w", err)
		}
	}

	img.RemoveMetadata()

	ep := vips.NewWebpExportParams()
	ep.Quality = pageQuality
	ep.Lossless = false
	ep.StripMetadata = true

	bytes, _, err := img.ExportWebp(ep)
	if err != nil {
		return nil, fmt.Errorf("export webp: %w", err)
	}

	return bytes, nil
}

// GenerateThumbnail() generates a thumbnail for a given archive page.
func (p *Processor) GenerateThumbnail(ctx context.Context, archiveID string, index int, r io.Reader) error {
	if err := util.EnsureDir(p.ThumbDir(archiveID)); err != nil {
		return err
	}

	dest := p.thumbPath(archiveID, index)
	if _, err := os.Stat(dest); err == nil {
		return nil
	}

	if err := p.processImage(r, dest, thumbWidth, thumbQuality); err != nil {
		return fmt.Errorf("generate thumbnail archive=%s page=%d: %w", archiveID, index, err)
	}

	p.log.Debug("thumbnail generated", "archive_id", archiveID, "index", index)

	return nil
}

// ComputePHash computes a 64-bit perceptual hash (DCT-based, see
// goimagehash.PerceptionHash) of an image - used for duplicate detection
// across archives whose cover looks the same even when the file doesn't
// match byte-for-byte (recompressed, re-cropped, different format, ...).
// Unlike processImage/ProcessToBytes this never touches disk or re-encodes:
// ToGoImage() converts the decoded vips image directly to a Go image.Image.
func (p *Processor) ComputePHash(data []byte) (uint64, error) {
	img, err := vips.NewImageFromBuffer(data)
	if err != nil {
		return 0, fmt.Errorf("decode image: %w", err)
	}
	defer img.Close()

	goImg, err := img.ToGoImage()
	if err != nil {
		return 0, fmt.Errorf("convert to go image: %w", err)
	}

	hash, err := goimagehash.PerceptionHash(goImg)
	if err != nil {
		return 0, fmt.Errorf("compute perceptual hash: %w", err)
	}

	return hash.GetHash(), nil
}

// EvictAll removes the entire cache directory for an archive.
// Called when an archive is deleted from the library.
func (p *Processor) EvictAll(archiveID string) error {
	dir := filepath.Join(p.cacheDir, archiveID)
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("evict archive=%s: %w", archiveID, err)
	}

	p.log.Info("cache evicted", "archive_id", archiveID)

	return nil
}
