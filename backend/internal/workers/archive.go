package workers

import (
	"context"
	"encoding/json"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
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
	var payload repository.GetArchiveByIDRow
	force := false
	s := w.app.Noti.Listen("notifyarchives")

	// Get list of Archives
	list := fsutil.ListArchives(w.app.Cfg.ContentDir)

	// File scan client
	w.Scan(list)

	for range list {

		arch := <-s.NotificationC()

		// fmt.Println(string(arch))

		if err := json.Unmarshal(arch, &payload); err != nil {
			w.app.Log.Error("could not unmarshal archive", "error:", err)
			return
		}

		c := NewClient(w.app.Client, w.app, &payload)
		// Create cover client
		// w.Covers(&payload, item)
		c.NewCover(force)

		// Thumbnail Client
		// workers.NewThumbClient(*cfg, &app, false)

		// Metadata Client
		// TODO: Metadata source should be changeable via configuration in ui/config file
		c.NewMetadata(config.MetadataFormats[w.app.Cfg.Sources.File.Format])
	}

	s.Unlisten(w.ctx)
}
