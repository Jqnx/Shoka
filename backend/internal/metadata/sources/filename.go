package sources

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"Shoka/internal/language"
	"Shoka/internal/metadata"

	"github.com/spf13/viper"
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

type FilenameSource struct{}

func NewFilenameSource() *FilenameSource {
	return &FilenameSource{}
}

func (s *FilenameSource) Name() string    { return "filename" }
func (s *FilenameSource) Priority() int   { return 3 }
func (s *FilenameSource) IsLocal() bool   { return true }

func (s *FilenameSource) Fetch(ctx context.Context, input metadata.Input) (*metadata.Result, error) {
	base := strings.TrimSuffix(filepath.Base(input.FilePath), filepath.Ext(input.FilePath))
	result := &metadata.Result{}

	base = numericBracketPattern.ReplaceAllString(base, "")
	base = strings.TrimSpace(eventPattern.ReplaceAllString(base, ""))
	base = strings.TrimSpace(sizePattern.ReplaceAllString(base, ""))

	// Match for [Circle (Artist)]
	if m := circleArtistPattern.FindStringSubmatch(base); m != nil {
		circle := strings.ToLower(strings.TrimSpace(m[1]))
		artist := strings.ToLower(strings.TrimSpace(m[2]))

		if circle != "" {
			circle = strings.Trim(circle, "[]() ")
			result.Circles = []string{circle}
		}

		if artist != "" {
			artist = strings.Trim(artist, "[]() ")
			result.Artists = []string{artist}
		}

		base = circleArtistPattern.ReplaceAllString(base, "")
	} else {
		// Match for [Artist]
		brackets := bracketPattern.FindAllStringSubmatch(base, -1)
		if len(brackets) > 0 {
			result.Artists = []string{strings.ToLower(strings.TrimSpace(brackets[0][1]))}

			lc := language.NewLanguageConverter()
			for _, part := range brackets {
				if lc.IsValidName(part[1]) {
					iso, _ := lc.ToISO(part[1])
					result.Language = &iso
				}
			}

		}

		base = bracketPattern.ReplaceAllString(base, "")
	}

	// Match for (Parody)
	parens := parenPattern.FindAllStringSubmatch(base, -1)
	if len(parens) > 0 {
		for _, m := range parens {
			candidate := strings.ToLower(strings.TrimSpace(m[1]))
			if !s.isMagazine(candidate) && !s.isMiscString(candidate) {
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

func (s *FilenameSource) isMagazine(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))

	for _, prefix := range knownMagazinePrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}

	for _, m := range viper.GetStringSlice("metadata.sources.filename.magazine_blocklist") {
		prefix := strings.ToLower(strings.TrimSpace(m))
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}

	return false
}

func (s *FilenameSource) isMiscString(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))

	for _, misc := range miscStrings {
		if strings.Contains(normalized, misc) {
			return true
		}
	}

	for _, misc := range viper.GetStringSlice("metadata.sources.filename.misc_blocklist") {
		contain := strings.ToLower(strings.TrimSpace(misc))
		if strings.Contains(normalized, contain) {
			return true
		}
	}

	return false
}
