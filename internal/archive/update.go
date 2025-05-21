package archive

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO:

func UpdateTransaction(c context.Context,
	db *pgxpool.Pool,
	q *repository.Queries,
	p *models.ArchivePayload,
	id string,
	log *slog.Logger,
) (*models.ArchiveResponse, error) {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)

	lang := strings.ToLower(p.Language)
	category := strings.ToLower(p.Category)
	archive, err := qtx.UpdateArchive(c, repository.UpdateArchiveParams{
		Title:     p.Title,
		Summary:   &p.Summary,
		Language:  &lang,
		Category:  &category,
		FilePath:  &p.FilePath,
		UpdatedAt: time.Now(),
		ArchiveID: id,
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Artist(c, qtx, p, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Tag(c, qtx, p, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Character(c, qtx, p, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Parody(c, qtx, p, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := URL(c, qtx, p, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result, err := Get(c, qtx, archive.ArchiveID, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, tx.Commit(c)
}
