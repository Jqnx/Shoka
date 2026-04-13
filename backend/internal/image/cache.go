package image

import (
	"fmt"
	"log/slog"

	"Shoka/internal/library/archive"

	lru "github.com/hashicorp/golang-lru/v2"
)

const lruSize = 128 // number of pages to keep in the LRU

type Cache struct {
	processor *Processor
	lru       *lru.Cache[string, []byte]
	log       *slog.Logger
}

func NewCache(processor *Processor, log *slog.Logger) (*Cache, error) {
	l, err := lru.New[string, []byte](lruSize)
	if err != nil {
		return nil, err
	}
	return &Cache{
		processor: processor,
		lru:       l,
		log:       log.With("component", "image_cache"),
	}, nil
}

// lruKey returns a unique key for an archive page
// format: archiveID:index
func lruKey(archiveID string, index int) string {
	return fmt.Sprintf("%s:%d", archiveID, index)
}

// GetPage returns processed WebP bytes for a page.
// Checks LRU first, then opens the archive on a miss.
func (c *Cache) GetPage(archiveID string, filePath string, index int) ([]byte, error) {
	key := lruKey(archiveID, index)

	if data, ok := c.lru.Get(key); ok {
		c.log.Debug("page cache hit", "archive_id", archiveID, "index", index)
		return data, nil
	}

	c.log.Debug("page cache miss, extracting", "archive_id", archiveID, "index", index)

	data, err := c.extractAndProcess(filePath, index)
	if err != nil {
		return nil, err
	}

	c.lru.Add(key, data)
	return data, nil
}

func (c *Cache) extractAndProcess(filePath string, index int) ([]byte, error) {
	a, err := archive.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer a.Close()

	pages, err := a.Pages()
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}

	if index < 0 || index >= len(pages) {
		return nil, fmt.Errorf("page index %d out of range", index)
	}

	r, err := a.Extract(pages[index])
	if err != nil {
		return nil, fmt.Errorf("extract page %d: %w", index, err)
	}
	defer r.Close()

	return c.processor.ProcessToBytes(r)
}
