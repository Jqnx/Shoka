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
	Series        string   `xml:"Series,omitempty"`
	Summary       string   `xml:"Summary,omitempty"`
	Year          int      `xml:"Year,omitempty"`
	Month         int      `xml:"Month,omitempty"`
	Day           int      `xml:"Day,omitempty"`
	Writer        string   `xml:"Writer,omitempty"` // Comma seperated
	Genre         string   `xml:"Genre,omitempty"`  // Comma seperated
	Tags          string   `xml:"Tags,omitempty"`   // Comma seperated
	Web           string   `xml:"Web,omitempty"`    // Space seperated, TODO: Spaces in url need to be hex encoded (%20 for space)
	PageCount     int      `xml:"PageCount"`
	Language      string   `xml:"LanguageISO"`
	Characters    string   `xml:"Characters,omitempty"` // Comma seperated
	BlackAndWhite bool     `xml:"BlackAndWhite,omitempty"`
	Manga         bool     `xml:"Manga,omitempty"`
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
