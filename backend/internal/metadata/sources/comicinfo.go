package sources

import (
	"Shoka/internal/library/archive"
	"Shoka/internal/metadata"
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type comicInfoXML struct {
	Title       string `xml:"Title"`
	Summary     string `xml:"Summary"`
	LanguageISO string `xml:"LanguageISO"`
	Genre       string `xml:"Genre"`
	Year        int    `xml:"Year"`
	Month       int    `xml:"Month"`
	Day         int    `xml:"Day"`
	PageCount   int64  `xml:"PageCount"`
	Writer      string `xml:"Writer"`
	Characters  string `xml:"Characters"`
	SeriesGroup string `xml:"SeriesGroup"`
}

type ComicInfoSource struct{}

func NewComicInfoSource() *ComicInfoSource {
	return &ComicInfoSource{}
}

func (s *ComicInfoSource) Name() string  { return "comicinfo" }
func (s *ComicInfoSource) Priority() int { return 1 }

func (s *ComicInfoSource) Fetch(ctx context.Context, input metadata.Input) (*metadata.Result, error) {
	a, err := archive.Open(input.FilePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer a.Close()

	data, err := a.ReadFile("ComicInfo.xml")
	if err != nil {
		return nil, fmt.Errorf("read ComicInfo.xml: %w", err)
	}

	if data == nil {
		return nil, nil
	}

	return s.parse(data)
}

func (s *ComicInfoSource) parse(data []byte) (*metadata.Result, error) {
	var info comicInfoXML
	if err := xml.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parse ComicInfo.xml: %w", err)
	}

	result := &metadata.Result{}

	if info.Title != "" {
		result.Title = &info.Title
	}
	if info.Summary != "" {
		result.Summary = &info.Summary
	}
	if info.LanguageISO != "" {
		result.Language = &info.LanguageISO
	}
	if info.PageCount > 0 {
		result.PageCount = &info.PageCount
	}
	if info.Writer != "" {
		artists := splitAndTrim(info.Writer, ",")
		result.Artists = artists
	}
	if info.Genre != "" {
		result.Tags = splitAndTrim(info.Genre, ",")
	}
	if info.Characters != "" {
		result.Characters = splitAndTrim(info.Characters, ",")
	}
	if info.SeriesGroup != "" {
		result.Parodies = []string{info.SeriesGroup}
	}
	if info.Year > 0 {
		t := time.Date(info.Year, time.Month(info.Month), info.Day, 0, 0, 0, 0, time.UTC)
		result.ReleaseDate = &t
	}

	return result, nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
