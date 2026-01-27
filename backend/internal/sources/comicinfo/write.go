package comicinfo

import (
	"encoding/xml"
	"os"
	"path/filepath"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
)

// Write writes the contents of the ComicInfo struct to file
// and adds it to the desired zip file
func (c *ComicInfo) Write(path, tempDir string) error {
	output, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	tempFile := filepath.Join(tempDir, config.ComicInfoFile)

	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(file.Name())

	if _, err := file.WriteString(xml.Header); err != nil {
		return err
	}

	if _, err := file.Write(output); err != nil {
		return err
	}

	if err = fsutil.AddToExistingZip(path, file.Name(), tempDir, config.ComicInfoFile); err != nil {
		return err
	}
	return nil
}
