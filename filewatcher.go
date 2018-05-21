package gonotify

import "path/filepath"

type FileWatcher struct {
	stopC chan struct{}
	C     chan InotifyEvent
}

func NewFileWatcher(mask uint32, files ...string) (*FileWatcher, error) {

	f := &FileWatcher{
		stopC: make(chan struct{}),
		C:     make(chan InotifyEvent),
	}

	inotify, err := NewInotify()
	if err != nil {
		return nil, err
	}

	expectedPaths := make(map[string]bool)

	for _, file := range files {
		err := inotify.AddWatch(filepath.Dir(file), mask)
		if err != nil {
			inotify.Close()
			return nil, err
		}
		expectedPaths[file] = true
	}

	events := make(chan InotifyEvent)

	go func() {
		for {
			raw, err := inotify.Read()

			if err != nil {
				return
			}

			for _, event := range raw {
				events <- event
			}
		}
	}()

	go func() {
		for {
			select {
			case <-f.stopC:
				inotify.Close()
				return
			case event := <-events:

				if !expectedPaths[event.Name] {
					continue
				}

				f.C <- event
			}
		}
	}()

	return f, nil
}

func (f *FileWatcher) Close() {
	f.stopC <- struct{}{}
}