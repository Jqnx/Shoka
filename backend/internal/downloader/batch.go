package downloader

/*
func (dm *Manager) downloadBatch(id string, ctx context.Context, requests []*grab.Request, activeDownload *ActiveDownload) ([]*grab.Response, error) {
	var current int
	// respch := dm.grab.DoBatch(1, requests...)
	respch := dm.doBatch(1, requests...)

	responses := make([]*grab.Response, 0)

	for {
		select {
		case <-ctx.Done():
			for _, resp := range responses {
				if !resp.IsComplete() {
					resp.Cancel()
				}
			}
			return responses, ctx.Err()

		case resp, ok := <-respch:
			if !ok {
				return responses, nil
			}

			responses = append(responses, resp)

			if err := resp.Err(); err != nil {
				dm.log.Error("Download failed", "file", resp.Filename, "err", err)
				return nil, err
			} else {

				dm.ActiveMutex.Lock()
				current++

				progress := int32((current * 100) / len(requests))
				activeDownload.Progress = progress

				now := time.Now()
				dm.ActiveMutex.Unlock()

				activeDownload.LastProgressUpdate = now

				dm.log.Info("download", "id", activeDownload.ID, "file", resp.Filename, "progress", activeDownload.Progress, "speed", activeDownload.DownloadSpeed)

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

func (dm *Manager) doBatch(workers int, requests ...*grab.Request) <-chan *grab.Response {
	if workers < 1 {
		workers = len(requests)
	}
	reqch := make(chan *grab.Request, len(requests))
	respch := make(chan *grab.Response, len(requests))
	wg := sync.WaitGroup{}
	for range workers {
		wg.Add(1)
		go func() {
			dm.grab.DoChannel(reqch, respch)
			wg.Done()
		}()
	}

	// queue requests
	go func() {
		for _, req := range requests {
			reqch <- req
			time.Sleep(time.Duration(dm.cfg.Downloader.RateLimit) * time.Millisecond)
		}
		close(reqch)
		wg.Wait()
		close(respch)
	}()
	return respch
}
*/
