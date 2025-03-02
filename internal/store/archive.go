package store

import (
	"Shoka/internal/fsutil"
	"Shoka/internal/metadata"
	"Shoka/internal/models"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TODO:
// 1. use unmarshalled data to create file with metadata
// 2. Make Update function
// 3. Add Metadata Function
// 4. Move migrations to its own thing

type ArchiveStore struct {
	db *gorm.DB
}

func (s *ArchiveStore) Migrate() error {
	err := s.db.AutoMigrate(
		&models.Archive{},
		&models.Character{},
		&models.Parody{},
		&models.URL{})
	if err != nil {
		return err
	}
	return nil
}

// Create

func (s *ArchiveStore) Create(ctx context.Context, a *models.Archive) error {
	err, aid := s.GetLastID()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err := s.db.WithContext(ctx).Create(a).Error
		if err != nil {
			log.Printf("error creating archive: %v", err)
			return err
		}
	} else {
		a.AID = aid.AID + 1
		// err := s.db.Debug().WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(a).Error
		err := s.db.WithContext(ctx).FirstOrCreate(a, models.Archive{Title: a.Title}).Error
		if err != nil {
			log.Printf("error creating archive: %v", err)
			return err
		}

	}

	return nil
}

func (s *ArchiveStore) CreateFromFile() error {
	ctx := context.Background()
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	d := filepath.Join(wd, "content")

	// Read directory contents
	dir, err := os.ReadDir(d)

	// TODO:
	// 1. Fix updates of database records

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
				s.Create(ctx, archive)
			} else {
				archive := &models.Archive{
					Title:     title,
					FilePath:  file.Name(),
					PageCount: pagecount,
				}
				s.Create(ctx, archive)
			}

		} else {
			log.Printf("File: %v is not a supported archive.", file.Name())
		}
	}
	return nil
}

// Read
func (s *ArchiveStore) GetAll() (error, *[]models.ArchiveSearch) {
	archive := []models.Archive{}
	err := s.db.Debug().Find(&archive).Error
	if err != nil {
		log.Printf("error getting archives: %v", err)
		return err, nil
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

	return nil, &archives
}

func (s *ArchiveStore) Get(id string) (error, *models.ArchiveSearch) {
	var archive models.Archive
	err := s.db.Where("a_id = ?", id).Find(&archive).Error
	if err != nil {
		log.Printf("error finding archive: %v", err)
		return err, nil
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

	return nil, search
}

func (s *ArchiveStore) GetLastID() (error, *models.AIDSearch) {
	var archive models.Archive
	var aidsearch models.AIDSearch
	err := s.db.Model(archive).Last(&aidsearch).Error

	return err, &aidsearch
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

// TODO:
// 1. Update worker?

// Update

func (s *ArchiveStore) Update(ctx context.Context, a *models.Archive) error {
	err := s.db.Debug().Clauses(clause.OnConflict{DoNothing: true}).Save(a).Error
	if err != nil {
		log.Printf("error updating archive: %v", err)
		return err
	}
	return nil
}
