package store

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type Tag struct {
	Name string `gorm:"many2many:archive_tags;primaryKey;not null" json:"tag" form:"tag"` // Tag name
}

type TagStore struct {
	db *gorm.DB
}

func (s *TagStore) Migrate() error {
	err := s.db.AutoMigrate(&Tag{})
	if err != nil {
		return err
	}

	return nil
}

func (s *TagStore) Create(ctx context.Context, a *Tag) error {
	err := s.db.WithContext(ctx).Create(a).Error
	if err != nil {
		log.Printf("error creating tag: %v", err)
	}
	return err
}
