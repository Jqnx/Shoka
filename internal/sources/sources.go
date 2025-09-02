package sources

import (
	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/sources/nhentai"
	"fmt"
	"net/url"
)

// TODO: custom types to use for declaring source

type Sources interface {
	Download()
	GetMetadata(method string) ([]models.Metadata, error)
	SetURL(u *url.URL)
	SetTitle(title string)
}

func NewSource(source string, cfg *config.Config) (Sources, error) {
	switch source {
	case config.SourceNhentai, config.SourceNhentaiSearch:
		if cfg.Sources.Flaresolverr.URL == "" {
			return nil, fmt.Errorf("this source requires flaresolverr")
		} else {
			return nhentai.NewNhentaiSource(cfg, source), nil
		}
	}
	return nil, nil
}
