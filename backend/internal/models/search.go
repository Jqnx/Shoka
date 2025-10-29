package models

import "time"

type AIDSearch struct {
	AID int `json:"archive_id"`
}

type TagSearch struct {
	Name string
}

type ArchiveSearch struct {
	AID        int
	CreatedAt  time.Time
	Title      string
	Summary    string
	Tags       []string
	Artists    []string
	Parodies   []string
	Characters []string
	Language   string
	Category   string
	Urls       []string
	PageCount  int
}

type ArtistSearch struct {
	Name    string
	Aliases []string
	Groups  []string
	Links   []string
}
