package artist

import (
	"context"
	"time"

	"Shoka/internal/repository"
)

type GetArtist struct {
	Name      string    `json:"name"`
	Aliases   []string  `json:"aliases"`
	Groups    []string  `json:"groups"`
	Links     []string  `json:"links"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetAll(c context.Context, q *repository.Queries) (*[]GetArtist, error) {
	artists, err := q.GetAllArtists(c)
	if err != nil {
		return nil, err
	}

	result := []GetArtist{}

	for _, item := range artists {
		aliases, err := q.GetArtistAliases(c, item.Name)
		if err != nil {
			return nil, err
		}
		links, err := q.GetArtistLinks(c, item.Name)
		if err != nil {
			return nil, err
		}
		groups, err := q.GetArtistGroups(c, item.Name)
		if err != nil {
			return nil, err
		}

		artist := GetArtist{
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

func Get(c context.Context, q *repository.Queries, name string) (*GetArtist, error) {
	artist, err := q.GetArtistByName(c, name)
	if err != nil {
		return nil, err
	}
	links, err := q.GetArtistLinks(c, name)
	if err != nil {
		return nil, err
	}
	groups, err := q.GetArtistGroups(c, name)
	if err != nil {
		return nil, err
	}
	aliases, err := q.GetArtistAliases(c, name)
	if err != nil {
		return nil, err
	}

	result := GetArtist{
		Name:      artist.Name,
		Aliases:   aliases,
		Groups:    groups,
		Links:     links,
		CreatedAt: artist.CreatedAt,
		UpdatedAt: artist.UpdatedAt,
	}
	return &result, nil
}
