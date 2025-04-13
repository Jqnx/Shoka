package archive

import (
	"Shoka/internal/repository"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func DeleteTransaction(c context.Context,
	db *pgx.Conn,
	q *repository.Queries,
	id int64,
	log *slog.Logger,
) error {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)

	archive, err := qtx.GetArchiveByAID(c, id)
	if err != nil {
		return err
	}

	if err := qtx.RemoveArtistFromArchive(c, archive.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	if err := qtx.RemoveTagFromArchive(c, archive.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	if err := qtx.RemoveCharacterFromArchive(c, archive.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	if err := qtx.RemoveParodyFromArchive(c, archive.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	if err := qtx.RemoveArchiveUrl(c, archive.ID); err != nil {
		log.Error(err.Error())
		return err
	}

	if err := qtx.DeleteArchive(c, id); err != nil {
		log.Error(err.Error())
		return err
	}
	return tx.Commit(c)
}
