package workers

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"context"
	"encoding/json"
)

type Workers struct {
	app   *config.App
	force bool
	ctx   context.Context
}

func NewWorkers(app *config.App, force bool, ctx context.Context) *Workers {
	return &Workers{app: app, force: force, ctx: ctx}
}

// TODO: Implement schedulers for scheduled scanning

// NewArchives scans for archives that are not in the database
// then adds them and generates covers for them
func (w *Workers) NewArchives() {
	var payload repository.Archive
	s := w.app.Noti.Listen("notifyarchives")

	ac := w.NewAsynqClient()
	defer ac.Close()

	// File scan client
	w.NewScanClient()

	list := fsutil.ListArchives(w.app.Cfg.ContentDir)
	for range list {

		arch := <-s.NotificationC()

		// fmt.Println(string(arch))

		if err := json.Unmarshal(arch, &payload); err != nil {
			w.app.Log.Error("could not unmarshal archive", "error:", err)
			return
		}

		c := NewClient(ac, w.app, &payload)
		// Create cover client
		// w.Covers(&payload, item)
		c.NewCover(false)

		// Thumbnail Client
		// workers.NewThumbClient(*cfg, &app, false)

		// Preview Client

		// Metadata Client
		c.NewMetadata("file")
	}

	s.Unlisten(w.ctx)
}
