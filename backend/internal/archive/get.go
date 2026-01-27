package archive

import (
	"context"
	"errors"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TODO: Also update the metadataToFileHandler and the getArchiveHandler to properly use this

type ArchiveWithMetadata struct {
	Archive    repository.GetArchiveByIDRow
	Artists    []repository.Artist
	Tags       []repository.Tag
	Characters []repository.Character
	Parodies   []repository.Parody
	URLs       []repository.GetArchiveUrlsRow
}

func GetWithMetadata(app *config.App, id string) (*ArchiveWithMetadata, error) {
	ctx := context.Background()
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := app.Repo.WithTx(tx)

	archive, err := qtx.GetArchiveByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tags, err := qtx.GetArchiveTag(ctx, id)
	if err != nil {
		return nil, err
	}
	characters, err := qtx.GetArchiveCharacters(ctx, id)
	if err != nil {
		return nil, err
	}
	parodies, err := qtx.GetArchiveParody(ctx, id)
	if err != nil {
		return nil, err
	}
	urls, err := qtx.GetArchiveUrls(ctx, id)
	if err != nil {
		return nil, err
	}
	artists, err := qtx.GetArchiveArtists(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &ArchiveWithMetadata{
		Archive:    archive,
		Artists:    artists,
		Tags:       tags,
		Characters: characters,
		Parodies:   parodies,
		URLs:       urls,
	}

	return result, tx.Commit(ctx)
}

func GetResponse(app *config.App, id string, userid uuid.UUID) (*models.ArchiveResponse, error) {
	ctx := context.Background()
	res, err := GetWithMetadata(app, id)
	if err != nil {
		return nil, err
	}

	isFav := false
	read := models.ReadingProgress{
		Progress: 0,
		LastRead: nil,
	}

	if userid != uuid.Nil {
		check, err := app.Repo.ArchiveIsFavorited(ctx, repository.ArchiveIsFavoritedParams{
			UserID:    userid,
			ArchiveID: id,
		})
		if err != nil {
			return nil, err
		}
		if check.RowsAffected() != 0 {
			isFav = true
		}

		rp, err := app.Repo.GetUserReadingProgress(ctx, repository.GetUserReadingProgressParams{
			UserID:    userid,
			ArchiveID: id,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			read.SetState(rp.Status)
			read.Progress = rp.Page
			read.LastRead = &rp.LastRead
		}
	}

	result := &models.ArchiveResponse{
		ID:          res.Archive.ID,
		Title:       res.Archive.Title,
		Summary:     res.Archive.Summary,
		Language:    res.Archive.Language,
		Category:    res.Archive.Category,
		PageCount:   res.Archive.PageCount,
		FileHash:    res.Archive.FileHash,
		CreatedAt:   res.Archive.CreatedAt,
		UpdatedAt:   res.Archive.UpdatedAt,
		ReleaseDate: res.Archive.ReleaseDate,
		PagesOnDisk: fsutil.CountPages(res.Archive),
		Tags:        res.Tags,
		Artist:      res.Artists,
		Parody:      res.Parodies,
		Character:   res.Characters,
		URL:         res.URLs,
		Status:      read.Status.String(),
		Progress:    read.Progress,
		LastRead:    read.LastRead,
		IsFavorite:  isFav,
	}

	return result, nil
}

//func GetAll(c context.Context, q *repository.Queries, log *slog.Logger, uid uuid.UUID) (*[]models.ArchiveResponse, error) {
//	archives, err := q.GetAllArchives(c, uid)
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//
//	result := []models.ArchiveResponse{}
//
//	for _, item := range archives {
//		tags, err := q.GetArchiveTag(c, item.ID)
//		if err != nil {
//			log.Error(err.Error())
//			return nil, err
//		}
//		characters, err := q.GetArchiveCharacters(c, item.ID)
//		if err != nil {
//			log.Error(err.Error())
//			return nil, err
//		}
//		parodies, err := q.GetArchiveParody(c, item.ID)
//		if err != nil {
//			log.Error(err.Error())
//			return nil, err
//		}
//		urls, err := q.GetArchiveUrls(c, item.ID)
//		if err != nil {
//			log.Error(err.Error())
//			return nil, err
//		}
//		artists, err := q.GetArchiveArtists(c, item.ID)
//		if err != nil {
//			log.Error(err.Error())
//			return nil, err
//		}
//		archive := models.ArchiveResponse{
//			ID:          item.ID,
//			Title:       item.Title,
//			Summary:     item.Summary,
//			Tags:        tags,
//			Artist:      artists,
//			Parody:      parodies,
//			Character:   characters,
//			Language:    item.Language,
//			Category:    item.Category,
//			PageCount:   item.PageCount,
//			URL:         urls,
//			FileHash:    item.FileHash,
//			CreatedAt:   item.CreatedAt,
//			UpdatedAt:   item.UpdatedAt,
//			ReleaseDate: item.ReleaseDate,
//		}
//
//		result = append(result, archive)
//	}
//	return &result, nil
//}
