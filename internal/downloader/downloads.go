package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/sources"

	"github.com/cavaliergopher/grab/v3"
	"github.com/google/uuid"
)

func (dm *Manager) reQueueInterruptedDownloads() error {
	ctx := context.Background()

	interruptedDownloads, err := dm.queries.GetDownloadsByStatus(ctx, []string{StatusPending, StatusDownloading})
	if err != nil {
		return err
	}

	for _, download := range interruptedDownloads {
		_, err := dm.queries.UpdateDownloadStatus(ctx, repository.UpdateDownloadStatusParams{
			ID:        download.ID,
			Status:    StatusPending,
			Progress:  download.Progress,
			Error:     nil,
			UpdatedAt: time.Now(),
		})
		if err != nil {
			dm.log.Error("Failed to reset download status", "id", download.ID, "error", err)
			continue
		}

		task, err := NewDownloadTask(download.ID.String(), download.Url, download.Source, nil)
		if err != nil {
			return err
		}

		if _, err := dm.asynqClient.Enqueue(task); err != nil {
			dm.log.Error("Failed to re-enqueue download", "id", download.ID, "error", err)
		}
	}

	if len(interruptedDownloads) > 0 {
		dm.log.Info(fmt.Sprintf("Re-queued %d interrupted downloads", len(interruptedDownloads)))
	}

	return nil
}

func (dm *Manager) AddDownload(url *url.URL, source string) (*repository.Download, error) {
	downloadID := uuid.New()
	now := time.Now()
	progress := int32(0)

	src, err := sources.NewSource(dm.cfg, source, &config.MethodID)
	if err != nil {
		return nil, err
	}
	src.SetURL(url)
	meta, err := src.GetMetadata()
	if err != nil {
		return nil, err
	}
	metadata := meta[0]

	filename := fmt.Sprintf("%s.%s", metadata.Title, dm.cfg.Downloader.SaveFileExt)

	ctx := context.Background()
	download, err := dm.queries.CreateDownload(ctx, repository.CreateDownloadParams{
		ID:        downloadID,
		Url:       url.String(),
		Source:    source,
		Filename:  filename,
		Status:    StatusPending,
		Progress:  &progress,
		Error:     nil,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}

	dm.log.Info("added download", "id", download.ID, "file", download.Filename, "status", download.Status)

	task, err := NewDownloadTask(download.ID.String(), download.Url, download.Source, &metadata)
	if err != nil {
		return nil, err
	}

	if _, err := dm.asynqClient.Enqueue(task); err != nil {
		dm.queries.DeleteDownload(ctx, download.ID)
		return nil, err
	}

	return &download, nil
}

func (dm *Manager) EnrichWithActiveProgress(downloads []repository.Download) {
	dm.ActiveMutex.RLock()
	defer dm.ActiveMutex.RUnlock()

	for i := range downloads {
		if activeDownload, exists := dm.ActiveDownloads[downloads[i].ID.String()]; exists {
			downloads[i].Progress = &activeDownload.Progress
			downloads[i].Speed = &activeDownload.DownloadSpeed
			downloads[i].Downloaded = &activeDownload.BytesDownloaded
			downloads[i].StartedAt = &activeDownload.StartTime
		}
	}
}

func (dm *Manager) PauseDownload(id string) error {
	dm.ActiveMutex.Lock()
	defer dm.ActiveMutex.Unlock()

	activeDownload, exists := dm.ActiveDownloads[id]
	if !exists {
		return fmt.Errorf("download not active")
	}

	activeDownload.Cancel()

	if err := dm.UpdateDownloadStatusInDB(id, StatusPaused, activeDownload.Progress, ""); err != nil {
		return err
	}

	return nil
}

func (dm *Manager) ResumeDownload(ctx context.Context, download *repository.Download) error {
	_, err := dm.queries.UpdateDownloadStatus(ctx, repository.UpdateDownloadStatusParams{
		ID:        download.ID,
		Status:    StatusPending,
		Progress:  download.Progress,
		Error:     nil,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	dm.log.Info("download resumed", "id", download.ID, "file", download.Filename, "status", download.Status)

	task, err := NewDownloadTask(download.ID.String(), download.Url, download.Source, nil)
	if err != nil {
		return err
	}

	if _, err := dm.asynqClient.Enqueue(task); err != nil {
		return err
	}
	return nil
}

func (dm *Manager) DeleteDownload(ctx context.Context, download *repository.Download, delFile bool) error {
	id := download.ID.String()

	dm.ActiveMutex.Lock()
	var filePathToClean string
	if activeDownload, exists := dm.ActiveDownloads[id]; exists {
		activeDownload.Cancelled = true
		filePathToClean = activeDownload.FilePath
		activeDownload.Cancel()
	}
	dm.ActiveMutex.Unlock()

	if filePathToClean != "" {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := fsutil.Remove(filePathToClean); err != nil && !os.IsNotExist(err) {
				dm.log.Error("Failed to remove partial file", "path", filePathToClean, "error", err)
			}
		}()
	}

	if download.Status == StatusCompleted && delFile {
		filePath := filepath.Join(dm.cfg.Downloader.DownloadDir, download.Source, download.Filename)
		if err := fsutil.Remove(filePath); err != nil {
			return err
		}
	}

	if err := dm.queries.DeleteDownload(ctx, download.ID); err != nil {
		return err
	}

	dm.BroadcastDeletion(id)

	return nil
}

func (dm *Manager) MonitorGrabProgress(id string, resp *grab.Response, activeDownload *ActiveDownload) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-resp.Done:
			return
		case <-ticker.C:
			bytesComplete := resp.BytesComplete()
			totalSize := resp.Size()

			dm.ActiveMutex.Lock()
			activeDownload.BytesDownloaded = bytesComplete
			activeDownload.TotalSize = totalSize

			var progress int32
			if totalSize > 0 {
				progress = int32((bytesComplete * 100) / totalSize)
			}

			now := time.Now()
			timeSinceLastUpdate := now.Sub(activeDownload.LastSpeedUpdate)
			if timeSinceLastUpdate >= time.Second {
				bytesSinceLastUpdate := bytesComplete - activeDownload.LastBytesCount

				if timeSinceLastUpdate > 0 {
					activeDownload.DownloadSpeed = int64(float64(bytesSinceLastUpdate) / timeSinceLastUpdate.Seconds())
				}

				activeDownload.LastSpeedUpdate = now
				activeDownload.LastBytesCount = bytesComplete
			}

			oldProgress := activeDownload.Progress

			progressCheck := progress - oldProgress
			if progressCheck >= 5 {
				activeDownload.Progress = progress
			}
			dm.ActiveMutex.Unlock()

			if progressCheck >= 5 || now.Sub(activeDownload.LastProgressUpdate) >= 5*time.Second {
				activeDownload.LastProgressUpdate = now

				dm.log.Info("download", "id", activeDownload.ID, "file", activeDownload.FilePath, "progress", activeDownload.Progress, "speed", activeDownload.DownloadSpeed)

				// Update database asynchronously
				go func() {
					ctx := context.Background()
					downloadID, err := uuid.Parse(id)
					if err != nil {
						return
					}

					dm.queries.UpdateDownloadProgress(ctx, repository.UpdateDownloadProgressParams{
						ID:        downloadID,
						Progress:  &progress,
						UpdatedAt: now,
					})
				}()

				// Broadcast update
				go func() {
					ctx := context.Background()
					downloadID, err := uuid.Parse(id)
					if err != nil {
						return
					}

					download, err := dm.queries.GetDownload(ctx, downloadID)
					if err != nil {
						return
					}

					download.Progress = &progress
					download.Speed = &activeDownload.DownloadSpeed
					download.Downloaded = &activeDownload.BytesDownloaded
					download.TotalSize = &activeDownload.TotalSize
					download.StartedAt = &activeDownload.StartTime
					download.CanResume = &activeDownload.CanResume
					download.ResumeSupported = &activeDownload.ResumeSupported
					dm.BroadcastUpdate(&download)
				}()
			}
		}
	}
}

func (dm *Manager) UpdateDownloadStatusInDB(id, status string, progress int32, errorMsg string) error {
	ctx := context.Background()
	downloadID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	var errorPtr *string
	if errorMsg != "" {
		errorPtr = &errorMsg
	}

	updatedDownload, err := dm.queries.UpdateDownloadStatus(ctx, repository.UpdateDownloadStatusParams{
		ID:        downloadID,
		Status:    status,
		Progress:  &progress,
		Error:     errorPtr,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	if status == StatusCancelled {
		go dm.BroadcastDeletion(updatedDownload.ID.String())
	} else {
		go dm.BroadcastUpdate(&updatedDownload)
	}

	return nil
}

func (dm *Manager) BroadcastUpdate(download *repository.Download) {
	message, _ := json.Marshal(map[string]any{
		"type": "download_update",
		"data": download,
	})
	dm.hub.Broadcast(message)
}

func (dm *Manager) BroadcastDeletion(id string) {
	message, _ := json.Marshal(map[string]any{
		"type": "download_deleted",
		"data": map[string]string{"id": id},
	})
	dm.hub.Broadcast(message)
}

func (dm *Manager) Close() {
	// Cancel all active downloads
	dm.ActiveMutex.Lock()
	for _, activeDownload := range dm.ActiveDownloads {
		activeDownload.Cancelled = true
		activeDownload.Cancel()
	}
	dm.ActiveMutex.Unlock()
}
