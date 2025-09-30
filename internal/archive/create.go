package archive

import (
	"context"

	"Shoka/internal/config"
	"Shoka/internal/repository"
)

// TODO: Create thumndir in archive creation
//func CreateFromFile(c context.Context, path string, app *config.App) (*repository.CreateArchiveRow, error) {
//	if fsutil.MatchExtension(path, config.ArchiveExtensions) {
//
//		//exists, err := app.Repo.FilePathExists(c, &path)
//		//if err != nil {
//		//	app.Log.ErrorContext(c, "error checking if file is already in db")
//		//}
//		//if exists.RowsAffected() == 0 {
//
//		archiveID := NewArchiveID()
//		title := fsutil.GetNameFromPath(path, true)
//		pagecount := fsutil.GetPageCount(path, config.ImageExtensions)
//
//		hash := fsutil.GenHash(path)
//
//		archive, err := app.Repo.CreateArchive(c, repository.CreateArchiveParams{
//			Title:     title,
//			ArchiveID: archiveID,
//			PageCount: pagecount,
//			FilePath:  &path,
//			Type:      "archive",
//			Hash:      hash,
//			CreatedAt: time.Now(),
//			UpdatedAt: time.Now(),
//		})
//		if err != nil {
//			app.Log.Error("failed to insert archive in db", "error", err)
//			// app.Log.ErrorContext(c, err.Error())
//			return nil, err
//		}
//		app.Log.Info("archive", "created:", title)
//		return &archive, nil
//	}
//	//} else {
//	//	title := fsutil.GetNameFromPath(path, true)
//	//	app.Log.Warn("failed to create", "archive", title)
//	//}
//	return nil, errors.New("error: extension does not match")
//}

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
		ID:         b.ID,
		Title:      *b.Title,
		PageCount:  b.PageCount,
		FilePath:   b.FilePath,
		FileName:   b.FileName,
		Hash:       b.Hash,
		ThumbsPath: b.ThumbsPath,
		// CoverPath:  b.CoverPath,
		Type:      b.Type,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	})
	if err != nil {
		app.Log.Error("failed to insert archive in db", "error", err)
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
