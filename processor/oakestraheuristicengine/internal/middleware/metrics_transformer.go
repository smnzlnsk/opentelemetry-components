package middleware

import (
	"time"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type metricsTransformer struct {
	logger *zap.Logger
}

func NewMetricsTransformer(logger *zap.Logger) domain.MetricsTransformer {
	return &metricsTransformer{
		logger: logger,
	}
}

// PMetricToMap transforms pmetric.Metrics into a map of host to pmetric.Metrics
// Since we assume metrics are always from a single host, this is simplified
func (t *metricsTransformer) PMetricToMap(md pmetric.Metrics) (map[string]pmetric.Metrics, error) {
	// Get the host from the first resource metrics (assuming all are from the same host)
	host := t.ExtractHost(md)

	// Create result map with the single host
	result := make(map[string]pmetric.Metrics)
	result[host] = md

	return result, nil
}

// ExtractHost extracts the host identifier from metrics
func (t *metricsTransformer) ExtractHost(md pmetric.Metrics) string {
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

// TransformToDBHostMetrics converts OpenTelemetry metrics to our DB format
func (t *metricsTransformer) TransformToDBHostMetrics(md pmetric.Metrics) (domain.DBHostMetrics, error) {
	// Get the host
	host := t.ExtractHost(md)

	hostMetrics := domain.DBHostMetrics{
		Host:                   host,
		SystemMetrics:          []domain.DBMetricDatapoint{},
		ServiceInstanceMetrics: []domain.DBServiceInstanceMetrics{},
	}

	// Services map to track service metrics
	serviceMap := make(map[string]*domain.DBServiceInstanceMetrics)

	// Process all resource metrics
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)

		// Extract service name from resource attributes
		serviceName := ""
		if svcAttr, ok := rm.Resource().Attributes().Get("service.name"); ok {
			serviceName = svcAttr.Str()
		}

		// Process all scope metrics for this resource
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			// Process all metrics in this scope
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)

				// Skip non-gauge and non-sum metrics
				if metric.Type() != pmetric.MetricTypeGauge && metric.Type() != pmetric.MetricTypeSum {
					continue
				}

				// Extract datapoints
				datapoint := t.ExtractDatapoint(metric)
				if datapoint == nil {
					continue
				}

				// Determine if this is a system metric or service metric
				isSystemMetric := IsSystemMetric(metric.Name())

				if isSystemMetric {
					// Add to system metrics
					hostMetrics.SystemMetrics = append(hostMetrics.SystemMetrics, *datapoint)
				} else if serviceName != "" {
					// Get or create service instance
					serviceInstance, exists := serviceMap[serviceName]
					if !exists {
						// Create new service reference in map
						hostMetrics.ServiceInstanceMetrics = append(hostMetrics.ServiceInstanceMetrics, domain.DBServiceInstanceMetrics{
							JobName:        serviceName,
							InstanceNumber: 0,
							Metrics:        []domain.DBMetricDatapoint{},
						})
						serviceInstance = &hostMetrics.ServiceInstanceMetrics[len(hostMetrics.ServiceInstanceMetrics)-1]
						serviceMap[serviceName] = serviceInstance
					}

					// Add metric to service
					serviceInstance.Metrics = append(serviceInstance.Metrics, *datapoint)
				}
			}
		}
	}

	return hostMetrics, nil
}

// mergeHostMetrics merges new metrics into existing host metrics
func (t *metricsTransformer) MergeHostMetrics(existing, new domain.DBHostMetrics) domain.DBHostMetrics {
	result := existing

	// Helper function to add new datapoints to existing metrics
	mergeMetricDatapoints := func(existing []domain.DBMetricDatapoint, new []domain.DBMetricDatapoint) []domain.DBMetricDatapoint {
		// Create a map for quick lookup by identifier
		metricMap := make(map[string]int)

		for i, dp := range existing {
			key := dp.Identifier.Name + ":" + dp.Identifier.State
			metricMap[key] = i
		}

		// Add new datapoints
		for _, newDP := range new {
			key := newDP.Identifier.Name + ":" + newDP.Identifier.State

			if idx, found := metricMap[key]; found {
				// Append datapoints to existing entry
				existing[idx].Datapoints = append(existing[idx].Datapoints, newDP.Datapoints...)

				// Keep only the latest N datapoints (e.g., 100)
				if len(existing[idx].Datapoints) > 100 {
					existing[idx].Datapoints = existing[idx].Datapoints[len(existing[idx].Datapoints)-100:]
				}
			} else {
				// Add new metric
				existing = append(existing, newDP)
			}
		}

		return existing
	}

	// Merge system metrics
	result.SystemMetrics = mergeMetricDatapoints(result.SystemMetrics, new.SystemMetrics)

	// Merge service metrics
	for _, newService := range new.ServiceInstanceMetrics {
		var found bool

		// Find matching service
		for i, existingService := range result.ServiceInstanceMetrics {
			if existingService.JobName == newService.JobName &&
				existingService.InstanceNumber == newService.InstanceNumber {
				// Merge metrics for this service
				result.ServiceInstanceMetrics[i].Metrics = mergeMetricDatapoints(
					result.ServiceInstanceMetrics[i].Metrics,
					newService.Metrics,
				)
				found = true
				break
			}
		}

		if !found {
			// Add new service
			result.ServiceInstanceMetrics = append(result.ServiceInstanceMetrics, newService)
		}
	}

	return result
}

// ExtractDatapoint extracts a single datapoint from a metric
func (t *metricsTransformer) ExtractDatapoint(metric pmetric.Metric) *domain.DBMetricDatapoint {
	name := metric.Name()
	now := time.Now()

	// Create a slice to hold datapoints for this metric
	var datapoints []domain.MetricDatapoint

	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		// For simplicity just use the first datapoint
		if metric.Gauge().DataPoints().Len() > 0 {
			dp := metric.Gauge().DataPoints().At(0)
			value := getValueFromDataPoint(dp)
			state := getStateAttribute(dp.Attributes())

			datapoints = append(datapoints, domain.MetricDatapoint{
				Value:     value,
				Timestamp: now,
			})

			return &domain.DBMetricDatapoint{
				Identifier: domain.MetricKey{
					Name:  name,
					State: state,
					Type:  domain.MetricValueTypeRaw,
				},
				Datapoints: datapoints,
			}
		}

	case pmetric.MetricTypeSum:
		// For simplicity just use the first datapoint
		if metric.Sum().DataPoints().Len() > 0 {
			dp := metric.Sum().DataPoints().At(0)
			value := getValueFromDataPoint(dp)
			state := getStateAttribute(dp.Attributes())

			datapoints = append(datapoints, domain.MetricDatapoint{
				Value:     value,
				Timestamp: now,
			})

			return &domain.DBMetricDatapoint{
				Identifier: domain.MetricKey{
					Name:  name,
					State: state,
					Type:  domain.MetricValueTypeRaw,
				},
				Datapoints: datapoints,
			}
		}
	}

	return nil
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
