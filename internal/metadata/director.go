package metadata

import (
	"Shoka/internal/models"
)

type Director struct {
	source IMetadata
}

func NewDirector(source IMetadata) *Director {
	return &Director{
		source: source,
	}
}

func (d *Director) FetchMetadata(data any) ([]models.Metadata, error) {
	if err := d.source.Unmarshal(data); err != nil {
		return nil, err
	}
	return d.source.GetMetadata(), nil
}
