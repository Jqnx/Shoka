package fsutil

import (
	"fmt"
	"io"
	"log"

	"Shoka/internal/config"
	"Shoka/internal/util"

	"github.com/bodgit/sevenzip"
	"github.com/gabriel-vasile/mimetype"
)

type SevenZipArchive struct {
	reader *sevenzip.ReadCloser
}

// NewSevenZipArchive creates a new SevenZipArchive struct containing the
// content of a 7zip archive
func NewSevenZipArchive(path string) (*SevenZipArchive, error) {
	r, err := sevenzip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	return &SevenZipArchive{reader: r}, nil
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

// GetFileNames gets a list of files contained within a 7zip archive.
// Can be naturally sorted by passing true
func (z *SevenZipArchive) GetFileNames(onlyImages bool) ([]string, error) {
	var fileNames []string
	for _, f := range z.reader.File {
		if !f.FileInfo().IsDir() {
			if onlyImages {
				if MatchExtension(f.Name, config.ImageExtensions) {
					fileNames = append(fileNames, f.Name)
				}
			} else {
				fileNames = append(fileNames, f.Name)
			}
		}
	}

	if len(fileNames) == 0 {
		return nil, fmt.Errorf("empty archive")
	}

	util.NaturalSort(fileNames)

	return fileNames, nil
}

func (z *SevenZipArchive) ImagesToMap() (map[int]string, error) {
	list, err := z.GetFileNames(true)
	if err != nil {
		return nil, err
	}

	m := make(map[int]string)

	for i, file := range list {
		m[i+1] = file
	}
	return m, nil
}

// ReadFile reads a file from within a 7zip file into memory
func (z *SevenZipArchive) ReadFile(name string) ([]byte, error) {
	for _, f := range z.reader.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, config.ErrFileNotFound
}

// Close closes a 7zip file reader
func (z *SevenZipArchive) Close() error {
	return z.reader.Close()
}
