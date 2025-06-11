package metadata

import (
	"Shoka/internal/models"
	"fmt"
	"strings"
	"time"
)

type Form struct {
	Title       string
	Summary     string
	URL         []string
	Category    string
	Parody      []string
	Character   []string
	Tags        []string
	Artist      []string
	Language    string
	ReleaseDate *time.Time
}

func newFormMetadata() *Form {
	return &Form{}
}

func (m *Form) Unmarshal(data any) {
	d := data.(models.ArchivePayload)
	m.Title = d.Title
	m.Summary = d.Summary
	m.URL = d.URL
	m.Category = d.Category
	m.Parody = d.Parody
	m.Character = d.Character
	m.Tags = d.Tags
	m.Artist = d.Artist
	m.Language = d.Language
	m.ReleaseDate = d.ReleaseDate
}

func (m *Form) getURL() *[]URL {
	var urls []URL

	for _, item := range m.URL {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		url := URL{URL: b}
		urls = append(urls, url)
	}
	return &urls
}

func (m *Form) getParody() *[]Parody {
	var parodies []Parody

	for _, item := range m.Parody {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		parody := Parody{Parody: b}
		parodies = append(parodies, parody)
	}
	return &parodies
}

func (m *Form) getCharacters() *[]Character {
	var characters []Character

	for _, item := range m.Character {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		character := Character{Character: b}
		characters = append(characters, character)
	}
	return &characters
}

func (m *Form) getTags() *[]Tag {
	var tags []Tag

	for _, item := range m.Tags {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		tag := Tag{Tag: b}
		tags = append(tags, tag)
	}
	return &tags
}

func (m *Form) getArtist() *[]Artist {
	var artists []Artist

	for _, item := range m.Artist {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		artist := Artist{Artist: b}
		artists = append(artists, artist)
	}
	return &artists
}

// TODO: Fix Release Date
func (m *Form) getReleaseDate(year, month, day int) *time.Time {
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

func (m *Form) getMetadata() Metadata {
	urls := m.getURL()
	parodies := m.getParody()
	characters := m.getCharacters()
	tags := m.getTags()
	artists := m.getArtist()
	return Metadata{
		Title:       m.Title,
		Summary:     m.Summary,
		URL:         *urls,
		Category:    m.Category,
		Parody:      *parodies,
		Character:   *characters,
		Tags:        *tags,
		Artist:      *artists,
		Language:    m.Language,
		ReleaseDate: m.ReleaseDate,
	}
}
