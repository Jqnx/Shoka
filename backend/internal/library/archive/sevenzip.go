package archive

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
)

// sevenZipArchive implements the Archive interface
type sevenZipArchive struct {
	reader *sevenzip.ReadCloser
	path   string
}

// openSevenZip() opens a 7zip file and returns zipArchive
func openSevenZip(path string) (Archive, error) {
	r, err := sevenzip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open sevenZip: %w", err)
	}
	return &sevenZipArchive{reader: r, path: path}, nil
}

// Pages() returns a sorted list of all pages in the 7zip file
func (z *sevenZipArchive) Pages() ([]Page, error) {
	var pages []Page
	for _, f := range z.reader.File {
		if f.FileInfo().IsDir() || !isImageFile(f.Name) {
			continue
		}
		pages = append(pages, Page{
			Filename: f.Name,
			Size:     int64(f.UncompressedSize),
		})
	}
	sortPages(pages)
	return pages, nil
}

// Extract() returns a reader for the page
func (z *sevenZipArchive) Extract(page Page) (io.ReadCloser, error) {
	for _, f := range z.reader.File {
		if f.Name == page.Filename {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("page not found in archive: %s", page.Filename)
}

// ReadFile() returns the contents of a file
func (z *sevenZipArchive) ReadFile(file string) ([]byte, error) {
	for _, f := range z.reader.File {
		name := strings.ToLower(filepath.Base(f.Name))
		if name == file {
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

// Close() closes the 7zip file reader
func (z *sevenZipArchive) Close() error {
	return z.reader.Close()
}
