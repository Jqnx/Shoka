package fsutil

import (
	"Shoka/internal/config"
	"archive/zip"
	"bytes"
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

	"github.com/bodgit/sevenzip"
	"github.com/gabriel-vasile/mimetype"
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

// ArchiveContents lists the contents of a zip file
func ArchiveContents(path string) []string {
	var files []string

	if Is7z(path) {
		zipFile, err := sevenzip.OpenReader(path)
		if err != nil {
			log.Fatal(err)
		}
		defer zipFile.Close()

		for _, zip := range zipFile.File {
			if MatchExtension(zip.FileInfo().Name(), config.ImageExtensions) {
				files = append(files, zip.FileInfo().Name())
				// fmt.Println(files)
			}
		}
	} else {

		zipFile, err := zip.OpenReader(path)
		if err != nil {
			log.Fatal(err)
		}
		defer zipFile.Close()

		for _, zip := range zipFile.File {
			if MatchExtension(zip.FileInfo().Name(), config.ImageExtensions) {
				files = append(files, zip.FileInfo().Name())
				// fmt.Println(files)
			}
		}
	}
	return files
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

// GetPageCount returns the amount of image files in an archive
func GetPageCount(path string, extensions []string) int64 {
	var count int64

	if Is7z(path) {
		archive, err := sevenzip.OpenReader(path)
		if err != nil {
			log.Println(err)
		}

		defer archive.Close()

		for _, file := range archive.File {
			if MatchExtension(file.Name, extensions) {
				count++
			}
		}

	} else {
		archive, err := zip.OpenReader(path)
		if err != nil {
			log.Println(err)
		}

		defer archive.Close()

		for _, file := range archive.File {
			if MatchExtension(file.Name, extensions) {
				count++
			}
		}

	}
	return count
}

// Is7z checks if archive is compressed with 7z or not
func Is7z(path string) bool {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		log.Println(err)
	}

	if mtype.Is("application/x-7z-compressed") {
		return true
	}
	return false
}

func Read7z(file sevenzip.File) (string, error) {
	openFile, err := file.Open()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(openFile)
	if err != nil {
		return "", err
	}
	content := buf.String()

	return content, nil
}

func ReadZip(file zip.File) (string, error) {
	openFile, err := file.Open()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(openFile)
	if err != nil {
		return "", err
	}
	content := buf.String()

	return content, nil
}

func ExtractFirstPage(path string, tempFile *os.File) string {
	if Is7z(path) {
		file, err := sevenzip.OpenReader(path)
		if err != nil {
			return ""
		}

		defer file.Close()

		for i, f := range file.File {
			if !f.FileInfo().IsDir() && i < 2 && strings.Contains(f.FileInfo().Name(), "1") {
				// fmt.Printf("i: %v, name: %v \n", i, f.Name)
				open, err := f.Open()
				if err != nil {
					return ""
				}
				_, err = io.Copy(tempFile, open)
				if err != nil {
					return ""
				}
				open.Close()
			}
		}

	} else {
		file, err := zip.OpenReader(path)
		if err != nil {
			return ""
		}

		defer file.Close()

		for i, f := range file.File {
			if !f.FileInfo().IsDir() && i < 2 && strings.Contains(f.FileInfo().Name(), "1") {
				// fmt.Printf("i: %v, name: %v \n", i, f.Name)
				open, err := f.Open()
				if err != nil {
					return ""
				}
				_, err = io.Copy(tempFile, open)
				if err != nil {
					return ""
				}
				open.Close()
			}
		}
	}
	return ""
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
	if err := os.MkdirAll(fp, 0777); err != nil {
		return err
	}
	return nil
}
