// Package fsutil implements utility pertaining to interacting with the filesystem
package fsutil

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"Shoka/internal/config"
	"Shoka/internal/repository"

	"github.com/bodgit/sevenzip"
)

// ListArchives lists all files in the given path
func ListArchives(path string) []string {
	list := []string{}
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			f := filepath.Join(path, d.Name())
			list = append(list, f)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return list
}

// MatchExtension returns true if the extension of the provided path
// matches any of the provided extensions.
func MatchExtension(path string, extensions []string) bool {
	ext := filepath.Ext(path)
	for _, e := range extensions {
		if strings.EqualFold(ext, (".")+e) {
			return true
		}
	}
	return false
}

// GetNameFromPath returns the name of a file from its path
// if stripExtension is true the extension is omitted from the name
func GetNameFromPath(path string, stripExtension bool) string {
	fn := filepath.Base(path)
	if stripExtension {
		fn = StripExtension(fn)
	}
	return fn
}

// StripExtension returns the name of a file without the extension
func StripExtension(file string) string {
	ext := filepath.Ext(file)
	fn := strings.TrimSuffix(file, ext)
	return fn
}

// GetFirstPage gets the first image from a zip/7zip archive
// and returns it as a byte
func GetFirstPage(path string) ([]byte, error) {
	arch, err := OpenArchive(path)
	if err != nil {
		return nil, err
	}
	list, err := arch.GetFileNames(true)
	if err != nil {
		return nil, err
	}

	file, err := arch.ReadFile(list[0])
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GenHash generates sha256 hash for a file
func GenHash(fileName string) string {
	var res []byte

	file, err := os.Open(fileName)
	if err != nil {
		return ""
	}

	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(res))
}

// FileExists checks is a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// CoverExists checks if a cover exists
func CoverExists(path string, hash string) bool {
	fp := filepath.Join(path, fmt.Sprintf("%s.webp", hash))
	return FileExists(fp)
}

func ExtractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	// os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.FileInfo().Name())

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		if !f.FileInfo().IsDir() {

			err := extractAndWriteFile(f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Extract7Zip(src, dest string) error {
	r, err := sevenzip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	// os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *sevenzip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.FileInfo().Name())

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		if !f.FileInfo().IsDir() {

			err := extractAndWriteFile(f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Unzip(src, dest string) error {
	if Is7z(src) {
		if err := Extract7Zip(src, dest); err != nil {
			return err
		}
	} else {
		if err := ExtractZip(src, dest); err != nil {
			return err
		}
	}
	return nil
}

// Zip creates a new zip archive at the destination containing all content from a given source directory
func Zip(src, dst string) error {
	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		relPath = strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		fileHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		fileHeader.Name = relPath
		fileHeader.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(fileHeader)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		fmt.Printf("Added: %s\n", relPath)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateDir takes in a relative path
// converts it to an absolute path
// and creates all necessary folders
// returns absolute path and an error
func CreateDir(dir string) error {
	// Convert relative path to absolute path
	fp, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	// Use absolute path to create all necessary folders
	if err := os.MkdirAll(fp, 0o777); err != nil {
		return err
	}
	return nil
}

func Remove(src string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.RemoveAll(src); err != nil {
			return err
		}
	} else {
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("error removing file: %v", err)
		}
	}
	return nil
}

// CountPages counts the amount of image files in the pages directory of an archive
func CountPages(archive repository.GetArchiveByIDRow) int {
	p, _ := filepath.Abs(*archive.ThumbPath)
	d, err := os.ReadDir(filepath.Join(p, "pages"))
	if err != nil {
		return 0
	}

	var pages int
	for _, i := range d {
		if MatchExtension(i.Name(), config.ImageExtensions) {
			pages++
		}
	}
	return pages
}
