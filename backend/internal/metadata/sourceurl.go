package metadata

import (
	"Shoka/internal/util"
	"net/url"
	"strings"
)

// hostToSource maps a known gallery host to the metadata source name that
// owns it, matching the corresponding Source.Name(). It starts empty and is
// populated explicitly by NewPipeline (see HostAware), which registers the
// hosts of every source it's constructed with - metadata can't import the
// sources package directly (it would be a cycle, since sources imports
// metadata for the Source interface), so this is how the two stay in sync
// instead of a second hardcoded map here. Deliberately not populated via an
// init() side effect: that would make SourceFromURL's behavior depend on
// which packages happen to be linked in, rather than on which sources the
// running pipeline actually has. Hosts not registered fall back to the bare
// host (see SourceFromURL).
var hostToSource = map[string]string{}

// HostAware is implemented by a Source whose gallery URLs should be
// recognized by SourceFromURL/NormalizeURLs. NewPipeline registers the hosts
// of every HostAware source it's given.
type HostAware interface {
	// Hosts returns the gallery hostnames this source owns (e.g.
	// "nhentai.net"), matching Source.Name().
	Hosts() []string
}

// RegisterSourceHost associates one or more gallery hostnames with the
// metadata source name that owns them. Called by NewPipeline for each
// HostAware source - not intended to be called after startup.
func RegisterSourceHost(name string, hosts ...string) {
	for _, host := range hosts {
		hostToSource[strings.ToLower(host)] = name
	}
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
// and dedupes by a canonicalized key (scheme/case/"www."/trailing-slash
// insensitive - see canonicalURLKey) while preserving first-seen order. Used
// by both the pipeline merge and the PATCH handler so they agree on what a
// URL set is.
func NormalizeURLs(urls []string) []string {
	if urls == nil {
		return nil
	}

	trimmed := util.DedupeTrimmed(urls)

	seen := make(map[string]struct{}, len(trimmed))
	out := make([]string, 0, len(trimmed))

	for _, raw := range trimmed {
		if _, ok := SourceFromURL(raw); !ok {
			continue
		}

		key := canonicalURLKey(raw)
		if _, dup := seen[key]; dup {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, raw)
	}

	return out
}

// canonicalURLKey reduces a URL to the parts that identify the same gallery
// regardless of scheme, host case, a "www." prefix, or a trailing slash, so
// e.g. "http://nhentai.net/g/1" and "https://www.nhentai.net/g/1/" dedupe as
// the same link. Falls back to the raw string for anything unparseable so it
// still participates in dedup rather than always surviving as unique.
func canonicalURLKey(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	path := strings.TrimSuffix(u.Path, "/")

	return host + path + "?" + u.RawQuery
}
