// Package archive contains logic/utility for creating, updating, getting,
// converting archives
package archive

import (
	"time"

	"Shoka/internal/models"
)

// Archive interface that implements builder functions
type IArchive interface {
	setArchiveID(id string)
	setTitle(title string)
	setSummary(summary string)
	setLanguage(language string)
	setCategory(category string)
	setPageCount()
	setFilePath(path string) // NOTE: SHOULD BE SET FIRST IN DIRECTOR CLASS, OTHER "set" FUNCTIONS USE IT e.g. setTitle, setPageCount, setHash, setThumbsPath, setCoverPath
	setFileName()
	setHash()
	setThumbsPath()
	setCoverPath()
	createPagesPath()
	setType()
	setCreatedAt()
	setUpdatedAt()

	setTags()
	setArtist()
	setCharacter()
	setParody()
	setURL()
	setReleaseDate()
	getArchive() Archive
}

// Archive struct that represents an archive
type Archive struct {
	ID          string
	Type        string
	Title       *string
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
	FilePath    *string
	FileName    *string
	Hash        string
	ThumbsPath  *string
	CoverPath   *string
	PagesPath   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
