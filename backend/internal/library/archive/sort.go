package archive

import (
	"path/filepath"
	"sort"

	"Shoka/internal/util"
)

// sortPages sorts pages into natural reading order by filename.
func sortPages(pages []Page) {
	sort.Slice(pages, func(i, j int) bool {
		return util.NaturalLess(
			filepath.Base(pages[i].Filename),
			filepath.Base(pages[j].Filename),
		)
	})

	for i := range pages {
		pages[i].Index = i
	}
}
