package util

import (
	"slices"
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
