package metadata

import (
	"Shoka/internal/config"
	"Shoka/internal/language"
	"Shoka/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type NHMetadata struct {
	ID         int    `json:"id"`
	MediaID    string `json:"media_id"`
	Title      Title  `json:"title"`
	Images     Images `json:"images"`
	Scanlator  string `json:"scanlator"`
	UploadDate int    `json:"upload_date"`
	Tags       []Tag  `json:"tags"`
	PageCount  int    `json:"num_pages"`
	FavCount   int    `json:"num_favorites"`
}

type Title struct {
	English  string `json:"english"`
	Japanese string `json:"japanese"`
	Pretty   string `json:"pretty"`
}

type Page struct {
	Type   string `json:"t"`
	Width  int    `json:"w"`
	Height int    `json:"h"`
}

type Images struct {
	Pages     []Page `json:"pages"`
	Cover     Page   `json:"cover"`
	Thumbnail Page   `json:"thumbnail"`
}

type Tag struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}

func newNHMetadata() *NHMetadata {
	return &NHMetadata{}
}

func (m *NHMetadata) Unmarshal(data any) error {
	if err := json.Unmarshal(data.([]byte), &m); err != nil {
		return err
	}
	return nil
}

// TODO: Implement groups
func (m *NHMetadata) splitTags() (*[]models.Artist, *[]models.Character, *[]models.Parody, *[]models.Tag, *string, *string) {
	var artists []models.Artist
	var characters []models.Character
	var parodies []models.Parody
	var tags []models.Tag
	var lang string
	var category string
	for _, item := range m.Tags {
		switch item.Type {
		case "artist":
			if strings.Contains(item.Name, "|") {
				for i := range strings.SplitSeq(item.Name, "|") {
					a := strings.TrimSpace(i)
					artist := models.Artist{Artist: a}
					artists = append(artists, artist)
				}
			} else {
				artist := models.Artist{Artist: item.Name}
				artists = append(artists, artist)
			}
		case "category":
			category = item.Name
		case "character":
			if strings.Contains(item.Name, "|") {
				for i := range strings.SplitSeq(item.Name, "|") {
					c := strings.TrimSpace(i)
					character := models.Character{Character: c}
					characters = append(characters, character)
				}
			} else {
				character := models.Character{Character: item.Name}
				characters = append(characters, character)
			}
		case "language":
			if item.Name != "translated" {
				conv := language.NewLanguageConverter()
				tempLang, _ := conv.ToISO(item.Name)
				lang = tempLang
			}
		case "parody":
			if strings.Contains(item.Name, "|") {
				for i := range strings.SplitSeq(item.Name, "|") {
					p := strings.TrimSpace(i)
					parody := models.Parody{Parody: p}
					parodies = append(parodies, parody)
				}
			} else {
				parody := models.Parody{Parody: item.Name}
				parodies = append(parodies, parody)
			}
		case "tag":
			tag := models.Tag{Tag: item.Name}
			tags = append(tags, tag)
		}
	}
	return &artists, &characters, &parodies, &tags, &category, &lang
}

func (m *NHMetadata) getUrl() *[]models.URL {
	var urls []models.URL
	url := models.URL{
		URL: fmt.Sprintf("https://nhentai.net/g/%v", m.ID),
	}
	urls = append(urls, url)
	return &urls
}

func (m *NHMetadata) getReleaseDate() (*time.Time, error) {
	tm := time.Unix(int64(m.UploadDate), 0)
	return &tm, nil
}

func (m *NHMetadata) getImageType() string {
	return config.NHFileTypes[m.Images.Cover.Type]
}

func (m *NHMetadata) GetMetadata() []models.Metadata {
	var metaSlice []models.Metadata
	artists, characters, parodies, tags, category, language := m.splitTags()
	releaseDate, _ := m.getReleaseDate()
	url := m.getUrl()
	imageType := m.getImageType()

	meta := &models.Metadata{
		Title:       m.Title.English,
		Summary:     m.Title.Japanese,
		Artist:      *artists,
		Category:    *category,
		Character:   *characters,
		Language:    *language,
		Parody:      *parodies,
		Tags:        *tags,
		ReleaseDate: releaseDate,
		URL:         *url,
		PageCount:   m.PageCount,
		NHID:        m.ID,
		NHMediaID:   m.MediaID,
		NHImageType: imageType,
	}
	metaSlice = append(metaSlice, *meta)
	return metaSlice
}
