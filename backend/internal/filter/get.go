package filter

import (
	"strings"

	"Shoka/internal/models"

	"github.com/gin-gonic/gin"
)

func GetFromQuery(c *gin.Context) (*models.ArchiveFilters, bool) {
	hasFilter := false

	artists := c.Query("artists")
	var artistsSlice []string
	if artists != "" {
		artistsSlice = strings.Split(artists, ",")
		hasFilter = true
	}

	categories := c.Query("categories")
	var categoriesSlice []string
	if categories != "" {
		categoriesSlice = strings.Split(categories, ",")
		hasFilter = true
	}

	characters := c.Query("characters")
	var charactersSlice []string
	if characters != "" {
		charactersSlice = strings.Split(characters, ",")
		hasFilter = true
	}

	languages := c.Query("languages")
	var languagesSlice []string
	if languages != "" {
		languagesSlice = strings.Split(languages, ",")
		hasFilter = true
	}

	parodies := c.Query("parodies")
	var parodiesSlice []string
	if parodies != "" {
		parodiesSlice = strings.Split(parodies, ",")
		hasFilter = true
	}
	tags := c.Query("tags")
	var tagsSlice []string
	if tags != "" {
		tagsSlice = strings.Split(tags, ",")
		hasFilter = true
	}

	return &models.ArchiveFilters{
		Artists:    artistsSlice,
		Categories: categoriesSlice,
		Characters: charactersSlice,
		Languages:  languagesSlice,
		Parodies:   parodiesSlice,
		Tags:       tagsSlice,
	}, hasFilter
}
