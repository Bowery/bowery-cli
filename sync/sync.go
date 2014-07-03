// Copyright 2013-2014 Bowery, Inc.
// Package sync implements routines to do file updates to services
// satellite instances.
package sync

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/tar"
	"github.com/Bowery/gopackages/schemas"
)

// Event describes a file event and the associated service name.
type Event struct {
	Service string
	Status  string
	Path    string
}

func (ev *Event) String() string {
	return "(" + ev.Service + "): " + strings.Title(ev.Status) + "d " + ev.Path
}

// Watcher contains an fs watcher and handles the syncing to a service.
type Watcher struct {
	Path    string
	Service *schemas.Service
	done    chan struct{}
}

// NewWatcher creates a watcher.
func NewWatcher(path string, service *schemas.Service) *Watcher {
	return &Watcher{
		Path:    path,
		Service: service,
		done:    make(chan struct{}),
	}
}

// Start handles file events and uploads the changes.
func (watcher *Watcher) Start(evChan chan *Event, errChan chan error) {
	stats := make(map[string]os.FileInfo)
	found := make([]string, 0)

	ignores, err := db.GetIgnores(watcher.Path)
	if err != nil {
		errChan <- watcher.wrapErr(err)
		return
	}

	// Get initial stats.
	err = filepath.Walk(watcher.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if ignoring.
		for _, ignore := range ignores {
			if ignore == path {
				if info.IsDir() {
					return filepath.SkipDir
				}

				return nil
			}
		}

		stats[path] = info
		return nil
	})
	if err != nil {
		errChan <- watcher.wrapErr(err)
		return
	}

	// Manager updates/creates.
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(watcher.Path, path)
		if err != nil {
			return err
		}

		// Check if ignoring.
		for _, ignore := range ignores {
			if ignore == path {
				for p := range stats {
					if p == path || strings.Contains(p, path+string(filepath.Separator)) {
						delete(stats, p)
					}
				}

				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		pstat, ok := stats[path]
		status := ""

		// Check if created/updated.
		if ok && (info.ModTime().After(pstat.ModTime()) || info.Mode() != pstat.Mode()) {
			status = "update"
		} else if !ok {
			status = "create"
		}

		// Ignore directory changes, and no event status.
		if info.IsDir() || status == "" {
			stats[path] = info
			found = append(found, path)
			return nil
		}

		err = watcher.Update(rel, status)
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug("Ignoring temp file", status, "event", rel)
				return nil
			}

			return err
		}

		evChan <- &Event{watcher.Service.Name, status, rel}
		stats[path] = info
		found = append(found, path)
		return nil
	}

	for {
		// Check if we're done.
		select {
		case <-watcher.done:
			return
		default:
		}

		ignores, err = db.GetIgnores(watcher.Path)
		if err != nil {
			errChan <- watcher.wrapErr(err)
			return
		}

		err = filepath.Walk(watcher.Path, walker)
		if err != nil {
			errChan <- watcher.wrapErr(err)
			return
		}

		// Check for deletes.
		for path := range stats {
			skip := false
			rel, err := filepath.Rel(watcher.Path, path)
			if err != nil {
				errChan <- watcher.wrapErr(err)
				return
			}

			for _, f := range found {
				if f == path {
					skip = true
					break
				}
			}

			if skip {
				continue
			}

			delete(stats, path)
			err = watcher.Update(rel, "delete")
			if err != nil {
				errChan <- watcher.wrapErr(err)
				return
			}

			evChan <- &Event{watcher.Service.Name, "delete", rel}
		}

		found = make([]string, 0)
		<-time.After(500 * time.Millisecond)
	}
}

// Upload sends the paths contents to the service compressed.
func (watcher *Watcher) Upload() error {
	var err error
	path := watcher.uploadPath()
	i := 0

	upload, err := tar.Tar(watcher.Path)
	if err != nil {
		return watcher.wrapErr(err)
	}

	// Write the tgz file to read from.
	file, err := os.Create(path)
	if err != nil {
		return watcher.wrapErr(err)
	}
	defer os.RemoveAll(path)
	defer file.Close()

	// Copy the contents to the file.
	_, err = io.Copy(file, upload)
	if err != nil {
		return watcher.wrapErr(err)
	}

	for i < 1000 {
		// If we've failed once, wait a bit.
		if err != nil {
			<-time.After(time.Millisecond * 50)
		}

		// Make sure we're at the beginning of the file.
		_, err = file.Seek(0, os.SEEK_SET)
		if err != nil {
			return watcher.wrapErr(err)
		}

		// Attempt to upload the file to the services satellite.
		err = requests.SatelliteUpload(watcher.Service.SatelliteAddr, watcher.Service.Name, file)
		if err == nil {
			return nil
		}

		i++
	}

	return watcher.wrapErr(err)
}

// Update updates a path to the service.
func (watcher *Watcher) Update(name, status string) error {
	return requests.SatelliteUpdate(watcher.Service.SatelliteAddr, watcher.Service.Name,
		filepath.Join(watcher.Path, name), name, status)
}

// Close closes the watcher and removes existing upload files.
func (watcher *Watcher) Close() error {
	close(watcher.done)

	return watcher.wrapErr(os.RemoveAll(watcher.uploadPath()))
}

// uploadPath points to the watchers upload file.
func (watcher *Watcher) uploadPath() string {
	return filepath.Join(".bowery", watcher.Service.Name+"_upload.tgz")
}

// wrapErr wraps an error with the given service name.
func (watcher *Watcher) wrapErr(err error) error {
	if err == nil {
		return nil
	}

	se := errors.IsStackError(err)
	if se != nil {
		se.Err = errors.Newf(errors.ErrSyncTmpl, watcher.Service.Name, se.Err)
		return se
	}

	return errors.Newf(errors.ErrSyncTmpl, watcher.Service.Name, err)
}

// Syncer syncs file changes to a list of given service satellite instances.
type Syncer struct {
	Event    chan *Event
	Upload   chan *schemas.Service
	Error    chan error
	Watchers []*Watcher
}

// NewSyncer creates a syncer.
func NewSyncer() *Syncer {
	return &Syncer{
		Event:    make(chan *Event),
		Upload:   make(chan *schemas.Service),
		Error:    make(chan error),
		Watchers: make([]*Watcher, 0),
	}
}

// Watch starts watching the given path and updates changes to the service.
func (syncer *Syncer) Watch(path string, service *schemas.Service) error {
	watcher := NewWatcher(path, service)
	syncer.Watchers = append(syncer.Watchers, watcher)

	// Do the actual event management, and the inital upload.
	go func() {
		err := watcher.Upload()
		if err != nil {
			syncer.Error <- err
			return
		}
		syncer.Upload <- service

		watcher.Start(syncer.Event, syncer.Error)
	}()

	return nil
}

// Close closes all the watchers.
func (syncer *Syncer) Close() error {
	for _, watcher := range syncer.Watchers {
		err := watcher.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
