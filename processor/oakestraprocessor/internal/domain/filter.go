package domain

import (
	"fmt"
	"strings"
	"sync"
)

type MetricFilter interface {
	MetricFiltersMap() map[string]*metricFilterStruct
	AddMetricFilter(string, string) error
	DeleteMetricFilter(string, string) error
}

type filter struct {
	// used to extract a set of metrics for calculations
	// read: map[metric]MetricFilterStruct
	MetricFilters map[string]*metricFilterStruct
}

func NewFilter() *filter {
	return &filter{
		MetricFilters: make(map[string]*metricFilterStruct),
	}
}

func (f *filter) MetricFiltersMap() map[string]*metricFilterStruct {
	return f.MetricFilters
}

func (f *filter) AddMetricFilter(key string, state string) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}

	if state == "" {
		return fmt.Errorf("state is empty")
	}

	// If the key contains parentheses for age, remove them
	if strings.Contains(key, "(") {
		parts := strings.Split(key, "(")
		key = parts[0]
	}

	if mf, exists := f.MetricFilters[key]; exists {
		// Each call to AddMetricFilter represents a new contract using this metric
		mf.mu.Lock()
		mf.activeContracts++
		mf.mu.Unlock()
		// Add state tracking
		mf.addState(state)
		return nil
	}

	mfs := newMetricFilterStruct()
	mfs.addState(state)
	f.MetricFilters[key] = mfs
	return nil
}

func (f *filter) DeleteMetricFilter(key string, state string) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}

	if state == "" {
		return fmt.Errorf("state is empty")
	}
	// If the key contains parentheses for age, remove them
	if strings.Contains(key, "(") {
		parts := strings.Split(key, "(")
		key = parts[0]
	}

	if mfs, exists := f.MetricFilters[key]; exists {
		// Each call to DeleteMetricFilter represents removing a contract using this metric
		mfs.mu.Lock()
		mfs.activeContracts--
		contractsLeft := mfs.activeContracts
		mfs.mu.Unlock()

		// Remove state tracking
		mfs.removeState(state)

		// Delete the metric filter if no more contracts use it
		if contractsLeft <= 0 {
			delete(f.MetricFilters, key)
		}
	}
	return nil
}

type metricFilterStruct struct {
	mu              sync.RWMutex
	activeContracts int
	StateFilter     map[string]int
}

func (mfs *metricFilterStruct) ActiveContracts() int {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()
	return mfs.activeContracts
}

func (mfs *metricFilterStruct) addState(state string) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	mfs.StateFilter[state]++
}

func (mfs *metricFilterStruct) removeState(state string) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()

	if count, exists := mfs.StateFilter[state]; exists && count > 0 {
		mfs.StateFilter[state]--
		if mfs.StateFilter[state] <= 0 {
			delete(mfs.StateFilter, state)
		}
	}
}

func newMetricFilterStruct() *metricFilterStruct {
	mfs := metricFilterStruct{
		activeContracts: 1, // Start at 1 since this is the first contract using this metric
		StateFilter:     make(map[string]int),
	}
	return &mfs
}
