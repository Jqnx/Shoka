package util

import (
	"path/filepath"
	"strings"
)

// StripExtension removes the final extension from filename, returning it
// unchanged when there is none.
func StripExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
