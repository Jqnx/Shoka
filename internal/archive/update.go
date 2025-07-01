package archive

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"context"
)

// TODO: ArchivePayload to Archive converter
//func UpdateTransaction(c context.Context,
//	db *pgxpool.Pool,
//	q *repository.Queries,
//	p *models.ArchivePayload,
//	id string,
//	log *slog.Logger,
//) (*models.ArchiveResponse, error) {
//	tx, err := db.Begin(c)
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//	defer tx.Rollback(c)
//	qtx := q.WithTx(tx)
//
//	lang := strings.ToLower(p.Language)
//	category := strings.ToLower(p.Category)
//	archive, err := qtx.UpdateArchive(c, repository.UpdateArchiveParams{
//		// Title:    p.Title,
//		Summary:  &p.Summary,
//		Language: &lang,
//		Category: &category,
//		// FilePath:  &p.FilePath,
//		UpdatedAt: time.Now(),
//		ArchiveID: id,
//	})
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//
//	//if err := Artist(c, qtx, p, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := Tag(c, qtx, p, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := Character(c, qtx, p, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := Parody(c, qtx, p, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := URL(c, qtx, p, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	result, err := Get(c, qtx, archive.ArchiveID, log)
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//
//	return result, tx.Commit(c)
//}

func (a *Archive) Update(ctx context.Context, app *config.App) error {
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := app.Repo.WithTx(tx)

	archive, err := qtx.UpdateArchive(ctx, repository.UpdateArchiveParams{
		Title:       a.Title,
		Summary:     a.Summary,
		Language:    a.Language,
		Category:    a.Category,
		UpdatedAt:   a.UpdatedAt,
		ReleaseDate: a.ReleaseDate,
		ArchiveID:   a.ArchiveID,
	})
	if err != nil {
		return err
	}

	m := NewMetadata(qtx, a, archive.ID)

	if err := m.Artist(ctx); err != nil {
		return err
	}

	if err := m.Tag(ctx); err != nil {
		return err
	}

	if err := m.Character(ctx); err != nil {
		return err
	}

	if err := m.Parody(ctx); err != nil {
		return err
	}

	if err := m.URL(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
