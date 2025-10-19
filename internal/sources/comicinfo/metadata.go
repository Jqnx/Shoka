package comicinfo

import (
	"fmt"
	"strings"
	"time"

	"Shoka/internal/language"
	"Shoka/internal/models"
	"Shoka/internal/repository"
)

// getBlackWhite checks if a slice of tags contains the "full color" tag
// if it does, returns false
// if it does not, returns true
func getBlackWhite(tags []repository.Tag) bool {
	for _, i := range tags {
		if strings.Contains(i.Name, "full color") {
			return false
		}
	}
	return true
}

// getManga checks if a slice of tags contains the "webtoon" tag
// if it does, returns false
// if it does not, returns true
func getManga(tags []repository.Tag) bool {
	for _, i := range tags {
		if strings.Contains(i.Name, "webtoon") {
			return false
		}
	}
	return true
}

func (c *ComicInfo) getURL() *[]models.URL {
	var urls []models.URL
	var list []string

	l1 := strings.SplitSeq(c.URL, ",")
	for item := range l1 {
		if len(item) != 0 {
			list = append(list, item)
		}
	}

	l2 := strings.SplitSeq(c.Web, ",")
	for item := range l2 {
		if len(item) != 0 {
			list = append(list, item)
		}
	}

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		url := models.URL{URL: b}
		urls = append(urls, url)
	}
	return &urls
}

func (c *ComicInfo) getSeries() *[]models.Parody {
	var series []models.Parody
	list := strings.SplitSeq(c.Series, ",")

	for item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		if b != "unknown" {
			se := models.Parody{Parody: b}
			series = append(series, se)
		}
	}
	return &series
}

func (c *ComicInfo) getCharacters() *[]models.Character {
	var characters []models.Character
	list := strings.SplitSeq(c.Characters, ",")

	for item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		if b != "unknown" {
			character := models.Character{Character: b}
			characters = append(characters, character)
		}
	}
	return &characters
}

func (c *ComicInfo) getTags() *[]models.Tag {
	var tags []models.Tag
	list := strings.SplitSeq(c.Tags, ",")

	for item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		tag := models.Tag{Tag: b}
		tags = append(tags, tag)
	}
	return &tags
}

func (c *ComicInfo) getWriter() *[]models.Artist {
	var writers []models.Artist
	list := strings.SplitSeq(c.Writer, ",")

	for item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		writer := models.Artist{Artist: b}
		writers = append(writers, writer)
	}
	return &writers
}

func (c *ComicInfo) getLanguage() string {
	conv := language.NewLanguageConverter()
	lang, _ := conv.ToISO(c.Language)
	return lang
}

func (c *ComicInfo) getReleaseDate(year, month, day int) *time.Time {
	switch {
	case day < 10 && month < 10:
		dayString := fmt.Sprintf("0%d", day)
		monthString := fmt.Sprintf("0%d", month)
		dateString := fmt.Sprintf("%d-%s-%s", year, monthString, dayString)
		date, err := time.Parse(time.DateOnly, dateString)
		if err != nil {
			fmt.Println(err)
		}
		return &date
	case day < 10:
		dayString := fmt.Sprintf("0%d", day)
		dateString := fmt.Sprintf("%d-%d-%s", year, month, dayString)
		date, err := time.Parse(time.DateOnly, dateString)
		if err != nil {
			fmt.Println(err)
		}
		return &date
	case month < 10:
		monthString := fmt.Sprintf("0%d", month)
		dateString := fmt.Sprintf("%d-%s-%d", year, monthString, day)
		date, err := time.Parse(time.DateOnly, dateString)
		if err != nil {
			fmt.Println(err)
		}
		return &date
	default:
		d := fmt.Sprintf("%d-%d-%d", year, month, day)
		date, err := time.Parse(time.DateOnly, d)
		if err != nil {
			fmt.Println(err)
		}
		return &date
	}
}

func (c *ComicInfo) GetMetadata() ([]models.Metadata, error) {
	var metaSlice []models.Metadata
	urls := c.getURL()
	series := c.getSeries()
	characters := c.getCharacters()
	tags := c.getTags()
	writers := c.getWriter()
	lang := c.getLanguage()
	releaseDate := c.getReleaseDate(c.Year, c.Month, c.Day)
	meta := &models.Metadata{
		Title:       c.Title,
		Summary:     c.Summary,
		URL:         *urls,
		Category:    strings.ToLower(c.Genre),
		Parody:      *series,
		Character:   *characters,
		Tags:        *tags,
		Artist:      *writers,
		Language:    lang,
		ReleaseDate: releaseDate,
		PageCount:   c.PageCount,
	}
	metaSlice = append(metaSlice, *meta)
	return metaSlice, nil
}
