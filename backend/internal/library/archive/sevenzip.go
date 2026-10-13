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

// sevenZipArchive implements the Archive interface.
type sevenZipArchive struct {
	reader *sevenzip.ReadCloser
	path   string
}

// openSevenZip() opens a 7zip file and returns zipArchive.
func openSevenZip(path string) (Archive, error) {
	r, err := sevenzip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open sevenZip: %w", err)
	}

	return &sevenZipArchive{reader: r, path: path}, nil
}

// Path() returns the path to the 7zip file.
func (z *sevenZipArchive) Path() string { return z.path }

// Pages() returns a sorted list of all pages in the 7zip file.
func (z *sevenZipArchive) Pages() ([]Page, error) {
	var pages []Page

	for _, f := range z.reader.File {
		if f.FileInfo().IsDir() || !isImageFile(f.Name) {
			continue
		}

		pages = append(pages, Page{
			Filename: f.Name,
			//nolint:gosec // archive-reported size, stored as page metadata only; never used for allocation
			Size: int64(f.UncompressedSize),
		})
	}

	sortPages(pages)

	return pages, nil
}

// Extract() returns a reader for the page.
func (z *sevenZipArchive) Extract(page Page) (io.ReadCloser, error) {
	for _, f := range z.reader.File {
		if f.Name == page.Filename {
			return f.Open()
		}
	}

	return nil, fmt.Errorf("page not found in archive: %s", page.Filename)
}

// ReadFile() returns the contents of a file.
func (z *sevenZipArchive) ReadFile(file string) ([]byte, error) {
	for _, f := range z.reader.File {
		name := strings.ToLower(filepath.Base(f.Name))
		if name == file {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer func() { _ = rc.Close() }()

			return io.ReadAll(rc)
		}
	}

	return nil, nil
}

// WriteFile() adds a new file to the archive.
func (z *sevenZipArchive) WriteFile(name string, data []byte) error {
	zipPath := strings.TrimSuffix(z.path, filepath.Ext(z.path)) + ".cbz"

	var buf bytes.Buffer

	w := zip.NewWriter(&buf)

	f, err := os.Open(z.path)
	if err != nil {
		return fmt.Errorf("open 7z: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat 7z: %w", err)
	}

	reader, err := sevenzip.NewReader(f, info.Size())
	if err != nil {
		return fmt.Errorf("open 7z reader: %w", err)
	}

	if err := copySevenZipEntries(w, reader.File, name); err != nil {
		return err
	}

	if err := finishCBZ(w, &buf, name, data, zipPath, z.path); err != nil {
		return err
	}

	z.path = zipPath

	return nil
}

// copySevenZipEntries copies every non-directory entry except skipName from
// the source archive into the zip writer.
func copySevenZipEntries(w *zip.Writer, entries []*sevenzip.File, skipName string) error {
	for _, entry := range entries {
		if entry.FileInfo().IsDir() {
			continue
		}
		// skip any existing metadata file
		if strings.EqualFold(filepath.Base(entry.Name), skipName) {
			continue
		}

		if err := copySevenZipEntry(w, entry); err != nil {
			return err
		}
	}

	return nil
}

// copySevenZipEntry copies one entry from the source 7z into the new zip
// writer, bounding how much it may decompress.
func copySevenZipEntry(w *zip.Writer, entry *sevenzip.File) error {
	rc, err := entry.Open()
	if err != nil {
		return fmt.Errorf("open entry %s: %w", entry.Name, err)
	}
	defer func() { _ = rc.Close() }()

	fw, err := w.Create(entry.Name)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", entry.Name, err)
	}

	return copyArchiveEntry(fw, rc, entry.Name)
}

// Close() closes the 7zip file reader.
func (z *sevenZipArchive) Close() error {
	return z.reader.Close()
}
