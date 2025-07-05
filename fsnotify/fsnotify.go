package fsnotify

import (
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher returns two channels: one for file system events and one for extra files to watch.
//
// The first channel emits a struct{}{} whenever a file is modified. The caller can use this channel to
// trigger a slide rebuild.
//
// The second channel is used to add extra files to watch. The caller can use this channel to add extra
// files to the watcher. The extra files are added to the watcher only if they are not already being
// watched.
//
// The function returns two channels, but also starts a goroutine that listens for events on the watcher.
// The goroutine is not stopped until the channel is closed.
//
// The function is not safe for concurrent use. It is the caller's responsibility to ensure that the
// function is not called concurrently.
func FileWatcher() (chan struct{}, chan string) { //revive:disable:function-length,cognitive-complexity
	filesModified := make(chan struct{})
	extraFilesCh := make(chan string)
	extraDirs := make(map[string]struct{})
	baseDir := ""
	extraFiles := make(map[string]struct{})
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	// defer watcher.Close()

	// Start listening for events.
	mu := sync.Mutex{}
	muExtraFiles := sync.Mutex{}
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
					dir := path.Dir(event.Name)
					if dir != baseDir {
						// check if its and extra file
						muExtraFiles.Lock()
						_, ok := extraFiles[event.Name]
						muExtraFiles.Unlock()
						if !ok {
							continue
						}
					}

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

	go func() {
		// read endlessly from extraFiles channel
		for file := range extraFilesCh {
			file = path.Join(baseDir, file)
			// get directory of a file
			dir := path.Dir(file)
			// if directory is not already being watched, add it
			muExtraFiles.Lock()
			extraFiles[file] = struct{}{}
			_, ok := extraDirs[dir]
			if ok {
				muExtraFiles.Unlock()
				continue
			}
			extraDirs[dir] = struct{}{}
			muExtraFiles.Unlock()
			// log.Println("adding directory to watcher:", dir)
			// add directory to watcher
			err := watcher.Add(dir)
			if err != nil {
				log.Println("error adding directory to watcher:", err)
				continue
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
	extraDirs[wd] = struct{}{}
	baseDir = wd

	return filesModified, extraFilesCh
}
