package domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	metric "github.com/smnzlnsk/opentelemetry-components/pkg/metric"
	datapoint "github.com/smnzlnsk/opentelemetry-components/pkg/metric/datapoint"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type DatapointManager interface {
	GetHost(metrics pmetric.Metrics) string
	GetDatapoint(key datapoint.Key, age int) (datapoint.Datapoint, bool)
	SaveMetrics(metrics pmetric.Metrics) error
	DeleteDatapointForService(service string) error
	GetDatapoints() map[datapoint.Key]map[int]datapoint.Datapoint
	GetCurrentIndex(key datapoint.Key) int
	SaveCalculationResults(metrics pmetric.Metrics) error
	Shutdown()
}

// datapointManager is a simple implementation of the DatapointManager interface
// It stores datapoints in memory
type datapointManager struct {
	Datapoints     map[datapoint.Key]map[int]datapoint.Datapoint
	indexTracker   map[datapoint.Key]int
	metricsService MetricsService

	// Buffering for async processing
	buffer        []database.HostMetrics
	bufferMutex   sync.Mutex
	bufferSize    int
	flushInterval time.Duration
	stopChan      chan struct{}
	logger        *zap.Logger
}

func NewDatapointManager(logger *zap.Logger, metricsService MetricsService) DatapointManager {
	dm := &datapointManager{
		Datapoints:     make(map[datapoint.Key]map[int]datapoint.Datapoint),
		indexTracker:   make(map[datapoint.Key]int),
		metricsService: metricsService,
		buffer:         make([]database.HostMetrics, 0),
		bufferSize:     100,
		flushInterval:  1 * time.Second,
		stopChan:       make(chan struct{}),
		logger:         logger,
	}

	// Start background flusher
	go dm.backgroundFlusher()

	return dm
}

// backgroundFlusher periodically flushes buffered metrics to database
func (d *datapointManager) backgroundFlusher() {
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.flushBuffer()
		case <-d.stopChan:
			// Final flush before stopping
			d.flushBuffer()
			return
		}
	}
}

// flushBuffer writes all buffered metrics to database
func (d *datapointManager) flushBuffer() {
	d.bufferMutex.Lock()
	defer d.bufferMutex.Unlock()

	if len(d.buffer) == 0 {
		return
	}

	// Create a copy and clear the buffer
	metricsToSave := make([]database.HostMetrics, len(d.buffer))
	copy(metricsToSave, d.buffer)
	d.buffer = d.buffer[:0] // Clear buffer

	// Save to database asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Use batch save if available, otherwise fall back to individual saves
		if batchSaver, ok := d.metricsService.(interface {
			SaveMetricsBatch(ctx context.Context, metrics []database.HostMetrics) error
		}); ok {
			if err := batchSaver.SaveMetricsBatch(ctx, metricsToSave); err != nil {
				if d.logger != nil {
					d.logger.Error("Failed to batch save metrics", zap.Error(err))
				}
			}
		} else {
			// Fallback to individual saves
			for _, metrics := range metricsToSave {
				if err := d.metricsService.SaveMetrics(ctx, metrics); err != nil {
					if d.logger != nil {
						d.logger.Error("Failed to save metrics", zap.Error(err))
					}
				}
			}
		}
	}()
}

// addToBuffer adds metrics to buffer and flushes if needed
func (d *datapointManager) addToBuffer(metrics database.HostMetrics) {
	d.bufferMutex.Lock()
	defer d.bufferMutex.Unlock()

	d.buffer = append(d.buffer, metrics)

	// Flush if buffer is full
	if len(d.buffer) >= d.bufferSize {
		go d.flushBuffer() // Async flush to avoid blocking
	}
}

// Shutdown gracefully shuts down the datapoint manager
func (d *datapointManager) Shutdown() {
	close(d.stopChan)
	// Give some time for final flush
	time.Sleep(100 * time.Millisecond)
}

// String returns a string representation of the DatapointManager state
// Implements the fmt.Stringer interface to allow for easy debugging / logging
func (d *datapointManager) String() string {
	str := "DatapointManager State:\nCurrent"
	for key, dps := range d.Datapoints {
		str += fmt.Sprintf("\tKey: %v\n", key)
		str += fmt.Sprintf("\t\tCurrent Index: %v\n", d.indexTracker[key])
		for age, dp := range dps {
			str += fmt.Sprintf("\t\tIndex: %v, Value: %v\n", age, dp.Value)
		}
	}
	return str
}

func (d *datapointManager) GetDatapoints() map[datapoint.Key]map[int]datapoint.Datapoint {
	return d.Datapoints
}

func (d *datapointManager) GetCurrentIndex(key datapoint.Key) int {
	return d.indexTracker[key]
}

func (d *datapointManager) GetDatapoint(key datapoint.Key, age int) (datapoint.Datapoint, bool) {
	if dps, exists := d.Datapoints[key]; exists {
		currentIdx := d.indexTracker[key]
		lookupIdx := (currentIdx - 1 - age + 5) % 5
		if dp, exists := dps[lookupIdx]; exists {
			return dp, true
		}
	}
	return datapoint.Datapoint{}, false
}

// SaveCalculationResults saves the calculation results to the database
func (d *datapointManager) SaveCalculationResults(metrics pmetric.Metrics) error {
	host := d.GetHost(metrics)

	dbMetrics := database.HostMetrics{
		Host:                   host,
		SystemMetrics:          []database.MetricDatapoints{},
		ServiceInstanceMetrics: []database.ServiceInstanceMetrics{},
	}

	// Services map to track service metrics - use map for grouping
	serviceMap := make(map[string]*database.ServiceInstanceMetrics)

	// Process all resource metrics
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)

		// Extract service name from resource attributes
		serviceName := ""
		if svcAttr, ok := rm.Resource().Attributes().Get("container_id"); ok {
			serviceName = svcAttr.Str()
		}

		jobName, instanceNumber := splitServiceName(serviceName)

		// Process all scope metrics for this resource
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			// Check if the scope metric is coming from the oakestraprocessor
			if !strings.Contains(sm.Scope().Name(), "oakestraprocessor/internal/processor") {
				continue
			}

			// Process all metrics in this scope
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)

				// Extract datapoints
				datapoints := d.extractDatapointsFromMetric(m)
				if datapoints == nil {
					continue
				}

				// Determine if this is a container metric
				if IsContainerMetric(m.Name()) {
					for _, dp := range datapoints {
						// Group metrics by service instance
						if serviceMap[serviceName] == nil {
							serviceMap[serviceName] = &database.ServiceInstanceMetrics{
								JobName:        jobName,
								InstanceNumber: instanceNumber,
								Metrics:        []database.MetricDatapoints{},
							}
						}

						// Add the metric datapoint to the service instance
						serviceMap[serviceName].Metrics = append(serviceMap[serviceName].Metrics, database.MetricDatapoints{
							Identifier: metric.Key{
								Name:  m.Name(),
								State: dp.state,
								Type:  metric.MetricValueTypeRaw,
							},
							Datapoints: []database.MetricDatapoint{
								{
									Value:     dp.value,
									Timestamp: time.Now(),
								},
							},
						})
					}
				}
			}
		}
	}

	// Convert the service map to a slice for BSON compatibility
	for _, serviceMetrics := range serviceMap {
		dbMetrics.ServiceInstanceMetrics = append(dbMetrics.ServiceInstanceMetrics, *serviceMetrics)
	}

	// Add metrics to buffer for async processing
	d.addToBuffer(dbMetrics)

	return nil
}

// SaveMetrics saves the incoming metrics to memory and simulatenously collects information to save to the database
func (d *datapointManager) SaveMetrics(metrics pmetric.Metrics) error {
	// Get the host
	host := d.GetHost(metrics)

	dbMetrics := database.HostMetrics{
		Host:                   host,
		SystemMetrics:          []database.MetricDatapoints{},
		ServiceInstanceMetrics: []database.ServiceInstanceMetrics{},
	}

	// Services map to track service metrics - use map for grouping
	serviceMap := make(map[string]*database.ServiceInstanceMetrics)

	// Process all resource metrics
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)

		// Extract service name from resource attributes
		serviceName := ""
		if svcAttr, ok := rm.Resource().Attributes().Get("container_id"); ok {
			serviceName = svcAttr.Str()
		}

		jobName, instanceNumber := splitServiceName(serviceName)

		// Process all scope metrics for this resource
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			// Process all metrics in this scope
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				metricType := m.Type()

				// Skip non-gauge and non-sum metrics
				if metricType != pmetric.MetricTypeGauge && metricType != pmetric.MetricTypeSum {
					continue
				}

				// Extract datapoints
				datapoints := d.extractDatapointsFromMetric(m)
				if datapoints == nil {
					continue
				}

				// Determine if this is a container metric
				if IsContainerMetric(m.Name()) {
					for _, dp := range datapoints {
						key := datapoint.Key{
							Service: serviceName,
							Metric:  m.Name(),
							State:   dp.state,
						}
						id := d.indexTracker[key]
						if d.Datapoints[key] == nil {
							d.Datapoints[key] = make(map[int]datapoint.Datapoint)
						}
						d.Datapoints[key][id] = datapoint.Datapoint{
							Value: dp.value,
						}
						d.indexTracker[key] = (id + 1) % 5

						// Group metrics by service instance
						if serviceMap[serviceName] == nil {
							serviceMap[serviceName] = &database.ServiceInstanceMetrics{
								JobName:        jobName,
								InstanceNumber: instanceNumber,
								Metrics:        []database.MetricDatapoints{},
							}
						}

						// Add the metric datapoint to the service instance
						serviceMap[serviceName].Metrics = append(serviceMap[serviceName].Metrics, database.MetricDatapoints{
							Identifier: metric.Key{
								Name:  m.Name(),
								State: dp.state,
								Type:  metric.MetricValueTypeRaw,
							},
							Datapoints: []database.MetricDatapoint{
								{
									Value:     dp.value,
									Timestamp: time.Now(),
								},
							},
						})
					}
				} else {
					for _, dp := range datapoints {
						key := datapoint.Key{
							Service: "",
							Metric:  m.Name(),
							State:   dp.state,
						}

						id := d.indexTracker[key]
						if d.Datapoints[key] == nil {
							d.Datapoints[key] = make(map[int]datapoint.Datapoint)
						}
						d.Datapoints[key][id] = datapoint.Datapoint{
							Value: dp.value,
						}
						d.indexTracker[key] = (id + 1) % 5

						dbMetrics.SystemMetrics = append(dbMetrics.SystemMetrics, database.MetricDatapoints{
							Identifier: metric.Key{
								Name:  m.Name(),
								State: dp.state,
								Type:  metric.MetricValueTypeRaw,
							},
							Datapoints: []database.MetricDatapoint{
								{
									Value:     dp.value,
									Timestamp: time.Now(),
								},
							},
						})
					}
				}
			}
		}
	}

	// Convert the service map to a slice for BSON compatibility
	for _, serviceMetrics := range serviceMap {
		dbMetrics.ServiceInstanceMetrics = append(dbMetrics.ServiceInstanceMetrics, *serviceMetrics)
	}

	// Add metrics to buffer for async processing
	d.addToBuffer(dbMetrics)

	// Publish to metrics broker for direct processor communication
	// Group metrics by job name for publishing
	jobMetrics := make(map[string]database.HostMetrics)

	d.logger.Info("Processing metrics for broker publishing",
		zap.String("host", dbMetrics.Host),
		zap.Int("total_service_instances", len(dbMetrics.ServiceInstanceMetrics)),
		zap.Int("total_system_metrics", len(dbMetrics.SystemMetrics)))

	for _, serviceInstance := range dbMetrics.ServiceInstanceMetrics {
		jobName := serviceInstance.JobName
		d.logger.Info("Processing service instance for broker",
			zap.String("job_name", jobName),
			zap.Int("instance_number", serviceInstance.InstanceNumber),
			zap.Int("metrics_count", len(serviceInstance.Metrics)))

		if _, exists := jobMetrics[jobName]; !exists {
			jobMetrics[jobName] = database.HostMetrics{
				Host:                   dbMetrics.Host,
				SystemMetrics:          dbMetrics.SystemMetrics, // Include system metrics for all jobs
				ServiceInstanceMetrics: []database.ServiceInstanceMetrics{},
			}
		}

		// Add this service instance to the job's metrics
		jobMetrics[jobName] = database.HostMetrics{
			Host:                   jobMetrics[jobName].Host,
			SystemMetrics:          jobMetrics[jobName].SystemMetrics,
			ServiceInstanceMetrics: append(jobMetrics[jobName].ServiceInstanceMetrics, serviceInstance),
		}
	}

	// Publish each job's metrics to the broker
	broker := database.GetGlobalMetricsBroker()

	if d.logger != nil {
		d.logger.Info("Publishing to broker",
			zap.Int("total_jobs", len(jobMetrics)))

		// Log all job names being published
		jobNames := make([]string, 0, len(jobMetrics))
		for jobName := range jobMetrics {
			jobNames = append(jobNames, jobName)
		}
		d.logger.Info("Job names being published", zap.Strings("job_names", jobNames))
	}

	for jobName, metrics := range jobMetrics {
		broker.PublishMetrics(jobName, metrics)
		if d.logger != nil {
			d.logger.Info("Published metrics to broker",
				zap.String("job_name", jobName),
				zap.Int("service_instances", len(metrics.ServiceInstanceMetrics)),
				zap.Int("system_metrics", len(metrics.SystemMetrics)))
		}

		// Verify the metrics were actually stored in the broker
		if verifyMetrics, exists := broker.GetMetrics(jobName); exists {
			if d.logger != nil {
				d.logger.Info("Verified metrics in broker",
					zap.String("job_name", jobName),
					zap.Int("verified_service_instances", len(verifyMetrics.ServiceInstanceMetrics)))
			}
		} else {
			if d.logger != nil {
				d.logger.Error("Failed to verify metrics in broker", zap.String("job_name", jobName))
			}
		}
	}

	return nil
}

func (d *datapointManager) DeleteDatapointForService(service string) error {
	for key := range d.Datapoints {
		if key.Service == service {
			delete(d.Datapoints, key)
		}
	}
	return nil
}

// IsContainerMetrics determines if a metric is a container metric based on its name
func IsContainerMetric(metricName string) bool {
	// Common prefixes for container metrics
	containerPrefixes := []string{
		"container_",
		"container.",
		"service_",
		"service.",
		"pod.",
		"node.",
		"namespace.",
	}

	for _, prefix := range containerPrefixes {
		if strings.HasPrefix(metricName, prefix) {
			return true
		}
	}
	return false
}

// IsSystemMetric determines if a metric is a system metric based on its name
func IsSystemMetric(metricName string) bool {
	// Common prefixes for system metrics
	systemPrefixes := []string{
		"system.",
		"host.",
		"os.",
		"cpu.",
		"memory.",
		"disk.",
		"network.",
		"process.",
	}

	for _, prefix := range systemPrefixes {
		if len(metricName) >= len(prefix) && metricName[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// Helper functions
func getStateAttribute(attrs pcommon.Map) string {
	if state, ok := attrs.Get("state"); ok {
		return state.Str()
	}
	return "default"
}

func getValueFromDataPoint(dp pmetric.NumberDataPoint) float64 {
	switch dp.ValueType() {
	case pmetric.NumberDataPointValueTypeDouble:
		return dp.DoubleValue()
	case pmetric.NumberDataPointValueTypeInt:
		return float64(dp.IntValue())
	default:
		return 0
	}
}

func splitServiceName(input string) (jobName string, instanceNum int) {
	lastDotIndex := strings.LastIndex(input, ".")
	if lastDotIndex != -1 {
		// instance number is the last part of the service name
		// we don't care about the error here because we assume the input is a valid service name
		// f not, check the Oakestra backend
		instanceNum, _ = strconv.Atoi(input[lastDotIndex+1:])

		// job name is the part before the last dot
		secondLastDotIndex := strings.LastIndex(input[:lastDotIndex], ".")
		if secondLastDotIndex != -1 {
			jobName = input[:secondLastDotIndex]
		}
	}
	return
}

// ExtractHost extracts the host identifier from metrics
func (d *datapointManager) GetHost(md pmetric.Metrics) string {
	// Default host if we can't find one
	host := "unknown"

	// Check only the first resource metrics
	if md.ResourceMetrics().Len() > 0 {
		rm := md.ResourceMetrics().At(0)

		// Try to extract host information from 'machine' attribute or fall back to 'host.name'
		if machineAttr, ok := rm.Resource().Attributes().Get("machine"); ok {
			host = machineAttr.Str()
		} else if hostAttr, ok := rm.Resource().Attributes().Get("host.name"); ok {
			host = hostAttr.Str()
		}
	}

	return host
}

type extractedDatapoint struct {
	state string
	value float64
}

// extractDatapointsFromMetric extracts the datapoints from a metric
// if a state attribute is present, it will be used to create a new datapoint
func (d *datapointManager) extractDatapointsFromMetric(metric pmetric.Metric) []extractedDatapoint {
	// Create a slice to hold datapoints for this metric
	var datapoints []extractedDatapoint

	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		if metric.Gauge().DataPoints().Len() > 0 {
			for i := 0; i < metric.Gauge().DataPoints().Len(); i++ {
				dp := metric.Gauge().DataPoints().At(i)
				value := getValueFromDataPoint(dp)
				state := getStateAttribute(dp.Attributes())

				datapoint := extractedDatapoint{
					state: state,
					value: value,
				}

				datapoints = append(datapoints, datapoint)
			}
		}

	case pmetric.MetricTypeSum:
		if metric.Sum().DataPoints().Len() > 0 {
			for i := 0; i < metric.Sum().DataPoints().Len(); i++ {
				dp := metric.Sum().DataPoints().At(i)
				value := getValueFromDataPoint(dp)
				state := getStateAttribute(dp.Attributes())

				datapoint := extractedDatapoint{
					state: state,
					value: value,
				}

				datapoints = append(datapoints, datapoint)
			}
		}
	}

	return datapoints
}
