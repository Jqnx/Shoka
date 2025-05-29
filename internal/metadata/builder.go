package metadata

import "time"

type IMetadata interface {
	Unmarshal(data string)
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
	getMetadata() Metadata
}

type Metadata struct {
	Title       string
	Summary     string
	URL         []URL
	Category    string
	Parody      []Parody
	Character   []Character
	Tags        []Tag
	Artist      []Artist
	Language    string
	ReleaseDate time.Time
}

type URL struct {
	URL string
}

type Parody struct {
	Parody string
}

type Character struct {
	Character string
}

type Tag struct {
	Tag string
}

type Artist struct {
	Artist string
}

func GetBuilder(builderType string) IMetadata {
	if builderType == "comicinfo" {
		return newComicInfo()
	}
	return nil
}
