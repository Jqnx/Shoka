package models

import (
	"time"

	"gorm.io/gorm"
)

// Database models
type Archive struct {
	gorm.Model
	Title     string      `json:"title" gorm:"unique"`
	Summary   string      `json:"summary"`
	Tags      []Tag       `json:"tags" gorm:"many2many:archive_tags;References:Name;constraint:OnUpdate:CASCADE"`
	Artist    []Artist    `json:"artists" gorm:"many2many:archive_artists;references:Name"`
	Parody    []Parody    `json:"parodies" gorm:"many2many:archive_parody;references:Name"`
	Character []Character `json:"characters" gorm:"many2many:archive_chars;references:Name"`
	Language  string      `json:"language"`
	Category  string      `json:"category"`
	PageCount int         `json:"pagecount"`
	URL       []URL       `json:"urls" gorm:"foreignKey:UrlID"`
	FilePath  string      `json:"filepath" gorm:"unique"`
	AID       int         `gorm:"uniqueIndex;default:1"`
}

type Tag struct {
	Name string `gorm:"many2many:archive_tags;primaryKey;not null" json:"tag" form:"tag"` // Tag name
}

type Character struct {
	Name string `gorm:"many2many:archive_chars;primaryKey;not null" json:"character" form:"character"` // Character name
}

type Parody struct {
	Name string `gorm:"many2many:archive_parody;primaryKey;not null" json:"parody" form:"parody"` // Series name
}

type URL struct {
	UrlID uint   `gorm:"primaryKey"`
	Url   string `json:"url" form:"url"`
}

type Artist struct {
	gorm.Model
	Name  string  `gorm:"many2many:archive_artists;uniqueIndex;not null"`
	Alias []Alias `gorm:"foreignKey:AliasID" json:"aliases" form:"aliases"`
	Group []Group `gorm:"many2many:artist_groups;references:Name" json:"groups" form:"groups"`
	Links []Links `gorm:"foreignKey:LinkID" json:"links" form:"links"`
}

type Alias struct {
	AliasID uint   `gorm:"primaryKey"`
	Alias   string `gorm:"uniqueIndex" json:"alias" form:"alias"`
}

type Links struct {
	LinkID uint   `gorm:"primaryKey"`
	Link   string `gorm:"unique" json:"link" form:"link"`
}

type Group struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"many2many:artist_groups;uniqueIndex;not null" json:"group" form:"group"`
}

// Metadata translation models
type ComicInfoXML struct {
	Title      string `xml:"Title"`
	Summary    string `xml:"Summary"`
	PageCount  int    `xml:"PageCount"`
	URL        string `xml:"URL"`
	Genre      string `xml:"Genre"`
	Series     string `xml:"Series"`
	Characters string `xml:"Characters"`
	Tags       string `xml:"Tags"`
	Writer     string `xml:"Writer"`
	Language   string `xml:"LanguageISO"`
}

// Search models
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
