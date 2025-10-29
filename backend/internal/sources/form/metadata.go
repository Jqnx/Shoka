package form

import (
	"strings"

	"Shoka/internal/models"
)

func (f *Form) getURL() *[]models.URL {
	var urls []models.URL

	for _, item := range f.URL {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		url := models.URL{URL: b}
		urls = append(urls, url)
	}
	return &urls
}

func (f *Form) getParody() *[]models.Parody {
	var parodies []models.Parody

	for _, item := range f.Parody {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		parody := models.Parody{Parody: b}
		parodies = append(parodies, parody)
	}
	return &parodies
}

func (f *Form) getCharacters() *[]models.Character {
	var characters []models.Character

	for _, item := range f.Character {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		character := models.Character{Character: b}
		characters = append(characters, character)
	}
	return &characters
}

func (f *Form) getTags() *[]models.Tag {
	var tags []models.Tag

	for _, item := range f.Tags {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		tag := models.Tag{Tag: b}
		tags = append(tags, tag)
	}
	return &tags
}

func (f *Form) getArtist() *[]models.Artist {
	var artists []models.Artist

	for _, item := range f.Artist {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		artist := models.Artist{Artist: b}
		artists = append(artists, artist)
	}
	return &artists
}

func (f *Form) GetMetadata() ([]models.Metadata, error) {
	var metaSlice []models.Metadata
	urls := f.getURL()
	parodies := f.getParody()
	characters := f.getCharacters()
	tags := f.getTags()
	artists := f.getArtist()
	meta := &models.Metadata{
		Title:       f.Title,
		Summary:     f.Summary,
		URL:         *urls,
		Category:    f.Category,
		Parody:      *parodies,
		Character:   *characters,
		Tags:        *tags,
		Artist:      *artists,
		Language:    f.Language,
		ReleaseDate: f.ReleaseDate,
	}
	metaSlice = append(metaSlice, *meta)
	return metaSlice, nil
}
