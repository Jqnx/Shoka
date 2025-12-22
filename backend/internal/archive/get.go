package archive

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"

	"github.com/google/uuid"
)

// TODO: Change function to not return an ArchiveResponse but instead just every item from db
// Archive:   arch,
// Artists:   artists,
// Tags:      tags,
// Parody:    parodies,
// Character: characters,
// URLs:      urls,
//
// TODO: Also update the metadataToFileHandler and the getArchiveHandler to properly use this

func GetResponse(ctx context.Context, app *config.App, id string) (*models.ArchiveResponse, error) {
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

	var pages int
	p, _ := filepath.Abs(*archive.ThumbPath)
	d, err := os.ReadDir(filepath.Join(p, "pages"))
	if err != nil {
		pages = 0
	}
	for _, i := range d {
		if fsutil.MatchExtension(i.Name(), config.ImageExtensions) {
			pages++
		}
	}

	result := &models.ArchiveResponse{
		ID:          archive.ID,
		Title:       archive.Title,
		Summary:     archive.Summary,
		Tags:        tags,
		Artist:      artists,
		Parody:      parodies,
		Character:   characters,
		Language:    archive.Language,
		Category:    archive.Category,
		PageCount:   archive.PageCount,
		URL:         urls,
		FileHash:    archive.FileHash,
		Pages:       pages,
		CreatedAt:   archive.CreatedAt,
		UpdatedAt:   archive.UpdatedAt,
		ReleaseDate: archive.ReleaseDate,
	}
	return result, tx.Commit(ctx)
}

func GetAll(c context.Context, q *repository.Queries, log *slog.Logger, uid uuid.UUID) (*[]models.ArchiveResponse, error) {
	archives, err := q.GetAllArchives(c, uid)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result := []models.ArchiveResponse{}

	for _, item := range archives {
		tags, err := q.GetArchiveTag(c, item.ID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		characters, err := q.GetArchiveCharacters(c, item.ID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		parodies, err := q.GetArchiveParody(c, item.ID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		urls, err := q.GetArchiveUrls(c, item.ID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		artists, err := q.GetArchiveArtists(c, item.ID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		archive := models.ArchiveResponse{
			ID:          item.ID,
			Title:       item.Title,
			Summary:     item.Summary,
			Tags:        tags,
			Artist:      artists,
			Parody:      parodies,
			Character:   characters,
			Language:    item.Language,
			Category:    item.Category,
			PageCount:   item.PageCount,
			URL:         urls,
			FileHash:    item.FileHash,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			ReleaseDate: item.ReleaseDate,
		}

		result = append(result, archive)
	}
	return &result, nil
}
