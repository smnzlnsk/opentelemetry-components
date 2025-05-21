package dirwatcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockDirectoryWatcher extends DirectoryWatcher and overrides isUnixSocket for testing
type mockDirectoryWatcher struct {
	*DirectoryWatcher
	mockSockets map[string]bool
	mockMutex   sync.Mutex
}

func newMockDirectoryWatcher(dir string, logger *zap.Logger, interval time.Duration, handle func(string)) *mockDirectoryWatcher {
	dw := NewDirectoryWatcher(dir, logger, interval, handle)
	return &mockDirectoryWatcher{
		DirectoryWatcher: dw,
		mockSockets:      make(map[string]bool),
	}
}

// Override the isUnixSocket method for testing
func (mdw *mockDirectoryWatcher) isUnixSocket(path string) bool {
	mdw.mockMutex.Lock()
	defer mdw.mockMutex.Unlock()
	return mdw.mockSockets[path]
}

// Override the Start method to create mock directory watchers for subdirectories
func (mdw *mockDirectoryWatcher) Start() {
	mdw.logger.Info("Starting mock directory watcher...", zap.String("directory", mdw.dir))

	// Override the original Start method logic
	go func() {
		for {
			select {
			case <-mdw.stopChan:
				mdw.logger.Info("Stopping mock directory watcher", zap.String("directory", mdw.dir))
				return
			default:
				// Call a custom poll method that creates mock watchers
				mdw.mockPollDirectory()
				time.Sleep(mdw.interval)
			}
		}
	}()
}

// Custom poll method that creates mock watchers for subdirectories
func (mdw *mockDirectoryWatcher) mockPollDirectory() {
	mdw.mutex.Lock()
	defer mdw.mutex.Unlock()

	// If stopped, don't continue
	if mdw.stopped {
		return
	}

	// Read the directory contents.
	entries, err := os.ReadDir(mdw.dir)
	if err != nil {
		mdw.logger.Error("error reading directory",
			zap.String("directory", mdw.dir),
			zap.Error(err))
		return
	}

	for _, entry := range entries {
		path := filepath.Join(mdw.dir, entry.Name())

		if entry.IsDir() {
			// Found a directory, check if we're already watching it
			if _, known := mdw.knownDirs[path]; !known {
				mdw.logger.Info("new directory detected", zap.String("path", path))
				// Create a new mock watcher for this directory
				subDirWatcher := newMockDirectoryWatcher(path, mdw.logger, mdw.interval, mdw.socketHandle)

				// Share the mockSockets map with the child watcher
				subDirWatcher.mockSockets = mdw.mockSockets

				mdw.knownDirs[path] = subDirWatcher.DirectoryWatcher
				subDirWatcher.Start()
			}
			continue
		}

		// check if the file is a new file and is a socket (using our mock method)
		if _, known := mdw.knownFiles[path]; !known && mdw.isUnixSocket(path) {
			// add new file to known files
			info, err := entry.Info()
			if err == nil {
				mdw.knownFiles[path] = info
				mdw.logger.Info("new unix socket detected",
					zap.String("path", path))
				// connect to the socket in a goroutine
				go mdw.socketHandle(path)
			}
		}
	}

	// cleanup: remove deleted directories from the known list
	for path, subWatcher := range mdw.knownDirs {
		if _, err = os.Stat(path); os.IsNotExist(err) {
			subWatcher.Stop()
			delete(mdw.knownDirs, path)
			mdw.logger.Info("directory removed",
				zap.String("path", path))
		}
	}

	// cleanup: remove deleted files from the known list
	for path := range mdw.knownFiles {
		if _, err = os.Stat(path); os.IsNotExist(err) {
			delete(mdw.knownFiles, path)
			mdw.logger.Info("file removed",
				zap.String("path", path))
		}
	}
}

// markAsSocket marks a file as a socket for testing purposes
func (mdw *mockDirectoryWatcher) markAsSocket(path string) {
	mdw.mockMutex.Lock()
	defer mdw.mockMutex.Unlock()
	mdw.mockSockets[path] = true
}

func TestDirectoryWatcher(t *testing.T) {
	// Create a temporary directory for testing
	testDir, err := os.MkdirTemp("", "dirwatcher_test")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	// Create a logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Create a channel to receive socket paths
	socketPaths := make(chan string, 10)
	var socketsMu sync.Mutex
	discoveredSockets := make(map[string]bool)

	// Create a socket handler function
	handleSocket := func(path string) {
		socketsMu.Lock()
		defer socketsMu.Unlock()
		discoveredSockets[path] = true
		socketPaths <- path
	}

	// Create a mock directory watcher
	watcher := newMockDirectoryWatcher(testDir, logger, 100*time.Millisecond, handleSocket)
	// Start the watcher
	watcher.Start()
	// Make sure to stop the watcher when the test is done
	defer watcher.Stop()

	// Test case 1: Create a socket file in the root directory
	t.Run("SocketInRootDir", func(t *testing.T) {
		socketPath := filepath.Join(testDir, "test.sock")

		// Create a dummy file
		_, err := os.Create(socketPath)
		require.NoError(t, err)

		// Mark it as a socket in our mock
		watcher.markAsSocket(socketPath)

		// Wait for the watcher to pick up the socket
		select {
		case path := <-socketPaths:
			assert.Equal(t, socketPath, path)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for socket to be detected")
		}
	})

	// Test case 2: Create a socket file in a subdirectory
	t.Run("SocketInSubDir", func(t *testing.T) {
		// Create a subdirectory
		subDir := filepath.Join(testDir, "subdir")
		err := os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		// Wait a bit for the directory to be detected
		time.Sleep(300 * time.Millisecond)

		// Create a socket in the subdirectory
		socketPath := filepath.Join(subDir, "subdir.sock")
		_, err = os.Create(socketPath)
		require.NoError(t, err)

		// Mark it as a socket in our mock
		watcher.markAsSocket(socketPath)

		// Wait for the watcher to pick up the socket
		select {
		case path := <-socketPaths:
			assert.Equal(t, socketPath, path)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for socket in subdirectory to be detected")
		}
	})

	// Test case 3: Create a nested subdirectory with a socket
	t.Run("SocketInNestedSubDir", func(t *testing.T) {
		// Create a nested subdirectory
		nestedDir := filepath.Join(testDir, "subdir", "nesteddir")
		err := os.Mkdir(nestedDir, 0755)
		require.NoError(t, err)

		// Wait a bit for the directory to be detected
		time.Sleep(300 * time.Millisecond)

		// Create a socket in the nested subdirectory
		socketPath := filepath.Join(nestedDir, "nested.sock")
		_, err = os.Create(socketPath)
		require.NoError(t, err)

		// Mark it as a socket in our mock
		watcher.markAsSocket(socketPath)

		// Wait for the watcher to pick up the socket
		select {
		case path := <-socketPaths:
			assert.Equal(t, socketPath, path)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for socket in nested subdirectory to be detected")
		}
	})

	// Test cleanup of removed directories and files
	t.Run("CleanupRemovedItems", func(t *testing.T) {
		// First, verify the subdirectories exist in knownDirs
		subDir := filepath.Join(testDir, "subdir")
		nestedDir := filepath.Join(testDir, "subdir", "nesteddir")

		time.Sleep(200 * time.Millisecond) // Wait for poll cycle

		watcher.mutex.Lock()
		assert.Contains(t, watcher.knownDirs, subDir)
		subWatcher := watcher.knownDirs[subDir]
		watcher.mutex.Unlock()

		subWatcher.mutex.Lock()
		assert.Contains(t, subWatcher.knownDirs, nestedDir)
		subWatcher.mutex.Unlock()

		// Now remove the nested directory
		err = os.RemoveAll(nestedDir)
		require.NoError(t, err)

		// Wait for cleanup
		time.Sleep(200 * time.Millisecond)

		// Verify the nested directory is removed from knownDirs
		subWatcher.mutex.Lock()
		assert.NotContains(t, subWatcher.knownDirs, nestedDir)
		subWatcher.mutex.Unlock()
	})
}
