package archive

import (
	"fmt"
	"io"

	"github.com/nwaples/rardecode/v2"
)

// rarArchive implements the Archive interface
type rarArchive struct {
	path  string
	pages []Page
}

// openRar() opens a RAR file and returns rarArchive.
// Since RAR requires streaming, also build page list.
func openRar(path string) (Archive, error) {
	r, err := rardecode.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open rar: %w", err)
	}
	defer r.Close()

	var pages []Page
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.IsDir || !isImageFile(header.Name) {
			continue
		}
		pages = append(pages, Page{
			Filename: header.Name,
			Size:     header.UnPackedSize,
		})
	}

	sortPages(pages)
	return &rarArchive{path: path, pages: pages}, nil
}

// Pages() returns a sorted list of all pages in the RAR
func (r *rarArchive) Pages() ([]Page, error) {
	return r.pages, nil
}

// Extract() re-opens the archive and streams to the requested page.
// RAR does not support random access so we must scan from the beginning.
func (r *rarArchive) Extract(page Page) (io.ReadCloser, error) {
	reader, err := rardecode.OpenReader(r.path)
	if err != nil {
		return nil, err
	}

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			reader.Close()
			return nil, err
		}
		if header.Name == page.Filename {
			// wrap reader so closing it also closes the rar reader
			return &rarPageReader{reader: reader, current: reader}, nil
		}
	}

	reader.Close()
	return nil, fmt.Errorf("page not found in archive: %s", page.Filename)
}

// Close() does nothing, only here to satisfy the Archive interface
func (r *rarArchive) Close() error { return nil }

// rarPageReader wraps the rar reader so the caller can close it normally.
type rarPageReader struct {
	reader  io.Closer
	current io.Reader
}

func (r *rarPageReader) Read(p []byte) (int, error) { return r.current.Read(p) }

// Close() closes the RAR reader
func (r *rarPageReader) Close() error { return r.reader.Close() }
