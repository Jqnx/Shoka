package fsutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	"github.com/gabriel-vasile/mimetype"
)

// ListArchives lists all files in the given path
func ListArchives(path string) {
	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range dir {
		fmt.Println(file.Name())
	}
}

// ArchiveContents lists the contents of a zip file
func ArchiveContents(path string) []string {
	var files []string

	zipFile, err := zip.OpenReader(path)
	if err != nil {
		log.Fatal(err)
	}
	defer zipFile.Close()

	for _, zip := range zipFile.File {
		files = append(files, zip.Name)
		// fmt.Println(files)
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
		ext := filepath.Ext(fn)
		fn = strings.TrimSuffix(fn, ext)
	}
	return fn
}

// GetPageCount returns the amount of image files in an archive
func GetPageCount(path string, extensions []string) int {
	var count int

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
