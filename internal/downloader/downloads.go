package downloader

import (
	"Shoka/internal/repository"
	"Shoka/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

		task, err := NewDownloadTask(download.ID.String(), download.Url)
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

func (dm *Manager) AddDownload(url string) (*repository.Download, error) {
	downloadID := uuid.New()
	now := time.Now()
	progress := int32(0)

	ctx := context.Background()
	download, err := dm.queries.CreateDownload(ctx, repository.CreateDownloadParams{
		ID:        downloadID,
		Url:       url,
		Filename:  util.GetFilenameFromURL(url),
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

	task, err := NewDownloadTask(download.ID.String(), download.Url)
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

	task, err := NewDownloadTask(download.ID.String(), download.Url)
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
			if err := os.Remove(filePathToClean); err != nil && !os.IsNotExist(err) {
				dm.log.Error("Failed to remove partial file", "path", filePathToClean, "error", err)
			}
		}()
	}

	// TODO: Add choice to delete file when deleting download or not
	if delFile {
		filePath := filepath.Join(dm.cfg.DownloadDir, download.Filename)
		os.Remove(filePath)
	}

	if err := dm.queries.DeleteDownload(ctx, download.ID); err != nil {
		return err
	}

	dm.BroadcastDeletion(id)

	return nil
}

// TODO: redo to handle the sites i want, probably use an interface so multiple sites can plug into that.
// TODO: Check if requested link is valid file (image file, zip file, cbz file, w/e file just not pure html)
func (dm *Manager) DownloadFileWithCancel(ctx context.Context, id, url string, activeDownload *ActiveDownload) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get filename and create file
	filename := util.GetFilenameFromURL(url)
	filePath := filepath.Join(dm.cfg.DownloadDir, filename)

	dm.ActiveMutex.Lock()
	activeDownload.FilePath = filePath
	dm.ActiveMutex.Unlock()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		file.Close()
		if ctx.Err() == context.Canceled {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				dm.log.Error("Failed to remove partial file on cancellation", "path", filePath, "error", err)
			} else {
				dm.log.Info("Removed partial file 3", "path", filePath)
			}
		}
	}()

	// Download with context cancellation and progress tracking
	contentLength := resp.ContentLength
	var downloaded int64
	buffer := make([]byte, 32*1024) // 32KB buffer

	dm.log.Info("starting download", "id", activeDownload.ID, "file", activeDownload.FilePath, "status", StatusDownloading)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			// Update Download Speed
			// dm.updateActiveDownloadBytes(downloaded, activeDownload)

			// Update progress
			if contentLength > 0 {
				// progress := int32((downloaded * 100) / contentLength)
				// dm.UpdateActiveDownloadProgress(id, progress, activeDownload)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

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
