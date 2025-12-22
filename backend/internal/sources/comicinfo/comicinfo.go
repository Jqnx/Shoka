// Package comicinfo implements utility related to using the ComicInfo standard
package comicinfo

import (
	"encoding/xml"
	"fmt"
	"net/url"

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
	URL           string   `xml:"URL,omitempty"`    // Space seperated, TODO: Spaces in url need to be hex encoded (%20 for space)
	PageCount     int      `xml:"PageCount"`
	Language      string   `xml:"LanguageISO"`
	Characters    string   `xml:"Characters,omitempty"` // Comma seperated
	BlackAndWhite string   `xml:"BlackAndWhite,omitempty"`
	Manga         string   `xml:"Manga,omitempty"`
}

type ComicInfoParams struct {
	Archive   repository.GetArchiveByIDRow
	Artists   []repository.Artist
	Tags      []repository.Tag
	Parody    []repository.Parody
	Character []repository.Character
	URLs      []repository.GetArchiveUrlsRow
}

// NewComicInfo creates a pointer to a new ComicInfo struct
// which gets filled metadata from the given ComicInfoParams
func NewComicInfo() *ComicInfo {
	return &ComicInfo{
		Schema: schema,
	}
}

func (c *ComicInfo) Unmarshal(data any) error {
	err := xml.Unmarshal([]byte(data.([]byte)), &c)
	if err != nil {
		return err
	}
	return nil
}

func (c *ComicInfo) SetMetadata(data any) error {
	switch data := data.(type) {
	case ComicInfoParams:
		ci := data

		c.Title = data.Archive.Title
		c.Summary = *data.Archive.Summary
		c.Genre = *data.Archive.Category
		c.PageCount = int(data.Archive.PageCount)
		c.Language = *data.Archive.Language
		c.Writer = util.ToString(ci.Artists)
		c.Tags = util.ToString(ci.Tags)
		c.Series = util.ToString(ci.Parody)
		c.Characters = util.ToString(ci.Character)
		c.Web = util.ToString(ci.URLs)
		c.BlackAndWhite = getBlackWhite(ci.Tags)
		c.Manga = getManga(ci.Tags)
		year, month, day := ci.Archive.ReleaseDate.Date()
		c.Year = year
		c.Month = int(month)
		c.Day = day
	default:
		return fmt.Errorf("data is of incorrect type, needs to be of type ComicInfoParams")
	}

	return nil
}

func (c *ComicInfo) Download() {}

func (c *ComicInfo) SetURL(u *url.URL) {}

func (c *ComicInfo) SetTitle(title string) {}
