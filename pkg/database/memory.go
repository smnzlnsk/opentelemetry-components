package database

import (
	"fmt"
	"sync"
	"time"
)

// MemoryDatabase provides direct communication between processors
type MemoryDatabase interface {
	// PublishMetrics publishes metrics for a job
	PublishMetrics(jobName string, metrics HostMetricsMap)
	// GetMetrics retrieves the latest metrics for a job as HostMetricsMap
	GetMetrics(jobName string) (HostMetricsMap, bool)
	// Clear old metrics
	Cleanup(maxAge time.Duration)
}

// memoryDatabase implements MemoryDatabase using in-memory storage
type memoryDatabase struct {
	mu         sync.RWMutex
	metrics    map[string]HostMetricsMap
	timestamps map[string]time.Time
}

// NewMemoryDatabase creates a new in-memory database
func NewMemoryDatabase() MemoryDatabase {
	return &memoryDatabase{
		metrics:    make(map[string]HostMetricsMap),
		timestamps: make(map[string]time.Time),
	}
}

func (db *memoryDatabase) PublishMetrics(jobName string, metrics HostMetricsMap) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.metrics[jobName] = metrics
	db.timestamps[jobName] = time.Now()
}

func (db *memoryDatabase) GetMetrics(jobName string) (HostMetricsMap, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	metrics, exists := db.metrics[jobName]
	return metrics, exists
}

func (db *memoryDatabase) Cleanup(maxAge time.Duration) {
	db.mu.Lock()
	defer db.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for jobName, timestamp := range db.timestamps {
		if timestamp.Before(cutoff) {
			delete(db.metrics, jobName)
			delete(db.timestamps, jobName)
		}
	}
}

// Helper functions for metric ID building
func buildMetricID(name string, state string, age int) string {
	return name + "(" + fmt.Sprintf("%d", age) + ")" + "{" + state + "}"
}

func calculateAge(length int, index int) int {
	return (length - 1) - index
}

// Global memory database instance
var (
	globalMemoryDB     MemoryDatabase
	globalMemoryDBOnce sync.Once
)

// GetGlobalMemoryDatabase returns the global memory database instance
func GetGlobalMemoryDatabase() MemoryDatabase {
	globalMemoryDBOnce.Do(func() {
		globalMemoryDB = NewMemoryDatabase()

		// Start cleanup routine
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				globalMemoryDB.Cleanup(10 * time.Minute) // Clean up metrics older than 10 minutes
			}
		}()
	})
	return globalMemoryDB
}
