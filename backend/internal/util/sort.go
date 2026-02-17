package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

func NaturalSort(files []string) {
	re := regexp.MustCompile(`\d+`)

	sort.Slice(files, func(i, j int) bool {
		return naturalLess(files[i], files[j], re)
	})
}

func naturalLess(a, b string, re *regexp.Regexp) bool {
	// Find all number sequences
	numsA := re.FindAllStringIndex(a, -1)
	numsB := re.FindAllStringIndex(b, -1)

	posA, posB := 0, 0

	for len(numsA) > 0 && len(numsB) > 0 {
		// Compare text before numbers
		if a[posA:numsA[0][0]] != b[posB:numsB[0][0]] {
			return a[posA:numsA[0][0]] < b[posB:numsB[0][0]]
		}

		// Compare numbers numerically
		numA, _ := strconv.Atoi(a[numsA[0][0]:numsA[0][1]])
		numB, _ := strconv.Atoi(b[numsB[0][0]:numsB[0][1]])

		if numA != numB {
			return numA < numB
		}

		posA, posB = numsA[0][1], numsB[0][1]
		numsA, numsB = numsA[1:], numsB[1:]
	}

	return a[posA:] < b[posB:]
}

func GetSortOrderFromRequest(c *gin.Context) string {
	sortby := c.Query("sortby")
	sortdir := c.Query("sortdir")
	if sortby == "" && sortdir == "" && filepath.Base(c.Request.URL.Path) == "favorites" {
		return "favorited_at_desc"
	} else if sortby == "" && sortdir == "" {
		return "title_asc"
	} else {
		return fmt.Sprintf("%s_%s", sortby, sortdir)
	}
}
