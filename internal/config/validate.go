package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func Validate(verr validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)

	for _, f := range verr {
		switch f.ActualTag() {
		case "bcp47_language_tag":
			field := strings.ToLower(f.Field())
			err := "invalid language tag."
			errs[field] = err
			return errs
		case "required":
			field := strings.ToLower(f.Field())
			err := fmt.Sprintf("%s is required.", field)
			errs[field] = err
			return errs
		}
	}
	return errs
}

// ValidateGroupArtists checks for empty string in slice and filters it out, returns a new slice
func ValidateGroupArtists(s []string) []string {
	result := []string{}
	for _, i := range s {
		if i != "" {
			result = append(result, i)
		}
	}
	return result
}
