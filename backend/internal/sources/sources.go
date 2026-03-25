// Package sources contains the interface for interacting with various other source packages
package sources

import (
	"fmt"
	"net/url"

	"Shoka/internal/config"
	"Shoka/internal/models"
	"Shoka/internal/sources/comicinfo"
	"Shoka/internal/sources/form"
	"Shoka/internal/sources/nhentai"
)

// TODO: Redo the way sources are defined. Use Source struct per source, with fields: name, file, type(remote/local)
// TODO: Add handler which returns only remote sources

type Sources interface {
	Download() // NOTE: Probably doable with gocron instead of asynq
	GetMetadata() ([]models.Metadata, error)
	SetMetadata(data any) error
	SetURL(u *url.URL)
	SetTitle(title string)
	Unmarshal(data any) error
}

func NewSource(cfg *config.Config, source string, method *string) (Sources, error) {
	switch source {
	case config.SourceNhentai:
		if cfg.Metadata.Flaresolverr.URL == "" {
			return nil, fmt.Errorf("this source requires flaresolverr")
		} else {
			NHSource, err := nhentai.NewNhentaiSource(cfg, method)
			if err != nil {
				return nil, err
			}
			return NHSource, nil
		}
	case config.SourceComicInfo:
		return comicinfo.NewComicInfo(), nil
	case config.SourceForm:
		return form.NewForm(), nil
	}
	return nil, nil
}
