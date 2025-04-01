package artist

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// TODO:

func CreateTransaction(ctx context.Context,
	db *pgx.Conn,
	queries *repository.Queries,
	payload *models.ArtistPayload,
) (*repository.Artist, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	qtx := queries.WithTx(tx)
	artist, err := qtx.CreateArtist(ctx, repository.CreateArtistParams{
		Name:      payload.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range payload.Aliases {
		i := strings.ToLower(item)
		exists, err := qtx.ArtistAliasExists(ctx, i)
		if err != nil {
			return nil, err
		}
		if exists.RowsAffected() == 0 {
			if err := qtx.CreateAlias(ctx, repository.CreateAliasParams{
				Alias:    i,
				ArtistID: artist.ID,
			}); err != nil {
				return nil, err
			}
		}
	}

	for _, item := range payload.Links {
		exists, err := qtx.ArtistLinkExists(ctx, item)
		if err != nil {
			return nil, err
		}
		if exists.RowsAffected() == 0 {
			if err := qtx.CreateArtistLink(ctx, repository.CreateArtistLinkParams{
				Link:     item,
				ArtistID: artist.ID,
			}); err != nil {
				return nil, err
			}
		}
	}

	for _, item := range payload.Group {
		i := strings.ToLower(item)
		group, err := qtx.GetGroup(ctx, i)
		if err != nil {
			return nil, err
		}
		if group.Name == i {
			if err := qtx.AddArtistToGroup(ctx, repository.AddArtistToGroupParams{
				ArtistID: artist.ID,
				GroupID:  group.ID,
			}); err != nil {
				return nil, err
			}
		} else {
			group, err := qtx.CreateGroup(ctx, i)
			if err != nil {
				return nil, err
			}

			if err := qtx.AddArtistToGroup(ctx, repository.AddArtistToGroupParams{
				ArtistID: artist.ID,
				GroupID:  group.ID,
			}); err != nil {
				return nil, err
			}
		}
	}

	return &artist, tx.Commit(ctx)
}
