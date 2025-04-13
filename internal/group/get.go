package group

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
)

func Get(c context.Context, q *repository.Queries, name string, log *slog.Logger) (*models.GroupResponse, error) {
	group, err := q.GetGroup(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	artists, err := q.GetGroupArtists(c, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result := &models.GroupResponse{
		Name:      group.Name,
		Artists:   artists,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
	return result, nil
}

func GetAll(c context.Context, q *repository.Queries, log *slog.Logger) (*[]models.GroupResponse, error) {
	groups, err := q.GetAllGroups(c)
	if err != nil {
		return nil, err
	}

	result := []models.GroupResponse{}

	for _, group := range groups {
		artists, err := q.GetGroupArtists(c, group.Name)
		if err != nil {
			return nil, err
		}

		group := &models.GroupResponse{
			Name:    group.Name,
			Artists: artists,
		}

		result = append(result, *group)
	}
	return &result, nil
}
