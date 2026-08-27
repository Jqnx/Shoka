package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// zipArchive implements the Archive interface.
type zipArchive struct {
	reader *zip.ReadCloser
	path   string
}

// openZip() opens a zip file and returns zipArchive.
func openZip(path string) (Archive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	return &zipArchive{reader: r, path: path}, nil
}

// Path() returns the path to the zip.
func (z *zipArchive) Path() string { return z.path }

// Pages() returns a sorted list of all pages in the zip.
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

// Extract() returns a reader for the page.
func (z *zipArchive) Extract(page Page) (io.ReadCloser, error) {
	for _, f := range z.reader.File {
		if f.Name == page.Filename {
			return f.Open()
		}
	}

	return nil, fmt.Errorf("page not found in archive: %s", page.Filename)
}

// ReadFile() returns the contents of a file.
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

// WriteFile() adds a new file to the archive.
func (z *zipArchive) WriteFile(name string, data []byte) error {
	var buf bytes.Buffer

	w := zip.NewWriter(&buf)

	for _, f := range z.reader.File {
		if strings.EqualFold(filepath.Base(f.Name), name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open entry %s: %w", f.Name, err)
		}

		header, err := zip.FileInfoHeader(f.FileInfo())
		if err != nil {
			rc.Close()
			return fmt.Errorf("copy header %s: %w", f.Name, err)
		}

		header.Name = f.Name
		header.Method = zip.Deflate

		fw, err := w.CreateHeader(header)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create entry %s: %w", f.Name, err)
		}

		if _, err := io.Copy(fw, rc); err != nil {
			rc.Close()
			return fmt.Errorf("copy entry %s: %w", f.Name, err)
		}

		rc.Close()
	}

	// write the new file at the root of the archive
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

	// close the reader before overwriting the file
	if err := z.reader.Close(); err != nil {
		return fmt.Errorf("close zip reader: %w", err)
	}

	// write atomically — write to a temp file then rename
	// this ensures the original is never left in a corrupted state
	tmp := z.path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmp, z.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replace archive: %w", err)
	}

	// re-open the reader so the archive remains usable after writing
	r, err := zip.OpenReader(z.path)
	if err != nil {
		return fmt.Errorf("reopen archive: %w", err)
	}

	z.reader = r

	return nil
}

// Close() closes the zip file reader.
func (z *zipArchive) Close() error {
	return z.reader.Close()
}
