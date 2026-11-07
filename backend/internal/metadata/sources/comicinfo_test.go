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

func TestComicInfoCirclesRoundTrip(t *testing.T) {
	t.Parallel()

	original := &metadata.Result{Circles: []string{"circle a", "circle b"}}

	data, err := MarshalComicInfo(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(data), "<Publisher>circle a, circle b</Publisher>") {
		t.Errorf("marshalled ComicInfo missing joined Publisher element:\n%s", data)
	}

	parsed, err := NewComicInfoSource().parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if !reflect.DeepEqual(parsed.Circles, original.Circles) {
		t.Errorf("round-trip Circles = %#v, want %#v", parsed.Circles, original.Circles)
	}
}

func TestComicInfoMeaninglessRelationsStayNil(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0"?>
<ComicInfo>
  <Writer>Unknown</Writer>
  <Tags> , unknown,   </Tags>
  <Characters>UNKNOWN</Characters>
  <Series> Unknown </Series>
  <Publisher>unknown</Publisher>
</ComicInfo>`

	result, err := NewComicInfoSource().parse([]byte(xml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Nil, not an empty slice: ApplyMetadata reads a non-nil empty slice as
	// "clear this relation".
	if result.Artists != nil {
		t.Errorf("Artists = %#v, want nil", result.Artists)
	}

	if result.Tags != nil {
		t.Errorf("Tags = %#v, want nil", result.Tags)
	}

	if result.Characters != nil {
		t.Errorf("Characters = %#v, want nil", result.Characters)
	}

	if result.Parodies != nil {
		t.Errorf("Parodies = %#v, want nil", result.Parodies)
	}

	if result.Circles != nil {
		t.Errorf("Circles = %#v, want nil", result.Circles)
	}
}

func TestComicInfoAbsentRelationsStayNil(t *testing.T) {
	t.Parallel()

	result, err := NewComicInfoSource().parse([]byte(`<ComicInfo><Title>Example</Title></ComicInfo>`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	for field, got := range relationFields(result) {
		if got != nil {
			t.Errorf("%s = %#v, want nil", field, got)
		}
	}
}

// relationFields exposes the parsed relation slices by field name so a test
// can assert one is populated while the rest stay nil.
func relationFields(r *metadata.Result) map[string][]string {
	return map[string][]string{
		"Artists":    r.Artists,
		"Tags":       r.Tags,
		"Characters": r.Characters,
		"Parodies":   r.Parodies,
		"Circles":    r.Circles,
	}
}

func TestComicInfoPartiallyMeaninglessRelationsKeepTheRest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		xml  string
		key  string
		want []string
	}{
		{"Writer", `<ComicInfo><Writer>Alice, Unknown</Writer></ComicInfo>`, "Artists", []string{"alice"}},
		{"Tags", `<ComicInfo><Tags>tag, unknown</Tags></ComicInfo>`, "Tags", []string{"tag"}},
		{"Characters", `<ComicInfo><Characters>char, UNKNOWN</Characters></ComicInfo>`, "Characters", []string{"char"}},
		{"Series", `<ComicInfo><Series>series, unknown</Series></ComicInfo>`, "Parodies", []string{"series"}},
		{"Publisher", `<ComicInfo><Publisher>circle, unknown</Publisher></ComicInfo>`, "Circles", []string{"circle"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := NewComicInfoSource().parse([]byte(tt.xml))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			fields := relationFields(result)
			if got := fields[tt.key]; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %#v, want %#v", tt.key, got, tt.want)
			}

			for field, got := range fields {
				if field != tt.key && got != nil {
					t.Errorf("%s = %#v, want nil", field, got)
				}
			}
		})
	}
}
