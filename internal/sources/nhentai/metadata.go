package nhentai

import (
	"fmt"
	"strings"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/language"
	"Shoka/internal/models"
	"Shoka/internal/sources/flaresolverr"
)

func (s *Nhentai) GetMetadata() ([]models.Metadata, error) {
	switch s.Method {
	case config.MethodID:
		byte, err := s.FetchMetadataByID()
		if err != nil {
			return nil, err
		}
		s.Unmarshal(byte)
		meta := s.Metadata.Get()
		return meta, nil
	case config.MethodTitle:
		byte, err := s.FetchMetadataByTitle(s.Title)
		if err != nil {
			return nil, err
		}
		s.Unmarshal(byte)
		meta := s.SearchMetadata.Get()
		return meta, nil
	default:
		return nil, fmt.Errorf("invalid method used")
	}
}

func (s *Nhentai) FetchMetadataByID() ([]byte, error) {
	if err := s.GetGalleryID(); err != nil {
		return nil, err
	}

	gallery := fmt.Sprintf("%v/gallery/%s", config.NHApi, s.GalleryID)

	byte, err := flaresolverr.Request(s.cfg, gallery)
	if err != nil {
		return nil, err
	}

	return byte, nil
}

func (s *Nhentai) FetchMetadataByTitle(title string) ([]byte, error) {
	gallery := fmt.Sprintf("%v/galleries/search?query=%s", config.NHApi, title)

	byte, err := flaresolverr.Request(s.cfg, gallery)
	if err != nil {
		return nil, err
	}

	return byte, nil
}

func (s *Nhentai) ToLocalMetadata() ([]models.Metadata, error) {
	return nil, nil
}

func (m *Metadata) splitTags() (*[]models.Artist, *[]models.Character, *[]models.Parody, *[]models.Tag, *string, *string) {
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

func (m *Metadata) getURL() *[]models.URL {
	var urls []models.URL
	url := models.URL{
		URL: fmt.Sprintf("https://nhentai.net/g/%v", m.ID),
	}
	urls = append(urls, url)
	return &urls
}

func (m *Metadata) getReleaseDate() (*time.Time, error) {
	tm := time.Unix(int64(m.UploadDate), 0)
	return &tm, nil
}

func (m *Metadata) getImageType() string {
	return config.NHFileTypes[m.Images.Cover.Type]
}

func (m *Metadata) Get() []models.Metadata {
	var metaSlice []models.Metadata
	artists, characters, parodies, tags, category, language := m.splitTags()
	releaseDate, _ := m.getReleaseDate()
	url := m.getURL()
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

func (m *NHSearch) Get() []models.Metadata {
	var metaSlice []models.Metadata

	for _, i := range m.Metadata {
		artists, characters, parodies, tags, category, language := i.splitTags()
		releaseDate, _ := i.getReleaseDate()
		url := i.getURL()
		imageType := i.getImageType()
		meta := &models.Metadata{
			Title:       i.Title.English,
			Summary:     i.Title.Japanese,
			Artist:      *artists,
			Category:    *category,
			Character:   *characters,
			Language:    *language,
			Parody:      *parodies,
			Tags:        *tags,
			ReleaseDate: releaseDate,
			URL:         *url,
			PageCount:   i.PageCount,
			NHID:        i.ID,
			NHMediaID:   i.MediaID,
			NHImageType: imageType,
		}
		metaSlice = append(metaSlice, *meta)
	}
	return metaSlice
}
