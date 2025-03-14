package metadata

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ComicInfoSuite struct {
	suite.Suite
	ctx context.Context
	dir string
}

func (suite *ComicInfoSuite) SetupSuite() {
	suite.ctx = context.Background()

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	suite.dir = filepath.Join(dir, "content")
}

func (suite *ComicInfoSuite) TeardownSuite() {
}

func (suite *ComicInfoSuite) TestUnmarshalXML() {
	// XML Test String here
	data := ``

	ComicInfoUnmarshal(data)
}

func TestComicInfoSuite(t *testing.T) {
	suite.Run(t, new(ComicInfoSuite))
}
