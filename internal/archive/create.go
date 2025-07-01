package archive

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"context"
	"errors"
	"time"
)

// TODO: ArchivePayload to Archive converter
//func CreateTransaction(c context.Context,
//	db *pgxpool.Pool,
//	q *repository.Queries,
//	payload *models.ArchivePayload,
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
//	//aid, err := qtx.GetArchiveLastAID(c)
//	//if err != nil {
//	//	if err == pgx.ErrNoRows {
//	//		aid = 0
//	//	} else {
//	//		log.Error(err.Error())
//	//		return nil, err
//	//	}
//	//}
//
//	archiveID := NewArchiveID()
//
//	lang := strings.ToLower(payload.Language)
//	category := strings.ToLower(payload.Category)
//	archive, err := qtx.CreateArchive(c, repository.CreateArchiveParams{
//		Title:    payload.Title,
//		Summary:  &payload.Summary,
//		Language: &lang,
//		Category: &category,
//		// FilePath:  &payload.FilePath,
//		ArchiveID: archiveID,
//		CreatedAt: time.Now(),
//		UpdatedAt: time.Now(),
//	})
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//
//	//if err := Artist(c, qtx, payload, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := Tag(c, qtx, payload, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := Character(c, qtx, payload, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := Parody(c, qtx, payload, &archive); err != nil {
//	//	log.Error(err.Error())
//	//	return nil, err
//	//}
//
//	//if err := URL(c, qtx, payload, &archive); err != nil {
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

// TODO: Create thumndir in archive creation
func CreateFromFile(c context.Context, path string, app *config.App) (*repository.CreateArchiveRow, error) {
	if fsutil.MatchExtension(path, config.ArchiveExtensions) {

		//exists, err := app.Repo.FilePathExists(c, &path)
		//if err != nil {
		//	app.Log.ErrorContext(c, "error checking if file is already in db")
		//}
		//if exists.RowsAffected() == 0 {

		archiveID := NewArchiveID()
		title := fsutil.GetNameFromPath(path, true)
		pagecount := fsutil.GetPageCount(path, config.ImageExtensions)

		hash := fsutil.GenHash(path)

		archive, err := app.Repo.CreateArchive(c, repository.CreateArchiveParams{
			Title:     title,
			ArchiveID: archiveID,
			PageCount: pagecount,
			FilePath:  &path,
			Type:      "archive",
			Hash:      hash,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			app.Log.ErrorContext(c, err.Error())
			return nil, err
		}
		app.Log.Info("archive", "created:", title)
		return &archive, nil
	}
	//} else {
	//	title := fsutil.GetNameFromPath(path, true)
	//	app.Log.Warn("failed to create", "archive", title)
	//}
	return nil, errors.New("error: extension does not match")
}

////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////

// Insert inserts an Archive struct into the db
func (b *Archive) Insert(c context.Context, app *config.App) error {
	_, err := app.Repo.CreateArchive(c, repository.CreateArchiveParams{
		Title:      *b.Title,
		PageCount:  b.PageCount,
		FilePath:   b.FilePath,
		ArchiveID:  b.ArchiveID,
		Hash:       b.Hash,
		ThumbsPath: b.ThumbsPath,
		// CoverPath:  b.CoverPath,
		// PagesPath:  b.PagesPath,
		Type:      b.Type,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	})
	if err != nil {
		app.Log.ErrorContext(c, err.Error())
		return err
	}
	return nil
}

// DIRECTOR
//type Director struct {
//	builder IArchive
//}
//
//func NewDirector(a IArchive) *Director {
//	return &Director{
//		builder: a,
//	}
//}

// NewArchive sets ArchiveBuilder fields and returns an Archive struct
//func (d *Director) NewArchive(path string) Archive {
//	// TODO:
//	// Call GetMetadata on filepath to fetch from comicinfo/info.json/info.txt
//	// should return a struct of metadata, pass that metadata into builder functions
//	// Example:
//	// struct := GetMetadata(path)
//	// d.builder.setTitle(struct.Title)
//
//	d.builder.setFilePath(path)
//	d.builder.setArchiveID()
//	d.builder.setTitle("")
//	d.builder.setSummary("")
//	d.builder.setLanguage("")
//	d.builder.setCategory("")
//	d.builder.setPageCount()
//	d.builder.setHash()
//	d.builder.setType()
//	d.builder.setThumbsPath()
//	// d.builder.setCoverPath()
//	d.builder.createPagesPath()
//	d.builder.setCreatedAt()
//	d.builder.setUpdatedAt()
//	return d.builder.getArchive()
//}
