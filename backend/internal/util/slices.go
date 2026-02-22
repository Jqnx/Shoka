package util

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"Shoka/internal/repository"
)

func RemoveDuplicatesStrPointer(in []*string) []string {
	var list []string
	for _, i := range in {
		list = append(list, *i)
	}
	slices.Sort(list)
	return slices.Compact(list)
}

func MatchStringsInSlices(in1 []string, in2 []string) []string {
	var list []string
	if in1 == nil {
		return in2
	} else if in2 == nil {
		return in1
	} else {
		for _, i := range in1 {
			if slices.Contains(in2, i) {
				list = append(list, i)
			}
		}
	}
	return list
}

func ToString(in any) string {
	var slice []string

	switch in := in.(type) {
	case []repository.Artist:
		for _, i := range in {
			slice = append(slice, i.Name)
		}
	case []repository.Tag:
		for _, i := range in {
			slice = append(slice, i.Name)
		}
	case []repository.Character:
		for _, i := range in {
			slice = append(slice, i.Name)
		}
	case []repository.Parody:
		for _, i := range in {
			slice = append(slice, i.Name)
		}
	case []repository.GetArchiveUrlsRow:
		for _, i := range in {
			slice = append(slice, i.Url)
		}
	default:
		return ""
	}

	out := strings.Join(slice, ", ")
	return out
}

func ToSliceString(in any) []string {
	v := reflect.ValueOf(in)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var result []string
	for i := 0; i < v.Len(); i++ {
		result = append(result, fmt.Sprint(v.Index(i).Field(0)))
	}
	return result
}

// Intersect compares multiple slices of type string with each other
// and returns a new slice containing only the items that are present
// in all input slices.
func Intersect(slices ...[]string) []string {
	var validSlices [][]string
	for _, s := range slices {
		if len(s) > 0 {
			validSlices = append(validSlices, s)
		}
	}

	if len(validSlices) == 0 {
		return []string{}
	}

	shortestIndex := 0
	minLen := len(validSlices[0])
	for i, s := range validSlices {
		if len(s) < minLen {
			minLen = len(s)
			shortestIndex = i
		}
	}

	candidates := make(map[string]bool)
	for _, item := range validSlices[shortestIndex] {
		candidates[item] = true
	}

	for i, s := range validSlices {
		if i == shortestIndex {
			continue
		}

		if len(candidates) == 0 {
			return []string{}
		}

		currentSliceMap := make(map[string]bool)
		for _, item := range s {
			currentSliceMap[item] = true
		}

		for item := range candidates {
			if !currentSliceMap[item] {
				delete(candidates, item)
			}
		}
	}

	result := make([]string, 0, len(candidates))
	for item := range candidates {
		result = append(result, item)
	}

	return result
}

// IntersectDifferent compares multiple slices of type string with each other
// and returns a new slice containing only the items that are not present
// in all input slices. Inverse of IntersectCommon.
func IntersectDifferent(slices ...[]string) []string {
	var validSlices [][]string
	for _, s := range slices {
		if len(s) > 0 {
			validSlices = append(validSlices, s)
		}
	}

	if len(validSlices) == 0 {
		return []string{}
	}

	shortestIndex := 0
	minLen := len(validSlices[0])
	for i, s := range validSlices {
		if len(s) < minLen {
			minLen = len(s)
			shortestIndex = i
		}
	}

	candidates := make(map[string]bool)
	for _, item := range validSlices[shortestIndex] {
		candidates[item] = true
	}

	for i, s := range validSlices {
		if i == shortestIndex {
			continue
		}

		if len(candidates) == 0 {
			return []string{}
		}

		currentSliceMap := make(map[string]bool)
		for _, item := range s {
			currentSliceMap[item] = true
		}

		for item := range candidates {
			if currentSliceMap[item] {
				delete(candidates, item)
			}
		}
	}

	result := make([]string, 0, len(candidates))
	for item := range candidates {
		result = append(result, item)
	}

	return result
}

// CalculateDiff takes two slices of any comparable type (like int, string, UUID)
// and returns the items that need to be inserted and the items to be removed.
func CalculateDiff[T comparable](oldIDs, newIDs []T) (toInsert []T, toRemove []T) {
	oldSet := make(map[T]struct{}, len(oldIDs))
	newSet := make(map[T]struct{}, len(newIDs))

	for _, id := range oldIDs {
		oldSet[id] = struct{}{}
	}
	for _, id := range newIDs {
		newSet[id] = struct{}{}
	}

	for _, id := range newIDs {
		if _, exists := oldSet[id]; !exists {
			toInsert = append(toInsert, id)
		}
	}

	for _, id := range oldIDs {
		if _, exists := newSet[id]; !exists {
			toRemove = append(toRemove, id)
		}
	}

	return toInsert, toRemove
}
