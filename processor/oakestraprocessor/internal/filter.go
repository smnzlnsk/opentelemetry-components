package internal

import "sync"

type MetricFilter interface {
	AddMetricFilter(string) error
	RemoveMetricFilter(string) error
	AddStateFilter(string) error
	RemoveStateFilter(string) error
}

type Filter struct {
	// used to extract a set of metrics for calculations
	// read: map[metric]MetricFilterStruct
	MetricFilters map[string]*MetricFilterStruct
}

func newFilter() *Filter {
	return &Filter{
		MetricFilters: make(map[string]*MetricFilterStruct),
	}
}

func (f *Filter) AddMetricFilter(key string, states map[string]bool) error {
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

func (f *Filter) DeleteMetricFilter(key string, states map[string]bool) error {
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

type MetricFilterStruct struct {
	mu              sync.RWMutex
	activeContracts int
	StateFilter     map[string]int
}

func (mfs *MetricFilterStruct) addStates(states map[string]bool) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	for state := range states {
		mfs.StateFilter[state]++
	}
}

func (mfs *MetricFilterStruct) removeStates(states map[string]bool) {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	for state := range states {
		mfs.StateFilter[state]--
		if mfs.StateFilter[state] <= 0 {
			delete(mfs.StateFilter, state)
		}
	}
}

func newMetricFilterStruct() *MetricFilterStruct {
	mfs := MetricFilterStruct{
		activeContracts: 1,
		StateFilter:     make(map[string]int),
	}
	return &mfs
}
