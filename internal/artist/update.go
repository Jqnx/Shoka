package artist

import (
	"Shoka/internal/logger"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateTransaction(c context.Context,
	db *pgxpool.Pool,
	q *repository.Queries,
	p *models.ArtistPayload,
	oldname string,
	log logger.Logger,
) (*models.ArtistResponse, error) {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)

	name := strings.ToLower(p.Name)
	artist, err := qtx.UpdateArtist(c, repository.UpdateArtistParams{
		Name:      name,
		UpdatedAt: time.Now(),
		OldName:   oldname,
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Alias(c, qtx, p, &artist); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Link(c, qtx, p, &artist); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Group(c, qtx, p, &artist); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result, err := Get(c, qtx, name, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, tx.Commit(c)
}
