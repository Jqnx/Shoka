package archive

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

// TODO:

func (a *Archive) Get(ctx context.Context, app *config.App) (*models.ArchiveResponse, error) {
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := app.Repo.WithTx(tx)

	archive, err := qtx.GetArchiveByID(ctx, a.ArchiveID)
	if err != nil {
		return nil, err
	}
	tags, err := qtx.GetArchiveTags(ctx, a.ArchiveID)
	if err != nil {
		return nil, err
	}
	characters, err := qtx.GetArchiveCharacters(ctx, a.ArchiveID)
	if err != nil {
		return nil, err
	}
	parodies, err := qtx.GetArchiveParodies(ctx, a.ArchiveID)
	if err != nil {
		return nil, err
	}
	urls, err := qtx.GetArchiveURLs(ctx, a.ArchiveID)
	if err != nil {
		return nil, err
	}
	artists, err := qtx.GetArchiveArtists(ctx, a.ArchiveID)
	if err != nil {
		return nil, err
	}

	var pages int
	p, _ := filepath.Abs(*archive.ThumbsPath)
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
		ArchiveID:   archive.ArchiveID,
		Title:       archive.Title,
		Summary:     archive.Summary,
		Tags:        tags,
		Artist:      artists,
		Parody:      parodies,
		Character:   characters,
		Language:    archive.Language,
		Category:    archive.Category,
		PageCount:   archive.PageCount,
		Url:         urls,
		Hash:        archive.Hash,
		Pages:       pages,
		Type:        archive.Type,
		CreatedAt:   archive.CreatedAt,
		UpdatedAt:   archive.UpdatedAt,
		ReleaseDate: archive.ReleaseDate,
	}
	return result, tx.Commit(ctx)
}

func GetAll(c context.Context, q *repository.Queries, log *slog.Logger) (*[]models.ArchiveResponse, error) {
	archives, err := q.GetAllArchives(c)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	result := []models.ArchiveResponse{}

	for _, item := range archives {
		tags, err := q.GetArchiveTags(c, item.ArchiveID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		characters, err := q.GetArchiveCharacters(c, item.ArchiveID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		parodies, err := q.GetArchiveParodies(c, item.ArchiveID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		urls, err := q.GetArchiveURLs(c, item.ArchiveID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		artists, err := q.GetArchiveArtists(c, item.ArchiveID)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		archive := models.ArchiveResponse{
			ArchiveID: item.ArchiveID,
			Title:     item.Title,
			Summary:   item.Summary,
			Tags:      tags,
			Artist:    artists,
			Parody:    parodies,
			Character: characters,
			Language:  item.Language,
			Category:  item.Category,
			PageCount: item.PageCount,
			Url:       urls,
			Hash:      item.Hash,
			// ThumbsPath: item.ThumbsPath,
			Type:      item.Type,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}

		result = append(result, archive)
	}
	return &result, nil
}

func GetByTag(c context.Context, q *repository.Queries, tag string, log *slog.Logger) ([]string, error) {
	archives, err := q.GetArchivesByTag(c, tag)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return archives, nil
}

func GetByCharacter(c context.Context, q *repository.Queries, character string, log *slog.Logger) ([]string, error) {
	archives, err := q.GetArchivesByCharacter(c, character)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return archives, nil
}

func GetByParody(c context.Context, q *repository.Queries, parody string, log *slog.Logger) ([]string, error) {
	archives, err := q.GetArchivesByParody(c, parody)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return archives, nil
}
