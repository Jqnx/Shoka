package metadata

import (
	"Shoka/internal/models"
)

type IMetadata interface {
	Unmarshal(data any) error
	// setTitle(title string)
	// setSummary()
	// setPageCount()
	// setURLs()
	// setGenre()
	// setSeries()
	// setCharacters()
	// setTags()
	// setWriter()
	// setLanguage()
	GetMetadata() []models.Metadata
}

func GetBuilder(builderType string) IMetadata {
	switch builderType {
	case "form":
		// TODO: Move to a "form" source package
		return newFormMetadata()
	}
	return nil
}
