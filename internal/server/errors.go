package server

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	// Archives
	ErrArchiveNotFound     = errors.New("archive not found")
	ErrArchiveNoDuplicates = errors.New("archive already exists")

	// Artists
	ErrArtistNotFound     = errors.New("artist not found")
	ErrArtistNoDuplicates = errors.New("artist already exists")

	// Group
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupNoDuplicates = errors.New("group already exists")
)

func Validate(verr validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)

	for _, f := range verr {
		switch f.ActualTag() {
		case "bcp47_language_tag":
			err := "Invalid language tag."
			errs[f.Field()] = err
			return errs
		case "required":
			err := fmt.Sprintf("%s is required.", f.Field())
			errs[f.Field()] = err
			return errs
		}
	}
	return errs
}
