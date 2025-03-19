package store

import (
	"Shoka/internal/models"
	"context"

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
		Migrate() error
		// Create
		Create(context.Context, *models.Archive) error
		CreateFromFile() error
		// Read
		GetLastID() (*models.AIDSearch, error)
		GetID(int) (uint, error)
		GetAll() (*[]models.ArchiveSearch, error)
		Get(int) (*models.ArchiveSearch, error)
		GetArtistList(*models.Archive) []string
		GetUrlList(*models.Archive) []string
		GetTagList(*models.Archive) []string
		GetParodyList(*models.Archive) []string
		GetCharacterList(*models.Archive) []string
		// Update
		Update(*models.Archive) error
		// Delete
		Delete(*models.Archive) error
	}

	Tags interface {
		Migrate() error
		Create(context.Context, *Tag) error
	}

	Artist interface {
		Migrate() error
		Create(context.Context, *models.Artist) error
	}

	Group interface {
		Migrate() error
		Create(context.Context, *Group) error
	}
}

func NewStorage(db *gorm.DB) Storage {
	return Storage{
		Archive: &ArchiveStore{db},
		Tags:    &TagStore{db},
		Artist:  &ArtistStore{db},
		Group:   &GroupStore{db},
	}
}
