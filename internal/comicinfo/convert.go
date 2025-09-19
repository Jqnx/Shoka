package comicinfo

import (
	"strings"

	"Shoka/internal/repository"
)

// getBlackWhite checks if a slice of tags contains the "full color" tag
// if it does, returns false
// if it does not, returns true
func getBlackWhite(tags []repository.Tag) bool {
	for _, i := range tags {
		if strings.Contains(i.Name, "full color") {
			return false
		}
	}
	return true
}

// getManga checks if a slice of tags contains the "webtoon" tag
// if it does, returns false
// if it does not, returns true
func getManga(tags []repository.Tag) bool {
	for _, i := range tags {
		if strings.Contains(i.Name, "webtoon") {
			return false
		}
	}
	return true
}
