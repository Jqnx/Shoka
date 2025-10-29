package config

import "errors"

var (
	// Archives
	ErrArchiveNotFound     = errors.New("archive not found")
	ErrArchiveNoDuplicates = errors.New("archive already exists")
	ErrNoArchive           = errors.New("no archives found")

	// Artists
	ErrArtistNotFound     = errors.New("artist not found")
	ErrArtistNoDuplicates = errors.New("artist already exists")
	ErrNoArtist           = errors.New("no artists found")

	// Group
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupNoDuplicates = errors.New("group already exists")
	ErrNoGroup           = errors.New("no groups found")
	ErrGroupHasMembers   = errors.New("group still has members")

	// Tag
	ErrTagNotFound = errors.New("tag not found")
	ErrNoTags      = errors.New("no tags found")

	// Character
	ErrCharacterNotFound = errors.New("character not found")
	ErrNoCharacter       = errors.New("no characters found")

	// Parody
	ErrParodyNotFound = errors.New("parody not found")
	ErrNoParody       = errors.New("no parodies found")

	// Language
	ErrLanguageNotFound = errors.New("language not found")
	ErrNoLanguage       = errors.New("no languages found")

	// Category
	ErrCategoryNotFound = errors.New("category not found")
	ErrNoCategory       = errors.New("no categories found")

	// Database
	ErrNoDBHost     = errors.New("missing db host")
	ErrNoDBPort     = errors.New("missing db port")
	ErrNoDBDatabase = errors.New("missing db database")
	ErrNoDBUser     = errors.New("missing db user")
	ErrNoDBPassword = errors.New("missing db password")

	// Redis
	ErrNoRedisHost = errors.New("missing redis host")
	ErrNoRedisPort = errors.New("missing redis port")
)
