package unixsocketreceiver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// PrometheusParser parses Prometheus exposition format data
type PrometheusParser struct {
	logger *zap.Logger
}

// NewPrometheusParser creates a new parser for Prometheus data
func NewPrometheusParser(logger *zap.Logger) *PrometheusParser {
	return &PrometheusParser{
		logger: logger,
	}
}

// ParsePrometheusMetrics parses Prometheus metrics from a byte array and returns pmetric.Metrics
func (p *PrometheusParser) ParsePrometheusMetrics(socketPath string, data []byte) (pmetric.Metrics, error) {
	if len(data) == 0 {
		return pmetric.NewMetrics(), fmt.Errorf("empty data received")
	}

	// Ensure the data ends with a newline
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}

	// Create a new reader with the data
	reader := bytes.NewReader(data)

	// Create a decoder for Prometheus text format
	decoder := expfmt.NewDecoder(reader, expfmt.NewFormat(expfmt.TypeTextPlain))

	metrics := pmetric.NewMetrics()

	// Create a resource metrics builder
	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "prometheus")
	rm.Resource().Attributes().PutStr("container_id", strings.Split(socketPath, "/")[len(strings.Split(socketPath, "/"))-1])

	// Add a scope metrics
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("prometheus")

	// Decode all metrics
	metricFamilies := make(map[string]*dto.MetricFamily)
	lineNumber := 1

	for {
		var mf dto.MetricFamily
		err := decoder.Decode(&mf)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Try to extract the line number from the error
			p.logger.Error("Error decoding Prometheus metric",
				zap.Error(err),
				zap.Int("line_number", lineNumber),
				zap.String("data_preview", string(data[:min(len(data), 200)])))
			return pmetric.NewMetrics(), fmt.Errorf("failed to decode Prometheus metric at line %d: %w", lineNumber, err)
		}

		metricFamilies[mf.GetName()] = &mf
		lineNumber++
	}

	if len(metricFamilies) == 0 {
		p.logger.Warn("No metric families found in the data",
			zap.String("data_preview", string(data[:min(len(data), 200)])))
	}

	// Convert metric families to pmetric.Metrics
	for name, mf := range metricFamilies {
		p.convertMetricFamily(name, mf, sm.Metrics())
	}

	return metrics, nil
}

// convertMetricFamily converts a Prometheus metric family to pmetric format
func (p *PrometheusParser) convertMetricFamily(name string, family *dto.MetricFamily, metrics pmetric.MetricSlice) {
	metricType := family.GetType()

	switch metricType {
	case dto.MetricType_COUNTER:
		p.convertCounter(name, family, metrics)
	case dto.MetricType_GAUGE:
		p.convertGauge(name, family, metrics)
	case dto.MetricType_HISTOGRAM:
		p.convertHistogram(name, family, metrics)
	case dto.MetricType_SUMMARY:
		p.convertSummary(name, family, metrics)
	case dto.MetricType_UNTYPED:
		p.convertUntyped(name, family, metrics)
	default:
		p.logger.Info("Unsupported metric type", zap.String("type", metricType.String()))
	}
}

// convertCounter converts Prometheus counter metrics to OpenTelemetry format
func (p *PrometheusParser) convertCounter(name string, family *dto.MetricFamily, metrics pmetric.MetricSlice) {
	metric := metrics.AppendEmpty()
	metric.SetName(name)

	if family.Help != nil {
		metric.SetDescription(*family.Help)
	}

	sum := metric.SetEmptySum()
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	sum.SetIsMonotonic(true)

	for _, m := range family.Metric {
		dp := sum.DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

		if m.Counter != nil && m.Counter.Value != nil {
			dp.SetDoubleValue(*m.Counter.Value)
		}

		// Add labels as attributes
		p.addLabelsAsAttributes(m.Label, dp.Attributes())
	}
}

// convertGauge converts Prometheus gauge metrics to OpenTelemetry format
func (p *PrometheusParser) convertGauge(name string, family *dto.MetricFamily, metrics pmetric.MetricSlice) {
	metric := metrics.AppendEmpty()
	metric.SetName(name)

	if family.Help != nil {
		metric.SetDescription(*family.Help)
	}

	gauge := metric.SetEmptyGauge()

	for _, m := range family.Metric {
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

		if m.Gauge != nil && m.Gauge.Value != nil {
			dp.SetDoubleValue(*m.Gauge.Value)
		}

		// Add labels as attributes
		p.addLabelsAsAttributes(m.Label, dp.Attributes())
	}
}

// convertHistogram converts Prometheus histogram metrics to OpenTelemetry format
func (p *PrometheusParser) convertHistogram(name string, family *dto.MetricFamily, metrics pmetric.MetricSlice) {
	metric := metrics.AppendEmpty()
	metric.SetName(name)

	if family.Help != nil {
		metric.SetDescription(*family.Help)
	}

	histogram := metric.SetEmptyHistogram()
	histogram.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	for _, m := range family.Metric {
		if m.Histogram == nil {
			continue
		}

		dp := histogram.DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

		if m.Histogram.SampleCount != nil {
			dp.SetCount(*m.Histogram.SampleCount)
		}

		if m.Histogram.SampleSum != nil {
			dp.SetSum(*m.Histogram.SampleSum)
		}

		// Convert bucket bounds and counts
		if len(m.Histogram.Bucket) > 0 {
			bucketCounts := make([]uint64, len(m.Histogram.Bucket))
			explicitBounds := make([]float64, len(m.Histogram.Bucket)-1) // Last bucket is +Inf

			var cumulativeCount uint64
			for i, bucket := range m.Histogram.Bucket {
				if i < len(explicitBounds) && bucket.UpperBound != nil {
					explicitBounds[i] = *bucket.UpperBound
				}

				if bucket.CumulativeCount != nil {
					cumulativeCount = *bucket.CumulativeCount
					bucketCounts[i] = cumulativeCount
				}
			}

			// Convert cumulative counts to counts per bucket
			var prevCount uint64
			for i := range bucketCounts {
				currentCount := bucketCounts[i]
				bucketCounts[i] = currentCount - prevCount
				prevCount = currentCount
			}

			dp.BucketCounts().FromRaw(bucketCounts)
			dp.ExplicitBounds().FromRaw(explicitBounds)
		}

		// Add labels as attributes
		p.addLabelsAsAttributes(m.Label, dp.Attributes())
	}
}

// convertSummary converts Prometheus summary metrics to OpenTelemetry format
func (p *PrometheusParser) convertSummary(name string, family *dto.MetricFamily, metrics pmetric.MetricSlice) {
	metric := metrics.AppendEmpty()
	metric.SetName(name)

	if family.Help != nil {
		metric.SetDescription(*family.Help)
	}

	summary := metric.SetEmptySummary()

	for _, m := range family.Metric {
		if m.Summary == nil {
			continue
		}

		dp := summary.DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

		if m.Summary.SampleCount != nil {
			dp.SetCount(*m.Summary.SampleCount)
		}

		if m.Summary.SampleSum != nil {
			dp.SetSum(*m.Summary.SampleSum)
		}

		// Convert quantiles
		if len(m.Summary.Quantile) > 0 {
			for _, q := range m.Summary.Quantile {
				if q.Quantile != nil && q.Value != nil {
					qv := dp.QuantileValues().AppendEmpty()
					qv.SetQuantile(*q.Quantile)
					qv.SetValue(*q.Value)
				}
			}
		}

		// Add labels as attributes
		p.addLabelsAsAttributes(m.Label, dp.Attributes())
	}
}

// convertUntyped converts Prometheus untyped metrics to OpenTelemetry gauge format
func (p *PrometheusParser) convertUntyped(name string, family *dto.MetricFamily, metrics pmetric.MetricSlice) {
	metric := metrics.AppendEmpty()
	metric.SetName(name)

	if family.Help != nil {
		metric.SetDescription(*family.Help)
	}

	gauge := metric.SetEmptyGauge()

	for _, m := range family.Metric {
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

		if m.Untyped != nil && m.Untyped.Value != nil {
			dp.SetDoubleValue(*m.Untyped.Value)
		}

		// Add labels as attributes
		p.addLabelsAsAttributes(m.Label, dp.Attributes())
	}
}

// addLabelsAsAttributes adds Prometheus labels as OpenTelemetry attributes
func (p *PrometheusParser) addLabelsAsAttributes(labels []*dto.LabelPair, attributes pcommon.Map) {
	for _, label := range labels {
		if label.Name != nil && label.Value != nil {
			attributes.PutStr(*label.Name, *label.Value)
		}
	}
}

// IsPrometheusFormat checks if the data appears to be in Prometheus exposition format
func IsPrometheusFormat(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Ensure the data ends with a newline
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	linesChecked := 0
	validLines := 0

	for scanner.Scan() && linesChecked < 10 {
		line := strings.TrimSpace(scanner.Text())
		linesChecked++

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for Prometheus format HELP/TYPE comments
		if strings.HasPrefix(line, "# HELP ") || strings.HasPrefix(line, "# TYPE ") {
			validLines++
			continue
		}

		// Check for metric name followed by value or labels
		if !strings.HasPrefix(line, "#") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				metricPart := parts[0]
				// Check if it contains a metric name pattern (letters, numbers, underscores)
				// and either curly braces for labels or is followed by a numeric value
				if strings.ContainsAny(metricPart, "{") {
					if len(parts) >= 2 && isNumeric(parts[1]) {
						validLines++
					}
				} else if len(parts) >= 2 && isNumeric(parts[1]) {
					validLines++
				}
			}
		}
	}

	// Log the format check results
	if validLines > 0 {
		return true
	}

	return false
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
