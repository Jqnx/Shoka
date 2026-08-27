package util

// DiffIDs computes which IDs to add and which to remove
// given the current set and the desired set.
func DiffIDs(current, desired []int64) (toAdd, toRemove []int64) {
	currentSet := make(map[int64]bool, len(current))
	desiredSet := make(map[int64]bool, len(desired))

	for _, id := range current {
		currentSet[id] = true
	}

	for _, id := range desired {
		desiredSet[id] = true
	}

	for _, id := range desired {
		if !currentSet[id] {
			toAdd = append(toAdd, id)
		}
	}

	for _, id := range current {
		if !desiredSet[id] {
			toRemove = append(toRemove, id)
		}
	}

	return toAdd, toRemove
}
