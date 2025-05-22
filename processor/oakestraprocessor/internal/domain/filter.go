package domain

import (
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
	// If the key contains parentheses for age, remove them
	if strings.Contains(key, "(") {
		parts := strings.Split(key, "(")
		key = parts[0]
	}

	if mf, exists := f.MetricFilters[key]; exists {
		// Increment activeContracts for overlapping metrics
		mf.activeContracts++
		// set states where necessary
		mf.addState(state)
		return nil
	}
	mfs := newMetricFilterStruct()
	mfs.addState(state)
	f.MetricFilters[key] = mfs
	return nil
}

func (f *filter) DeleteMetricFilter(key string, state string) error {
	// If the key contains parentheses for age, remove them
	if strings.Contains(key, "(") {
		parts := strings.Split(key, "(")
		key = parts[0]
	}

	if mfs, exists := f.MetricFilters[key]; exists {
		// First remove states
		mfs.removeState(state)

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

func (mfs *metricFilterStruct) addState(state string) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	mfs.StateFilter[state]++
}

func (mfs *metricFilterStruct) removeState(state string) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	mfs.StateFilter[state]--
	if mfs.StateFilter[state] <= 0 {
		delete(mfs.StateFilter, state)
	}
}

func newMetricFilterStruct() *metricFilterStruct {
	mfs := metricFilterStruct{
		activeContracts: 1,
		StateFilter:     make(map[string]int),
	}
	return &mfs
}
