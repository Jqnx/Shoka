package metadata

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"time"
)

// TODO: Language and LanguageISO db table?
// TODO: Add unmarshalled metadata to database

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

func (m *ComicInfo) Unmarshal(data string) {
	err := xml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Println(err)
	}
}

func getURL(u string) *[]URL {
	var urls []URL
	list := strings.Split(u, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		url := URL{URL: b}
		urls = append(urls, url)
	}
	return &urls
}

func getSeries(s string) *[]Parody {
	var series []Parody
	list := strings.Split(s, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		se := Parody{Parody: b}
		series = append(series, se)
	}
	return &series
}

func getCharacters(c string) *[]Character {
	var characters []Character
	list := strings.Split(c, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		character := Character{Character: b}
		characters = append(characters, character)
	}
	return &characters
}

func getTags(t string) *[]Tag {
	var tags []Tag
	list := strings.Split(t, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		tag := Tag{Tag: b}
		tags = append(tags, tag)
	}
	return &tags
}

func getWriter(w string) *[]Artist {
	var writers []Artist
	list := strings.Split(w, ",")

	for _, item := range list {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		writer := Artist{Artist: b}
		writers = append(writers, writer)
	}
	return &writers
}

func getReleaseDate(year, month, day int) *time.Time {
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

func (m *ComicInfo) getMetadata() Metadata {
	urls := getURL(m.URL)
	series := getSeries(m.Series)
	characters := getCharacters(m.Characters)
	tags := getTags(m.Tags)
	writers := getWriter(m.Writer)
	releaseDate := getReleaseDate(m.Year, m.Month, m.Day)
	return Metadata{
		Title:       m.Title,
		Summary:     m.Summary,
		URL:         *urls,
		Category:    strings.ToLower(m.Genre),
		Parody:      *series,
		Character:   *characters,
		Tags:        *tags,
		Artist:      *writers,
		Language:    strings.ToLower(m.Language),
		ReleaseDate: *releaseDate,
	}
}
