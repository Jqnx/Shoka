package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// Path() returns the path to the RAR
func (r *rarArchive) Path() string { return r.path }

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

func (r *rarArchive) ReadFile(file string) ([]byte, error) {
	lowerFile := strings.ToLower(file)

	reader, err := rardecode.OpenReader(r.path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if strings.ToLower(filepath.Base(header.Name)) == lowerFile {
			return io.ReadAll(reader)
		}
	}

	return nil, nil
}

func (r *rarArchive) WriteFile(name string, data []byte) error {
	zipPath := strings.TrimSuffix(r.path, filepath.Ext(r.path)) + ".cbz"

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	reader, err := rardecode.OpenReader(r.path)
	if err != nil {
		return fmt.Errorf("open rar: %w", err)
	}
	defer reader.Close()

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read rar entry: %w", err)
		}
		if header.IsDir {
			continue
		}
		// skip any existing metadata file
		if strings.EqualFold(filepath.Base(header.Name), name) {
			continue
		}

		fw, err := w.Create(header.Name)
		if err != nil {
			return fmt.Errorf("create zip entry %s: %w", header.Name, err)
		}
		if _, err := io.Copy(fw, reader); err != nil {
			return fmt.Errorf("copy entry %s: %w", header.Name, err)
		}
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

	if err := os.Remove(r.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove original archive: %w", err)
	}

	r.path = zipPath
	return nil
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
