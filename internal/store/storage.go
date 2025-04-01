package store

import (
	"Shoka/internal/models"

	"gorm.io/gorm"
)

var (
	imageExtensions   = []string{"png", "jpg", "jpeg", "gif", "webp"}
	archiveExtensions = []string{"zip", "cbz", "7z"}
	xml               = []string{"xml"}
	json              = []string{"json"}
	txt               = []string{"txt"}
)

type Storage struct {
	Archive interface {
		// Create
		Create(*models.Archive) error
		CreateFromFile() error
		// Read
		GetLastID() (*models.AIDSearch, error)
		GetID(int) int
		GetAll() (*[]models.ArchiveSearch, error)
		Get(int) (*models.ArchiveSearch, error)
		GetArtistList(*models.Archive) []string
		GetUrlList(*models.Archive) []string
		GetTagList(*models.Archive) []string
		GetParodyList(*models.Archive) []string
		GetCharacterList(*models.Archive) []string
		TitleExists(*models.Archive) bool
		// Update
		Update(*models.Archive) error
		// Delete
		Delete(*models.Archive) error
	}
}

func NewStorage(db *gorm.DB) Storage {
	return Storage{
		Archive: &ArchiveStore{db},
	}
}
