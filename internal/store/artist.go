package store

import (
	"Shoka/internal/models"
	"context"
	"log"

	"gorm.io/gorm"
)

type ArtistStore struct {
	db *gorm.DB
}

func (s *ArtistStore) Migrate() error {
	err := s.db.AutoMigrate(
		&models.Artist{},
		&models.Alias{},
		&models.Links{})
	if err != nil {
		return err
	}
	return nil
}

// Create
func (s *ArtistStore) Create(ctx context.Context, a *models.Artist) error {
	err := s.db.Debug().WithContext(ctx).Create(a).Error
	if err != nil {
		log.Printf("error creating artist: %v", err)
	}

	return err
}

// Read
