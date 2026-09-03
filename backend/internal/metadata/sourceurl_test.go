package metadata

import (
	"reflect"
	"testing"
)

func TestSourceFromURL(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantSource string
		wantOK     bool
	}{
		{"nhentai gallery", "https://nhentai.net/g/123/", "nhentai", true},
		{"nhentai with www", "https://www.nhentai.net/g/123/", "nhentai", true},
		{"e-hentai", "https://e-hentai.org/g/456/abc/", "e-hentai", true},
		{"exhentai maps to e-hentai", "https://exhentai.org/g/456/abc/", "e-hentai", true},
		{"http scheme allowed", "http://nhentai.net/g/1/", "nhentai", true},
		{"uppercase host", "https://NHentai.NET/g/1/", "nhentai", true},
		{"unknown host returns bare host", "https://example.com/x", "example.com", true},
		{"unknown host strips www", "https://www.example.com/x", "example.com", true},
		{"non-http scheme rejected", "ftp://nhentai.net/g/1/", "", false},
		{"mailto rejected", "mailto:someone@nhentai.net", "", false},
		{"garbage rejected", "not a url at all", "", false},
		{"empty rejected", "", "", false},
		{"scheme-relative rejected", "//nhentai.net/g/1/", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSource, gotOK := SourceFromURL(tt.raw)
			if gotSource != tt.wantSource || gotOK != tt.wantOK {
				t.Errorf("SourceFromURL(%q) = (%q, %v), want (%q, %v)",
					tt.raw, gotSource, gotOK, tt.wantSource, tt.wantOK)
			}
		})
	}
}

func TestNormalizeURLs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil stays nil", nil, nil},
		{"empty stays empty non-nil", []string{}, []string{}},
		{
			"trims, drops invalid, dedupes by url preserving order",
			[]string{
				"  https://nhentai.net/g/1/  ",
				"not-a-url",
				"https://e-hentai.org/g/2/x/",
				"https://nhentai.net/g/1/",
			},
			[]string{"https://nhentai.net/g/1/", "https://e-hentai.org/g/2/x/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeURLs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NormalizeURLs(%#v) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
