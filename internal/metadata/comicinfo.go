package metadata

import (
	"Shoka/internal/models"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"time"
)

type ComicInfo struct {
	Title      string `xml:"Title"`
	Summary    string `xml:"Summary"`
	URL        string `xml:"URL"`
	Genre      string `xml:"Genre"`
	Series     string `xml:"Series"`
	Characters string `xml:"Characters"`
	Tags       string `xml:"Tags"`
	Writer     string `xml:"Writer"`
	Language   string `xml:"LanguageISO"`
	Year       int    `xml:"Year"`
	Month      int    `xml:"Month"`
	Day        int    `xml:"Day"`
}

func newComicInfo() *ComicInfo {
	return &ComicInfo{}
}

func (m *ComicInfo) Unmarshal(data any) {
	err := xml.Unmarshal([]byte(data.(string)), &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *ComicInfo) getURL() *[]models.URL {
	var urls []models.URL
	list := strings.Split(m.URL, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		url := models.URL{URL: b}
		urls = append(urls, url)
	}
	return &urls
}

func (m *ComicInfo) getSeries() *[]models.Parody {
	var series []models.Parody
	list := strings.Split(m.Series, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		se := models.Parody{Parody: b}
		series = append(series, se)
	}
	return &series
}

func (m *ComicInfo) getCharacters() *[]models.Character {
	var characters []models.Character
	list := strings.Split(m.Characters, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		character := models.Character{Character: b}
		characters = append(characters, character)
	}
	return &characters
}

func (m *ComicInfo) getTags() *[]models.Tag {
	var tags []models.Tag
	list := strings.Split(m.Tags, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		tag := models.Tag{Tag: b}
		tags = append(tags, tag)
	}
	return &tags
}

func (m *ComicInfo) getWriter() *[]models.Artist {
	var writers []models.Artist
	list := strings.Split(m.Writer, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		writer := models.Artist{Artist: b}
		writers = append(writers, writer)
	}
	return &writers
}

func (m *ComicInfo) getReleaseDate(year, month, day int) *time.Time {
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

func (m *ComicInfo) GetMetadata() models.Metadata {
	urls := m.getURL()
	series := m.getSeries()
	characters := m.getCharacters()
	tags := m.getTags()
	writers := m.getWriter()
	releaseDate := m.getReleaseDate(m.Year, m.Month, m.Day)
	return models.Metadata{
		Title:       m.Title,
		Summary:     m.Summary,
		URL:         *urls,
		Category:    strings.ToLower(m.Genre),
		Parody:      *series,
		Character:   *characters,
		Tags:        *tags,
		Artist:      *writers,
		Language:    strings.ToLower(m.Language),
		ReleaseDate: releaseDate,
	}
}
