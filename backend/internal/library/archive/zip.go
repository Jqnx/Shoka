package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// zipArchive implements the Archive interface
type zipArchive struct {
	reader *zip.ReadCloser
	path   string
}

// openZip() opens a zip file and returns zipArchive
func openZip(path string) (Archive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	return &zipArchive{reader: r, path: path}, nil
}

// Pages() returns a sorted list of all pages in the zip
func (z *zipArchive) Pages() ([]Page, error) {
	var pages []Page
	for _, f := range z.reader.File {
		if f.FileInfo().IsDir() || !isImageFile(f.Name) {
			continue
		}
		pages = append(pages, Page{
			Filename: f.Name,
			Size:     int64(f.UncompressedSize64),
		})
	}
	sortPages(pages)
	return pages, nil
}

// Extract() returns a reader for the page
func (z *zipArchive) Extract(page Page) (io.ReadCloser, error) {
	for _, f := range z.reader.File {
		if f.Name == page.Filename {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("page not found in archive: %s", page.Filename)
}

// ReadFile() returns the contents of a file
func (z *zipArchive) ReadFile(file string) ([]byte, error) {
	lowerFile := strings.ToLower(file)
	for _, f := range z.reader.File {
		name := strings.ToLower(filepath.Base(f.Name))
		if name == lowerFile {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}

	return nil, nil
}

// Close() closes the zip file reader
func (z *zipArchive) Close() error {
	return z.reader.Close()
}
