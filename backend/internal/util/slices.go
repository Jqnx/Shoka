package util

import (
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
