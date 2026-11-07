package handlers

import (
	"Shoka/internal/database/sqlc"
	"math/bits"
	"net/url"
	"sort"
	"strconv"
)

// duplicateThresholdDefault is the per-sample-point Hamming distance within
// which a pair of archives counts as matching, used when the request doesn't
// ask for another one.
const duplicateThresholdDefault = 8

// samplePointLabels name the sample points in GetAllArchivePHashes column
// order, for the per-point distance breakdown in the response.
var samplePointLabels = [4]string{"p0", "p25", "p50", "p75"}

// DuplicateArchive is one archive within a detected duplicate group,
// carrying its closest-match distance and per-sample-point breakdown.
type DuplicateArchive struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	LibraryID string `json:"library_id"`
	// CoverPage lets the frontend build the same cache-busted cover URL used
	// everywhere else (see ArchiveResponse.CoverPage).
	CoverPage int `json:"cover_page"`
	// Distance is the mean per-point Hamming distance to whichever *other*
	// member of the group this archive matches most closely - lower means
	// more visually similar. In a chained group (A close to B, B close to C)
	// this reports the real match, not the distance to an arbitrary anchor.
	Distance int `json:"distance"`
	// Distances breaks that down per sample point (p0 = cover, p25/p50/p75 =
	// that fraction through the archive), against that same closest member.
	// Points either archive hasn't hashed are omitted rather than reported
	// as 0.
	Distances map[string]int `json:"distances"`
}

// DuplicateGroup is a connected component of archives whose perceptual
// hashes match at a majority of comparable sample points.
type DuplicateGroup struct {
	Archives []DuplicateArchive `json:"archives"`
}

// duplicateCandidate is one phash-hashed archive reduced to the fields the
// grouping and reporting below need.
type duplicateCandidate struct {
	id, title, libraryID string
	coverPage            int
	hashes               [4]*int64
}

// duplicateThreshold reads the max per-point distance a sample point may have
// and still count as matching, ignoring anything outside the documented 0-20
// range.
func duplicateThreshold(q url.Values) int {
	raw := q.Get("threshold")
	if raw == "" {
		return duplicateThresholdDefault
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > 20 {
		return duplicateThresholdDefault
	}

	return n
}

// duplicateCandidates flattens the phash rows into the shape the pairwise
// comparison works on.
func duplicateCandidates(rows []sqlc.GetAllArchivePHashesRow) []duplicateCandidate {
	candidates := make([]duplicateCandidate, 0, len(rows))

	for _, row := range rows {
		candidates = append(candidates, duplicateCandidate{
			id:        row.ID,
			title:     row.Title,
			libraryID: row.LibraryID,
			coverPage: int(row.CoverPage),
			hashes:    [4]*int64{row.PhashP0, row.PhashP25, row.PhashP50, row.PhashP75},
		})
	}

	return candidates
}

// comparePHashes returns the mean Hamming distance across the sample points
// both archives have hashed, how many points were comparable, and whether
// the pair counts as a match.
//
// A match needs a *majority* of comparable points within threshold rather
// than all of them: page counts rarely line up exactly between two releases
// of the same work (bonus pages, a title page, a different scan), so one
// sample point landing on misaligned content is expected and shouldn't veto
// an otherwise strong match. Requiring a majority still rules out the
// cover-coincidence false positives that a cover-only comparison produced.
func comparePHashes(a, b [4]*int64, threshold int) (mean, compared int, matched bool) {
	var total, within int

	for i := range a {
		if a[i] == nil || b[i] == nil {
			continue
		}

		//nolint:gosec // bit reinterpretation of a 64-bit perceptual hash stored as signed int64, not an arithmetic conversion
		d := bits.OnesCount64(uint64(*a[i]) ^ uint64(*b[i]))
		total += d
		compared++

		if d <= threshold {
			within++
		}
	}

	if compared == 0 {
		return 0, 0, false
	}

	// Two comparable points minimum, so a pair can never match on the cover
	// alone - that's the weak signal this replaced.
	return total / compared, compared, compared >= 2 && within*2 > compared
}

// archivePartition is a union-find over candidate indexes, used to turn the
// pairwise match scan into connected components.
type archivePartition struct {
	parent []int
}

// newArchivePartition returns a partition where every index is its own
// component.
func newArchivePartition(n int) *archivePartition {
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}

	return &archivePartition{parent: parent}
}

// find returns the representative of x, halving the path to it on the way so
// repeated lookups in one group stay cheap.
func (p *archivePartition) find(x int) int {
	for p.parent[x] != x {
		p.parent[x] = p.parent[p.parent[x]]
		x = p.parent[x]
	}

	return x
}

// union merges the components containing a and b.
func (p *archivePartition) union(a, b int) {
	ra, rb := p.find(a), p.find(b)
	if ra != rb {
		p.parent[ra] = rb
	}
}

// duplicateComponents groups candidates into connected components of
// phash-matched archives (A close to B, B close to C groups all three even if
// A and C aren't directly close) and drops the singletons - a component of
// one is just an archive nothing duplicates.
//
// The pairwise scan is O(n^2), fine at this app's scale (a personal library,
// not millions of archives) and far simpler than an approximate
// nearest-neighbor index.
func duplicateComponents(candidates []duplicateCandidate, threshold int) [][]int {
	partition := newArchivePartition(len(candidates))

	for i := range candidates {
		for j := i + 1; j < len(candidates); j++ {
			if _, _, matched := comparePHashes(candidates[i].hashes, candidates[j].hashes, threshold); matched {
				partition.union(i, j)
			}
		}
	}

	membersByRoot := make(map[int][]int, len(candidates))

	for i := range candidates {
		root := partition.find(i)
		membersByRoot[root] = append(membersByRoot[root], i)
	}

	components := make([][]int, 0, len(membersByRoot))

	for _, members := range membersByRoot {
		if len(members) < 2 {
			continue
		}

		components = append(components, members)
	}

	return components
}

// duplicateGroupFor renders one component as a response group. Each archive
// is reported against whichever other member of the group it is actually
// closest to, rather than a fixed anchor: union-find chains A-B and B-C into
// one group even when A and C are a poor direct match, and reporting A's
// distance to that anchor would look like a false positive.
func duplicateGroupFor(members []int, candidates []duplicateCandidate, threshold int) DuplicateGroup {
	archives := make([]DuplicateArchive, 0, len(members))

	for _, idx := range members {
		c := candidates[idx]
		mean, nearest := nearestMember(idx, members, candidates, threshold)

		archives = append(archives, DuplicateArchive{
			ID:        c.id,
			Title:     c.title,
			LibraryID: c.libraryID,
			CoverPage: c.coverPage,
			Distance:  mean,
			Distances: perPointDistances(&c, nearest),
		})
	}

	sort.Slice(archives, func(i, j int) bool { return archives[i].Distance < archives[j].Distance })

	return DuplicateGroup{Archives: archives}
}

// nearestMember returns the mean per-point distance from candidates[idx] to
// whichever other member of its group it matches most closely, plus that
// member - nil when no other member shares a sample point with it.
func nearestMember(idx int, members []int, candidates []duplicateCandidate, threshold int) (int, *duplicateCandidate) {
	c := candidates[idx]

	mean := 0
	nearest := -1

	for _, otherIdx := range members {
		if otherIdx == idx {
			continue
		}

		m, compared, _ := comparePHashes(c.hashes, candidates[otherIdx].hashes, threshold)
		if compared > 0 && (nearest == -1 || m < mean) {
			mean = m
			nearest = otherIdx
		}
	}

	if nearest == -1 {
		return 0, nil
	}

	return mean, &candidates[nearest]
}

// perPointDistances breaks the mean distance between two archives down by
// sample point (p0 = cover, p25/p50/p75 = that fraction through the archive),
// omitting the points either archive hasn't hashed rather than reporting them
// as 0. b is nil when nothing in the group was comparable, leaving the map
// empty.
func perPointDistances(a, b *duplicateCandidate) map[string]int {
	distances := make(map[string]int, len(samplePointLabels))

	if b == nil {
		return distances
	}

	for k, label := range samplePointLabels {
		if a.hashes[k] == nil || b.hashes[k] == nil {
			continue
		}

		//nolint:gosec // bit reinterpretation of a 64-bit perceptual hash stored as signed int64, not an arithmetic conversion
		distances[label] = bits.OnesCount64(uint64(*a.hashes[k]) ^ uint64(*b.hashes[k]))
	}

	return distances
}

// sortDuplicateGroupsBySize orders groups with the most members first, then
// by how tightly their closest pair matches: bigger groups (more copies of
// the same thing) and tighter matches are the most actionable, so they
// surface first.
func sortDuplicateGroupsBySize(groups []DuplicateGroup) {
	sort.Slice(groups, func(i, j int) bool {
		if len(groups[i].Archives) != len(groups[j].Archives) {
			return len(groups[i].Archives) > len(groups[j].Archives)
		}

		return groups[i].Archives[1].Distance < groups[j].Archives[1].Distance
	})
}
