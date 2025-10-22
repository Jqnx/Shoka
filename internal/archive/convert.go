package archive

import (
	"Shoka/internal/repository"
)

func RepoToArchive(a any) *Archive {
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
		}
	default:
		return nil
	}
}
