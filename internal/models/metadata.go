package models

import (
	"time"
)

type Metadata struct {
	Title       string
	Summary     string
	URL         []URL
	Category    string
	Parody      []Parody
	Character   []Character
	Tags        []Tag
	Artist      []Artist
	Language    string
	ReleaseDate *time.Time
	PageCount   int
	NHID        int
	NHMediaID   string
	NHImageType string
}

type URL struct {
	URL string
}

type Parody struct {
	Parody string
}

type Character struct {
	Character string
}

type Tag struct {
	Tag string
}

type Artist struct {
	Artist string
}
