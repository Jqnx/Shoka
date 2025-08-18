package downloader

import (
	"Shoka/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cavaliergopher/grab/v3"
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

	filename := util.GetFilenameFromURL(payload.URL)
	filepath := filepath.Join(d.dm.cfg.DownloadDir, filename)

	req, err := grab.NewRequest(filepath, payload.URL)
	if err != nil {
		return err
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	req = req.WithContext(downloadCtx)
	req.NoResume = false

	now := time.Now()
	activeDownload := &ActiveDownload{
		ID:                 payload.ID,
		Cancel:             cancel,
		Progress:           0,
		LastProgressUpdate: now,
		Cancelled:          false,
		FilePath:           filepath,
		BytesDownloaded:    0,
		TotalSize:          0,
		LastSpeedUpdate:    now,
		LastBytesCount:     0,
		DownloadSpeed:      0,
		StartTime:          now,
		CanResume:          false,
		ResumeSupported:    false,
	}

	d.dm.ActiveMutex.Lock()
	d.dm.ActiveDownloads[payload.ID] = activeDownload
	d.dm.ActiveMutex.Unlock()

	defer func() {
		d.dm.ActiveMutex.Lock()
		activeDownload, exists := d.dm.ActiveDownloads[payload.ID]
		if exists {
			if activeDownload.Cancelled && activeDownload.FilePath != "" && !activeDownload.CanResume {
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

	resp := d.dm.grab.Do(req)

	d.dm.ActiveMutex.Lock()
	activeDownload.Response = resp
	activeDownload.TotalSize = resp.Size()
	activeDownload.ResumeSupported = resp.CanResume
	activeDownload.CanResume = resp.CanResume
	d.dm.ActiveMutex.Unlock()

	go d.dm.MonitorGrabProgress(payload.ID, resp, activeDownload)

	if err := resp.Err(); err != nil {
		d.dm.ActiveMutex.RLock()
		wasCancelled := activeDownload.Cancelled
		d.dm.ActiveMutex.RUnlock()

		if downloadCtx.Err() == context.Canceled && wasCancelled {
			d.dm.UpdateDownloadStatusInDB(payload.ID, StatusCancelled, activeDownload.Progress, "Download cancelled by user")
			return nil
		} else if downloadCtx.Err() == context.Canceled {
			if activeDownload.ResumeSupported {
				d.dm.UpdateDownloadStatusInDB(payload.ID, StatusPaused, activeDownload.Progress, "Download interrupted")
			} else {
				d.dm.UpdateDownloadStatusInDB(payload.ID, StatusFailed, activeDownload.Progress, "Download interrupted")
			}
			return err
		} else {
			if activeDownload.ResumeSupported {
				d.dm.UpdateDownloadStatusInDB(payload.ID, StatusPaused, activeDownload.Progress, err.Error())
			} else {
				d.dm.UpdateDownloadStatusInDB(payload.ID, StatusFailed, activeDownload.Progress, err.Error())
			}
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
