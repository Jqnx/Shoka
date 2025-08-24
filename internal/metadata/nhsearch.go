package metadata

import (
	"Shoka/internal/models"
	"encoding/json"
)

type NHSearch struct {
	Metadata []NHMetadata `json:"result"`
	NumPages int          `json:"num_pages"`
	PerPage  int          `json:"per_page"`
}

func newNHSearchMetadata() *NHSearch {
	return &NHSearch{}
}

func (m *NHSearch) Unmarshal(data any) error {
	if err := json.Unmarshal(data.([]byte), &m); err != nil {
		return err
	}
	return nil
}

func (m *NHSearch) GetMetadata() []models.Metadata {
	var metaSlice []models.Metadata

	for _, i := range m.Metadata {
		artists, characters, parodies, tags, category, language := i.splitTags()
		releaseDate, _ := i.getReleaseDate()
		url := i.getUrl()
		meta := &models.Metadata{
			Title:       i.Title.English,
			Summary:     i.Title.Japanese,
			Artist:      *artists,
			Category:    *category,
			Character:   *characters,
			Language:    *language,
			Parody:      *parodies,
			Tags:        *tags,
			ReleaseDate: releaseDate,
			URL:         *url,
			PageCount:   i.PageCount,
		}
		metaSlice = append(metaSlice, *meta)
	}
	return metaSlice
}
