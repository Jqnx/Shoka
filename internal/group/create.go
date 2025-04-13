package group

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func CreateTransaction(c context.Context,
	db *pgx.Conn,
	q *repository.Queries,
	p *models.GroupPayload,
	log *slog.Logger,
) (*models.GroupResponse, error) {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)

	name := strings.ToLower(p.Name)
	group, err := qtx.CreateGroup(c, repository.CreateGroupParams{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	for _, i := range p.Artists {
		artistName := strings.ToLower(i)
		artist, _ := qtx.GetArtistByName(c, artistName)
		if artist.Name == artistName {
			if err := qtx.AddArtistToGroup(c, repository.AddArtistToGroupParams{
				ArtistID: artist.ID,
				GroupID:  group.ID,
			}); err != nil {
				log.Error(err.Error())
				return nil, err
			}
		}
	}

	result, err := Get(c, qtx, group.Name, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, tx.Commit(c)
}
