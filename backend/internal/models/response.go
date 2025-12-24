package models

import (
	"time"

	"Shoka/internal/repository"
)

type ResponseFail struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type Response struct {
	Status  string `json:"status"`
	Message any    `json:"message"`
}

type ArchiveResponse struct {
	ID          string                         `json:"id"`
	Title       string                         `json:"title"`
	Summary     *string                        `json:"summary"`
	Tags        []repository.Tag               `json:"tags"`
	Artist      []repository.Artist            `json:"artist"`
	Parody      []repository.Parody            `json:"parody"`
	Character   []repository.Character         `json:"character"`
	Language    *string                        `json:"language"`
	Category    *string                        `json:"category"`
	PageCount   int16                          `json:"page_count"`
	URL         []repository.GetArchiveUrlsRow `json:"url"`
	FileHash    string                         `json:"file_hash"`
	Pages       int                            `json:"pages"`
	CreatedAt   time.Time                      `json:"created_at"`
	UpdatedAt   time.Time                      `json:"updated_at"`
	ReleaseDate *time.Time                     `json:"release_date"`
}

type ArchiveResponseFavorite struct {
	ID          string                         `json:"id"`
	Title       string                         `json:"title"`
	Summary     *string                        `json:"summary"`
	Tags        []repository.Tag               `json:"tags"`
	Artist      []repository.Artist            `json:"artist"`
	Parody      []repository.Parody            `json:"parody"`
	Character   []repository.Character         `json:"character"`
	Language    *string                        `json:"language"`
	Category    *string                        `json:"category"`
	PageCount   int16                          `json:"page_count"`
	URL         []repository.GetArchiveUrlsRow `json:"url"`
	FileHash    string                         `json:"file_hash"`
	Pages       int                            `json:"pages"`
	Type        string                         `json:"type"`
	Status      string                         `json:"read_state"`
	Progress    int16                          `json:"progress"`
	LastRead    *time.Time                     `json:"last_read"`
	CreatedAt   time.Time                      `json:"created_at"`
	UpdatedAt   time.Time                      `json:"updated_at"`
	ReleaseDate *time.Time                     `json:"release_date"`
	IsFavorite  bool                           `json:"is_favorite"`
}

type ArtistResponse struct {
	Name    string                           `json:"name"`
	Aliases []repository.GetArtistAliasesRow `json:"aliases"`
	// Groups    []repository.GetArtistGroupsRow  `json:"groups"`
	URLs []repository.GetArtistUrlsRow `json:"links"`
}

//type GroupResponse struct {
//	Name      string                          `json:"name"`
//	Artists   []repository.GetGroupArtistsRow `json:"artists"`
//	CreatedAt time.Time                       `json:"created_at"`
//	UpdatedAt time.Time                       `json:"updated_at"`
//}

//type UserResponse struct {
//	ID        uuid.UUID `json:"user_id"`
//	Name      string    `json:"username"`
//	CreatedAt time.Time `json:"created_at"`
//}

type Page struct {
	Height int `json:"h"`
	Width  int `json:"w"`
}
