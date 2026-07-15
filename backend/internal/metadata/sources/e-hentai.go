package sources

import (
	"Shoka/internal/metadata"
	"context"
	"net/http"
	"time"
)

type EHentaiSource struct {
	client *http.Client
	apiURL string
}

func NewEHentaiSource() *EHentaiSource {
	return &EHentaiSource{
		client: &http.Client{Timeout: 15 * time.Second},
		apiURL: "https://api.e-hentai.org/api.php",
	}
}

func (s *EHentaiSource) Name() string  { return "e-hentai" }
func (s *EHentaiSource) Priority() int { return 11 }
func (s *EHentaiSource) IsLocal() bool { return false }

func (s *EHentaiSource) Fetch(ctx context.Context, input metadata.Input) (*metadata.Result, error) {
	// search by title, authenticating with this library's configured cookies
	galleryID, galleryToken, err := s.search(ctx, input.Title, input.SourceConfig.Cookies)
	if err != nil {
		return nil, err
	}
	if galleryID == "" {
		return nil, nil // not found
	}

	return s.fetchGallery(ctx, galleryID, galleryToken, input.SourceConfig.Cookies)
}

func (s *EHentaiSource) search(ctx context.Context, title, cookies string) (string, string, error) {
	// implementation depends on e-hentai search API
	// returns galleryID and token on success
	return "", "", nil
}

func (s *EHentaiSource) fetchGallery(ctx context.Context, id, token, cookies string) (*metadata.Result, error) {
	// call e-hentai API with gallery id+token
	// parse response into metadata.Result
	return nil, nil
}
