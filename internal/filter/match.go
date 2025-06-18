package filter

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"
)

func matchArtistsTags(in models.ArchiveFilters, qtx *repository.Queries) ([]string, error) {
	tags, err := fetchTags(in.Tags, qtx)
	if err != nil {
		return nil, err
	}
	artists, err := fetchArtists(in.Artists, qtx)
	if err != nil {
		return nil, err
	}

	matched := util.MatchStringsInSlices(tags, artists)
	return matched, nil
}

func matchCharactersParodies(in models.ArchiveFilters, qtx *repository.Queries) ([]string, error) {
	characters, err := fetchCharacters(in.Characters, qtx)
	if err != nil {
		return nil, err
	}
	parodies, err := fetchParodies(in.Parodies, qtx)
	if err != nil {
		return nil, err
	}

	matched := util.MatchStringsInSlices(characters, parodies)
	return matched, nil
}

func matchLanguagesCategories(in models.ArchiveFilters, qtx *repository.Queries) ([]string, error) {
	languages, err := fetchLanguages(in.Languages, qtx)
	if err != nil {
		return nil, err
	}
	categories, err := fetchCategories(in.Categories, qtx)
	if err != nil {
		return nil, err
	}

	matched := util.MatchStringsInSlices(languages, categories)
	return matched, nil
}

func Match(in models.ArchiveFilters, qtx *repository.Queries) ([]string, error) {
	artistsTags, err := matchArtistsTags(in, qtx)
	if err != nil {
		return nil, err
	}
	charactersParodies, err := matchCharactersParodies(in, qtx)
	if err != nil {
		return nil, err
	}
	languagesCategories, err := matchLanguagesCategories(in, qtx)
	if err != nil {
		return nil, err
	}

	aTcP := util.MatchStringsInSlices(artistsTags, charactersParodies)
	match := util.MatchStringsInSlices(aTcP, languagesCategories)
	return match, nil
}
