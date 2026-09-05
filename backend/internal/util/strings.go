package util

import "strings"

// DedupeTrimmed trims each entry, drops blanks, and removes duplicates
// while preserving first-seen order. Shared by callers that need to clean
// up a caller-supplied string list (bulk archive ids, source URLs, ...)
// before applying their own validation on top.
func DedupeTrimmed(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))

	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		if _, dup := seen[s]; dup {
			continue
		}

		seen[s] = struct{}{}
		out = append(out, s)
	}

	return out
}
