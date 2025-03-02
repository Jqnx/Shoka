package store

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type Group struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"many2many:artist_groups;uniqueIndex;not null" json:"group" form:"group"`
}

type GroupStore struct {
	db *gorm.DB
}

func (s *GroupStore) Migrate() error {
	err := s.db.AutoMigrate(&Group{})
	if err != nil {
		return err
	}
	return nil
}

func (s *GroupStore) Create(ctx context.Context, a *Group) error {
	err := s.db.Debug().WithContext(ctx).Create(a).Error
	if err != nil {
		log.Printf("error creating artist: %v", err)
	}

	return err
}
