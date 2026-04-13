package util

import (
	"regexp"
	"strconv"
)

var digitRun = regexp.MustCompile(`\d+`)

func NaturalLess(a, b string) bool {
	for {
		if b == "" {
			return false
		}
		if a == "" {
			return true
		}

		aDigit := digitRun.FindStringIndex(a)
		bDigit := digitRun.FindStringIndex(b)

		// if neither starts with a digit segment, compare as strings
		if aDigit == nil || bDigit == nil {
			if a[0] != b[0] {
				return a[0] < b[0]
			}
			a, b = a[1:], b[1:]
			continue
		}

		// compare the non-digit prefix first
		aPrefix, bPrefix := a[:aDigit[0]], b[:bDigit[0]]
		if aPrefix != bPrefix {
			return aPrefix < bPrefix
		}

		// compare the numeric segments as integers
		aNum, _ := strconv.Atoi(a[aDigit[0]:aDigit[1]])
		bNum, _ := strconv.Atoi(b[bDigit[0]:bDigit[1]])
		if aNum != bNum {
			return aNum < bNum
		}

		a, b = a[aDigit[1]:], b[bDigit[1]:]
	}
}
