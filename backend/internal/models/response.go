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
	Language    *string                        `json:"language"`
	Category    *string                        `json:"category"`
	PageCount   int16                          `json:"pageCount"`
	FileHash    string                         `json:"fileHash"`
	Type        string                         `json:"type"`
	CreatedAt   time.Time                      `json:"createdAt"`
	UpdatedAt   time.Time                      `json:"updatedAt"`
	ReleaseDate *time.Time                     `json:"releaseDate"`
	PagesOnDisk int                            `json:"pagesOnDisk"`
	Tags        []repository.Tag               `json:"tags"`
	Artist      []repository.Artist            `json:"artists"`
	Parody      []repository.Parody            `json:"parodies"`
	Character   []repository.Character         `json:"characters"`
	URL         []repository.GetArchiveUrlsRow `json:"url"`
	Status      string                         `json:"readState"`
	Progress    int16                          `json:"progress"`
	LastRead    *time.Time                     `json:"lastRead"`
	IsFavorite  bool                           `json:"isFavorite"`
}

type ArchiveListResponse[T any] struct {
	Archives []T `json:"archives"`
	Count    int `json:"total"`
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
