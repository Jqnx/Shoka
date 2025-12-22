// Package archive contains logic/utility for creating, updating, getting,
// converting archives
package archive

import (
	"time"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
)

// Archive struct that represents an archive
type Archive struct {
	ID          string
	Title       string
	Summary     *string
	Language    *string
	Category    *string
	Tags        *[]models.Tag
	URL         *[]models.URL
	Parody      *[]models.Parody
	Character   *[]models.Character
	Artist      *[]models.Artist
	ReleaseDate *time.Time
	PageCount   int16
	FilePath    string
	FileHash    string
	ThumbPath   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	app         *config.App
	archiveDir  *fsutil.ArchiveDir
}

// NewArchive returns a new empty Archive struct
// that contains only app config
func NewArchive(app *config.App) *Archive {
	return &Archive{app: app}
}
