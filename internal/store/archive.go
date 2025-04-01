package store

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"Shoka/internal/fsutil"
	"Shoka/internal/metadata"
	"Shoka/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TODO:
// 1. use unmarshalled data to create file with metadata
// 3. Add Metadata Function
// 4. Move migrations to its own thing

type ArchiveStore struct {
	db *gorm.DB
}

// Create

func (s *ArchiveStore) Create(a *models.Archive) error {
	ctx := context.Background()
	aid, err := s.GetLastID()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err := s.db.WithContext(ctx).Create(a).Error
		if err != nil {
			log.Printf("error creating archive: %v", err)
			return err
		}
	} else {
		a.AID = aid.AID + 1
		// err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&a).Error
		err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).FirstOrCreate(&a, models.Archive{Title: a.Title}).Error
		if err != nil {
			log.Printf("error creating archive: %v", err)
			return err
		}

	}

	return nil
}

func (s *ArchiveStore) CreateFromFile() error {
	// ctx := context.Background()
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	d := filepath.Join(wd, "content")

	// Read directory contents
	dir, err := os.ReadDir(d)
	if err != nil {
		return err
	}

	// TODO:
	// 1. Fix updates of database records on scan

	// Add contents to database
	for _, file := range dir {
		// Matches for correct archive extensions
		if fsutil.MatchExtension(file.Name(), archiveExtensions) {
			fullpath := filepath.Join(d, file.Name())
			title := fsutil.GetNameFromPath(file.Name(), true)
			pagecount := fsutil.GetPageCount(fullpath, imageExtensions)
			hasMetadata, metadata, err := metadata.GetMetadata(fullpath)
			if err != nil {
				return err
			}
			if hasMetadata {
				archive := &models.Archive{
					Title:     metadata.Title,
					Summary:   metadata.Summary,
					Tags:      metadata.Tags,
					Artist:    metadata.Artist,
					Parody:    metadata.Parody,
					Character: metadata.Character,
					Language:  metadata.Language,
					Category:  metadata.Category,
					PageCount: pagecount,
					URL:       metadata.URL,
					FilePath:  file.Name(),
				}
				s.Create(archive)
			} else {
				archive := &models.Archive{
					Title:     title,
					FilePath:  file.Name(),
					PageCount: pagecount,
				}
				s.Create(archive)
			}

		} else {
			log.Printf("File: %v is not a supported archive.", file.Name())
		}
	}
	return nil
}

// Read
func (s *ArchiveStore) GetAll() (*[]models.ArchiveSearch, error) {
	archive := []models.Archive{}
	err := s.db.Find(&archive).Error
	if err != nil {
		log.Printf("error getting archives: %v", err)
		return nil, err
	}

	archives := []models.ArchiveSearch{}

	for _, item := range archive {
		tags := s.GetTagList(&item)
		artists := s.GetArtistList(&item)
		urls := s.GetUrlList(&item)
		parodies := s.GetParodyList(&item)
		chars := s.GetCharacterList(&item)

		model := &models.ArchiveSearch{
			AID:        item.AID,
			CreatedAt:  item.CreatedAt,
			Title:      item.Title,
			Summary:    item.Summary,
			Tags:       tags,
			Artists:    artists,
			Parodies:   parodies,
			Characters: chars,
			Language:   item.Language,
			Category:   item.Category,
			Urls:       urls,
			PageCount:  item.PageCount,
		}

		archives = append(archives, *model)
	}

	return &archives, nil
}

func (s *ArchiveStore) Get(id int) (*models.ArchiveSearch, error) {
	var archive models.Archive
	err := s.db.Where("a_id = ?", id).Find(&archive).Error
	if err != nil {
		log.Printf("error finding archive: %v", err)
		return nil, err
	}

	artists := s.GetArtistList(&archive)
	urls := s.GetUrlList(&archive)
	tags := s.GetTagList(&archive)
	parodies := s.GetParodyList(&archive)
	chars := s.GetCharacterList(&archive)

	search := &models.ArchiveSearch{
		AID:        archive.AID,
		CreatedAt:  archive.CreatedAt,
		Title:      archive.Title,
		Summary:    archive.Summary,
		Tags:       tags,
		Artists:    artists,
		Parodies:   parodies,
		Characters: chars,
		Language:   archive.Language,
		Category:   archive.Category,
		Urls:       urls,
		PageCount:  archive.PageCount,
	}

	return search, nil
}

func (s *ArchiveStore) GetID(aid int) int {
	var archive models.Archive
	s.db.Where("a_id = ?", aid).Find(&archive)

	return archive.ID
}

func (s *ArchiveStore) GetLastID() (*models.AIDSearch, error) {
	var archive models.Archive
	var aidsearch models.AIDSearch
	err := s.db.Model(archive).Last(&aidsearch).Error

	return &aidsearch, err
}

func (s *ArchiveStore) GetArtistList(archive *models.Archive) []string {
	artists := []models.Artist{}
	err := s.db.Model(archive).Association("Artist").Find(&artists)
	if err != nil {
		log.Printf("error getting archive artists: %v", err)
		return nil
	}

	list := []string{}
	for _, artist := range artists {
		list = append(list, artist.Name)
	}

	return list
}

func (s *ArchiveStore) GetUrlList(archive *models.Archive) []string {
	urls := []models.URL{}
	err := s.db.Model(archive).Association("URL").Find(&urls)
	if err != nil {
		log.Printf("error getting archive urls: %v", err)
		return nil
	}

	list := []string{}
	for _, url := range urls {
		list = append(list, url.Url)
	}

	return list
}

func (s *ArchiveStore) GetTagList(archive *models.Archive) []string {
	tags := []models.Tag{}
	err := s.db.Model(archive).Association("Tags").Find(&tags)
	if err != nil {
		log.Printf("error getting archive tags: %v", err)
		return nil
	}

	list := []string{}
	for _, tag := range tags {
		list = append(list, tag.Name)
	}

	return list
}

func (s *ArchiveStore) GetParodyList(archive *models.Archive) []string {
	parodies := []models.Parody{}
	err := s.db.Model(archive).Association("Parody").Find(&parodies)
	if err != nil {
		log.Printf("error getting archive parodies: %v", err)
		return nil
	}

	list := []string{}
	for _, parody := range parodies {
		list = append(list, parody.Name)
	}

	return list
}

func (s *ArchiveStore) GetCharacterList(archive *models.Archive) []string {
	chars := []models.Character{}
	err := s.db.Model(archive).Association("Character").Find(&chars)
	if err != nil {
		log.Printf("error getting archive characters: %v", err)
		return nil
	}

	list := []string{}
	for _, char := range chars {
		list = append(list, char.Name)
	}

	return list
}

func (s *ArchiveStore) TitleExists(archive *models.Archive) bool {
	newarchive := models.Archive{}
	s.db.Where("title = ?", archive.Title).Find(&newarchive)
	return newarchive.Title == archive.Title
}

// TODO:
// 1. Update worker?

// Update

func (s *ArchiveStore) Update(a *models.Archive) error {
	// Update or create Title, Summary, Language and Category
	if err := s.db.Select("title", "summary", "language", "category").Save(a).Error; err != nil {
		return err
	}
	// Replace old Tags with new Tags
	if err := s.db.Model(a).Association("Tags").Replace(a.Tags); err != nil {
		return err
	}
	// Replace old Artists with new Artists
	if err := s.db.Model(a).Association("Artist").Replace(a.Artist); err != nil {
		return err
	}
	// Replace old Parodies with new Parodies
	if err := s.db.Model(a).Association("Parody").Replace(a.Parody); err != nil {
		return err
	}
	// Replace old Characters with new Characters
	if err := s.db.Model(a).Association("Character").Replace(a.Character); err != nil {
		return err
	}
	// Replace old URLs with new URLs
	if err := s.db.Model(a).Association("URL").Replace(a.URL); err != nil {
		return err
	}

	return nil
}

// Delete

func (s *ArchiveStore) Delete(a *models.Archive) error {
	err := s.db.Delete(a).Error
	if err != nil {
		return err
	}
	return nil
}
