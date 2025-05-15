package domain

import (
	"strings"
	"sync"
)

type MetricFilter interface {
	MetricFiltersMap() map[string]*metricFilterStruct
	AddMetricFilter(string, map[string]bool) error
	DeleteMetricFilter(string, map[string]bool) error
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

func (f *filter) AddMetricFilter(key string, states map[string]bool) error {
	// if the key contains a |, we only want to work with the left part
	if strings.Contains(key, "|") {
		parts := strings.Split(key, "|")
		key = parts[0]
	}
	if mf, exists := f.MetricFilters[key]; exists {
		// Increment activeContracts for overlapping metrics
		mf.activeContracts++
		// set states where necessary
		mf.addStates(states)
		return nil
	}
	mfs := newMetricFilterStruct()
	mfs.addStates(states)
	f.MetricFilters[key] = mfs
	return nil
}

func (f *filter) DeleteMetricFilter(key string, states map[string]bool) error {
	// if the key contains a |, we only want to work with the left part
	if strings.Contains(key, "|") {
		parts := strings.Split(key, "|")
		key = parts[0]
	}
	if mfs, exists := f.MetricFilters[key]; exists {
		// First remove states
		mfs.removeStates(states)

		// Only decrement activeContracts if all states are removed
		if len(mfs.StateFilter) == 0 {
			mfs.activeContracts--

			// If no more active contracts, delete the entire metric filter
			if mfs.activeContracts <= 0 {
				delete(f.MetricFilters, key)
			}
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

func (mfs *metricFilterStruct) addStates(states map[string]bool) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	for state := range states {
		mfs.StateFilter[state]++
	}
}

func (mfs *metricFilterStruct) removeStates(states map[string]bool) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	for state := range states {
		mfs.StateFilter[state]--
		if mfs.StateFilter[state] <= 0 {
			delete(mfs.StateFilter, state)
		}
	}
}

func newMetricFilterStruct() *metricFilterStruct {
	mfs := metricFilterStruct{
		activeContracts: 1,
		StateFilter:     make(map[string]int),
	}
	return &mfs
}
