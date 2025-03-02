package metadata

import (
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"archive/zip"
	"log"
	"path/filepath"

	"github.com/bodgit/sevenzip"
)

func GetMetadata(path string) (bool, *models.Archive, error) {
	if fsutil.Is7z(path) {
		zip, err := sevenzip.OpenReader(path)
		if err != nil {
			log.Println(err)
		}
		defer zip.Close()

		for _, file := range zip.File {
			switch filepath.Base(file.Name) {
			case "ComicInfo.xml":
				content, err := fsutil.Read7z(*file)
				if err != nil {
					return false, nil, err
				}

				data, err := ComicInfoUnmarshal(content)
				if err != nil {
					return false, nil, err
				}

				return true, data, nil
			}
		}
	} else {
		zip, err := zip.OpenReader(path)
		if err != nil {
			log.Println(err)
		}
		defer zip.Close()

		for _, file := range zip.File {
			switch filepath.Base(file.Name) {
			case "ComicInfo.xml":
				content, err := fsutil.ReadZip(*file)
				if err != nil {
					return false, nil, err
				}

				data, err := ComicInfoUnmarshal(content)
				if err != nil {
					return false, nil, err
				}

				return true, data, nil
			}
		}
	}
	return false, nil, nil
}
