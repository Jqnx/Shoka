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
		switch {
		case cfg.Sources.NHentai.CSRFToken == "":
			return nil, fmt.Errorf("missing CSRFToken")
		case cfg.Sources.NHentai.UserAgent == "":
			return nil, fmt.Errorf("missing UserAgent")
		default:
			return nhentai.NewNhentaiSource(cfg, source), nil
		}
	}
	return nil, nil
}
