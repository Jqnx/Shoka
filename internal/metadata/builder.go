package metadata

import (
	"Shoka/internal/config"
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
	case config.SourceComicInfo:
		return newComicInfo()
	case "form":
		return newFormMetadata()
	case config.SourceNhentai:
		return newNHMetadata()
	case config.SourceNhentaiSearch:
		return newNHSearchMetadata()
	}
	return nil
}
