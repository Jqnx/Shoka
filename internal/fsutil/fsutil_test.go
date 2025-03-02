package fsutil

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FSutilSuite struct {
	suite.Suite
	ctx context.Context
	dir string
}

var (
	imageExtensions   = []string{"png", "jpg", "jpeg", "gif", "webp"}
	archiveExtensions = []string{"zip", "cbz"}
)

func (suite *FSutilSuite) SetupSuite() {
	suite.ctx = context.Background()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	suite.dir = filepath.Join(dir, "content")
}

func (suite *FSutilSuite) TeardownSuite() {
}

func (suite *FSutilSuite) TestListArchives() {
	ListArchives(suite.dir)
}

//func (suite *FSutilSuite) TestArchiveContents() {
//	file := ""
//	path := filepath.Join(suite.dir, file)
//	ArchiveContents(path)
//}

func (suite *FSutilSuite) TestMatchExtensionsArchive() {
	// file := ""
	// path := filepath.Join(suite.dir, file)
	MatchExtension(suite.dir, archiveExtensions)
}

func (suite *FSutilSuite) TestMatchExtensionsImage() {
	file := ""
	path := filepath.Join(suite.dir, file)
	MatchExtension(path, imageExtensions)
}

func (suite *FSutilSuite) TestGetNameFromPath() {
	file := ""
	fn := filepath.Join(suite.dir, file)
	GetNameFromPath(fn, true)
}

func (suite *FSutilSuite) TestGetPageCount() {
	file := ""
	fn := filepath.Join(suite.dir, file)
	GetPageCount(fn, imageExtensions)
}

func (suite *FSutilSuite) TestCheck7z() {
	file := ""
	fn := filepath.Join(suite.dir, file)
	Is7z(fn)
}

func TestFSutilSuite(t *testing.T) {
	suite.Run(t, new(FSutilSuite))
}
