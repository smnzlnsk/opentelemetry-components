package domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type DatapointManager interface {
	GetHost(metrics pmetric.Metrics) string
	GetDatapoint(key DatapointKey, age int) (Datapoint, bool)
	SaveMetrics(metrics pmetric.Metrics) error
	DeleteDatapointForService(service string) error
	GetDatapoints() map[DatapointKey]map[int]Datapoint
	GetCurrentIndex(key DatapointKey) int
	SaveCalculationResults(metrics pmetric.Metrics) error
}

// datapointManager is a simple implementation of the DatapointManager interface
// It stores datapoints in memory
type datapointManager struct {
	Datapoints     map[DatapointKey]map[int]Datapoint
	indexTracker   map[DatapointKey]int
	metricsService MetricsService
}

func NewDatapointManager(metricsService MetricsService) DatapointManager {
	return &datapointManager{
		Datapoints:     make(map[DatapointKey]map[int]Datapoint),
		indexTracker:   make(map[DatapointKey]int),
		metricsService: metricsService,
	}
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

func (d *datapointManager) GetDatapoints() map[DatapointKey]map[int]Datapoint {
	return d.Datapoints
}

func (d *datapointManager) GetCurrentIndex(key DatapointKey) int {
	return d.indexTracker[key]
}

func (d *datapointManager) GetDatapoint(key DatapointKey, age int) (Datapoint, bool) {
	if dps, exists := d.Datapoints[key]; exists {
		currentIdx := d.indexTracker[key]
		lookupIdx := (currentIdx - 1 - age + 5) % 5
		if dp, exists := dps[lookupIdx]; exists {
			return dp, true
		}
	}
	return Datapoint{}, false
}

// SaveCalculationResults saves the calculation results to the database
func (d *datapointManager) SaveCalculationResults(metrics pmetric.Metrics) error {
	host := d.GetHost(metrics)

	dbMetrics := DBHostMetrics{
		Host:                   host,
		SystemMetrics:          []DBMetricDatapoints{},
		ServiceInstanceMetrics: []DBServiceInstanceMetrics{},
	}

	// Services map to track service metrics - use map for grouping
	serviceMap := make(map[string]*DBServiceInstanceMetrics)

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
				metric := sm.Metrics().At(k)

				// Extract datapoints
				datapoints := d.extractDatapointsFromMetric(metric)
				if datapoints == nil {
					continue
				}

				// Determine if this is a container metric
				if IsContainerMetric(metric.Name()) {
					for _, dp := range datapoints {
						// Group metrics by service instance
						if serviceMap[serviceName] == nil {
							serviceMap[serviceName] = &DBServiceInstanceMetrics{
								JobName:        jobName,
								InstanceNumber: instanceNumber,
								Metrics:        []DBMetricDatapoints{},
							}
						}

						// Add the metric datapoint to the service instance
						serviceMap[serviceName].Metrics = append(serviceMap[serviceName].Metrics, DBMetricDatapoints{
							Identifier: MetricKey{
								Name:  metric.Name(),
								State: dp.state,
								Type:  MetricValueTypeRaw,
							},
							Datapoints: []DBMetricDatapoint{
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

	// Save metrics to database
	err := d.metricsService.SaveMetrics(context.Background(), dbMetrics)
	if err != nil {
		return fmt.Errorf("failed to save metrics to database: %w", err)
	}

	return nil
}

// SaveMetrics saves the incoming metrics to memory and simulatenously collects information to save to the database
func (d *datapointManager) SaveMetrics(metrics pmetric.Metrics) error {
	// Get the host
	host := d.GetHost(metrics)

	dbMetrics := DBHostMetrics{
		Host:                   host,
		SystemMetrics:          []DBMetricDatapoints{},
		ServiceInstanceMetrics: []DBServiceInstanceMetrics{},
	}

	// Services map to track service metrics - use map for grouping
	serviceMap := make(map[string]*DBServiceInstanceMetrics)

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
				metric := sm.Metrics().At(k)
				metricType := metric.Type()

				// Skip non-gauge and non-sum metrics
				if metricType != pmetric.MetricTypeGauge && metricType != pmetric.MetricTypeSum {
					continue
				}

				// Extract datapoints
				datapoints := d.extractDatapointsFromMetric(metric)
				if datapoints == nil {
					continue
				}

				// Determine if this is a container metric
				if IsContainerMetric(metric.Name()) {
					for _, dp := range datapoints {
						key := DatapointKey{
							Service: serviceName,
							Metric:  metric.Name(),
							State:   dp.state,
						}
						id := d.indexTracker[key]
						if d.Datapoints[key] == nil {
							d.Datapoints[key] = make(map[int]Datapoint)
						}
						d.Datapoints[key][id] = Datapoint{
							Value: dp.value,
						}
						d.indexTracker[key] = (id + 1) % 5

						// Group metrics by service instance
						if serviceMap[serviceName] == nil {
							serviceMap[serviceName] = &DBServiceInstanceMetrics{
								JobName:        jobName,
								InstanceNumber: instanceNumber,
								Metrics:        []DBMetricDatapoints{},
							}
						}

						// Add the metric datapoint to the service instance
						serviceMap[serviceName].Metrics = append(serviceMap[serviceName].Metrics, DBMetricDatapoints{
							Identifier: MetricKey{
								Name:  metric.Name(),
								State: dp.state,
								Type:  MetricValueTypeRaw,
							},
							Datapoints: []DBMetricDatapoint{
								{
									Value:     dp.value,
									Timestamp: time.Now(),
								},
							},
						})
					}
				} else {
					for _, dp := range datapoints {
						key := DatapointKey{
							Service: "",
							Metric:  metric.Name(),
							State:   dp.state,
						}

						id := d.indexTracker[key]
						if d.Datapoints[key] == nil {
							d.Datapoints[key] = make(map[int]Datapoint)
						}
						d.Datapoints[key][id] = Datapoint{
							Value: dp.value,
						}
						d.indexTracker[key] = (id + 1) % 5

						dbMetrics.SystemMetrics = append(dbMetrics.SystemMetrics, DBMetricDatapoints{
							Identifier: MetricKey{
								Name:  metric.Name(),
								State: dp.state,
								Type:  MetricValueTypeRaw,
							},
							Datapoints: []DBMetricDatapoint{
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

	// Save metrics to database
	err := d.metricsService.SaveMetrics(context.Background(), dbMetrics)
	if err != nil {
		return fmt.Errorf("failed to save metrics to database: %w", err)
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
