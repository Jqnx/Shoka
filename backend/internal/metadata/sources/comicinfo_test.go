package sources

import (
	"Shoka/internal/metadata"
	"bytes"
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

func TestMarshalComicInfoRootElement(t *testing.T) {
	t.Parallel()

	title := "Example"

	data, err := MarshalComicInfo(&metadata.Result{Title: &title})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	dec := xml.NewDecoder(bytes.NewReader(data))

	var start xml.StartElement

	for {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("decode: %v", err)
		}

		if s, ok := tok.(xml.StartElement); ok {
			start = s
			break
		}
	}

	if start.Name.Local != "ComicInfo" {
		t.Errorf("root element = %q, want %q", start.Name.Local, "ComicInfo")
	}
}

func TestComicInfoParseLenientRoot(t *testing.T) {
	t.Parallel()

	result, err := NewComicInfoSource().parse([]byte("<notcomicinfo><Title>x</Title></notcomicinfo>"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if result.Title == nil || *result.Title != "x" {
		t.Errorf("Title = %v, want %q", result.Title, "x")
	}
}

func TestComicInfoParseWeb(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0"?>
<ComicInfo>
  <Title>Example</Title>
  <Web>https://nhentai.net/g/1/ https://e-hentai.org/g/2/abc/</Web>
</ComicInfo>`

	s := NewComicInfoSource()

	result, err := s.parse([]byte(xml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	want := []string{"https://nhentai.net/g/1/", "https://e-hentai.org/g/2/abc/"}

	if !reflect.DeepEqual(result.URLs, want) {
		t.Errorf("URLs = %#v, want %#v", result.URLs, want)
	}
}

func TestMarshalComicInfoWeb(t *testing.T) {
	t.Parallel()

	result := &metadata.Result{
		URLs: []string{"https://nhentai.net/g/1/", "https://e-hentai.org/g/2/abc/"},
	}

	data, err := MarshalComicInfo(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(data), "<Web>https://nhentai.net/g/1/ https://e-hentai.org/g/2/abc/</Web>") {
		t.Errorf("marshalled ComicInfo missing joined Web element:\n%s", data)
	}
}

func TestComicInfoWebRoundTrip(t *testing.T) {
	t.Parallel()

	original := &metadata.Result{
		URLs: []string{"https://nhentai.net/g/1/", "https://e-hentai.org/g/2/abc/"},
	}

	data, err := MarshalComicInfo(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := NewComicInfoSource()

	parsed, err := s.parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if !reflect.DeepEqual(parsed.URLs, original.URLs) {
		t.Errorf("round-trip URLs = %#v, want %#v", parsed.URLs, original.URLs)
	}
}
