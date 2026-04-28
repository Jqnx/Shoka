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
	pdfMimes      = []string{"application/pdf", "application/x-pdf"}
)

type Page struct {
	Index    int
	Filename string
	Size     int64
}

type Archive interface {
	Pages() ([]Page, error)
	Extract(page Page) (io.ReadCloser, error)
	ReadFile(file string) ([]byte, error)
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
