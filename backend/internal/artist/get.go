package artist

import (
	"context"

	"Shoka/internal/logger"
	"Shoka/internal/models"
	"Shoka/internal/repository"
)

// TODO:

func GetAll(c context.Context, q *repository.Queries, log logger.Logger) (*[]models.ArtistResponse, error) {
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
		links, err := q.GetArtistUrls(c, item.Name)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		artist := models.ArtistResponse{
			Name:    item.Name,
			Aliases: aliases,
			URLs:    links,
		}

		result = append(result, artist)
	}
	return &result, nil
}

func Get(c context.Context, q *repository.Queries, name string, log logger.Logger) (*models.ArtistResponse, error) {
	artist, err := q.GetArtistByName(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	links, err := q.GetArtistUrls(c, name)
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
		Name:    artist.Name,
		Aliases: aliases,
		URLs:    links,
	}
	return result, nil
}
