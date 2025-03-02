package metadata

import (
	"Shoka/internal/models"
	"encoding/xml"
	"log"
	"strings"
)

// TODO:
// 1. Language and LanguageISO db table?
// 2. Add unmarshalled metadata to database

func ComicInfoUnmarshal(data string) (*models.Archive, error) {
	var archive models.ComicInfoXML
	var tags []models.Tag
	var artists []models.Artist
	var characters []models.Character
	var parodies []models.Parody

	err := xml.Unmarshal([]byte(data), &archive)
	if err != nil {
		log.Println(err)
	}

	// Tags
	taglist := strings.Split(archive.Tags, ",")

	for _, item := range taglist {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		tag := models.Tag{Name: b}
		tags = append(tags, tag)
	}

	// Artists
	artistlist := strings.Split(archive.Writer, ",")

	for _, item := range artistlist {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		artist := models.Artist{Name: b}
		artists = append(artists, artist)
	}

	// Characters
	charlist := strings.Split(archive.Characters, ",")

	for _, item := range charlist {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		if !strings.Contains(b, "unknown") {
			char := models.Character{Name: b}
			characters = append(characters, char)

		}
	}

	// Parodies
	parodylist := strings.Split(archive.Series, ",")

	for _, item := range parodylist {
		a := strings.TrimSpace(item)
		b := strings.ToLower(a)
		if !strings.Contains(b, "unknown") {
			parody := models.Parody{Name: b}
			parodies = append(parodies, parody)
		}
	}

	// URLs
	urls := []models.URL{}
	url := models.URL{Url: archive.URL}
	urls = append(urls, url)

	// Set Archive contents
	updates := &models.Archive{
		Title:     archive.Title,
		Summary:   archive.Summary,
		Tags:      tags,
		Parody:    parodies,
		Artist:    artists,
		Character: characters,
		Language:  strings.ToLower(archive.Language),
		Category:  strings.ToLower(archive.Genre),
		URL:       urls,
	}

	// fmt.Printf("Title: %v\n", archive.Title)
	// fmt.Printf("Summary: %v\n", archive.Summary)
	// fmt.Printf("PageCount: %v\n", archive.PageCount)
	// fmt.Printf("URL: %v\n", urls)
	// fmt.Printf("Genre: %v\n", archive.Genre)
	// fmt.Printf("Series: %v\n", parodies)
	// fmt.Printf("Characters: %v\n", characters)
	// fmt.Printf("Tags: %v\n", tags)
	// fmt.Printf("Writer: %v\n", artists)
	// fmt.Printf("Language: %v\n", archive.Language)

	return updates, nil
}
