package archive

import (
	"time"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
)

// GetBuilder creates a new ArchiveBuilder and passes app config
// returns the Archive interface
func GetBuilder(app *config.App) *ArchiveBuilder {
	return newArchiveBuilder(app)
}

// The ArchiveBuilder struct containing data about an archive
// that can be set using the built-on functions
type ArchiveBuilder struct {
	ID          string
	Type        string
	Title       *string
	Summary     *string
	Language    *string
	Category    *string
	Tags        *[]models.Tag
	URL         *[]models.URL
	Parody      *[]models.Parody
	Character   *[]models.Character
	Artist      *[]models.Artist
	ReleaseDate *time.Time
	PageCount   int16
	FilePath    *string
	FileName    *string
	Hash        string
	ThumbsPath  *string
	CoverPath   *string
	PagesPath   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	app         *config.App
	td          *fsutil.ThumbDir
}

// newArchiveBuilder creates a new empty ArchiveBuilder
// returns a pointer to that ArchiveBuilder
func newArchiveBuilder(app *config.App) *ArchiveBuilder {
	return &ArchiveBuilder{app: app}
}

// setArchiveID creates an archive's unique ID
// which is the first 4 bytes of a random UUID
func (a *ArchiveBuilder) setArchiveID(id string) {
	if id == "" {
		// newId := util.GenShortenedUUID()
		newID := NewArchiveID()
		a.ID = newID
	} else {
		a.ID = id
	}
}

// setTitle sets archive's title
// if title is empty uses fileName
func (a *ArchiveBuilder) setTitle(title string) {
	if title == "" {
		name := fsutil.GetNameFromPath(*a.FilePath, true)
		a.Title = &name
	} else {
		a.Title = &title
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

// setFileName sets archive's FileName
func (a *ArchiveBuilder) setFileName() {
	name := fsutil.GetNameFromPath(*a.FilePath, false)
	a.FileName = &name
}

// setHash sets archive's FileHash
func (a *ArchiveBuilder) setHash() {
	hash := fsutil.GenHash(*a.FilePath)
	a.Hash = hash
}

// setThumbsPath sets archive's ThumbsPath
// creates if necessary
func (a *ArchiveBuilder) setThumbsPath() {
	d := fsutil.NewThumbDir(a.app.Cfg.ThumbDir, a.Type, a.Hash)
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
		Type:        a.Type,
		ID:          a.ID,
		Title:       a.Title,
		Summary:     a.Summary,
		Language:    a.Language,
		Category:    a.Category,
		Tags:        a.Tags,
		URL:         a.URL,
		Parody:      a.Parody,
		Character:   a.Character,
		Artist:      a.Artist,
		ReleaseDate: a.ReleaseDate,
		PageCount:   a.PageCount,
		FilePath:    a.FilePath,
		FileName:    a.FileName,
		Hash:        a.Hash,
		ThumbsPath:  a.ThumbsPath,
		CoverPath:   a.CoverPath,
		PagesPath:   a.PagesPath,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func (a *ArchiveBuilder) NewArchive(path string) Archive {
	a.setFilePath(path)
	a.setFileName()
	a.setHash()
	a.setArchiveID("")
	a.setTitle("")
	a.setSummary("")
	a.setLanguage("")
	a.setCategory("")
	a.setPageCount()
	a.setType()
	a.setThumbsPath()
	a.setCoverPath()
	a.createPagesPath()
	a.setCreatedAt()
	a.setUpdatedAt()
	return a.getArchive()
}

func (a *ArchiveBuilder) UpdateArchive(id string, meta *models.Metadata) Archive {
	a.setArchiveID(id)
	a.setTitle(meta.Title)
	a.setSummary(meta.Summary)
	a.setLanguage(meta.Language)
	a.setCategory(meta.Category)
	a.setTags(&meta.Tags)
	a.setURL(&meta.URL)
	a.setParody(&meta.Parody)
	a.setCharacter(&meta.Character)
	a.setArtist(&meta.Artist)
	a.setReleaseDate(meta.ReleaseDate)
	a.setUpdatedAt()
	return a.getArchive()
}

func (a *ArchiveBuilder) setTags(t *[]models.Tag) {
	a.Tags = t
}

func (a *ArchiveBuilder) setURL(u *[]models.URL) {
	a.URL = u
}

func (a *ArchiveBuilder) setParody(p *[]models.Parody) {
	a.Parody = p
}

func (a *ArchiveBuilder) setCharacter(c *[]models.Character) {
	a.Character = c
}

func (a *ArchiveBuilder) setArtist(b *[]models.Artist) {
	a.Artist = b
}

func (a *ArchiveBuilder) setReleaseDate(t *time.Time) {
	a.ReleaseDate = t
}
