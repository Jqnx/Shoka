package nhentai

import (
	"Shoka/internal/config"
	"Shoka/internal/metadata"
	"Shoka/internal/models"
	"fmt"
)

func (s *Nhentai) GetMetadata(method string) ([]models.Metadata, error) {
	switch method {
	case config.MethodID:
		byte, err := s.GetMetadataByID()
		if err != nil {
			return nil, err
		}
		director := metadata.NewDirector(metadata.GetBuilder(s.Source))
		meta, err := director.FetchMetadata(byte)
		if err != nil {
			return nil, err
		}
		return meta, nil
	case config.MethodTitle:
		byte, err := s.GetMetadataByTitle(s.Title)
		if err != nil {
			return nil, err
		}
		director := metadata.NewDirector(metadata.GetBuilder(s.Source))
		meta, err := director.FetchMetadata(byte)
		if err != nil {
			return nil, err
		}
		return meta, nil
	default:
		return nil, fmt.Errorf("invalid method used")
	}
}

func (s *Nhentai) GetMetadataByID() ([]byte, error) {
	if err := s.GetGalleryID(); err != nil {
		return nil, err
	}

	gallery := fmt.Sprintf("%v/gallery/%s", config.NHApi, s.GalleryID)

	byte, err := s.Request(gallery)
	if err != nil {
		return nil, err
	}

	return byte, nil
}

func (s *Nhentai) GetMetadataByTitle(title string) ([]byte, error) {
	gallery := fmt.Sprintf("%v/galleries/search?query=%s", config.NHApi, title)

	byte, err := s.Request(gallery)
	if err != nil {
		return nil, err
	}

	return byte, nil
}
