package artist

import (
	"context"
	"strings"

	"Shoka/internal/models"
	"Shoka/internal/repository"
)

func Alias(c context.Context, q *repository.Queries, p *models.ArtistPayload, artist *repository.Artist) error {
	if err := q.RemoveArtistAliases(c, artist.ID); err != nil {
		return err
	}
	for _, item := range p.Aliases {
		i := strings.ToLower(item)
		exists, err := q.ArtistAliasExists(c, i)
		if err != nil {
			return err
		}
		if exists.RowsAffected() == 0 {
			if err := q.CreateAlias(c, repository.CreateAliasParams{
				Alias:    i,
				ArtistID: artist.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func Link(c context.Context, q *repository.Queries, p *models.ArtistPayload, artist *repository.Artist) error {
	if err := q.RemoveArtistUrls(c, artist.ID); err != nil {
		return err
	}
	for _, item := range p.Links {
		exists, err := q.ArtistUrlExists(c, item)
		if err != nil {
			return err
		}
		if exists.RowsAffected() == 0 {
			if err := q.CreateArtistUrl(c, repository.CreateArtistUrlParams{
				Url:      item,
				ArtistID: artist.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

//func Group(c context.Context, q *repository.Queries, p *models.ArtistPayload, artist *repository.Artist) error {
//	if err := q.RemoveArtistFromGroup(c, artist.ID); err != nil {
//		return err
//	}
//	for _, item := range p.Group {
//		i := strings.ToLower(item)
//		group, _ := q.GetGroup(c, i)
//
//		if group.Name == i {
//			if err := q.AddArtistToGroup(c, repository.AddArtistToGroupParams{
//				ArtistID: artist.ID,
//				GroupID:  group.ID,
//			}); err != nil {
//				return err
//			}
//		} else {
//			group, err := q.CreateGroup(c, repository.CreateGroupParams{
//				Name:      i,
//				CreatedAt: time.Now(),
//				UpdatedAt: time.Now(),
//			})
//			if err != nil {
//				return err
//			}
//
//			if err := q.AddArtistToGroup(c, repository.AddArtistToGroupParams{
//				ArtistID: artist.ID,
//				GroupID:  group.ID,
//			}); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
