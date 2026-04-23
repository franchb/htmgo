package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/franchb/htmgo/cli/htmgo/internal"
	"github.com/franchb/htmgo/cli/htmgo/internal/dirutil"
	"github.com/franchb/htmgo/cli/htmgo/tasks/module"
)

func startWatcher(cb func(version string, file []*fsnotify.Event)) {
	var eventsMu sync.Mutex
	events := make([]*fsnotify.Event, 0)
	debouncer := internal.NewDebouncer(500 * time.Millisecond)
	config := dirutil.GetConfig()

	defer func() {
		if r := recover(); r != nil {
			slog.Debug("Recovered from fatal error:", slog.String("error", r.(error).Error()))
		}
	}()
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Remove) {
					if dirutil.IsGlobMatch(event.Name, config.WatchFiles, config.WatchIgnore) {
						watcher.Remove(event.Name)
						continue
					}
				}

				if event.Has(fsnotify.Create) {
					if dirutil.IsGlobMatch(event.Name, config.WatchFiles, config.WatchIgnore) {
						watcher.Add(event.Name)
						continue
					}
					info, err := os.Stat(event.Name)
					if err != nil {
						slog.Error("Error getting file info:", slog.String("path", event.Name), slog.String("error", err.Error()))
						continue
					}
					if info.IsDir() {
						err = watcher.Add(event.Name)
						if err != nil {
							slog.Error("Error adding directory to watcher:", slog.String("path", event.Name), slog.String("error", err.Error()))
						} else {
							slog.Debug("Watching directory:", slog.String("path", event.Name))
						}
					}
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					if !dirutil.IsGlobMatch(event.Name, config.WatchFiles, config.WatchIgnore) {
						continue
					}
					eventsMu.Lock()
					events = append(events, &event)
					eventsMu.Unlock()
					debouncer.Do(func() {
						eventsMu.Lock()
						toProcess := events
						events = make([]*fsnotify.Event, 0)
						eventsMu.Unlock()
						seen := make(map[string]bool)
						dedupe := make([]*fsnotify.Event, 0)
						for _, e := range toProcess {
							if _, ok := seen[e.Name]; !ok {
								seen[e.Name] = true
								dedupe = append(dedupe, e)
							}
						}
						cb(fmt.Sprintf("%06x", rand.IntN(1<<24)), dedupe)
					})
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("error:", slog.String("error", err.Error()))
			}
		}
	}()

	rootDir := "."

	frameworkPath := module.GetDependencyPath("github.com/franchb/htmgo/framework/v2")

	if !strings.HasPrefix(frameworkPath, "github.com/") {
		assetPath := filepath.Join(frameworkPath, "assets", "dist")
		slog.Debug("Watching directory:", slog.String("path", assetPath))
		watcher.Add(assetPath)
	}

	// Walk through the root directory and add all subdirectories to the watcher
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore directories in the ignoredDirs list
		if dirutil.IsGlobExclude(path, config.WatchIgnore) {
			return filepath.SkipDir
		}

		// Only watch directories
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				slog.Error("Error adding directory to watcher:", slog.String("error", err.Error()))
			} else {
				slog.Debug("Watching directory:", slog.String("path", path))
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	done := RegisterSignals()
	<-done
	println("process exited")
}
