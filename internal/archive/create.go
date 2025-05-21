package archive

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO:

func CreateTransaction(c context.Context,
	db *pgxpool.Pool,
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

	//aid, err := qtx.GetArchiveLastAID(c)
	//if err != nil {
	//	if err == pgx.ErrNoRows {
	//		aid = 0
	//	} else {
	//		log.Error(err.Error())
	//		return nil, err
	//	}
	//}

	archiveID := NewArchiveID()

	lang := strings.ToLower(payload.Language)
	category := strings.ToLower(payload.Category)
	archive, err := qtx.CreateArchive(c, repository.CreateArchiveParams{
		Title:     payload.Title,
		Summary:   &payload.Summary,
		Language:  &lang,
		Category:  &category,
		FilePath:  &payload.FilePath,
		ArchiveID: archiveID,
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

	result, err := Get(c, qtx, archive.ArchiveID, log)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, tx.Commit(c)
}

// TODO: Create thumndir in archive creation
func CreateFromFile(c context.Context, path string, app *config.App) (*repository.Archive, error) {
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
			Hash:      &hash,
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

// Archive interface that implements builder functions
type IArchive interface {
	setArchiveID()
	setTitle(title string)
	setSummary(summary string)
	setLanguage(language string)
	setCategory(category string)
	setPageCount()
	setFilePath(path string) // NOTE: SHOULD BE SET FIRST IN DIRECTOR CLASS, OTHER "set" FUNCTIONS USE IT e.g. setTitle, setPageCount, setHash, setThumbsPath, setCoverPath
	setHash()
	setThumbsPath()
	setCoverPath()
	createPagesPath()
	setType()
	setCreatedAt()
	setUpdatedAt()

	// TODO: Add metadata builders to struct
	// setTags()
	// setArtist()
	// setCharacter()
	// setParody()
	// setURL()
	getArchive() Archive
}

// Archive struct that represents an archive
type Archive struct {
	ArchiveID  string
	Title      string
	Summary    *string
	Language   *string
	Category   *string
	PageCount  int64
	FilePath   *string
	Hash       *string
	ThumbsPath *string
	CoverPath  *string
	PagesPath  *string
	Type       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// getBuilder creates a new ArchiveBuilder and passes app config
// returns the Archive interface
func GetBuilder(app *config.App) *ArchiveBuilder {
	return newArchiveBuilder(app)
}

// The ArchiveBuilder struct containing data about an archive
// that can be set using the built-on functions
type ArchiveBuilder struct {
	ArchiveID  string
	Title      string
	Summary    *string
	Language   *string
	Category   *string
	PageCount  int64
	FilePath   *string
	Hash       *string
	ThumbsPath *string
	CoverPath  *string
	PagesPath  *string
	Type       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	app        *config.App
	td         *fsutil.ThumbDir
}

// newArchiveBuilder creates a new empty ArchiveBuilder
// returns a pointer to that ArchiveBuilder
func newArchiveBuilder(app *config.App) *ArchiveBuilder {
	return &ArchiveBuilder{app: app}
}

// setArchiveID creates an archive's unique ID
// which is the first 4 bytes of a random UUID
func (a *ArchiveBuilder) setArchiveID() {
	id := util.GenShortenedUUID()
	a.ArchiveID = id
}

// setTitle sets archive's title
// if title is empty uses fileName
func (a *ArchiveBuilder) setTitle(title string) {
	if title == "" {
		name := fsutil.GetNameFromPath(*a.FilePath, true)
		a.Title = name
	} else {
		a.Title = title
	}
}

// setSummary sets archive's summary
// if passed summary is empty sets to nil
func (a *ArchiveBuilder) setSummary(summary string) {
	if summary != "" {
		a.Summary = &summary
	} else {
		a.Summary = nil
	}
}

// setLanguage sets archive's language
// if passed language is empty sets to nil
func (a *ArchiveBuilder) setLanguage(language string) {
	if language != "" {
		a.Language = &language
	} else {
		a.Language = nil
	}
}

// setCategory sets archive's category
// if passed category is empty sets to nil
func (a *ArchiveBuilder) setCategory(category string) {
	if category != "" {
		a.Category = &category
	} else {
		a.Category = nil
	}
}

// setPageCount sets archive's PageCount
func (a *ArchiveBuilder) setPageCount() {
	count := fsutil.GetPageCount(*a.FilePath, config.ImageExtensions)
	a.PageCount = count
}

// setFilePath sets archive's FilePath
func (a *ArchiveBuilder) setFilePath(path string) {
	a.FilePath = &path
}

// setHash sets archive's FileHash
func (a *ArchiveBuilder) setHash() {
	hash := fsutil.GenHash(*a.FilePath)
	a.Hash = &hash
}

// setThumbsPath sets archive's ThumbsPath
// creates if necessary
func (a *ArchiveBuilder) setThumbsPath() {
	d := fsutil.NewThumbDir(a.app.Cfg.ThumbDir, a.Type, *a.Hash)
	p, err := d.CreateThumbDir()
	if err != nil {
		return
	}
	a.td = d
	a.ThumbsPath = &p
}

// setCoverPath sets archive's CoverPath
// creates if necessary
func (a *ArchiveBuilder) setCoverPath() {
	p, err := a.td.CreateCoverDir(*a.ThumbsPath)
	if err != nil {
		return
	}
	a.CoverPath = &p
}

// setPagesPath sets archive's PagesPath
// creates if necessary
func (a *ArchiveBuilder) createPagesPath() {
	p, err := a.td.CreatePageDir(*a.ThumbsPath)
	if err != nil {
		return
	}
	a.PagesPath = &p
}

// setType sets archive's media type
func (a *ArchiveBuilder) setType() {
	a.Type = "archive"
}

// setCreatedAt sets time at which archive is created and put in db
func (a *ArchiveBuilder) setCreatedAt() {
	a.CreatedAt = time.Now()
}

// setUpdatedAt sets time at which archive is updated
func (a *ArchiveBuilder) setUpdatedAt() {
	a.UpdatedAt = time.Now()
}

// getArchive returns a filled in repository.Archive struct
func (a *ArchiveBuilder) getArchive() Archive {
	return Archive{
		Title:      a.Title,
		Summary:    a.Summary,
		Language:   a.Language,
		Category:   a.Category,
		PageCount:  a.PageCount,
		FilePath:   a.FilePath,
		ArchiveID:  a.ArchiveID,
		Hash:       a.Hash,
		ThumbsPath: a.ThumbsPath,
		// CoverPath:  a.CoverPath,
		PagesPath: a.PagesPath,
		Type:      a.Type,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// Insert inserts an Archive struct into the db
func (b *Archive) Insert(c context.Context, app *config.App) error {
	_, err := app.Repo.CreateArchive(c, repository.CreateArchiveParams{
		Title:      b.Title,
		Summary:    b.Summary,
		Language:   b.Language,
		Category:   b.Category,
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
type Director struct {
	builder IArchive
}

func NewDirector(a IArchive) *Director {
	return &Director{
		builder: a,
	}
}

// NewArchive sets ArchiveBuilder fields and returns an Archive struct
func (d *Director) NewArchive(path string) Archive {
	// TODO:
	// Call GetMetadata on filepath to fetch from comicinfo/info.json/info.txt
	// should return a struct of metadata, pass that metadata into builder functions
	// Example:
	// struct := GetMetadata(path)
	// d.builder.setTitle(struct.Title)

	d.builder.setFilePath(path)
	d.builder.setArchiveID()
	d.builder.setTitle("")
	d.builder.setSummary("")
	d.builder.setLanguage("")
	d.builder.setCategory("")
	d.builder.setPageCount()
	d.builder.setHash()
	d.builder.setType()
	d.builder.setThumbsPath()
	// d.builder.setCoverPath()
	d.builder.createPagesPath()
	d.builder.setCreatedAt()
	d.builder.setUpdatedAt()
	return d.builder.getArchive()
}

func (a *ArchiveBuilder) NewArchive(path string) Archive {
	// TODO:
	// Call GetMetadata on filepath to fetch from comicinfo/info.json/info.txt
	// should return a struct of metadata, pass that metadata into builder functions
	// Example:
	// struct := GetMetadata(path)
	// d.builder.setTitle(struct.Title)

	a.setFilePath(path)
	a.setArchiveID()
	a.setTitle("")
	a.setSummary("")
	a.setLanguage("")
	a.setCategory("")
	a.setPageCount()
	a.setHash()
	a.setType()
	a.setThumbsPath()
	// a.setCoverPath()
	a.createPagesPath()
	a.setCreatedAt()
	a.setUpdatedAt()
	return a.getArchive()
}
