package archive

import (
	"fmt"
	"time"

	"Shoka/internal/fsutil"
	"Shoka/internal/models"
)

// setArchiveID creates an archive's unique ID
// which is the first 4 bytes of a random UUID
func (a *Archive) setArchiveID(id string) {
	if id == "" {
		newID := NewArchiveID()
		a.ID = newID
	} else {
		a.ID = id
	}
}

// setTitle sets archive's title
// if title is empty uses fileName
func (a *Archive) setTitle(title string) {
	if title == "" {
		name := fsutil.GetNameFromPath(a.FilePath, true)
		a.Title = name
	} else {
		a.Title = title
	}
}

// setSummary sets archive's summary
// if passed summary is empty sets to nil
func (a *Archive) setSummary(summary string) {
	if summary != "" {
		a.Summary = &summary
	} else {
		a.Summary = nil
	}
}

// setLanguage sets archive's language
// if passed language is empty sets to nil
func (a *Archive) setLanguage(language string) {
	if language != "" {
		a.Language = &language
	} else {
		a.Language = nil
	}
}

// setCategory sets archive's category
// if passed category is empty sets to nil
func (a *Archive) setCategory(category string) {
	if category != "" {
		a.Category = &category
	} else {
		a.Category = nil
	}
}

// setPageCount sets archive's PageCount
func (a *Archive) setPageCount() {
	arch, _ := fsutil.OpenArchive(a.FilePath)
	pages, _ := arch.GetFileNames(true)
	a.PageCount = int16(len(pages))
}

// setFilePath sets archive's FilePath
func (a *Archive) setFilePath(path string) {
	a.FilePath = path
}

// setFileName sets archive's FileName
func (a *Archive) setFileName() {
	name := fsutil.GetNameFromPath(a.FilePath, false)
	a.FileName = name
}

// setHash sets archive's FileHash
func (a *Archive) setHash() {
	hash := fsutil.GenHash(a.FilePath)
	a.Hash = hash
}

// setThumbsPath sets archive's ThumbsPath
// creates if necessary
func (a *Archive) setThumbsPath() {
	d := fsutil.NewArchiveDir(a.app.Cfg.ThumbDir, a.Type, a.Hash)
	archiveDir, err := d.CreateDirs()
	if err != nil {
		a.app.Log.Error("error creating archive directories:", "err", err.Error(), "archive", a.ID)
		return
	}
	a.ThumbsPath = &archiveDir
	a.archiveDir = d
}

// CreateCoverDir is a helper function for
// creating the cover directory if it does
// not exist. This should almost never be
// necessary as it should already get created
// during setThumbsPath.
func (a *Archive) CreateCoverDir() error {
	if err := a.archiveDir.CreateCoverDir(*a.ThumbsPath); err != nil {
		return fmt.Errorf("failed to create cover directory for %v: %v", a.Type, a.ID)
	}
	return nil
}

// CreatePagesDir is a helper function for
// creating the cover directory if it does
// not exist. This should almost never be
// necessary as it should already get created
// during setThumbsPath.
func (a *Archive) CreatePagesDir() error {
	if err := a.archiveDir.CreatePagesDir(*a.ThumbsPath); err != nil {
		return fmt.Errorf("failed to create pages directory for %v: %v", a.Type, a.ID)
	}
	return nil
}

// setType sets archive's media type
func (a *Archive) setType() {
	a.Type = "archive"
}

// setCreatedAt sets time at which archive is created and put in db
func (a *Archive) setCreatedAt() {
	a.CreatedAt = time.Now()
}

// setUpdatedAt sets time at which archive is updated
func (a *Archive) setUpdatedAt() {
	a.UpdatedAt = time.Now()
}

// Get returns a filled in repository.Archive struct
func (a *Archive) Get() Archive {
	return *a
}

func (a *Archive) New(path string) {
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
	a.setCreatedAt()
	a.setUpdatedAt()
}

func (a *Archive) Update(id string, meta *models.Metadata) {
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
}

func (a *Archive) setTags(t *[]models.Tag) {
	a.Tags = t
}

func (a *Archive) setURL(u *[]models.URL) {
	a.URL = u
}

func (a *Archive) setParody(p *[]models.Parody) {
	a.Parody = p
}

func (a *Archive) setCharacter(c *[]models.Character) {
	a.Character = c
}

func (a *Archive) setArtist(b *[]models.Artist) {
	a.Artist = b
}

func (a *Archive) setReleaseDate(t *time.Time) {
	a.ReleaseDate = t
}
