package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/pkg/database"
	datapoint "github.com/smnzlnsk/opentelemetry-components/pkg/metric/datapoint"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// mockMetricsService implements the MetricsService interface for testing
type mockMetricsService struct {
	persistent bool
}

func (m *mockMetricsService) Persistent() bool {
	return m.persistent
}

func (m *mockMetricsService) SaveMetrics(ctx context.Context, dbHostMetrics database.HostMetrics) error {
	return nil
}

func (m *mockMetricsService) SaveMetricsBatch(ctx context.Context, hostMetricsList []database.HostMetrics) error {
	return nil
}

func (m *mockMetricsService) GetJobMetrics(ctx context.Context, jobName string) (database.HostMetrics, error) {
	return database.HostMetrics{}, nil
}

func (m *mockMetricsService) GetJobMetricsBatch(ctx context.Context, jobNames []string) (map[string]database.HostMetrics, error) {
	return make(map[string]database.HostMetrics), nil
}

func (m *mockMetricsService) EnsureIndexes(ctx context.Context) error {
	return nil
}

// newMockMetricsService creates a new mock metrics service
func newMockMetricsService(persistent bool) MetricsService {
	return &mockMetricsService{
		persistent: persistent,
	}
}

// generateTestMetrics creates a pmetric.Metrics object for testing purposes
// This function creates realistic metrics that match the structure expected by DatapointManager
// System metrics and container metrics are separated at the resource level
func generateTestMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	// 1. Create system-level resource metrics
	systemRM := metrics.ResourceMetrics().AppendEmpty()
	systemRM.Resource().Attributes().PutStr("host.name", "test-host-01")
	systemRM.Resource().Attributes().PutStr("machine", "worker-node-1")
	// No container-specific attributes for system metrics

	systemSM := systemRM.ScopeMetrics().AppendEmpty()
	systemSM.Scope().SetName("system-scope")
	systemSM.Scope().SetVersion("1.0.0")

	// System CPU utilization metric (Gauge)
	cpuMetric := systemSM.Metrics().AppendEmpty()
	cpuMetric.SetName("system.cpu.utilization")
	cpuMetric.SetDescription("System CPU utilization")
	cpuMetric.SetUnit("percent")

	cpuGauge := cpuMetric.SetEmptyGauge()

	// CPU with different states
	cpuUserDp := cpuGauge.DataPoints().AppendEmpty()
	cpuUserDp.SetTimestamp(timestamp)
	cpuUserDp.SetDoubleValue(45.5)
	cpuUserDp.Attributes().PutStr("state", "user")

	cpuSystemDp := cpuGauge.DataPoints().AppendEmpty()
	cpuSystemDp.SetTimestamp(timestamp)
	cpuSystemDp.SetDoubleValue(25.2)
	cpuSystemDp.Attributes().PutStr("state", "system")

	cpuIdleDp := cpuGauge.DataPoints().AppendEmpty()
	cpuIdleDp.SetTimestamp(timestamp)
	cpuIdleDp.SetDoubleValue(29.3)
	cpuIdleDp.Attributes().PutStr("state", "idle")

	// System memory utilization metric (Gauge)
	memMetric := systemSM.Metrics().AppendEmpty()
	memMetric.SetName("system.memory.utilization")
	memMetric.SetDescription("System memory utilization")
	memMetric.SetUnit("percent")

	memGauge := memMetric.SetEmptyGauge()

	memUsedDp := memGauge.DataPoints().AppendEmpty()
	memUsedDp.SetTimestamp(timestamp)
	memUsedDp.SetDoubleValue(78.4)
	memUsedDp.Attributes().PutStr("state", "used")

	memFreeDp := memGauge.DataPoints().AppendEmpty()
	memFreeDp.SetTimestamp(timestamp)
	memFreeDp.SetDoubleValue(21.6)
	memFreeDp.Attributes().PutStr("state", "free")

	// 2. Create first container resource metrics
	container1RM := metrics.ResourceMetrics().AppendEmpty()
	container1RM.Resource().Attributes().PutStr("host.name", "test-host-01")
	container1RM.Resource().Attributes().PutStr("machine", "worker-node-1")
	container1RM.Resource().Attributes().PutStr("service.name", "a.b.c.d.instance.1")
	container1RM.Resource().Attributes().PutStr("container_id", "a.b.c.d.instance.1")

	container1SM := container1RM.ScopeMetrics().AppendEmpty()
	container1SM.Scope().SetName("container-scope")
	container1SM.Scope().SetVersion("1.0.0")

	// Container CPU utilization metric (Sum)
	containerCpuMetric := container1SM.Metrics().AppendEmpty()
	containerCpuMetric.SetName("container.cpu.utilization")
	containerCpuMetric.SetDescription("Container CPU utilization")
	containerCpuMetric.SetUnit("percent")

	containerCpuSum := containerCpuMetric.SetEmptySum()
	containerCpuSum.SetIsMonotonic(false)
	containerCpuSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	containerCpuDp := containerCpuSum.DataPoints().AppendEmpty()
	containerCpuDp.SetTimestamp(timestamp)
	containerCpuDp.SetDoubleValue(15.7)
	containerCpuDp.Attributes().PutStr("state", "active")

	// Container memory utilization metric (Gauge)
	containerMemMetric := container1SM.Metrics().AppendEmpty()
	containerMemMetric.SetName("container.memory.utilization")
	containerMemMetric.SetDescription("Container memory utilization")
	containerMemMetric.SetUnit("percent")

	containerMemGauge := containerMemMetric.SetEmptyGauge()

	containerMemDp := containerMemGauge.DataPoints().AppendEmpty()
	containerMemDp.SetTimestamp(timestamp)
	containerMemDp.SetDoubleValue(65.3)
	containerMemDp.Attributes().PutStr("state", "used")

	// 3. Create second container resource metrics
	container2RM := metrics.ResourceMetrics().AppendEmpty()
	container2RM.Resource().Attributes().PutStr("host.name", "test-host-01")
	container2RM.Resource().Attributes().PutStr("machine", "worker-node-1")
	container2RM.Resource().Attributes().PutStr("service.name", "a.b.c.d.instance.2")
	container2RM.Resource().Attributes().PutStr("container_id", "a.b.c.d.instance.2")

	container2SM := container2RM.ScopeMetrics().AppendEmpty()
	container2SM.Scope().SetName("container-scope")
	container2SM.Scope().SetVersion("1.0.0")

	// Container CPU utilization metric (Sum)
	container2CpuMetric := container2SM.Metrics().AppendEmpty()
	container2CpuMetric.SetName("container.cpu.utilization")
	container2CpuMetric.SetDescription("Container CPU utilization")
	container2CpuMetric.SetUnit("percent")

	container2CpuSum := container2CpuMetric.SetEmptySum()
	container2CpuSum.SetIsMonotonic(false)
	container2CpuSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	container2CpuDp := container2CpuSum.DataPoints().AppendEmpty()
	container2CpuDp.SetTimestamp(timestamp)
	container2CpuDp.SetDoubleValue(22.1)
	container2CpuDp.Attributes().PutStr("state", "active")

	// Container memory utilization metric (Gauge)
	container2MemMetric := container2SM.Metrics().AppendEmpty()
	container2MemMetric.SetName("container.memory.utilization")
	container2MemMetric.SetDescription("Container memory utilization")
	container2MemMetric.SetUnit("percent")

	container2MemGauge := container2MemMetric.SetEmptyGauge()

	container2MemDp := container2MemGauge.DataPoints().AppendEmpty()
	container2MemDp.SetTimestamp(timestamp)
	container2MemDp.SetDoubleValue(42.8)
	container2MemDp.Attributes().PutStr("state", "used")

	// 3. Create third container resource metrics for second host
	container3RM := metrics.ResourceMetrics().AppendEmpty()
	container3RM.Resource().Attributes().PutStr("host.name", "test-host-01")
	container3RM.Resource().Attributes().PutStr("machine", "worker-node-1")
	container3RM.Resource().Attributes().PutStr("service.name", "w.x.y.z.instance.2")
	container3RM.Resource().Attributes().PutStr("container_id", "w.x.y.z.instance.2")

	container3SM := container3RM.ScopeMetrics().AppendEmpty()
	container3SM.Scope().SetName("container-scope")
	container3SM.Scope().SetVersion("1.0.0")

	// Container CPU utilization metric (Sum)
	container3CpuMetric := container3SM.Metrics().AppendEmpty()
	container3CpuMetric.SetName("container.cpu.utilization")
	container3CpuMetric.SetDescription("Container CPU utilization")
	container3CpuMetric.SetUnit("percent")

	container3CpuSum := container3CpuMetric.SetEmptySum()
	container3CpuSum.SetIsMonotonic(false)
	container3CpuSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	container3CpuDp := container3CpuSum.DataPoints().AppendEmpty()
	container3CpuDp.SetTimestamp(timestamp)
	container3CpuDp.SetDoubleValue(28.9)
	container3CpuDp.Attributes().PutStr("state", "active")

	// Container memory utilization metric (Gauge)
	container3MemMetric := container3SM.Metrics().AppendEmpty()
	container3MemMetric.SetName("container.memory.utilization")
	container3MemMetric.SetDescription("Container memory utilization")
	container3MemMetric.SetUnit("percent")

	container3MemGauge := container3MemMetric.SetEmptyGauge()

	container3MemDp := container3MemGauge.DataPoints().AppendEmpty()
	container3MemDp.SetTimestamp(timestamp)
	container3MemDp.SetDoubleValue(56.1)
	container3MemDp.Attributes().PutStr("state", "used")

	return metrics
}

// generateTestMetricsFromSecondHost creates a pmetric.Metrics object for testing from a different host
// This function creates metrics from a second host to test multi-host scenarios
func generateTestMetricsFromSecondHost() pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	// 1. Create system-level resource metrics for second host
	systemRM := metrics.ResourceMetrics().AppendEmpty()
	systemRM.Resource().Attributes().PutStr("host.name", "test-host-02")
	systemRM.Resource().Attributes().PutStr("machine", "worker-node-2")
	// No container-specific attributes for system metrics

	systemSM := systemRM.ScopeMetrics().AppendEmpty()
	systemSM.Scope().SetName("system-scope")
	systemSM.Scope().SetVersion("1.0.0")

	// System CPU utilization metric (Gauge)
	cpuMetric := systemSM.Metrics().AppendEmpty()
	cpuMetric.SetName("system.cpu.utilization")
	cpuMetric.SetDescription("System CPU utilization")
	cpuMetric.SetUnit("percent")

	cpuGauge := cpuMetric.SetEmptyGauge()

	// CPU with different states (different values from first host)
	cpuUserDp := cpuGauge.DataPoints().AppendEmpty()
	cpuUserDp.SetTimestamp(timestamp)
	cpuUserDp.SetDoubleValue(62.3)
	cpuUserDp.Attributes().PutStr("state", "user")

	cpuSystemDp := cpuGauge.DataPoints().AppendEmpty()
	cpuSystemDp.SetTimestamp(timestamp)
	cpuSystemDp.SetDoubleValue(18.7)
	cpuSystemDp.Attributes().PutStr("state", "system")

	cpuIdleDp := cpuGauge.DataPoints().AppendEmpty()
	cpuIdleDp.SetTimestamp(timestamp)
	cpuIdleDp.SetDoubleValue(19.0)
	cpuIdleDp.Attributes().PutStr("state", "idle")

	// System memory utilization metric (Gauge)
	memMetric := systemSM.Metrics().AppendEmpty()
	memMetric.SetName("system.memory.utilization")
	memMetric.SetDescription("System memory utilization")
	memMetric.SetUnit("percent")

	memGauge := memMetric.SetEmptyGauge()

	memUsedDp := memGauge.DataPoints().AppendEmpty()
	memUsedDp.SetTimestamp(timestamp)
	memUsedDp.SetDoubleValue(85.2)
	memUsedDp.Attributes().PutStr("state", "used")

	memFreeDp := memGauge.DataPoints().AppendEmpty()
	memFreeDp.SetTimestamp(timestamp)
	memFreeDp.SetDoubleValue(14.8)
	memFreeDp.Attributes().PutStr("state", "free")

	// 2. Create first container resource metrics for second host
	container1RM := metrics.ResourceMetrics().AppendEmpty()
	container1RM.Resource().Attributes().PutStr("host.name", "test-host-02")
	container1RM.Resource().Attributes().PutStr("machine", "worker-node-2")
	container1RM.Resource().Attributes().PutStr("service.name", "w.x.y.z.instance.1")
	container1RM.Resource().Attributes().PutStr("container_id", "w.x.y.z.instance.1")

	container1SM := container1RM.ScopeMetrics().AppendEmpty()
	container1SM.Scope().SetName("container-scope")
	container1SM.Scope().SetVersion("1.0.0")

	// Container CPU utilization metric (Sum)
	containerCpuMetric := container1SM.Metrics().AppendEmpty()
	containerCpuMetric.SetName("container.cpu.utilization")
	containerCpuMetric.SetDescription("Container CPU utilization")
	containerCpuMetric.SetUnit("percent")

	containerCpuSum := containerCpuMetric.SetEmptySum()
	containerCpuSum.SetIsMonotonic(false)
	containerCpuSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	containerCpuDp := containerCpuSum.DataPoints().AppendEmpty()
	containerCpuDp.SetTimestamp(timestamp)
	containerCpuDp.SetDoubleValue(31.4)
	containerCpuDp.Attributes().PutStr("state", "active")

	// Container memory utilization metric (Gauge)
	containerMemMetric := container1SM.Metrics().AppendEmpty()
	containerMemMetric.SetName("container.memory.utilization")
	containerMemMetric.SetDescription("Container memory utilization")
	containerMemMetric.SetUnit("percent")

	containerMemGauge := containerMemMetric.SetEmptyGauge()

	containerMemDp := containerMemGauge.DataPoints().AppendEmpty()
	containerMemDp.SetTimestamp(timestamp)
	containerMemDp.SetDoubleValue(72.6)
	containerMemDp.Attributes().PutStr("state", "used")

	// 4. Create third container resource metrics for second host (additional container)
	container3RM := metrics.ResourceMetrics().AppendEmpty()
	container3RM.Resource().Attributes().PutStr("host.name", "test-host-02")
	container3RM.Resource().Attributes().PutStr("machine", "worker-node-2")
	container3RM.Resource().Attributes().PutStr("service.name", "w.x.y.z.instance.3")
	container3RM.Resource().Attributes().PutStr("container_id", "w.x.y.z.instance.3")

	container3SM := container3RM.ScopeMetrics().AppendEmpty()
	container3SM.Scope().SetName("container-scope")
	container3SM.Scope().SetVersion("1.0.0")

	// Container CPU utilization metric (Sum)
	container3CpuMetric := container3SM.Metrics().AppendEmpty()
	container3CpuMetric.SetName("container.cpu.utilization")
	container3CpuMetric.SetDescription("Container CPU utilization")
	container3CpuMetric.SetUnit("percent")

	container3CpuSum := container3CpuMetric.SetEmptySum()
	container3CpuSum.SetIsMonotonic(false)
	container3CpuSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	container3CpuDp := container3CpuSum.DataPoints().AppendEmpty()
	container3CpuDp.SetTimestamp(timestamp)
	container3CpuDp.SetDoubleValue(12.5)
	container3CpuDp.Attributes().PutStr("state", "active")

	// Container memory utilization metric (Gauge)
	container3MemMetric := container3SM.Metrics().AppendEmpty()
	container3MemMetric.SetName("container.memory.utilization")
	container3MemMetric.SetDescription("Container memory utilization")
	container3MemMetric.SetUnit("percent")

	container3MemGauge := container3MemMetric.SetEmptyGauge()

	container3MemDp := container3MemGauge.DataPoints().AppendEmpty()
	container3MemDp.SetTimestamp(timestamp)
	container3MemDp.SetDoubleValue(38.7)
	container3MemDp.Attributes().PutStr("state", "used")

	return metrics
}

// setupDatapointManager creates a DatapointManager for testing with proper dependencies
func setupDatapointManager(persistent bool) DatapointManager {
	logger := zap.NewNop() // Use no-op logger for tests to avoid noise
	mockService := newMockMetricsService(persistent)
	return NewDatapointManager(logger, mockService)
}

func TestDatapointManager(t *testing.T) {
	// Example test using the generated metrics
	metrics := generateTestMetrics()

	t.Run("test-metric-creation", func(t *testing.T) {
		// Verify the metrics were created correctly
		if metrics.ResourceMetrics().Len() == 0 {
			t.Error("Expected metrics to have resource metrics")
		}

		// Verify we have 3 resource metrics (1 system + 2 containers)
		if metrics.ResourceMetrics().Len() != 3 {
			t.Errorf("Expected 3 resource metrics, got %d", metrics.ResourceMetrics().Len())
		}

		// Check system resource metrics (first one)
		systemRM := metrics.ResourceMetrics().At(0)
		if hostName, exists := systemRM.Resource().Attributes().Get("host.name"); !exists || hostName.Str() != "test-host-01" {
			t.Error("Expected host.name to be 'test-host-01'")
		}

		// System resource should not have container_id
		if _, exists := systemRM.Resource().Attributes().Get("container_id"); exists {
			t.Error("Expected system resource to not have container_id")
		}

		// Verify system scope metrics
		if systemRM.ScopeMetrics().Len() == 0 {
			t.Error("Expected system scope metrics")
		}

		systemSM := systemRM.ScopeMetrics().At(0)
		if systemSM.Metrics().Len() == 0 {
			t.Error("Expected system metrics to be present")
		}

		// Check the first system metric
		firstSystemMetric := systemSM.Metrics().At(0)
		if firstSystemMetric.Name() != "system.cpu.utilization" {
			t.Errorf("Expected first system metric name to be 'system.cpu.utilization', got '%s'", firstSystemMetric.Name())
		}

		// Check container resource metrics (second one)
		containerRM := metrics.ResourceMetrics().At(1)
		if containerID, exists := containerRM.Resource().Attributes().Get("container_id"); !exists || containerID.Str() != "a.b.c.d.instance.1" {
			t.Error("Expected container_id to be 'a.b.c.d.instance.1'")
		}

		containerSM := containerRM.ScopeMetrics().At(0)
		if containerSM.Metrics().Len() == 0 {
			t.Error("Expected container metrics to be present")
		}

		// Check the first container metric
		firstContainerMetric := containerSM.Metrics().At(0)
		if firstContainerMetric.Name() != "container.cpu.utilization" {
			t.Errorf("Expected first container metric name to be 'container.cpu.utilization', got '%s'", firstContainerMetric.Name())
		}
	})

	t.Run("test-datapoint-manager-initialization", func(t *testing.T) {
		// Test with persistent metrics enabled
		dm := setupDatapointManager(true)
		defer dm.Shutdown()

		if dm == nil {
			t.Error("Expected DatapointManager to be initialized")
		}

		// Test host extraction (machine attribute takes priority over host.name)
		host := dm.GetHost(metrics)
		if host != "worker-node-1" {
			t.Errorf("Expected host to be 'worker-node-1', got '%s'", host)
		}
	})

	t.Run("test-datapoint-manager-save-metrics", func(t *testing.T) {
		dm := setupDatapointManager(false) // Non-persistent for faster testing
		defer dm.Shutdown()

		// Test saving metrics
		err := dm.SaveMetrics(metrics)
		if err != nil {
			t.Errorf("Expected no error saving metrics, got: %v", err)
		}

		// Verify datapoints were stored
		datapoints := dm.GetDatapoints()
		if len(datapoints) == 0 {
			t.Error("Expected datapoints to be stored after saving metrics")
		}
	})

	t.Run("test-multi-host-metrics", func(t *testing.T) {
		dm := setupDatapointManager(false)
		defer dm.Shutdown()

		// Generate metrics from second host
		metricsFromSecondHost := generateTestMetricsFromSecondHost()

		// Verify second host metrics structure
		if metricsFromSecondHost.ResourceMetrics().Len() != 4 {
			t.Errorf("Expected 4 resource metrics from second host, got %d", metricsFromSecondHost.ResourceMetrics().Len())
		}

		// Check host identification
		systemRM := metricsFromSecondHost.ResourceMetrics().At(0)
		if hostName, exists := systemRM.Resource().Attributes().Get("host.name"); !exists || hostName.Str() != "test-host-02" {
			t.Error("Expected host.name to be 'test-host-02'")
		}

		if machine, exists := systemRM.Resource().Attributes().Get("machine"); !exists || machine.Str() != "worker-node-2" {
			t.Error("Expected machine to be 'worker-node-2'")
		}

		// Test host extraction for second host
		host := dm.GetHost(metricsFromSecondHost)
		if host != "worker-node-2" {
			t.Errorf("Expected host to be 'worker-node-2', got '%s'", host)
		}

		// Test saving metrics from both hosts
		err := dm.SaveMetrics(metrics) // First host
		if err != nil {
			t.Errorf("Expected no error saving metrics from first host, got: %v", err)
		}

		err = dm.SaveMetrics(metricsFromSecondHost) // Second host
		if err != nil {
			t.Errorf("Expected no error saving metrics from second host, got: %v", err)
		}

		// Verify datapoints from both hosts were stored
		datapoints := dm.GetDatapoints()
		if len(datapoints) == 0 {
			t.Error("Expected datapoints to be stored after saving metrics from both hosts")
		}
	})
}

func TestDatapointManager_SaveMetrics(t *testing.T) {
	metrics := generateTestMetrics()
	metrics2 := generateTestMetricsFromSecondHost()
	dm := setupDatapointManager(false)
	defer dm.Shutdown()

	err := dm.SaveMetrics(metrics)
	if err != nil {
		t.Errorf("Expected no error saving metrics, got: %v", err)
	}
	err = dm.SaveMetrics(metrics2)
	if err != nil {
		t.Errorf("Expected no error saving metrics, got: %v", err)
	}

	datapoints := dm.GetDatapoints()
	if len(datapoints) == 0 {
		t.Error("Expected datapoints to be stored after saving metrics from both hosts")
	}

	t.Run("test-datapoint-manager-get-datapoints", func(t *testing.T) {
		datapoints := dm.GetDatapoints()
		if len(datapoints) == 0 {
			t.Errorf("Expected datapoints to be stored after saving metrics from both hosts")
		}

		// Check if the datapoints are correct
		if _, exists := datapoints[datapoint.Key{Service: "a.b.c.d.instance.1", Metric: "container.cpu.utilization", State: "active"}]; !exists {
			t.Error("Expected datapoints to be stored for a.b.c.d.instance.1")
		}
		if _, exists := datapoints[datapoint.Key{Service: "a.b.c.d.instance.1", Metric: "container.memory.utilization", State: "used"}]; !exists {
			t.Error("Expected datapoints to be stored for a.b.c.d.instance.1")
		}
		if _, exists := datapoints[datapoint.Key{Service: "w.x.y.z.instance.1", Metric: "container.cpu.utilization", State: "active"}]; !exists {
			t.Error("Expected datapoints to be stored for w.x.y.z.instance.1")
		}
		if _, exists := datapoints[datapoint.Key{Service: "w.x.y.z.instance.1", Metric: "container.memory.utilization", State: "used"}]; !exists {
			t.Error("Expected datapoints to be stored for w.x.y.z.instance.1")
		}
	})

	t.Run("test-datapoint-manager-get-datapoints-from-database", func(t *testing.T) {
		datapoints := dm.GetDatapoints()
		if len(datapoints) == 0 {
			t.Errorf("Expected datapoints to be stored after saving metrics from both hosts")
		}

		err := dm.SaveMetrics(metrics)
		if err != nil {
			t.Errorf("Expected no error saving metrics, got: %v", err)
		}

		metrics, exists := database.GetGlobalMemoryDatabase().GetMetrics("w.x.y.z")
		if !exists {
			t.Errorf("Expected metrics to be stored after saving metrics from both hosts")
		}
		if len(metrics) == 0 {
			t.Errorf("Expected metrics to be stored after saving metrics from both hosts")
		}

		fmt.Println(dm)
	})
}
