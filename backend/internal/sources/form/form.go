// Package form contains functionality for converting
// metadata from HTTP forms to the local Metadata model
package form

import (
	"fmt"
	"net/url"
	"time"

	"Shoka/internal/models"
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

func NewForm() *Form {
	return &Form{}
}

func (f *Form) Unmarshal(data any) error {
	d := data.(models.ArchivePayload)
	f.Title = d.Title
	f.Summary = d.Summary
	f.URL = d.URL
	f.Category = d.Category
	f.Parody = d.Parody
	f.Character = d.Character
	f.Tags = d.Tags
	f.Artist = d.Artist
	f.Language = d.Language
	f.ReleaseDate = d.ReleaseDate
	return nil
}

func (f *Form) Download() {
	// Does nothing, is only here because interface requires it
}

func (f *Form) SetMetadata(data any) error {
	return fmt.Errorf("SetMetadata not supported for form source")
}

func (f *Form) SetURL(u *url.URL) {
	var urls []string
	urls = append(urls, u.String())
	f.URL = urls
}

func (f *Form) SetTitle(title string) {
	f.Title = title
}
