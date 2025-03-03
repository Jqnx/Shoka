package server

import (
	"fmt"

	"github.com/go-playground/validator/v10"
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
