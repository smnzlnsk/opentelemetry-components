package database

import (
	"sync"
	"time"
)

// MemoryDatabase provides direct communication between processors
type MemoryDatabase interface {
	// SaveMetrics saves metrics for a host
	SaveMetrics(host string, metrics HostMetricsMap)
	// GetMetrics retrieves the latest metrics for a job as HostMetricsMap
	GetMetrics(jobName string) (HostMetricsMap, bool)
	// GetHostFromService retrieves the host for a service
	GetHostFromService(service string) string
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

func (db *memoryDatabase) SaveMetrics(host string, metrics HostMetricsMap) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.metrics[host] = metrics
	db.timestamps[host] = time.Now()

	/*
		fmt.Println("Snapshot of memory database:")
		for host, metrics := range db.metrics {
			fmt.Println(host)
			fmt.Println(metrics.String())
		}
	*/
}

func (db *memoryDatabase) GetMetrics(jobName string) (HostMetricsMap, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(HostMetricsMap)
	found := false

	// Search through all stored metrics for hosts that have service instances matching the job pattern
	for _, hostMetricsMap := range db.metrics {
		// Iterate through all entries in the HostMetricsMap (which is map[string]MetricsMap)
		for hostId, metricsMap := range hostMetricsMap {
			// Check if this host has any service instance metrics matching the job pattern
			hasMatchingInstances := false
			filteredServiceInstanceMetrics := make(map[string]map[string]float64)

			// Filter service instance metrics to only include those matching the job pattern
			for serviceIdentifier, serviceMetrics := range metricsMap.ServiceInstanceMetrics {
				// Check if the service identifier starts with the job name pattern
				// Service identifiers are in format: job_name.instance.instance_number
				// We want to match against the job_name part
				if len(serviceIdentifier) > len(jobName) &&
					serviceIdentifier[:len(jobName)] == jobName &&
					(len(serviceIdentifier) == len(jobName) || serviceIdentifier[len(jobName)] == '.') {
					filteredServiceInstanceMetrics[serviceIdentifier] = serviceMetrics
					hasMatchingInstances = true
				}
			}

			// If this host entry has matching service instances, include it with filtered metrics
			if hasMatchingInstances {
				// Create a new MetricsMap with all host metrics but only filtered service instance metrics
				filteredMetricsMap := MetricsMap{
					HostMetrics:            metricsMap.HostMetrics,         // Include all host-level system metrics
					ServiceInstanceMetrics: filteredServiceInstanceMetrics, // Only matching service instances
				}
				result[hostId] = filteredMetricsMap
				found = true
			}
		}
	}
	return result, found
}

func (db *memoryDatabase) GetHostFromService(service string) string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for host, metrics := range db.metrics {
		for _, metricsMap := range metrics {
			for serviceIdentifier := range metricsMap.ServiceInstanceMetrics {
				if serviceIdentifier == service {
					return host
				}
			}
		}
	}
	return ""
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
