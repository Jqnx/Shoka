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
	Series      string `xml:"Series"`
	Tags        string `xml:"Tags"`
	Web         string `xml:"Web"`
}

type ComicInfoSource struct{}

func NewComicInfoSource() *ComicInfoSource {
	return &ComicInfoSource{}
}

func (s *ComicInfoSource) Name() string  { return "comicinfo" }
func (s *ComicInfoSource) Priority() int { return 1 }
func (s *ComicInfoSource) IsLocal() bool { return true }

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

	if info.Genre != "" {
		result.Category = &info.Genre
	}

	if info.PageCount > 0 {
		result.PageCount = &info.PageCount
	}

	if info.Writer != "" {
		artists := splitAndTrim(info.Writer, ",")
		result.Artists = artists
	}

	if info.Tags != "" {
		result.Tags = splitAndTrim(info.Tags, ",")
	}

	if info.Characters != "" {
		result.Characters = splitAndTrim(info.Characters, ",")
	}

	if info.Series != "" {
		result.Parodies = splitAndTrim(info.Series, ",")
	}

	if info.Year > 0 {
		t := time.Date(info.Year, time.Month(info.Month), info.Day, 0, 0, 0, 0, time.UTC)
		result.ReleaseDate = &t
	}

	// The ComicInfo spec allows <Web> to hold multiple space-separated URLs.
	// The source name is derived from each link's host downstream, so a
	// nhentai link here still shows the nhentai icon, not a "comicinfo" one.
	if info.Web != "" {
		result.URLs = strings.Fields(info.Web)
	}

	return result, nil
}

func MarshalComicInfo(result *metadata.Result) ([]byte, error) {
	info := comicInfoXML{}

	if result.Title != nil {
		info.Title = *result.Title
	}

	if result.Summary != nil {
		info.Summary = *result.Summary
	}

	if result.Language != nil {
		info.LanguageISO = *result.Language
	}

	if result.Category != nil {
		info.Genre = *result.Category
	}

	if result.PageCount != nil {
		info.PageCount = *result.PageCount
	}

	if result.ReleaseDate != nil {
		info.Year = result.ReleaseDate.Year()
		info.Month = int(result.ReleaseDate.Month())
		info.Day = result.ReleaseDate.Day()
	}

	if len(result.Artists) > 0 {
		info.Writer = strings.Join(result.Artists, ", ")
	}

	if len(result.Tags) > 0 {
		info.Tags = strings.Join(result.Tags, ", ")
	}

	if len(result.Characters) > 0 {
		info.Characters = strings.Join(result.Characters, ", ")
	}

	if len(result.Parodies) > 0 {
		info.Series = result.Parodies[0]
	}

	if len(result.URLs) > 0 {
		info.Web = strings.Join(result.URLs, " ")
	}

	output, err := xml.MarshalIndent(info, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal ComicInfo.xml: %w", err)
	}

	return append([]byte(xml.Header), output...), nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)

	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.ToLower(strings.TrimSpace(p)); t != "" && t != "unknown" {
			result = append(result, t)
		}
	}

	return result
}
