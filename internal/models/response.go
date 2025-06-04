package models

import (
	"Shoka/internal/repository"
	"time"
)

type ResponseFail struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type ResponseError struct {
	Status  string `json:"status"`
	Message any    `json:"message"`
}

type ArchiveResponse struct {
	ArchiveID   string                            `json:"archive_id"`
	Title       string                            `json:"title"`
	Summary     *string                           `json:"summary"`
	Tags        []repository.Tag                  `json:"tags"`
	Artist      []repository.GetArchiveArtistsRow `json:"artist"`
	Parody      []repository.Parody               `json:"parody"`
	Character   []repository.Character            `json:"character"`
	Language    *string                           `json:"language"`
	Category    *string                           `json:"category"`
	PageCount   int64                             `json:"page_count"`
	Url         []repository.GetArchiveURLsRow    `json:"url"`
	Hash        string                            `json:"hash"`
	Pages       int                               `json:"pages"`
	Type        string                            `json:"type"`
	CreatedAt   time.Time                         `json:"created_at"`
	UpdatedAt   time.Time                         `json:"updated_at"`
	ReleaseDate *time.Time                        `json:"release_date"`
}

type ArtistResponse struct {
	Name      string                           `json:"name"`
	Aliases   []repository.GetArtistAliasesRow `json:"aliases"`
	Groups    []repository.GetArtistGroupsRow  `json:"groups"`
	Links     []repository.GetArtistLinksRow   `json:"links"`
	CreatedAt time.Time                        `json:"created_at"`
	UpdatedAt time.Time                        `json:"updated_at"`
}

type GroupResponse struct {
	Name      string                          `json:"name"`
	Artists   []repository.GetGroupArtistsRow `json:"artists"`
	CreatedAt time.Time                       `json:"created_at"`
	UpdatedAt time.Time                       `json:"updated_at"`
}

type Page struct {
	Height int `json:"h"`
	Width  int `json:"w"`
}
