package util

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
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

func ExtractDomainFromURL(u string) (string, *url.URL, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", nil, err
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return "", nil, fmt.Errorf("no hostname in url")
	}

	parts := strings.Split(hostname, ".")

	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid domain format")
	}

	return parts[len(parts)-2], parsed, nil
}
