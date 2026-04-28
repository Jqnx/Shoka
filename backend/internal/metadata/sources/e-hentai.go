package sources

import (
	"Shoka/internal/metadata"
	"context"
	"net/http"
	"time"
)

type EHentaiSource struct {
	client  *http.Client
	apiURL  string
	cookies string
}

func NewEHentaiSource(cookies string) *EHentaiSource {
	return &EHentaiSource{
		client:  &http.Client{Timeout: 15 * time.Second},
		apiURL:  "https://api.e-hentai.org/api.php",
		cookies: cookies,
	}
}

func (s *EHentaiSource) Name() string  { return "e-hentai" }
func (s *EHentaiSource) Priority() int { return 11 }

func (s *EHentaiSource) Fetch(ctx context.Context, input metadata.Input) (*metadata.Result, error) {
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

func (s *EHentaiSource) search(ctx context.Context, title string) (string, string, error) {
	// implementation depends on e-hentai search API
	// returns galleryID and token on success
	return "", "", nil
}

func (s *EHentaiSource) fetchGallery(ctx context.Context, id, token string) (*metadata.Result, error) {
	// call e-hentai API with gallery id+token
	// parse response into metadata.Result
	return nil, nil
}
