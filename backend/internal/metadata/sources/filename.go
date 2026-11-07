package sources

import (
	"Shoka/internal/language"
	"Shoka/internal/metadata"
	"context"
	"path/filepath"
	"regexp"
	"strings"
)

var knownMagazinePrefixes = []string{
	"angel club",
	"comic lo",
	"comic hotmilk",
	"comic kairakuten",
	"comic megastore",
	"comic bavel",
	"comic exe",
	"comic anthurium",
	"comic unreal",
	"comic masyo",
	"comic tenma",
	"comic gamma",
	"comic europa",
	"comic koh",
	"comic moemax",
	"comic penguin club",
	"comic rin",
	"comic x-eros",
	"comic gucho",
	"comic aun",
	"comic saseco",
	"comic shitsurakuten",
	"comic aoha",
	"monthly comic dengeki daioh",
	"young animal",
	"young magazine",
	"sunday gene-x",
	"action pizazz",
	"megastore alpha",
	"weekly kairakuten",
	"2d dream",
	"haiboku otome ecstasy",
}

var miscStrings = []string{
	"zenpen",
	"kouhen",
	"saimin",
	"hypnotized",
}

// TODO:
// 1. Add book/tankoubon support

// matches patterns like:
// [Artist] Title (Parody) [Circle]
// (Circle) [Artist] Title [Tag1][Tag2]
// [Circle (Artist)] Title (Parody)
var (
	bracketPattern        = regexp.MustCompile(`\[([^\]]+)\]`)
	parenPattern          = regexp.MustCompile(`\(([^)]+)\)`)
	curlyPattern          = regexp.MustCompile(`\{([^}]*)\}`)
	circleArtistPattern   = regexp.MustCompile(`\[([^\(^\]]+)\s*\(([^)]+)\)\s*\]`)
	numericBracketPattern = regexp.MustCompile(`^\[(\d{3,})\]\s*`)
	eventPattern          = regexp.MustCompile(`(?i)\((C\d+|CT\d+|COMIC1[^)]*|Comiket\s*\d+)\)`)
	sizePattern           = regexp.MustCompile(`\([x]\d+\)\s*`)
)

// FilenameSource extracts artist, circle, parody and title information from
// the release filename using bracket/parenthesis heuristics. It is a
// best-effort fallback for archives without an embedded ComicInfo.xml.
type FilenameSource struct{}

// NewFilenameSource returns a FilenameSource with no external dependencies.
func NewFilenameSource() *FilenameSource {
	return &FilenameSource{}
}

// Name returns the stable source identifier, "filename".
func (s *FilenameSource) Name() string { return "filename" }

// Priority returns 3: heuristics rank below an explicit ComicInfo.xml.
func (s *FilenameSource) Priority() int { return 3 }

// IsLocal reports true: the source only inspects the input filename.
func (s *FilenameSource) IsLocal() bool { return true }

// Fetch parses the input filename into metadata, stripping magazine, event and
// resolution noise before splitting the remainder into title, artist, circle
// and parodies. It always returns a non-nil result and never fails.
func (s *FilenameSource) Fetch(_ context.Context, input metadata.Input) (*metadata.Result, error) {
	base := strings.TrimSuffix(filepath.Base(input.FilePath), filepath.Ext(input.FilePath))
	result := &metadata.Result{}

	base = numericBracketPattern.ReplaceAllString(base, "")
	base = strings.TrimSpace(eventPattern.ReplaceAllString(base, ""))
	base = strings.TrimSpace(sizePattern.ReplaceAllString(base, ""))

	// Match for [Circle (Artist)]
	if m := circleArtistPattern.FindStringSubmatch(base); m != nil {
		applyCircleArtist(result, m)

		base = circleArtistPattern.ReplaceAllString(base, "")
	} else {
		base = applyArtistBrackets(result, base)
	}

	// Match for (Parody)
	parens := parenPattern.FindAllStringSubmatch(base, -1)
	if len(parens) > 0 {
		for _, m := range parens {
			candidate := strings.ToLower(strings.TrimSpace(m[1]))
			if !s.isMagazine(candidate, input.SourceConfig.MagazineBlocklist) && !s.isMiscString(candidate, input.SourceConfig.MiscBlocklist) {
				result.Parodies = []string{candidate}
				break
			}
		}
	}

	// Get title by removing every other pattern
	title := bracketPattern.ReplaceAllString(base, "")
	title = parenPattern.ReplaceAllString(title, "")
	title = curlyPattern.ReplaceAllString(title, "")

	title = strings.Trim(title, "[]() ")
	if title != "" {
		result.Title = &title
	}

	return result, nil
}

// applyCircleArtist records a "[Circle (Artist)]" match, lower-casing and
// trimming the bracket/paren residue each capture may still carry. An empty
// capture leaves its field unset.
func applyCircleArtist(result *metadata.Result, m []string) {
	circle := strings.ToLower(strings.TrimSpace(m[1]))
	if circle != "" {
		result.Circles = []string{strings.Trim(circle, "[]() ")}
	}

	artist := strings.ToLower(strings.TrimSpace(m[2]))
	if artist != "" {
		result.Artists = []string{strings.Trim(artist, "[]() ")}
	}
}

// applyArtistBrackets handles the "[Artist] [Language]" fallback: the first
// bracket names the artist and any bracket naming a language sets the
// language field. It returns base with every bracket stripped, which is what
// the subsequent parody/title parsing expects.
func applyArtistBrackets(result *metadata.Result, base string) string {
	brackets := bracketPattern.FindAllStringSubmatch(base, -1)
	if len(brackets) > 0 {
		result.Artists = []string{strings.ToLower(strings.TrimSpace(brackets[0][1]))}

		lc := language.NewConverter()
		for _, part := range brackets {
			if lc.IsValidName(part[1]) {
				iso, _ := lc.ToISO(part[1])
				result.Language = &iso
			}
		}
	}

	return bracketPattern.ReplaceAllString(base, "")
}

func (s *FilenameSource) isMagazine(name string, blocklist []string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))

	for _, prefix := range knownMagazinePrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}

	for _, m := range blocklist {
		prefix := strings.ToLower(strings.TrimSpace(m))
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}

	return false
}

func (s *FilenameSource) isMiscString(name string, blocklist []string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))

	for _, misc := range miscStrings {
		if strings.Contains(normalized, misc) {
			return true
		}
	}

	for _, misc := range blocklist {
		contain := strings.ToLower(strings.TrimSpace(misc))
		if strings.Contains(normalized, contain) {
			return true
		}
	}

	return false
}
