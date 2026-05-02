package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
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

// Path() returns the path to the 7zip file
func (z *sevenZipArchive) Path() string { return z.path }

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

// WriteFile() adds a new file to the archive
func (s *sevenZipArchive) WriteFile(name string, data []byte) error {
	zipPath := strings.TrimSuffix(s.path, filepath.Ext(s.path)) + ".cbz"

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("open 7z: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat 7z: %w", err)
	}

	reader, err := sevenzip.NewReader(f, info.Size())
	if err != nil {
		return fmt.Errorf("open 7z reader: %w", err)
	}

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		// skip any existing metadata file
		if strings.EqualFold(filepath.Base(entry.Name), name) {
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("open entry %s: %w", entry.Name, err)
		}

		fw, err := w.Create(entry.Name)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create zip entry %s: %w", entry.Name, err)
		}

		if _, err := io.Copy(fw, rc); err != nil {
			rc.Close()
			return fmt.Errorf("copy entry %s: %w", entry.Name, err)
		}
		rc.Close()
	}

	fw, err := w.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	if _, err := fw.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close zip writer: %w", err)
	}

	tmp := zipPath + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmp, zipPath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("write cbz: %w", err)
	}

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove original archive: %w", err)
	}

	s.path = zipPath
	return nil
}

// Close() closes the 7zip file reader
func (z *sevenZipArchive) Close() error {
	return z.reader.Close()
}
