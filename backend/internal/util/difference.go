package util

// Diff computes which elements to add and which to remove given the
// current set and the desired set. Works over any comparable element type
// (int64 IDs, string URLs, ...) so callers with a small set don't need
// their own hand-rolled diff.
func Diff[T comparable](current, desired []T) (toAdd, toRemove []T) {
	currentSet := make(map[T]bool, len(current))
	desiredSet := make(map[T]bool, len(desired))

	for _, v := range current {
		currentSet[v] = true
	}

	for _, v := range desired {
		desiredSet[v] = true
	}

	for _, v := range desired {
		if !currentSet[v] {
			toAdd = append(toAdd, v)
		}
	}

	for _, v := range current {
		if !desiredSet[v] {
			toRemove = append(toRemove, v)
		}
	}

	return toAdd, toRemove
}
