package models

import (
	"time"
)

type Metadata struct {
	Title       string      `json:"title"`
	Summary     string      `json:"summary"`
	URL         []URL       `json:"urls"`
	Category    string      `json:"category"`
	Parody      []Parody    `json:"parodies"`
	Character   []Character `json:"characters"`
	Tags        []Tag       `json:"tags"`
	Artist      []Artist    `json:"artists"`
	Language    string      `json:"language"`
	ReleaseDate *time.Time  `json:"release_date"`
	PageCount   int         `json:"page_count"`
	NHID        int         `json:"nhid"`
	NHMediaID   string      `json:"nhmediaid"`
	NHImageType string      `json:"nhimagetype"`
}

type URL struct {
	URL string `json:"url"`
}

type Parody struct {
	Parody string `json:"parody"`
}

type Character struct {
	Character string `json:"character"`
}

type Tag struct {
	Tag string `json:"tag"`
}

type Artist struct {
	Artist string `json:"artist"`
}
