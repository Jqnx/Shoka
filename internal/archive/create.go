package archive

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// TODO:

func CreateTransaction(c context.Context,
	db *pgx.Conn,
	q *repository.Queries,
	payload *models.ArchivePayload,
	log *slog.Logger,
) (*models.ArchiveResponse, error) {
	tx, err := db.Begin(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer tx.Rollback(c)
	qtx := q.WithTx(tx)

	aid, err := qtx.GetArchiveLastAID(c)
	if err != nil {
		if err == pgx.ErrNoRows {
			aid = 0
		} else {
			log.Error(err.Error())
			return nil, err
		}
	}

	lang := strings.ToLower(payload.Language)
	category := strings.ToLower(payload.Category)
	archive, err := qtx.CreateArchive(c, repository.CreateArchiveParams{
		Title:     payload.Title,
		Summary:   &payload.Summary,
		Lang:      &lang,
		Category:  &category,
		FilePath:  &payload.FilePath,
		AID:       aid + 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Artist(c, qtx, payload, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Tag(c, qtx, payload, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Character(c, qtx, payload, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := Parody(c, qtx, payload, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err := URL(c, qtx, payload, &archive); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result, err := Get(c, qtx, archive.AID, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, tx.Commit(c)
}

func CreateFromFile(c context.Context, path string, q *repository.Queries, log *slog.Logger) error {
	if fsutil.MatchExtension(path, config.ArchiveExtensions) {
		aid, err := q.GetArchiveLastAID(c)
		if err != nil {
			if err == pgx.ErrNoRows {
				aid = 0
			} else {
				log.Error(err.Error())
				return err
			}
		}

		exists, err := q.FilePathExists(c, &path)
		if err != nil {
			log.ErrorContext(c, "error checking if file is already in db")
		}
		if exists.RowsAffected() == 0 {
			title := fsutil.GetNameFromPath(path, true)
			pagecount := fsutil.GetPageCount(path, config.ImageExtensions)

			_, err = q.CreateArchive(c, repository.CreateArchiveParams{
				Title:     title,
				PageCount: pagecount,
				FilePath:  &path,
				AID:       aid + 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				log.ErrorContext(c, err.Error())
				return err
			}
			log.Info(title)
		}
	} else {
		title := fsutil.GetNameFromPath(path, true)
		log.Warn("failed to create", "archive", title)
	}
	return nil
}
