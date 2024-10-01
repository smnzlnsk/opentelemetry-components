package dirwatcher

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DirectoryWatcher watches a directory for newly added Unix sockets
type DirectoryWatcher struct {
	dir          string
	logger       *zap.Logger
	knownFiles   map[string]os.FileInfo
	socketHandle func(string)
	interval     time.Duration
	mutex        sync.Mutex
}

func NewDirectoryWatcher(dir string, logger *zap.Logger, interval time.Duration, handle func(string)) *DirectoryWatcher {
	return &DirectoryWatcher{
		dir:          dir,
		logger:       logger,
		knownFiles:   make(map[string]os.FileInfo),
		socketHandle: handle,
		interval:     interval,
	}
}

// Start starts the directory watcher, polling for changes every interval seconds
func (dw *DirectoryWatcher) Start() {
	dw.logger.Info("Starting directory watcher...")
	for {
		// Poll the directory for new files.
		dw.pollDirectory()
		time.Sleep(dw.interval)
	}
}

func (dw *DirectoryWatcher) pollDirectory() {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	// Read the directory contents.
	files, err := os.ReadDir(dw.dir)
	if err != nil {
		dw.logger.Error("error reading directory",
			zap.String("directory", dw.dir),
			zap.Error(err))
		return
	}

	for _, entry := range files {
		path := filepath.Join(dw.dir, entry.Name())

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
