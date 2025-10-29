package archive

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
)

func RepoToArchive(a any, app *config.App) *Archive {
	switch a := a.(type) {
	case repository.GetArchiveByIDRow:
		return &Archive{
			ID:         a.ID,
			Title:      a.Title,
			Summary:    a.Summary,
			Language:   a.Language,
			Category:   a.Category,
			PageCount:  a.PageCount,
			FilePath:   a.FilePath,
			Hash:       a.Hash,
			ThumbsPath: a.ThumbsPath,
			Type:       a.Type,
			CreatedAt:  a.CreatedAt,
			UpdatedAt:  a.UpdatedAt,
			app:        app,
		}
	case *repository.GetArchiveByIDRow:
		return &Archive{
			ID:         a.ID,
			Title:      a.Title,
			Summary:    a.Summary,
			Language:   a.Language,
			Category:   a.Category,
			PageCount:  a.PageCount,
			FilePath:   a.FilePath,
			Hash:       a.Hash,
			ThumbsPath: a.ThumbsPath,
			Type:       a.Type,
			CreatedAt:  a.CreatedAt,
			UpdatedAt:  a.UpdatedAt,
			app:        app,
		}
	default:
		return nil
	}
}
