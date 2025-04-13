package archive

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
)

// TODO:

func Get(c context.Context, q *repository.Queries, id int64, log *slog.Logger) (*models.ArchiveResponse, error) {
	archive, err := q.GetArchiveByAID(c, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	tags, err := q.GetArchiveTags(c, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	characters, err := q.GetArchiveCharacters(c, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	parodies, err := q.GetArchiveParodies(c, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	urls, err := q.GetArchiveURLs(c, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	artists, err := q.GetArchiveArtists(c, id)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	result := &models.ArchiveResponse{
		AID:       id,
		Title:     archive.Title,
		Summary:   *archive.Summary,
		Tags:      tags,
		Artist:    artists,
		Parody:    parodies,
		Character: characters,
		Language:  *archive.Lang,
		Category:  *archive.Category,
		Url:       urls,
		CreatedAt: archive.CreatedAt,
		UpdatedAt: archive.UpdatedAt,
	}
	return result, nil
}

func GetAll(c context.Context, q *repository.Queries, log *slog.Logger) (*[]models.ArchiveResponse, error) {
	archives, err := q.GetAllArchives(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result := []models.ArchiveResponse{}

	for _, item := range archives {
		tags, err := q.GetArchiveTags(c, item.AID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		characters, err := q.GetArchiveCharacters(c, item.AID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		parodies, err := q.GetArchiveParodies(c, item.AID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		urls, err := q.GetArchiveURLs(c, item.AID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		artists, err := q.GetArchiveArtists(c, item.AID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		archive := models.ArchiveResponse{
			AID:       item.AID,
			Title:     item.Title,
			Summary:   *item.Summary,
			Tags:      tags,
			Artist:    artists,
			Parody:    parodies,
			Character: characters,
			Language:  *item.Lang,
			Category:  *item.Category,
			Url:       urls,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}

		result = append(result, archive)
	}
	return &result, nil
}

func GetByTag(c context.Context, q *repository.Queries, tag string, log *slog.Logger) ([]int64, error) {
	archives, err := q.GetArchivesByTag(c, tag)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return archives, nil
}

func GetByCharacter(c context.Context, q *repository.Queries, character string, log *slog.Logger) ([]int64, error) {
	archives, err := q.GetArchivesByCharacter(c, character)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return archives, nil
}

func GetByParody(c context.Context, q *repository.Queries, parody string, log *slog.Logger) ([]int64, error) {
	archives, err := q.GetArchivesByParody(c, parody)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return archives, nil
}
