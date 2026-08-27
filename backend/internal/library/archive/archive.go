package archive

import (
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
)

var (
	zipMimes      = []string{"application/zip", "application/x-zip", "application/x-zip-compressed"}
	sevenZipMimes = []string{"application/x-7z-compressed"}
	rarMimes      = []string{"application/x-rar", "application/x-rar-compressed"}
	//nolint:unused // referenced by the commented-out PDF case in Open; kept for when PDF support returns
	pdfMimes = []string{"application/pdf", "application/x-pdf"}
)

// maxArchiveEntrySize bounds how large a single entry may decompress to when
// rewriting an archive, so a crafted "zip bomb" entry can't exhaust memory or
// disk. 512 MiB is far beyond any real scanned page.
const maxArchiveEntrySize = 512 << 20

// copyArchiveEntry copies a single archive entry into dst, rejecting entries
// that decompress to more than maxArchiveEntrySize.
func copyArchiveEntry(dst io.Writer, src io.Reader, name string) error {
	n, err := io.Copy(dst, io.LimitReader(src, maxArchiveEntrySize+1))
	if err != nil {
		return fmt.Errorf("copy entry %s: %w", name, err)
	}

	if n > maxArchiveEntrySize {
		return fmt.Errorf("entry %s exceeds the %d-byte decompression limit", name, maxArchiveEntrySize)
	}

	return nil
}

type Page struct {
	Index    int
	Filename string
	Size     int64
}

type Archive interface {
	Path() string
	Pages() ([]Page, error)
	Extract(page Page) (io.ReadCloser, error)
	ReadFile(file string) ([]byte, error)
	WriteFile(name string, data []byte) error
	Close() error
}

// Open() opens an archive.
func Open(path string) (Archive, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return nil, err
	}

	switch {
	case mimetype.EqualsAny(mtype.String(), zipMimes...):
		return openZip(path)
	case mimetype.EqualsAny(mtype.String(), rarMimes...):
		return openRar(path)
	case mimetype.EqualsAny(mtype.String(), sevenZipMimes...):
		return openSevenZip(path)
	// case mimetype.EqualsAny(mtype.String(), pdfMimes...):
	//	return openPDF(path)
	default:
		return nil, fmt.Errorf("unsupported format: %s", mtype.Extension())
	}
}
