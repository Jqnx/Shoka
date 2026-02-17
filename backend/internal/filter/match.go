package filter

import (
	"context"

	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"

	"github.com/google/uuid"
)

type MatchAndGetParams struct {
	Filters  models.ArchiveFilters
	Page     int
	PageSize int
	Order    string
	UserID   uuid.UUID
}

func Match(in models.ArchiveFilters, qtx *repository.Queries) ([]string, error) {
	ctx := context.Background()

	idsByArtists, err := qtx.GetArchiveIDsByArtists(ctx, repository.GetArchiveIDsByArtistsParams{
		Artists: in.Artists,
		Amount:  int32(len(in.Artists)),
	})
	if err != nil {
		return nil, err
	}

	idsByCategory, err := qtx.GetArchiveIDsByCategories(ctx, repository.GetArchiveIDsByCategoriesParams{
		Categories: in.Categories,
		Amount:     int32(len(in.Categories)),
	})
	if err != nil {
		return nil, err
	}

	idsByCharacters, err := qtx.GetArchiveIDsByCharacters(ctx, repository.GetArchiveIDsByCharactersParams{
		Characters: in.Characters,
		Amount:     int32(len(in.Characters)),
	})
	if err != nil {
		return nil, err
	}

	idsByLanguages, err := qtx.GetArchiveIDsByLanguages(ctx, repository.GetArchiveIDsByLanguagesParams{
		Languages: in.Languages,
		Amount:    int32(len(in.Languages)),
	})
	if err != nil {
		return nil, err
	}

	idsByParodies, err := qtx.GetArchiveIDsByParodies(ctx, repository.GetArchiveIDsByParodiesParams{
		Parodies: in.Parodies,
		Amount:   int32(len(in.Parodies)),
	})
	if err != nil {
		return nil, err
	}

	idsByTags, err := qtx.GetArchiveIDsByTags(ctx, repository.GetArchiveIDsByTagsParams{
		Tags:   in.Tags,
		Amount: int32(len(in.Tags)),
	})
	if err != nil {
		return nil, err
	}

	match := util.Intersect(idsByArtists, idsByCategory, idsByCharacters, idsByLanguages, idsByParodies, idsByTags)
	return match, nil
}

func MatchAndGet(qtx *repository.Queries, arg MatchAndGetParams) (*models.ArchiveListResponse[repository.GetArchiveFilterSortListRow], error) {
	ctx := context.Background()
	list, err := Match(arg.Filters, qtx)
	if err != nil {
		return nil, err
	}

	archives, err := qtx.GetArchiveFilterSortList(ctx, repository.GetArchiveFilterSortListParams{
		Uid:     arg.UserID,
		Ids:     list,
		OrderBy: arg.Order,
		Limit:   int32(arg.PageSize),
		Offset:  (int32(arg.Page) - 1) * int32(arg.PageSize),
	})
	if err != nil {
		return nil, err
	}

	return &models.ArchiveListResponse[repository.GetArchiveFilterSortListRow]{
		Archives: archives,
		Count:    len(list),
	}, nil
}

func MatchAndGetFavorites(qtx *repository.Queries, arg MatchAndGetParams) (*models.ArchiveListResponse[repository.GetFavoriteArchiveFilterSortListRow], error) {
	ctx := context.Background()
	list, err := Match(arg.Filters, qtx)
	if err != nil {
		return nil, err
	}

	archives, err := qtx.GetFavoriteArchiveFilterSortList(ctx, repository.GetFavoriteArchiveFilterSortListParams{
		UserID:  arg.UserID,
		Ids:     list,
		OrderBy: arg.Order,
		Limit:   int32(arg.PageSize),
		Offset:  (int32(arg.Page) - 1) * int32(arg.PageSize),
	})
	if err != nil {
		return nil, err
	}

	count, err := qtx.CountFavoriteFilteredArchive(ctx, repository.CountFavoriteFilteredArchiveParams{
		Ids:    list,
		UserID: arg.UserID,
	})
	if err != nil {
		return nil, err
	}

	results := &models.ArchiveListResponse[repository.GetFavoriteArchiveFilterSortListRow]{
		Archives: archives,
		Count:    int(count),
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
		archive, err := qtx.GetFavoriteArchiveFilter(ctx, repository.GetFavoriteArchiveFilterParams{
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
		archive, err := qtx.GetArchiveFilter(ctx, repository.GetArchiveFilterParams{
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
