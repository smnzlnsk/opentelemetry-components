package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestExtractHost(t *testing.T) {
	logger := zap.NewNop()
	transformer := NewMetricsTransformer(logger)

	t.Run("empty metrics", func(t *testing.T) {
		md := pmetric.NewMetrics()
		host := transformer.ExtractHost(md)
		assert.Equal(t, "unknown", host)
	})

	t.Run("extract from machine attribute", func(t *testing.T) {
		md := pmetric.NewMetrics()
		rm := md.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("machine", "test-host-1")
		host := transformer.ExtractHost(md)
		assert.Equal(t, "test-host-1", host)
	})

	t.Run("extract from host.name attribute", func(t *testing.T) {
		md := pmetric.NewMetrics()
		rm := md.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "test-host-2")
		host := transformer.ExtractHost(md)
		assert.Equal(t, "test-host-2", host)
	})

	t.Run("machine attribute takes precedence", func(t *testing.T) {
		md := pmetric.NewMetrics()
		rm := md.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("machine", "machine-host")
		rm.Resource().Attributes().PutStr("host.name", "hostname-host")
		host := transformer.ExtractHost(md)
		assert.Equal(t, "machine-host", host)
	})
}

func TestPMetricToMap(t *testing.T) {
	logger := zap.NewNop()
	transformer := NewMetricsTransformer(logger)

	t.Run("simple conversion", func(t *testing.T) {
		md := pmetric.NewMetrics()
		rm := md.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "testhost")

		resultMap, err := transformer.PMetricToMap(md)
		assert.NoError(t, err)
		assert.Len(t, resultMap, 1)
		assert.Contains(t, resultMap, "testhost")
		assert.Equal(t, md, resultMap["testhost"])
	})
}

func TestTransformToDBHostMetrics(t *testing.T) {
	logger := zap.NewNop()
	transformer := NewMetricsTransformer(logger)

	t.Run("empty metrics", func(t *testing.T) {
		md := pmetric.NewMetrics()
		hostMetrics, err := transformer.TransformToDBHostMetrics(md)
		assert.NoError(t, err)
		assert.Equal(t, "unknown", hostMetrics.Host)
		assert.Empty(t, hostMetrics.SystemMetrics)
		assert.Empty(t, hostMetrics.ServiceInstanceMetrics)
	})

	t.Run("system gauge metric", func(t *testing.T) {
		md := pmetric.NewMetrics()
		rm := md.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "test-host")
		sm := rm.ScopeMetrics().AppendEmpty()
		metric := sm.Metrics().AppendEmpty()
		metric.SetName("system.cpu.usage")
		dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetDoubleValue(0.75)
		dp.Attributes().PutStr("state", "idle")

		hostMetrics, err := transformer.TransformToDBHostMetrics(md)
		assert.NoError(t, err)
		assert.Equal(t, "test-host", hostMetrics.Host)
		assert.Len(t, hostMetrics.SystemMetrics, 1)
		assert.Equal(t, "system.cpu.usage", hostMetrics.SystemMetrics[0].Identifier.Name)
		assert.Equal(t, "idle", hostMetrics.SystemMetrics[0].Identifier.State)
		assert.Len(t, hostMetrics.SystemMetrics[0].Datapoints, 1)
		assert.Equal(t, 0.75, hostMetrics.SystemMetrics[0].Datapoints[0].Value)
	})

	t.Run("container metrics", func(t *testing.T) {
		md := pmetric.NewMetrics()
		rm := md.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "test-host")
		rm.Resource().Attributes().PutStr("container_id", "service1.job1.1")
		sm := rm.ScopeMetrics().AppendEmpty()
		metric := sm.Metrics().AppendEmpty()
		metric.SetName("container.memory.usage")
		dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetDoubleValue(256.0)
		dp.Attributes().PutStr("state", "used")

		hostMetrics, err := transformer.TransformToDBHostMetrics(md)
		assert.NoError(t, err)
		assert.Equal(t, "test-host", hostMetrics.Host)
		assert.Empty(t, hostMetrics.SystemMetrics)
		assert.Len(t, hostMetrics.ServiceInstanceMetrics, 1)
		assert.Equal(t, "service1", hostMetrics.ServiceInstanceMetrics[0].JobName)
		assert.Equal(t, 1, hostMetrics.ServiceInstanceMetrics[0].InstanceNumber)
		assert.Len(t, hostMetrics.ServiceInstanceMetrics[0].Metrics, 1)
		assert.Equal(t, "container.memory.usage", hostMetrics.ServiceInstanceMetrics[0].Metrics[0].Identifier.Name)
		assert.Equal(t, "used", hostMetrics.ServiceInstanceMetrics[0].Metrics[0].Identifier.State)
		assert.Len(t, hostMetrics.ServiceInstanceMetrics[0].Metrics[0].Datapoints, 1)
		assert.Equal(t, 256.0, hostMetrics.ServiceInstanceMetrics[0].Metrics[0].Datapoints[0].Value)
	})
}

func TestMergeHostMetrics(t *testing.T) {
	logger := zap.NewNop()
	transformer := NewMetricsTransformer(logger)

	t.Run("merge system metrics", func(t *testing.T) {
		// Create two sets of metrics to merge
		existing := DBHostMetrics{
			Host: "test-host",
			SystemMetrics: []DBMetricDatapoints{
				{
					Identifier: MetricKey{Name: "system.cpu", State: "idle", Type: MetricValueTypeRaw},
					Datapoints: []DBMetricDatapoint{
						{Value: 0.8, Timestamp: time.Now().Add(-time.Minute)},
					},
				},
			},
		}

		new := DBHostMetrics{
			Host: "test-host",
			SystemMetrics: []DBMetricDatapoints{
				{
					Identifier: MetricKey{Name: "system.cpu", State: "idle", Type: MetricValueTypeRaw},
					Datapoints: []DBMetricDatapoint{
						{Value: 0.9, Timestamp: time.Now()},
					},
				},
				{
					Identifier: MetricKey{Name: "system.memory", State: "used", Type: MetricValueTypeRaw},
					Datapoints: []DBMetricDatapoint{
						{Value: 1024.0, Timestamp: time.Now()},
					},
				},
			},
		}

		merged := transformer.MergeHostMetrics(existing, new)
		assert.Equal(t, "test-host", merged.Host)
		assert.Len(t, merged.SystemMetrics, 2)

		// Check that cpu metric has two datapoints
		var cpuMetrics *DBMetricDatapoints
		var memoryMetrics *DBMetricDatapoints
		for i := range merged.SystemMetrics {
			if merged.SystemMetrics[i].Identifier.Name == "system.cpu" {
				cpuMetrics = &merged.SystemMetrics[i]
			} else if merged.SystemMetrics[i].Identifier.Name == "system.memory" {
				memoryMetrics = &merged.SystemMetrics[i]
			}
		}

		assert.NotNil(t, cpuMetrics)
		assert.NotNil(t, memoryMetrics)
		assert.Len(t, cpuMetrics.Datapoints, 2)
		assert.Len(t, memoryMetrics.Datapoints, 1)
		assert.Equal(t, 0.8, cpuMetrics.Datapoints[0].Value)
		assert.Equal(t, 0.9, cpuMetrics.Datapoints[1].Value)
		assert.Equal(t, 1024.0, memoryMetrics.Datapoints[0].Value)
	})

	t.Run("merge service instance metrics", func(t *testing.T) {
		// Create test data with service instances
		existing := DBHostMetrics{
			Host: "test-host",
			ServiceInstanceMetrics: []DBServiceInstanceMetrics{
				{
					JobName:        "job1",
					InstanceNumber: 1,
					Metrics: []DBMetricDatapoints{
						{
							Identifier: MetricKey{Name: "container.cpu", State: "user", Type: MetricValueTypeRaw},
							Datapoints: []DBMetricDatapoint{
								{Value: 0.5, Timestamp: time.Now().Add(-time.Minute)},
							},
						},
					},
				},
			},
		}

		new := DBHostMetrics{
			Host: "test-host",
			ServiceInstanceMetrics: []DBServiceInstanceMetrics{
				{
					JobName:        "job1",
					InstanceNumber: 1,
					Metrics: []DBMetricDatapoints{
						{
							Identifier: MetricKey{Name: "container.cpu", State: "user", Type: MetricValueTypeRaw},
							Datapoints: []DBMetricDatapoint{
								{Value: 0.6, Timestamp: time.Now()},
							},
						},
					},
				},
				{
					JobName:        "job2",
					InstanceNumber: 1,
					Metrics: []DBMetricDatapoints{
						{
							Identifier: MetricKey{Name: "container.memory", State: "used", Type: MetricValueTypeRaw},
							Datapoints: []DBMetricDatapoint{
								{Value: 512.0, Timestamp: time.Now()},
							},
						},
					},
				},
			},
		}

		merged := transformer.MergeHostMetrics(existing, new)
		assert.Equal(t, "test-host", merged.Host)
		assert.Len(t, merged.ServiceInstanceMetrics, 2)

		// Find the service instances
		var job1Instance *DBServiceInstanceMetrics
		var job2Instance *DBServiceInstanceMetrics
		for i := range merged.ServiceInstanceMetrics {
			if merged.ServiceInstanceMetrics[i].JobName == "job1" {
				job1Instance = &merged.ServiceInstanceMetrics[i]
			} else if merged.ServiceInstanceMetrics[i].JobName == "job2" {
				job2Instance = &merged.ServiceInstanceMetrics[i]
			}
		}

		assert.NotNil(t, job1Instance)
		assert.NotNil(t, job2Instance)

		// Check the metrics for job1
		assert.Len(t, job1Instance.Metrics, 1)
		assert.Equal(t, "container.cpu", job1Instance.Metrics[0].Identifier.Name)
		assert.Len(t, job1Instance.Metrics[0].Datapoints, 2)
		assert.Equal(t, 0.5, job1Instance.Metrics[0].Datapoints[0].Value)
		assert.Equal(t, 0.6, job1Instance.Metrics[0].Datapoints[1].Value)

		// Check the metrics for job2
		assert.Len(t, job2Instance.Metrics, 1)
		assert.Equal(t, "container.memory", job2Instance.Metrics[0].Identifier.Name)
		assert.Len(t, job2Instance.Metrics[0].Datapoints, 1)
		assert.Equal(t, 512.0, job2Instance.Metrics[0].Datapoints[0].Value)
	})
}

func TestTransformDBHostMetricsToMap(t *testing.T) {
	logger := zap.NewNop()
	transformer := NewMetricsTransformer(logger)

	t.Run("empty metrics", func(t *testing.T) {
		hostMetrics := DBHostMetrics{
			Host: "test-host",
		}

		mapMetrics, err := transformer.TransformDBHostMetricsToMap(hostMetrics)
		assert.NoError(t, err)
		assert.Contains(t, mapMetrics, "test-host")
		assert.Empty(t, mapMetrics["test-host"].HostMetrics)
		assert.Empty(t, mapMetrics["test-host"].ServiceInstanceMetrics)
	})

	t.Run("complex transformation", func(t *testing.T) {
		now := time.Now()
		hostMetrics := DBHostMetrics{
			Host: "test-host",
			SystemMetrics: []DBMetricDatapoints{
				{
					Identifier: MetricKey{Name: "system.cpu", State: "idle", Type: MetricValueTypeRaw},
					Datapoints: []DBMetricDatapoint{
						{Value: 0.8, Timestamp: now},
						{Value: 0.9, Timestamp: now.Add(time.Second)},
					},
				},
				{
					Identifier: MetricKey{Name: "system.memory", State: "used", Type: MetricValueTypeRaw},
					Datapoints: []DBMetricDatapoint{
						{Value: 1024.0, Timestamp: now},
					},
				},
			},
			ServiceInstanceMetrics: []DBServiceInstanceMetrics{
				{
					JobName:        "job1",
					InstanceNumber: 1,
					Metrics: []DBMetricDatapoints{
						{
							Identifier: MetricKey{Name: "container.cpu", State: "user", Type: MetricValueTypeRaw},
							Datapoints: []DBMetricDatapoint{
								{Value: 0.5, Timestamp: now},
								{Value: 0.6, Timestamp: now.Add(time.Second)},
							},
						},
					},
				},
				{
					JobName:        "job2",
					InstanceNumber: 1,
					Metrics: []DBMetricDatapoints{
						{
							Identifier: MetricKey{Name: "container.memory", State: "used", Type: MetricValueTypeRaw},
							Datapoints: []DBMetricDatapoint{
								{Value: 512.0, Timestamp: now},
							},
						},
					},
				},
			},
		}

		mapMetrics, err := transformer.TransformDBHostMetricsToMap(hostMetrics)
		assert.NoError(t, err)

		// Check overall structure
		assert.Contains(t, mapMetrics, "test-host")
		assert.Len(t, mapMetrics["test-host"].HostMetrics, 3)            // system.cpu|idle|0, system.cpu|idle|1, system.memory|used|0
		assert.Len(t, mapMetrics["test-host"].ServiceInstanceMetrics, 2) // job1.instance.1, job2.instance.1

		// Check system metrics
		assert.Equal(t, 0.8, mapMetrics["test-host"].HostMetrics["system.cpu|idle|0"])
		assert.Equal(t, 0.9, mapMetrics["test-host"].HostMetrics["system.cpu|idle|1"])
		assert.Equal(t, 1024.0, mapMetrics["test-host"].HostMetrics["system.memory|used|0"])

		// Check service metrics
		job1ServiceID := "job1.instance.1"
		job2ServiceID := "job2.instance.1"

		assert.Contains(t, mapMetrics["test-host"].ServiceInstanceMetrics, job1ServiceID)
		assert.Contains(t, mapMetrics["test-host"].ServiceInstanceMetrics, job2ServiceID)

		assert.Len(t, mapMetrics["test-host"].ServiceInstanceMetrics[job1ServiceID], 2)
		assert.Equal(t, 0.5, mapMetrics["test-host"].ServiceInstanceMetrics[job1ServiceID]["container.cpu|user|0"])
		assert.Equal(t, 0.6, mapMetrics["test-host"].ServiceInstanceMetrics[job1ServiceID]["container.cpu|user|1"])

		assert.Len(t, mapMetrics["test-host"].ServiceInstanceMetrics[job2ServiceID], 1)
		assert.Equal(t, 512.0, mapMetrics["test-host"].ServiceInstanceMetrics[job2ServiceID]["container.memory|used|0"])
	})
}

// Complex integration test covering multiple transform steps
func TestMetricsTransformerIntegration(t *testing.T) {
	logger := zap.NewNop()
	transformer := NewMetricsTransformer(logger)

	// Create test metrics
	md := pmetric.NewMetrics()

	// Add resource for system metrics
	systemRM := md.ResourceMetrics().AppendEmpty()
	systemRM.Resource().Attributes().PutStr("host.name", "integration-host")
	systemSM := systemRM.ScopeMetrics().AppendEmpty()

	// Add CPU metric
	cpuMetric := systemSM.Metrics().AppendEmpty()
	cpuMetric.SetName("system.cpu.usage")
	cpuDP := cpuMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	cpuDP.SetDoubleValue(0.75)
	cpuDP.Attributes().PutStr("state", "idle")

	// Add memory metric
	memMetric := systemSM.Metrics().AppendEmpty()
	memMetric.SetName("system.memory.usage")
	memDP := memMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	memDP.SetDoubleValue(2048.0)
	memDP.Attributes().PutStr("state", "used")

	// Add resource for service metrics
	serviceRM := md.ResourceMetrics().AppendEmpty()
	serviceRM.Resource().Attributes().PutStr("host.name", "integration-host")
	serviceRM.Resource().Attributes().PutStr("container_id", "application1.backend.1")
	serviceSM := serviceRM.ScopeMetrics().AppendEmpty()

	// Add container memory metric
	containerMemMetric := serviceSM.Metrics().AppendEmpty()
	containerMemMetric.SetName("container.memory.usage")
	containerMemDP := containerMemMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	containerMemDP.SetDoubleValue(512.0)
	containerMemDP.Attributes().PutStr("state", "used")

	// Step 1: Test the PMetricToMap
	metricsMap, err := transformer.PMetricToMap(md)
	assert.NoError(t, err)
	assert.Len(t, metricsMap, 1)
	assert.Contains(t, metricsMap, "integration-host")

	// Step 2: Transform to DBHostMetrics
	dbHostMetrics, err := transformer.TransformToDBHostMetrics(md)
	assert.NoError(t, err)
	assert.Equal(t, "integration-host", dbHostMetrics.Host)
	assert.Len(t, dbHostMetrics.SystemMetrics, 2)          // CPU and memory
	assert.Len(t, dbHostMetrics.ServiceInstanceMetrics, 1) // application1.backend.1

	// Step 3: Create a second batch of metrics with new values
	md2 := pmetric.NewMetrics()

	// System metrics
	systemRM2 := md2.ResourceMetrics().AppendEmpty()
	systemRM2.Resource().Attributes().PutStr("host.name", "integration-host")
	systemSM2 := systemRM2.ScopeMetrics().AppendEmpty()

	// Update CPU metric
	cpuMetric2 := systemSM2.Metrics().AppendEmpty()
	cpuMetric2.SetName("system.cpu.usage")
	cpuDP2 := cpuMetric2.SetEmptyGauge().DataPoints().AppendEmpty()
	cpuDP2.SetDoubleValue(0.85)
	cpuDP2.Attributes().PutStr("state", "idle")

	// Add disk metric (new)
	diskMetric := systemSM2.Metrics().AppendEmpty()
	diskMetric.SetName("system.disk.usage")
	diskDP := diskMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	diskDP.SetDoubleValue(75.0)
	diskDP.Attributes().PutStr("state", "percent")

	// Service metrics
	serviceRM2 := md2.ResourceMetrics().AppendEmpty()
	serviceRM2.Resource().Attributes().PutStr("host.name", "integration-host")
	serviceRM2.Resource().Attributes().PutStr("container_id", "application1.backend.1")
	serviceSM2 := serviceRM2.ScopeMetrics().AppendEmpty()

	// Update container memory
	containerMemMetric2 := serviceSM2.Metrics().AppendEmpty()
	containerMemMetric2.SetName("container.memory.usage")
	containerMemDP2 := containerMemMetric2.SetEmptyGauge().DataPoints().AppendEmpty()
	containerMemDP2.SetDoubleValue(600.0)
	containerMemDP2.Attributes().PutStr("state", "used")

	// Add new service
	serviceRM3 := md2.ResourceMetrics().AppendEmpty()
	serviceRM3.Resource().Attributes().PutStr("host.name", "integration-host")
	serviceRM3.Resource().Attributes().PutStr("container_id", "application2.frontend.1")
	serviceSM3 := serviceRM3.ScopeMetrics().AppendEmpty()

	// Add frontend container CPU
	frontendCPUMetric := serviceSM3.Metrics().AppendEmpty()
	frontendCPUMetric.SetName("container.cpu.usage")
	frontendCPUDP := frontendCPUMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	frontendCPUDP.SetDoubleValue(0.3)
	frontendCPUDP.Attributes().PutStr("state", "user")

	// Transform the second batch
	dbHostMetrics2, err := transformer.TransformToDBHostMetrics(md2)
	assert.NoError(t, err)

	// Step 4: Merge the metrics
	mergedMetrics := transformer.MergeHostMetrics(dbHostMetrics, dbHostMetrics2)

	// Verify the merge
	assert.Equal(t, "integration-host", mergedMetrics.Host)
	assert.Len(t, mergedMetrics.SystemMetrics, 3)          // CPU, memory, and disk
	assert.Len(t, mergedMetrics.ServiceInstanceMetrics, 2) // backend and frontend

	// Step 5: Transform to map
	mapMetrics, err := transformer.TransformDBHostMetricsToMap(mergedMetrics)
	assert.NoError(t, err)

	// Verify final map structure
	assert.Contains(t, mapMetrics, "integration-host")

	// Check system metrics
	systemMetrics := mapMetrics["integration-host"].HostMetrics
	assert.Contains(t, systemMetrics, "system.cpu.usage|idle|0")
	assert.Contains(t, systemMetrics, "system.cpu.usage|idle|1")
	assert.Contains(t, systemMetrics, "system.memory.usage|used|0")
	assert.Contains(t, systemMetrics, "system.disk.usage|percent|0")

	// Check service metrics
	serviceMetrics := mapMetrics["integration-host"].ServiceInstanceMetrics
	assert.Contains(t, serviceMetrics, "application1.instance.1")
	assert.Contains(t, serviceMetrics, "application2.instance.1")

	backendMetrics := serviceMetrics["application1.instance.1"]
	assert.Contains(t, backendMetrics, "container.memory.usage|used|0")
	assert.Contains(t, backendMetrics, "container.memory.usage|used|1")

	frontendMetrics := serviceMetrics["application2.instance.1"]
	assert.Contains(t, frontendMetrics, "container.cpu.usage|user|0")
}
