package metadata

import (
	"net/url"
	"strings"
)

// hostToSource maps a known gallery host to the metadata source name that
// owns it, matching the corresponding Source.Name(). Hosts not listed here
// fall back to the bare host (see SourceFromURL).
var hostToSource = map[string]string{
	"nhentai.net":  "nhentai",
	"e-hentai.org": "e-hentai",
	"exhentai.org": "e-hentai",
}

// SourceFromURL maps a URL to the metadata source that owns it, so a link's
// icon reflects where the gallery actually lives rather than which pipeline
// source happened to report it (a ComicInfo <Web> nhentai link is 'nhentai').
//
// Returns ok == false for anything that isn't a parseable http(s) URL, so the
// caller can drop it. An unknown but otherwise valid host returns the bare
// host as the source name with ok == true — the frontend renders unknown
// source names with a generic globe icon.
func SourceFromURL(raw string) (source string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}

	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")

	if host == "" {
		return "", false
	}

	if name, known := hostToSource[host]; known {
		return name, true
	}

	return host, true
}

// NormalizeURLs trims each entry, drops any that aren't valid http(s) URLs,
// and dedupes by URL string while preserving first-seen order. Used by both
// the pipeline merge and the PATCH handler so they agree on what a URL set
// is.
func NormalizeURLs(urls []string) []string {
	if urls == nil {
		return nil
	}

	seen := make(map[string]struct{}, len(urls))
	out := make([]string, 0, len(urls))

	for _, raw := range urls {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		if _, ok := SourceFromURL(trimmed); !ok {
			continue
		}

		if _, dup := seen[trimmed]; dup {
			continue
		}

		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}

	return out
}
