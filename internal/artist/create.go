package artist

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO:
// Fix issue where you can't create if not every option is set.
// e.g. Make it so you can create object without having to set every parameter.

func CreateTransaction(c context.Context,
	db *pgxpool.Pool,
	q *repository.Queries,
	p *models.ArtistPayload,
	log *slog.Logger,
) (*models.ArtistResponse, error) {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)
	artist, err := qtx.CreateArtist(c, repository.CreateArtistParams{
		Name:      p.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Alias(c, qtx, p, &artist); err != nil {
		return nil, err
	}

	if err := Link(c, qtx, p, &artist); err != nil {
		return nil, err
	}

	if err := Group(c, qtx, p, &artist); err != nil {
		return nil, err
	}

	result, err := Get(c, qtx, artist.Name, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, tx.Commit(c)
}
