package dirwatcher

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DirectoryWatcher watches a directory for newly added Unix sockets
// and also monitors for new subdirectories to watch for sockets
type DirectoryWatcher struct {
	dir          string
	logger       *zap.Logger
	knownFiles   map[string]os.FileInfo
	knownDirs    map[string]*DirectoryWatcher
	socketHandle func(string)
	interval     time.Duration
	mutex        sync.Mutex
	stopChan     chan struct{}
	stopped      bool
}

func NewDirectoryWatcher(dir string, logger *zap.Logger, interval time.Duration, handle func(string)) *DirectoryWatcher {
	return &DirectoryWatcher{
		dir:          dir,
		logger:       logger,
		knownFiles:   make(map[string]os.FileInfo),
		knownDirs:    make(map[string]*DirectoryWatcher),
		socketHandle: handle,
		interval:     interval,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the directory watcher, polling for changes every interval seconds
func (dw *DirectoryWatcher) Start() {
	dw.logger.Info("Starting directory watcher...", zap.String("directory", dw.dir))
	go func() {
		for {
			select {
			case <-dw.stopChan:
				dw.logger.Info("Stopping directory watcher", zap.String("directory", dw.dir))
				return
			default:
				// Poll the directory for new files and subdirectories.
				dw.pollDirectory()
				time.Sleep(dw.interval)
			}
		}
	}()
}

// Stop stops the directory watcher and all child watchers
func (dw *DirectoryWatcher) Stop() {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	if dw.stopped {
		return
	}

	// Stop all subdirectory watchers
	for _, subWatcher := range dw.knownDirs {
		subWatcher.Stop()
	}

	// Signal the polling goroutine to stop
	close(dw.stopChan)
	dw.stopped = true
}

func (dw *DirectoryWatcher) pollDirectory() {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	// If stopped, don't continue
	if dw.stopped {
		return
	}

	// Read the directory contents.
	entries, err := os.ReadDir(dw.dir)
	if err != nil {
		dw.logger.Error("error reading directory",
			zap.String("directory", dw.dir),
			zap.Error(err))
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dw.dir, entry.Name())

		if entry.IsDir() {
			// Found a directory, check if we're already watching it
			if _, known := dw.knownDirs[path]; !known {
				dw.logger.Info("new directory detected", zap.String("path", path))
				// Create a new watcher for this directory
				subDirWatcher := NewDirectoryWatcher(path, dw.logger, dw.interval, dw.socketHandle)
				dw.knownDirs[path] = subDirWatcher
				subDirWatcher.Start()
			}
			continue
		}

		// check if the file is a new file and is a socket
		if _, known := dw.knownFiles[path]; !known && dw.isUnixSocket(path) {
			// add new file to known files
			info, err := entry.Info()
			if err == nil {
				dw.knownFiles[path] = info
				dw.logger.Info("new unix socket detected",
					zap.String("path", path))
				// connect to the socket in a goroutine
				go dw.socketHandle(path)
			}
		}
	}

	// cleanup: remove deleted directories from the known list
	for path, subWatcher := range dw.knownDirs {
		if _, err = os.Stat(path); os.IsNotExist(err) {
			subWatcher.Stop()
			delete(dw.knownDirs, path)
			dw.logger.Info("directory removed",
				zap.String("path", path))
		}
	}

	// cleanup: remove deleted files from the known list
	for path := range dw.knownFiles {
		if _, err = os.Stat(path); os.IsNotExist(err) {
			delete(dw.knownFiles, path)
			dw.logger.Info("file removed",
				zap.String("path", path))
		}
	}
}

// isUnixSocket checks if the given path is a socket
func (dw *DirectoryWatcher) isUnixSocket(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// check if it's a socket (mode bit: S_IFSOCK).
	return info.Mode()&os.ModeSocket != 0
}
