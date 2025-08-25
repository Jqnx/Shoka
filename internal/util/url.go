package util

import (
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

func GetFilenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "download_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	filename := filepath.Base(u.Path)
	if filename == "." || filename == "/" {
		return "download_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	return filename
}
