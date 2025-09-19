// Package comicinfo implements utility related to using the ComicInfo standard
package comicinfo

import (
	"encoding/xml"

	"Shoka/internal/repository"
	"Shoka/internal/util"
)

var schema = "http://www.w3.org/2001/XMLSchema"

type ComicInfo struct {
	XMLName       xml.Name `xml:"ComicInfo"`
	Schema        string   `xml:"xmlns:xs,attr"`
	Title         string   `xml:"Title"`
	Series        string   `xml:"Series"`
	Summary       string   `xml:"Summary"`
	Year          int      `xml:"Year"`
	Month         int      `xml:"Month"`
	Day           int      `xml:"Day"`
	Writer        string   `xml:"Writer"` // Comma seperated
	Genre         string   `xml:"Genre"`  // Comma seperated
	Tags          string   `xml:"Tags"`   // Comma seperated
	Web           string   `xml:"Web"`    // Space seperated, TODO: Spaces in url need to be hex encoded (%20 for space)
	PageCount     int      `xml:"PageCount"`
	Language      string   `xml:"LanguageISO"`
	Characters    string   `xml:"Characters"` // Comma seperated
	BlackAndWhite bool     `xml:"BlackAndWhite"`
	Manga         bool     `xml:"Manga"`
}

type ComicInfoParams struct {
	Archive   repository.GetArchiveByIDRow
	Artists   []repository.Artist
	Tags      []repository.Tag
	Parody    []repository.Parody
	Character []repository.Character
	URLs      []repository.GetArchiveURLsRow
}

// NewComicInfo creates a pointer to a new ComicInfo struct
// which gets filled metadata from the given ComicInfoParams
func NewComicInfo(p ComicInfoParams) *ComicInfo {
	artists := util.ToString(p.Artists)
	tags := util.ToString(p.Tags)
	parodies := util.ToString(p.Parody)
	characters := util.ToString(p.Character)
	urls := util.ToString(p.URLs)

	bw := getBlackWhite(p.Tags)
	manga := getManga(p.Tags)
	year, month, day := p.Archive.ReleaseDate.Date()

	return &ComicInfo{
		Schema:        schema,
		Title:         p.Archive.Title,
		Series:        parodies,
		Summary:       *p.Archive.Summary,
		Year:          year,
		Month:         int(month),
		Day:           day,
		Writer:        artists,
		Genre:         *p.Archive.Category,
		Tags:          tags,
		Web:           urls,
		PageCount:     int(p.Archive.PageCount),
		Language:      *p.Archive.Language,
		Characters:    characters,
		BlackAndWhite: bw,
		Manga:         manga,
	}
}
