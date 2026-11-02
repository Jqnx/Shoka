// Package media probes media files for the facts Shoka stores about them.
package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Duration returns the media duration reported by ffprobe. It asks for the
// container duration first, then falls back to the longest stream duration:
// some containers (notably several raw/adts variants) report no format
// duration while their streams still carry one.
//
// No startup check: callers treat a missing or failing ffprobe as a zero
// duration and keep working.
func Duration(ctx context.Context, path string) (time.Duration, error) {
	seconds, err := probe(ctx, path, "format=duration")
	if err == nil {
		return time.Duration(seconds * float64(time.Second)), nil
	}

	seconds, err = probeMax(ctx, path, "stream=duration")
	if err != nil {
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// probe runs ffprobe for one -show_entries value and returns its first
// parseable number.
func probe(ctx context.Context, path, entries string) (float64, error) {
	values, err := ffprobe(ctx, path, entries)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("ffprobe reported no duration for %s", path)
	}

	return values[0], nil
}

// probeMax is probe for multi-stream output, returning the largest parseable
// duration (the longest track of a multi-file container).
func probeMax(ctx context.Context, path, entries string) (float64, error) {
	values, err := ffprobe(ctx, path, entries)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("ffprobe reported no duration for %s", path)
	}

	maxValue := values[0]
	for _, v := range values[1:] {
		if v > maxValue {
			maxValue = v
		}
	}

	return maxValue, nil
}

// ffprobe invokes the binary and returns every parseable number it printed.
func ffprobe(ctx context.Context, path, entries string) ([]float64, error) {
	// #nosec G204 -- path is a file inside a configured library; args are
	// constants plus that path, never a shell string.
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", entries,
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe %s: %w", path, err)
	}

	var values []float64

	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		v, err := strconv.ParseFloat(line, 64)
		if err != nil {
			continue
		}

		values = append(values, v)
	}

	return values, nil
}
