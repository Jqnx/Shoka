// Package nhentai is a source package that contains utility
// for interacting with the nhentai API to fetch metadata
package nhentai

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"Shoka/internal/config"
)

type Nhentai struct {
	cfg            *config.Config
	URL            *url.URL
	Method         string
	GalleryID      string
	Title          string
	Metadata       Metadata
	SearchMetadata NHSearch
}

func NewNhentaiSource(cfg *config.Config, method *string) (*Nhentai, error) {
	if method == nil || *method == "" {
		return nil, fmt.Errorf("no method given")
	}

	return &Nhentai{
		cfg:    cfg,
		Method: *method,
	}, nil
}

func (s *Nhentai) SetMetadata(data any) error {
	return fmt.Errorf("SetMetadata is not supported on this source")
}

func (s *Nhentai) SetURL(u *url.URL) {
	s.URL = u
}

func (s *Nhentai) SetTitle(title string) {
	newTitle := strings.ReplaceAll(title, " ", "%20")
	s.Title = newTitle
}

func (s *Nhentai) GetGalleryID() error {
	id := strings.Split(s.URL.Path, "/")

	_, err := strconv.Atoi(id[2])
	if err != nil || len(id[2]) > 6 {
		return fmt.Errorf("invalid url")
	}

	s.GalleryID = id[2]
	return nil
}

func (s *Nhentai) Unmarshal(data any) error {
	switch s.Method {
	case config.MethodID:
		if err := json.Unmarshal(data.([]byte), &s.Metadata); err != nil {
			return err
		}
	case config.MethodTitle:
		if err := json.Unmarshal(data.([]byte), &s.SearchMetadata); err != nil {
			return err
		}
	}

	return nil
}
