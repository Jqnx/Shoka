// Package language contains logic/utility used for converting between
// full-length language names and their respective ISO-639-1 codes
package language

import (
	"fmt"
	"strings"
)

type LanguageConverter struct {
	nameToCode map[string]string
	codeToName map[string]string
}

func NewLanguageConverter() *LanguageConverter {
	nameToCode := map[string]string{
		"english":    "en",
		"spanish":    "es",
		"french":     "fr",
		"german":     "de",
		"italian":    "it",
		"portuguese": "pt",
		"russian":    "ru",
		"chinese":    "zh",
		"japanese":   "ja",
		"korean":     "ko",
		"arabic":     "ar",
		"hindi":      "hi",
		"dutch":      "nl",
		"swedish":    "sv",
		"norwegian":  "no",
		"danish":     "da",
		"finnish":    "fi",
		"polish":     "pl",
		"turkish":    "tr",
		"greek":      "el",
		"hebrew":     "he",
		"thai":       "th",
		"vietnamese": "vi",
		"czech":      "cs",
		"hungarian":  "hu",
		"romanian":   "ro",
		"bulgarian":  "bg",
		"croatian":   "hr",
		"slovak":     "sk",
		"slovenian":  "sl",
		"estonian":   "et",
		"latvian":    "lv",
		"lithuanian": "lt",
		"ukrainian":  "uk",
		"serbian":    "sr",
		"bosnian":    "bs",
		"macedonian": "mk",
		"albanian":   "sq",
		"maltese":    "mt",
		"icelandic":  "is",
		"irish":      "ga",
		"welsh":      "cy",
		"catalan":    "ca",
		"basque":     "eu",
		"galician":   "gl",
	}

	// Create reverse mapping
	codeToName := make(map[string]string)
	for name, code := range nameToCode {
		codeToName[code] = name
	}

	return &LanguageConverter{
		nameToCode: nameToCode,
		codeToName: codeToName,
	}
}

func (lc *LanguageConverter) ToISO(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty language input")
	}

	normalized := strings.ToLower(strings.TrimSpace(name))

	if len(normalized) == 2 {
		if _, exists := lc.codeToName[normalized]; exists {
			return normalized, nil
		}
		return "", fmt.Errorf("unknown ISO code: %s", name)
	}

	if code, exists := lc.nameToCode[normalized]; exists {
		return code, nil
	}

	return "", fmt.Errorf("unknown language: %s", name)
}

func (lc *LanguageConverter) ToName(code string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(code))
	if name, exists := lc.codeToName[normalized]; exists {
		return name, nil
	}
	return "", fmt.Errorf("unknown ISO code: %s", code)
}

func (lc *LanguageConverter) IsValidISO(code string) bool {
	normalized := strings.ToLower(strings.TrimSpace(code))
	_, exists := lc.codeToName[normalized]
	return exists
}

func (lc *LanguageConverter) IsValidName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	_, exists := lc.nameToCode[normalized]
	return exists
}
