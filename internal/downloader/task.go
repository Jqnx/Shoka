package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

const TypeDownload = "file:download"

type DownloadPayload struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func NewDownloadTask(id, url string) (*asynq.Task, error) {
	payload := DownloadPayload{
		ID:  id,
		URL: url,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDownload, bytes, asynq.MaxRetry(0)), nil
}

type DownloadProcessor struct {
	dm *Manager
}

func (d *DownloadProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload DownloadPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	downloadCtx, cancel := context.WithCancel(ctx)

	now := time.Now()
	activeDownload := &ActiveDownload{
		ID:                 payload.ID,
		Cancel:             cancel,
		Progress:           0,
		LastProgressUpdate: now,
		Cancelled:          false,
		FilePath:           "",
		BytesDownloaded:    0,
		LastSpeedUpdate:    now,
		LastBytesCount:     0,
		DownloadSpeed:      0,
		StartTime:          now,
	}

	d.dm.ActiveMutex.Lock()
	d.dm.ActiveDownloads[payload.ID] = activeDownload
	d.dm.ActiveMutex.Unlock()

	defer func() {
		d.dm.ActiveMutex.Lock()
		activeDownload, exists := d.dm.ActiveDownloads[payload.ID]
		if exists {
			if activeDownload.Cancelled && activeDownload.FilePath != "" {
				if err := os.Remove(activeDownload.FilePath); err != nil && !os.IsNotExist(err) {
					d.dm.log.Error("Failed to remove partial file", "path", activeDownload.FilePath, "error", err)
				} else {
					d.dm.log.Info("Removed partial file 2", "path", activeDownload.FilePath)
				}
			}
			delete(d.dm.ActiveDownloads, payload.ID)
		}
		d.dm.ActiveMutex.Unlock()
	}()

	if err := d.dm.UpdateDownloadStatusInDB(payload.ID, StatusDownloading, 0, ""); err != nil {
		d.dm.log.Error("Failed to update download status to downloading", "error", err)
	}

	if err := d.dm.DownloadFileWithCancel(downloadCtx, payload.ID, payload.URL, activeDownload); err != nil {
		d.dm.ActiveMutex.RLock()
		wasCancelled := activeDownload.Cancelled
		d.dm.ActiveMutex.RUnlock()

		if downloadCtx.Err() == context.Canceled && wasCancelled {
			d.dm.UpdateDownloadStatusInDB(payload.ID, StatusCancelled, 0, "Download cancelled by user")
			return nil
		} else if downloadCtx.Err() == context.Canceled {
			d.dm.UpdateDownloadStatusInDB(payload.ID, StatusFailed, 0, "Download interrupted")
			return err
		} else {
			d.dm.UpdateDownloadStatusInDB(payload.ID, StatusFailed, 0, err.Error())
			return err
		}
	}

	d.dm.UpdateDownloadStatusInDB(payload.ID, StatusCompleted, 100, "")
	d.dm.log.Info("Download completed", "id", activeDownload.ID, "path", activeDownload.FilePath)

	return nil
}

func NewDownloadProcessor(dm *Manager) *DownloadProcessor {
	return &DownloadProcessor{
		dm: dm,
	}
}
