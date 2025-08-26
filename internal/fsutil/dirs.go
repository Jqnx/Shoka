package fsutil

import (
	"Shoka/internal/config"
	"os"
)

func CreateDirs(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.ContentDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.Downloader.DownloadDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.ThumbDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		return err
	}

	return nil
}
