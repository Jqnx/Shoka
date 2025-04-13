package artist

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
)

// TODO:

func GetAll(c context.Context, q *repository.Queries, log *slog.Logger) (*[]models.ArtistResponse, error) {
	artists, err := q.GetAllArtists(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result := []models.ArtistResponse{}

	for _, item := range artists {
		aliases, err := q.GetArtistAliases(c, item.Name)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		links, err := q.GetArtistLinks(c, item.Name)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		groups, err := q.GetArtistGroups(c, item.Name)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		artist := models.ArtistResponse{
			Name:      item.Name,
			Aliases:   aliases,
			Groups:    groups,
			Links:     links,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}

		result = append(result, artist)
	}
	return &result, nil
}

func Get(c context.Context, q *repository.Queries, name string, log *slog.Logger) (*models.ArtistResponse, error) {
	artist, err := q.GetArtistByName(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	links, err := q.GetArtistLinks(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	groups, err := q.GetArtistGroups(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	aliases, err := q.GetArtistAliases(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result := &models.ArtistResponse{
		Name:      artist.Name,
		Aliases:   aliases,
		Groups:    groups,
		Links:     links,
		CreatedAt: artist.CreatedAt,
		UpdatedAt: artist.UpdatedAt,
	}
	return result, nil
}
