package fsnotify

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func FileWatcher() chan struct{} {
	filesModified := make(chan struct{})
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	// defer watcher.Close()

	// Start listening for events.
	mu := sync.Mutex{}
	events := map[string]struct{}{}
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					mu.Lock()
					_, ok := events[event.Name]
					mu.Unlock()
					if ok {
						continue
					}
					mu.Lock()
					events[event.Name] = struct{}{}
					go func() {
						<-time.After(1 * time.Second)
						mu.Lock()
						delete(events, event.Name)
						mu.Unlock()
					}()
					mu.Unlock()
					log.Println("modified file:", event.Name)
					<-time.After(100 * time.Millisecond)
					filesModified <- struct{}{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err) //revive:disable:deep-exit
	}

	// Add a path.
	err = watcher.Add(wd)
	if err != nil {
		log.Fatal(err) //revive:disable:deep-exit
	}

	return filesModified
}
