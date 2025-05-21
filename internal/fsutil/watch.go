package fsutil

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchFile(dir, file string) (string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("error:", err)
	}
	defer watcher.Close()

	ch := make(chan string)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) && event.Name == file {
					ch <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Println("error:", err)
	}

	name := <-ch

	return name, nil
}
