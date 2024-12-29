package internal

type MetricFilter interface {
	AddMetricFilter(string) error
	RemoveMetricFilter(string) error
	AddStateFilter(string) error
	RemoveStateFilter(string) error
}

type Filter struct {
	// used to extract a set of metrics for calculations
	// read: map[metric]MetricFilterStruct
	MetricFilters map[string]MetricFilterStruct
}

func newFilter() *Filter {
	return &Filter{
		MetricFilters: make(map[string]MetricFilterStruct),
	}
}

func (f *Filter) AddMetricFilter(key string, states map[string]bool) error {
	//fmt.Printf("setting metric filter for metric %s with states %v\n", key, states)
	if mf, exists := f.MetricFilters[key]; exists {
		// set states where necessary
		//fmt.Printf("updating metric filters %v of service %s\n", states, key)
		mf.addStates(states)
		//fmt.Println("result:", key, mf)
		return nil
	}
	//fmt.Printf("creating new metric filter for metric %s with states %v\n", key, states)
	f.MetricFilters[key] = newMetricFilterStruct()
	f.MetricFilters[key].addStates(states)
	//fmt.Println("result:", key, f.MetricFilters[key])
	return nil
}

type MetricFilterStruct struct {
	// will be relevant for deletion
	// amount of contracts using this filter
	activeContracts int
	// read map[state]filterCount
	StateFilter map[string]int
}

func (mfs MetricFilterStruct) addStates(states map[string]bool) {
	for state := range states {
		// directly increment the counter
		// if uninitialized, it will be initialized with 0
		mfs.StateFilter[state]++
	}
}

func (mfs MetricFilterStruct) removeStates(states map[string]bool) {
	for state := range states {
		if _, exists := mfs.StateFilter[state]; exists {
			mfs.StateFilter[state]--
			// remove the key, if it is not active anymore
			if mfs.StateFilter[state] <= 0 {
				delete(mfs.StateFilter, state)
			}
		}
	}
}

func newMetricFilterStruct() MetricFilterStruct {
	mfs := MetricFilterStruct{
		activeContracts: 1,
		StateFilter:     make(map[string]int),
	}
	return mfs
}
