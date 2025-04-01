package models

type ArchivePayload struct {
	Title     string   `json:"title" form:"title" binding:"required"`
	Summary   string   `json:"summary" form:"summary"`
	Tags      []string `json:"tags" form:"tags"`
	Artist    []string `json:"artist" form:"artist"`
	Parody    []string `json:"parody" form:"parody"`
	Character []string `json:"character" form:"character"`
	Language  string   `json:"language" form:"language" binding:"required,bcp47_language_tag"`
	Category  string   `json:"category" form:"category"`
	URL       []string `json:"url" form:"url"`
	FilePath  string   `json:"filepath" form:"filepath" binding:"required"`
}

type ArtistPayload struct {
	Name    string   `json:"name" binding:"required"`
	Aliases []string `json:"aliases"`
	Group   []string `json:"groups"`
	Links   []string `json:"links"`
}
