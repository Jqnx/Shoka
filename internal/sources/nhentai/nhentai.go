package nhentai

import (
	"Shoka/internal/config"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Nhentai struct {
	cfg       *config.Config
	URL       *url.URL
	Source    string
	GalleryID string
	Title     string
}

func NewNhentaiSource(cfg *config.Config, source string) *Nhentai {
	return &Nhentai{
		cfg:    cfg,
		Source: source,
	}
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
