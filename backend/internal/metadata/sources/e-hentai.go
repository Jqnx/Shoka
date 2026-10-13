package sources

import (
	"Shoka/internal/metadata"
	"context"
	"net/http"
	"time"
)

// EHentaiSource scrapes gallery metadata from the e-hentai API, authenticating
// with the per-library cookies configured by the user. The same gallery is
// reachable through the exhentai.org mirror.
type EHentaiSource struct {
	client *http.Client
	apiURL string
}

// NewEHentaiSource returns an EHentaiSource with a 15 s timeout HTTP client
// pointed at the public e-hentai API endpoint.
func NewEHentaiSource() *EHentaiSource {
	return &EHentaiSource{
		client: &http.Client{Timeout: 15 * time.Second},
		apiURL: "https://api.e-hentai.org/api.php",
	}
}

// Name returns the stable source identifier, "e-hentai".
func (s *EHentaiSource) Name() string { return "e-hentai" }

// Priority returns 11: a title search is less precise than an ID-based match.
func (s *EHentaiSource) Priority() int { return 11 }

// IsLocal reports false: the source needs network access to the e-hentai API.
func (s *EHentaiSource) IsLocal() bool { return false }

// Hosts implements metadata.HostAware - a gallery can live on either mirror.
func (s *EHentaiSource) Hosts() []string {
	return []string{"e-hentai.org", "exhentai.org"}
}

// Fetch searches e-hentai by the input title and, when a gallery is found,
// loads its full metadata. A missing gallery yields a nil result and nil error
// so later sources can run.
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

// search resolves a gallery ID and token from a title. It is not implemented
// yet; the signature is kept so Fetch can wire it in once the API client lands.
func (s *EHentaiSource) search(_ context.Context, _, _ string) (string, string, error) {
	// implementation depends on e-hentai search API
	// returns galleryID and token on success
	return "", "", nil
}

// fetchGallery loads one gallery's metadata from its ID and token. It is not
// implemented yet; the signature is kept so Fetch can wire it in later.
func (s *EHentaiSource) fetchGallery(_ context.Context, _, _, _ string) (*metadata.Result, error) {
	// call e-hentai API with gallery id+token
	// parse response into metadata.Result
	return nil, nil
}
