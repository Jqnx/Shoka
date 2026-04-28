package sources

import (
	"Shoka/internal/metadata"
	"context"
	"net/http"
	"time"
)

// API Docs: https://nhentai.net/api/v2/docs

type NHentaiSource struct {
	client *http.Client
	apiURL string
}

func NewNHentaiSource() *NHentaiSource {
	return &NHentaiSource{
		client: &http.Client{Timeout: 15 * time.Second},
		apiURL: "",
	}
}

func (s *NHentaiSource) Name() string  { return "nhentai" }
func (s *NHentaiSource) Priority() int { return 10 }

func (s *NHentaiSource) Fetch(ctx context.Context, input metadata.Input) (*metadata.Result, error) {
	// search by title
	galleryID, galleryToken, err := s.search(ctx, input.Title)
	if err != nil {
		return nil, err
	}
	if galleryID == "" {
		return nil, nil // not found
	}

	return s.fetchGallery(ctx, galleryID, galleryToken)
}

func (s *NHentaiSource) search(ctx context.Context, title string) (string, string, error) {
	// implementation depends on nhentai search API
	// returns galleryID and token on success
	return "", "", nil
}

func (s *NHentaiSource) fetchGallery(ctx context.Context, id, token string) (*metadata.Result, error) {
	// call nhentai API with gallery id+token
	// parse response into metadata.Result
	return nil, nil
}
