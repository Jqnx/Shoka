package models

import "time"

type ResponseSuccess struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type ResponseFail struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type ResponseError struct {
	Status  string `json:"status"`
	Message any    `json:"message"`
}

type ArchiveResponse struct {
	AID       int64
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Tags      []string  `json:"tags"`
	Artist    []string  `json:"artist"`
	Parody    []string  `json:"parody"`
	Character []string  `json:"character"`
	Language  string    `json:"language"`
	Category  string    `json:"category"`
	Url       []string  `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ArtistResponse struct {
	Name      string    `json:"name"`
	Aliases   []string  `json:"aliases"`
	Groups    []string  `json:"groups"`
	Links     []string  `json:"links"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupResponse struct {
	Name      string    `json:"name"`
	Artists   []string  `json:"artists"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
