package artist

import (
	"Shoka/internal/repository"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func DeleteTransaction(c context.Context,
	db *pgx.Conn,
	q *repository.Queries,
	name string,
	log *slog.Logger,
) error {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)

	artist, err := qtx.GetArtistByName(c, name)
	if err != nil {
		return err
	}

	if err := qtx.RemoveArtistAliases(c, artist.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	if err := qtx.RemoveArtistLinks(c, artist.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	if err := qtx.RemoveArtistFromGroup(c, artist.ID); err != nil {
		log.Error(err.Error())
		return err
	}

	if err := qtx.DeleteArtist(c, artist.ID); err != nil {
		log.Error(err.Error())
		return err
	}
	return tx.Commit(c)
}
