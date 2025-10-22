package fsutil

import (
	"errors"
	"os"

	"Shoka/internal/config"
)

func CreateDirs(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.ContentDir, 0o755); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.Downloader.DownloadDir, 0o755); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.ThumbDir, 0o755); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.TempDir, 0o755); err != nil {
		return err
	}

	return nil
}

func DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}
