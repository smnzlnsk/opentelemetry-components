package transformers

import (
	"strconv"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/database"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
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

// TransformDBHostMetricsToMap transforms DBHostMetrics to a MapHostMetrics
func (t *metricsTransformer) TransformDBHostMetricsToMap(dbHostMetrics database.HostMetrics) (database.MapHostMetrics, error) {
	hostMetrics := make(database.MapHostMetrics)

	// Initialize the host entry with empty maps
	hostMetrics[dbHostMetrics.Host] = database.MetricsStruct{
		HostMetrics:            make(map[string]float64),
		ServiceInstanceMetrics: make(map[string]database.ServiceInstanceMetricsMap),
	}

	// Process system metrics
	for _, systemMetric := range dbHostMetrics.SystemMetrics {
		for i, datapoint := range systemMetric.Datapoints {
			// Create a unique identifier for the metric
			metricID := systemMetric.Identifier.Name + "|" + systemMetric.Identifier.State + "|" + strconv.Itoa(i)
			hostMetrics[dbHostMetrics.Host].HostMetrics[metricID] = datapoint.Value
		}
	}

	// Process service instance metrics
	for _, serviceInstance := range dbHostMetrics.ServiceInstanceMetrics {
		// Create the service identifier
		serviceID := serviceInstance.JobName + ".instance." + strconv.Itoa(serviceInstance.InstanceNumber)

		// Initialize the service metrics map if it doesn't exist
		if _, exists := hostMetrics[dbHostMetrics.Host].ServiceInstanceMetrics[serviceID]; !exists {
			hostMetrics[dbHostMetrics.Host].ServiceInstanceMetrics[serviceID] = make(database.ServiceInstanceMetricsMap)
		}

		// Add each metric datapoint
		for _, metric := range serviceInstance.Metrics {
			for i, datapoint := range metric.Datapoints {
				// Create a unique identifier for the metric
				metricID := metric.Identifier.Name + "|" + metric.Identifier.State + "|" + strconv.Itoa(i)
				hostMetrics[dbHostMetrics.Host].ServiceInstanceMetrics[serviceID][metricID] = datapoint.Value
			}
		}
	}

	return hostMetrics, nil
}

// mergeHostMetrics merges new metrics into existing host metrics
func (t *metricsTransformer) MergeHostMetrics(existing, new database.HostMetrics) database.HostMetrics {
	result := existing

	// Helper function to add new datapoints to existing metrics
	mergeMetricDatapoints := func(existing []database.MetricDatapoints, new []database.MetricDatapoints) []database.MetricDatapoints {
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
