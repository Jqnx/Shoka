package archive

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/jxskiss/base62"
)

func TestNewArchiveID(t *testing.T) {
	t.Run("generates non-empty string", func(t *testing.T) {
		id := NewArchiveID()
		if id == "" {
			t.Error("NewArchiveID() returned empty string")
		}
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		id1 := NewArchiveID()
		id2 := NewArchiveID()

		if id1 == id2 {
			t.Errorf("NewArchiveID() generated duplicate IDs: %s", id1)
		}
	})

	t.Run("generates consistent length", func(t *testing.T) {
		ids := make([]string, 100)
		lengths := make(map[int]int)

		for i := 0; i < 100; i++ {
			ids[i] = NewArchiveID()
			lengths[len(ids[i])]++
		}

		if len(lengths) > 2 { // Allow for slight variation due to encoding
			t.Errorf("NewArchiveID() produced inconsistent lengths: %v", lengths)
		}
	})

	t.Run("contains only valid base62 characters", func(t *testing.T) {
		id := NewArchiveID()

		validBase62 := regexp.MustCompile(`^[0-9a-zA-Z]+$`)

		if !validBase62.MatchString(id) {
			t.Errorf("NewArchiveID() generated invalid base62 string: %s", id)
		}
	})

	t.Run("generates many unique IDs", func(t *testing.T) {
		seen := make(map[string]bool)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			id := NewArchiveID()
			if seen[id] {
				t.Errorf("NewArchiveID() generated duplicate ID after %d iterations: %s", i+1, id)
				break
			}
			seen[id] = true
		}
	})
}

// Benchmark test to measure performance
func BenchmarkNewArchiveID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewArchiveID()
	}
}

// Test helper to verify the function works with the expected dependencies
func TestNewArchiveIDDependencies(t *testing.T) {
	t.Run("uuid.New() works", func(t *testing.T) {
		u := uuid.New()
		if u == uuid.Nil {
			t.Error("uuid.New() returned nil UUID")
		}
	})

	t.Run("base62 encoding works", func(t *testing.T) {
		testUUID := uuid.New()
		encoded := base62.EncodeToString(testUUID[:])
		if encoded == "" {
			t.Error("base62.EncodeToString() returned empty string")
		}
	})
}
