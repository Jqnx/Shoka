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

	dm.broadcastUpdate(&download)

	return &download, nil
}

func (dm *Manager) EnrichWithActiveProgress(downloads []repository.Download) {
	dm.ActiveMutex.RLock()
	defer dm.ActiveMutex.RUnlock()

	for i := range downloads {
		if activeDownload, exists := dm.ActiveDownloads[downloads[i].ID.String()]; exists {
			downloads[i].Progress = &activeDownload.Progress
		}
	}
}

func (dm *Manager) DeleteDownload(ctx context.Context, download *repository.Download) error {
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
	if download.Status == StatusCompleted {
		filePath := filepath.Join(dm.cfg.DownloadDir, download.Filename)
		os.Remove(filePath)
	}

	if err := dm.queries.DeleteDownload(ctx, download.ID); err != nil {
		return err
	}

	dm.broadcastDeletion(id)

	return nil
}

// TODO: redo to handle the sites i want, probably use an interface so multiple sites can plug into that.
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

			// Update progress
			if contentLength > 0 {
				progress := int32((downloaded * 100) / contentLength)
				dm.UpdateActiveDownloadProgress(id, progress, activeDownload)
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

func (dm *Manager) UpdateActiveDownloadProgress(id string, progress int32, activeDownload *ActiveDownload) {
	now := time.Now()

	// Update in-memory progress
	activeDownload.Progress = progress

	// Only update database and broadcast every 5% progress change or every 5 seconds
	shouldUpdate := progress-activeDownload.Progress >= 5 ||
		now.Sub(activeDownload.LastProgressUpdate) >= 5*time.Second ||
		progress == 100

	if shouldUpdate {
		activeDownload.LastProgressUpdate = now

		dm.log.Info("download", "id", activeDownload.ID, "file", activeDownload.FilePath, "progress", activeDownload.Progress)

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

			// Use current progress
			download.Progress = &progress
			dm.broadcastUpdate(&download)
		}()
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

	go dm.broadcastUpdate(&updatedDownload)

	return nil
}

func (dm *Manager) broadcastUpdate(download *repository.Download) {
	message, _ := json.Marshal(map[string]interface{}{
		"type": "download_update",
		"data": download,
	})
	dm.hub.Broadcast(message)
}

func (dm *Manager) broadcastDeletion(id string) {
	message, _ := json.Marshal(map[string]interface{}{
		"type": "download_deleted",
		"data": map[string]string{"id": id},
	})
	dm.hub.Broadcast(message)
}

func (dm *Manager) Close() {
	// Cancel all active downloads
	dm.ActiveMutex.Lock()
	for _, activeDownload := range dm.ActiveDownloads {
		activeDownload.Cancel()
	}
	dm.ActiveMutex.Unlock()
}
