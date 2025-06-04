package util

import "slices"

func RemoveDuplicatesStrPointer(in []*string) []string {
	var list []string
	for _, i := range in {
		list = append(list, *i)
	}
	slices.Sort(list)
	return slices.Compact(list)
}
