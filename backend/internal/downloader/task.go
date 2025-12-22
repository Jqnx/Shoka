package downloader

/*
const TypeDownload = "file:download"

type DownloadPayload struct {
	ID       string           `json:"id"`
	URL      string           `json:"url"`
	Source   string           `json:"source"`
	Metadata *models.Metadata `json:"metadata"`
}

func NewDownloadTask(id, url, source string, metadata *models.Metadata) (*asynq.Task, error) {
	payload := DownloadPayload{
		ID:       id,
		URL:      url,
		Source:   source,
		Metadata: metadata,
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

	path := filepath.Join(d.dm.cfg.TempDir, payload.Metadata.Title)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	reqs := make([]*grab.Request, 0)
	for i := range payload.Metadata.PageCount {
		url := fmt.Sprintf("%s/galleries/%s/%d%s", config.NHImages, payload.Metadata.NHMediaID, i+1, payload.Metadata.NHImageType)
		req, err := grab.NewRequest(path, url)
		if err != nil {
			return err
		}
		reqs = append(reqs, req)
	}

	now := time.Now()
	activeDownload := &ActiveDownload{
		ID:                 payload.ID,
		Cancel:             cancel,
		Progress:           0,
		LastProgressUpdate: now,
		Cancelled:          false,
		FilePath:           path,
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
				if err := fsutil.Remove(activeDownload.FilePath); err != nil && !os.IsNotExist(err) {
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

	// resp := d.dm.grab.Do(req)
	response, err := d.dm.downloadBatch(payload.ID, downloadCtx, reqs, activeDownload)
	if err != nil {
		return err
	}

	// d.dm.ActiveMutex.Lock()
	// activeDownload.Response = resp
	// activeDownload.TotalSize = resp.Size()
	// activeDownload.ResumeSupported = resp.CanResume
	// activeDownload.CanResume = resp.CanResume
	// d.dm.ActiveMutex.Unlock()

	// go d.dm.MonitorGrabProgress(payload.ID, resp, activeDownload)

	for _, resp := range response {
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
	}

	sourceDir := filepath.Join(d.dm.cfg.Downloader.DownloadDir, payload.Source)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.%s", payload.Metadata.Title, d.dm.cfg.Downloader.SaveFileExt)
	dst := filepath.Join(sourceDir, filename)
	if err := fsutil.Zip(path, dst); err != nil {
		return err
	}

	if err := fsutil.Remove(path); err != nil {
		return err
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
*/
