package archive

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nwaples/rardecode/v2"
)

// finishCBZ writes the metadata entry into the open zip writer, flushes the
// buffer to zipPath through a temporary file, and finally removes the
// original archive at srcPath. It is the shared tail of every WriteFile
// implementation: the container formats differ only in how their existing
// entries are copied into w before this point.
func finishCBZ(w *zip.Writer, buf *bytes.Buffer, name string, data []byte, zipPath, srcPath string) error {
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
	//nolint:gosec // library archive file, deliberately group/other-readable so the frontend process can serve it
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmp, zipPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write cbz: %w", err)
	}

	if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove original archive: %w", err)
	}

	return nil
}

// copyRarEntries streams every non-directory RAR entry except skipName into
// the zip writer. RAR has no random access, so the reader is consumed from
// the start of the archive.
func copyRarEntries(w *zip.Writer, reader *rardecode.ReadCloser, skipName string) error {
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("read rar entry: %w", err)
		}

		if header.IsDir {
			continue
		}
		// skip any existing metadata file
		if strings.EqualFold(filepath.Base(header.Name), skipName) {
			continue
		}

		fw, err := w.Create(header.Name)
		if err != nil {
			return fmt.Errorf("create zip entry %s: %w", header.Name, err)
		}

		if err := copyArchiveEntry(fw, reader, header.Name); err != nil {
			return err
		}
	}
}
