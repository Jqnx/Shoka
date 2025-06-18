package filter

import (
	"Shoka/internal/repository"
	"context"
)

func fetchTags(in []string, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()
	var tags []string
	var err error
	for _, i := range in {
		tags, err = qtx.GetArchiveIDsByTag(ctx, i)
		if err != nil {
			return nil, err
		}
	}
	return tags, nil
}

func fetchArtists(in []string, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()
	var artists []string
	var err error
	for _, i := range in {
		artists, err = qtx.GetArchiveIDsByArtist(ctx, i)
		if err != nil {
			return nil, err
		}
	}
	return artists, nil
}

func fetchCharacters(in []string, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()
	var characters []string
	var err error
	for _, i := range in {
		characters, err = qtx.GetArchiveIDsByCharacter(ctx, i)
		if err != nil {
			return nil, err
		}
	}
	return characters, nil
}

func fetchParodies(in []string, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()
	var parodies []string
	var err error
	for _, i := range in {
		parodies, err = qtx.GetArchiveIDsByParody(ctx, i)
		if err != nil {
			return nil, err
		}
	}
	return parodies, nil
}

func fetchLanguages(in []string, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()
	var languages []string
	var err error
	for _, i := range in {
		languages, err = qtx.GetArchiveIDsByLanguage(ctx, &i)
		if err != nil {
			return nil, err
		}
	}
	return languages, nil
}

func fetchCategories(in []string, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()
	var categories []string
	var err error
	for _, i := range in {
		categories, err = qtx.GetArchiveIDsByCategory(ctx, &i)
		if err != nil {
			return nil, err
		}
	}
	return categories, nil
}
