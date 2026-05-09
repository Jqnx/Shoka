package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"Shoka/internal/language"
	"Shoka/internal/metadata"
	"golang.org/x/time/rate"
)

// API Docs: https://nhentai.net/api/v2/docs

const (
	nhentaiBaseURL    = "https://nhentai.net"
	nhentaiSearchURL  = "https://nhentai.net/api/v2/search"
	nhentaiGalleryURL = "https://nhentai.net/api/v2/galleries"
	nhentaiUserAgent  = "Shoka/1.0 (https://github.com/Jqnx/Shoka)"

	// Anonymous search endpoint: 10 requests/minute per IP
	nhentaiRateInterval = 6 * time.Second
)

type nhentaiSearchResult struct {
	Result   []nhentaiSearchResultItem `json:"result"`
	NumPages int                       `json:"num_pages"`
	PerPage  int                       `json:"per_page"`
	Total    int                       `json:"total"`
}

type nhentaiSearchResultItem struct {
	ID              int    `json:"id"`
	MediaID         string `json:"media_id"`
	EnglishTitle    string `json:"english_title"`
	JapaneseTitle   string `json:"japanese_title"`
	Thumbnail       string `json:"thumbnail"`
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
	NumPages        int    `json:"num_pages"`
	TagIDs          []int  `json:"tag_ids"`
}

type nhentaiGalleryDetail struct {
	ID         int          `json:"id"`
	MediaID    string       `json:"media_id"`
	Title      nhentaiTitle `json:"title"`
	UploadDate int64        `json:"upload_date"`
	Tags       []nhentaiTag `json:"tags"`
	NumPages   int64        `json:"num_pages"`
	Scanlator  string       `json:"scanlator"`
}

type nhentaiTitle struct {
	English  string `json:"english"`
	Japanese string `json:"japanese"`
	Pretty   string `json:"pretty"`
}

type nhentaiTag struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type NhentaiError struct {
	Error string `json:"error"`
}

type NHentaiSource struct {
	client  *http.Client
	apiKey  string
	limiter *rate.Limiter
}

func NewNHentaiSource(apiKey string) *NHentaiSource {
	return &NHentaiSource{
		client:  &http.Client{Timeout: 15 * time.Second},
		apiKey:  apiKey,
		limiter: rate.NewLimiter(rate.Every(nhentaiRateInterval), 1),
	}
}

func (s *NHentaiSource) Name() string    { return "nhentai" }
func (s *NHentaiSource) Priority() int   { return 10 }
func (s *NHentaiSource) IsLocal() bool   { return false }

// Search implements SearchableSource, returns all results for manual selection.
func (s *NHentaiSource) Search(ctx context.Context, input metadata.Input) ([]*metadata.SearchResult, error) {
	params := url.Values{}
	params.Set("query", input.Title)
	params.Set("sort", "date")

	resp, err := s.doRequest(ctx, nhentaiSearchURL+"?"+params.Encode())
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("nhentai rate limit exceeded")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nhentai search returned %d", resp.StatusCode)
	}

	var result nhentaiSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	candidates := make([]*metadata.SearchResult, 0, len(result.Result))
	for _, item := range result.Result {
		candidates = append(candidates, s.listItemToSearchResult(&item))
	}

	return candidates, nil
}

// FetchByID implements RemoteSource, fetches full details for a specific gallery ID.
func (s *NHentaiSource) FetchByID(ctx context.Context, id string) (*metadata.Result, error) {
	galleryID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid nhentai gallery id: %s", id)
	}
	return s.fetchGallery(ctx, galleryID)
}

// Fetch implements Source, picks the first result automatically.
func (s *NHentaiSource) Fetch(ctx context.Context, input metadata.Input) (*metadata.Result, error) {
	galleryID, err := s.search(ctx, input.Title)
	if err != nil {
		return nil, err
	}
	if galleryID == 0 {
		return nil, nil
	}
	return s.fetchGallery(ctx, galleryID)
}

// search queries nhentai search API for a gallery ID.
// Returns the first result.
func (s *NHentaiSource) search(ctx context.Context, title string) (int, error) {
	params := url.Values{}
	params.Set("query", title)
	params.Set("sort", "date")

	resp, err := s.doRequest(ctx, nhentaiSearchURL+"?"+params.Encode())
	if err != nil {
		return 0, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, fmt.Errorf("nhentai rate limit exceeded")
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("nhentai search returned %d", resp.StatusCode)
	}

	var result nhentaiSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode search response: %w", err)
	}

	if len(result.Result) == 0 {
		return 0, nil
	}

	return result.Result[0].ID, nil
}

// fetchGallery queries nhentai gallery API for a specific gallery's metadata.
func (s *NHentaiSource) fetchGallery(ctx context.Context, id int) (*metadata.Result, error) {
	resp, err := s.doRequest(ctx, fmt.Sprintf("%s/%d", nhentaiGalleryURL, id))
	if err != nil {
		return nil, fmt.Errorf("gallery request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("nhentai rate limit exceeded")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nhentai gallery returned %d", resp.StatusCode)
	}

	var gallery nhentaiGalleryDetail
	if err := json.NewDecoder(resp.Body).Decode(&gallery); err != nil {
		return nil, fmt.Errorf("decode gallery response: %w", err)
	}

	return s.toResult(&gallery), nil
}

// toResult converts nhentaiGalleryDetail to metadata.Result
func (s *NHentaiSource) toResult(g *nhentaiGalleryDetail) *metadata.Result {
	result := &metadata.Result{}

	// prefer english title, fall back to pretty, then japanese
	title := g.Title.English
	if title == "" {
		title = g.Title.Pretty
	}
	if title == "" {
		title = g.Title.Japanese
	}
	if title != "" {
		result.Title = &title
	}

	if g.NumPages > 0 {
		result.PageCount = &g.NumPages
	}

	if g.UploadDate > 0 {
		t := time.Unix(g.UploadDate, 0).UTC()
		result.ReleaseDate = &t
	}

	for _, tag := range g.Tags {
		name := tag.Name
		switch tag.Type {
		case "artist":
			result.Artists = append(result.Artists, name)
		case "group":
			result.Circles = append(result.Circles, name)
		case "parody":
			result.Parodies = append(result.Parodies, name)
		case "character":
			result.Characters = append(result.Characters, name)
		case "language":
			// normalise language tag — nhentai uses "english", "japanese" etc.
			lang := strings.ToLower(name)
			if lang != "translated" && lang != "" {
				lc := language.NewLanguageConverter()
				langISO, _ := lc.ToISO(lang)
				result.Language = &langISO
			} else {
				result.Language = nil
			}
		case "category":
			result.Category = &name
		case "tag":
			result.Tags = append(result.Tags, name)
		}
	}

	// NOTE: Don't know for sure yet if i'm keeping this

	// scanlator maps to circle if no group tag was present
	if g.Scanlator != "" && len(result.Circles) == 0 {
		result.Circles = []string{g.Scanlator}
	}

	return result
}

// doRequest waits for the rate limiter, executes the request, and retries once
// on HTTP 429 after honouring the Retry-After header (default 60 s).
func (s *NHentaiSource) doRequest(ctx context.Context, rawURL string) (*http.Response, error) {
	build := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		s.setHeaders(req)
		return req, nil
	}

	for attempt := 0; attempt < 2; attempt++ {
		if err := s.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		req, err := build()
		if err != nil {
			return nil, err
		}
		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests || attempt == 1 {
			return resp, nil
		}
		resp.Body.Close()

		wait := 60 * time.Second
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				wait = time.Duration(secs) * time.Second
			}
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
			timer.Stop()
		}
	}

	return nil, fmt.Errorf("nhentai: unexpected retry loop exit")
}

// setHeaders sets the user agent and API key headers for the request
func (s *NHentaiSource) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", nhentaiUserAgent)
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Key "+s.apiKey)
	}
}

// listItemToSearchResult converts a nhentaiSearchResultItem to a SearchResult
func (s *NHentaiSource) listItemToSearchResult(item *nhentaiSearchResultItem) *metadata.SearchResult {
	title := item.EnglishTitle
	if title == "" {
		title = item.JapaneseTitle
	}

	return &metadata.SearchResult{
		ID:        strconv.Itoa(item.ID),
		Title:     title,
		CoverURL:  item.Thumbnail,
		PageCount: item.NumPages,
	}
}
