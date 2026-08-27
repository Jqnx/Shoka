package util

import "os"

func EnsureDir(path string) error {
	// 0o755: data/cache directories are deliberately traversable by the
	// separate frontend process (see project docs), which may run as a
	// different user.
	//nolint:gosec // intentional: directory is shared with the frontend process
	return os.MkdirAll(path, 0o755)
}
