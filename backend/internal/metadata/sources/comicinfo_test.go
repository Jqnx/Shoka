package sources

import (
	"Shoka/internal/metadata"
	"reflect"
	"strings"
	"testing"
)

func TestComicInfoParseWeb(t *testing.T) {
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
