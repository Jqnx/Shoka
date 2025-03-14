package store

import (
	"Shoka/internal/fsutil"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Models
// Archive
type TestArchive struct {
	gorm.Model
	Title     string     `json:"title" gorm:"unique"`
	Summary   string     `json:"summary"`
	Tags      TestTag    `json:"tags" gorm:"many2many:archive_tags;References:Name"`
	Artist    TestArtist `json:"artist" gorm:"many2many:archive_artists;references:Name"`
	Parody    string     `json:"parody"`
	Character string     `json:"character"`
	Language  string     `json:"language"`
	Category  string     `json:"category"`
	PageCount int        `json:"pagecount"`
	URL       string     `json:"URL"`
	FilePath  string     `json:"filepath"`
	AID       int        `gorm:"uniqueIndex;default:1"`
}

type TestAIDSearch struct {
	AID int `json:"archive_id"`
}

// Tags
type TestTag struct {
	Name string `gorm:"many2many:archive_tags;primaryKey;not null" json:"tag" form:"tag"` // Tag name
}

// Artist
type TestArtist struct {
	gorm.Model
	Name  string    `gorm:"many2many:archive_artists;uniqueIndex;not null"`
	Alias TestAlias `gorm:"foreignKey:AliasID" json:"aliases" form:"aliases"`
	Group TestGroup `gorm:"many2many:artist_groups;references:Name" json:"groups" form:"groups"`
	Links TestLinks `gorm:"foreignKey:LinkID" json:"links" form:"links"`
}

type TestAlias struct {
	AliasID uint   `gorm:"primaryKey"`
	Alias   string `gorm:"uniqueIndex" json:"alias" form:"alias"`
}

type TestLinks struct {
	LinkID uint   `gorm:"primaryKey"`
	Link   string `gorm:"unique" json:"link" form:"link"`
}

// Group
type TestGroup struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"many2many:artist_groups;uniqueIndex;not null" json:"group" form:"group"`
}

type ArchiveSuite struct {
	suite.Suite
	pgContainer *PostgresContainer
	repository  *Repository
	ctx         context.Context
}

// Setup test suite
func (suite *ArchiveSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := createPostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}

	suite.pgContainer = pgContainer
	repository, err := NewRepo(suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	suite.repository = repository

	fmt.Println("connection to database established")

	// Migrations
	err = repository.db.AutoMigrate(&TestArchive{}, &TestArtist{}, &TestAlias{}, &TestLinks{}, &TestGroup{}, &TestTag{})
	suite.Require().NoError(err, "Error auto-migrating database tables")
}

// Teardown test suite
func (suite *ArchiveSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

// Test the creating of an archive
//func (suite *ArchiveSuite) TestCreateArchive() {
//	testtags := Tag{}
//
//	testartist := TestArtist{
//		Name: "",
//		Alias: TestAlias{
//			Alias: "",
//		},
//		Group: TestGroup{
//			Name: "",
//		},
//		Links: TestLinks{
//			Link: "",
//		},
//	}
//
//	testarchive := TestArchive{
//		Title:     "",
//		Summary:   "",
//		Tags:      testtags,
//		Artist:    testartist,
//		Parody:    "",
//		Language:  "",
//		Category:  "",
//		Character: "",
//		URL:       "",
//		PageCount: 59,
//		FilePath:  "test/path",
//	}
//
//	err := suite.repository.db.Create(&testarchive).Error
//	if err != nil {
//		log.Fatalf("error creating archive: %s", err)
//	}
//
//	var retrievedArchive TestArchive
//	err = suite.repository.db.First(&retrievedArchive, "title = ?", "").Error
//	if err != nil {
//		log.Fatalf("error retrieving archive: %s", err)
//	}
//
//	suite.Equal(testarchive.Title, retrievedArchive.Title, "Names should match")
//}

func (suite *ArchiveSuite) TestCreateFromFile() {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	d := filepath.Join(wd, "content")

	// Read directory contents
	dir, err := os.ReadDir(d)

	// Add contents to database
	for _, file := range dir {
		// Matches for correct archive extensions
		if fsutil.MatchExtension(file.Name(), archiveExtensions) {
			fullpath := filepath.Join(d, file.Name())
			title := fsutil.GetNameFromPath(file.Name(), true)
			pagecount := fsutil.GetPageCount(fullpath, imageExtensions)
			archive := &TestArchive{
				Title:     title,
				FilePath:  file.Name(),
				PageCount: pagecount,
			}
			suite.repository.db.Create(archive)
		} else {
			log.Printf("File: %v is not a supported archive.", file.Name())
		}
	}
}

// Test the updating of an archive
func (suite *ArchiveSuite) TestUpdateArchive() {
	testtags := TestTag{
		Name: "",
	}

	testartist := TestArtist{
		Name: "",
		Alias: TestAlias{
			Alias: "",
		},
		Group: TestGroup{
			Name: "",
		},
		Links: TestLinks{
			Link: "",
		},
	}

	testarchive := TestArchive{
		Title:     "",
		Summary:   "test",
		Tags:      testtags,
		Artist:    testartist,
		Parody:    "",
		Language:  "",
		Category:  "",
		Character: "",
		URL:       "",
		PageCount: 183,
		FilePath:  "test/newpath",
	}

	err := suite.repository.db.Where("title = ?", "").Updates(&testarchive).Error
	if err != nil {
		log.Fatalf("error creating archive: %s", err)
	}

	var retrievedArchive TestArchive
	err = suite.repository.db.First(&retrievedArchive, "title = ?", "").Error
	if err != nil {
		log.Fatalf("error retrieving archive: %s", err)
	}

	suite.Equal(testarchive.Title, retrievedArchive.Title, "Names should match")
}

// Run the test suite
func TestArchiveSuite(t *testing.T) {
	suite.Run(t, new(ArchiveSuite))
}
