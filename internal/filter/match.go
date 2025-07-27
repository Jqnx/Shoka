package filter

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"
	"context"

	"github.com/google/uuid"
)

type ArchiveList struct {
	Archives []repository.GetArchivesFilterSortListRow `json:"archives"`
	Count    int                                       `json:"total"`
}

type FavoriteArchiveList struct {
	Archives []repository.GetFavoriteArchivesFilterSortListRow `json:"archives"`
	Count    int64                                             `json:"total"`
}

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

func MatchAndGet(in models.ArchiveFilters, qtx *repository.Queries, page, pageSize int, order string) (*ArchiveList, error) {
	ctx := context.Background()
	list, err := Match(in, qtx)
	if err != nil {
		return nil, err
	}

	archives, err := qtx.GetArchivesFilterSortList(ctx, repository.GetArchivesFilterSortListParams{
		Ids:     list,
		OrderBy: order,
		Limit:   int32(pageSize),
		Offset:  (int32(page) - 1) * int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	results := &ArchiveList{
		Archives: archives,
		Count:    len(list),
	}
	return results, nil
}

func MatchAndGetFavorites(in models.ArchiveFilters, qtx *repository.Queries, page, pageSize int, userid uuid.UUID, order string) (*FavoriteArchiveList, error) {
	ctx := context.Background()
	list, err := Match(in, qtx)
	if err != nil {
		return nil, err
	}

	archives, err := qtx.GetFavoriteArchivesFilterSortList(ctx, repository.GetFavoriteArchivesFilterSortListParams{
		UserID:  userid,
		Ids:     list,
		OrderBy: order,
		Limit:   int32(pageSize),
		Offset:  (int32(page) - 1) * int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	count, err := qtx.CountFavoriteFilteredArchives(ctx, repository.CountFavoriteFilteredArchivesParams{
		Ids:    list,
		UserID: userid,
	})
	if err != nil {
		return nil, err
	}

	results := &FavoriteArchiveList{
		Archives: archives,
		Count:    count,
	}

	return results, nil
}

func MatchAndGetShuffle(in models.ArchiveFilters, qtx *repository.Queries, rng int, userid uuid.UUID, favorite bool) (string, error) {
	ctx := context.Background()
	list, err := Match(in, qtx)
	if err != nil {
		return "", err
	}

	if favorite {
		archive, err := qtx.GetFavoriteArchivesFilter(ctx, repository.GetFavoriteArchivesFilterParams{
			UserID: userid,
			Ids:    list,
			Limit:  1,
			Offset: int32(rng),
		})
		if err != nil {
			return "", err
		}

		return archive, nil
	} else {
		archive, err := qtx.GetArchivesFilter(ctx, repository.GetArchivesFilterParams{
			Ids:    list,
			Limit:  1,
			Offset: int32(rng),
		})
		if err != nil {
			return "", err
		}

		return archive, nil
	}
}
