package oakestraheuristicengine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processortest"
	"go.uber.org/zap"
)

func TestHeuristicEngineProcessorWithAlertNotification(t *testing.T) {
	// Setup test notification server for alerts
	var alertWg sync.WaitGroup
	var receivedAlertNotification []byte

	alertServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Alert notification received: %s %s", r.Method, r.URL.Path)

		// Read and store the request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		t.Logf("Alert notification body: %s", string(body))
		receivedAlertNotification = body

		// Send a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))

		alertWg.Done()
	}))
	defer alertServer.Close()

	// Extract host and port from test server
	alertHost, alertPort := extractHostPort(t, alertServer.URL)
	t.Logf("Alert server at %s:%d", alertHost, alertPort)

	// Create processor config
	cfg := &Config{}

	// Create the processor
	nextConsumer := consumertest.NewNop()
	settings := processortest.NewNopSettings()
	settings.Logger = zap.NewExample()

	proc, err := newProcessor(cfg, settings, nextConsumer)
	require.NoError(t, err)
	require.NotNil(t, proc)

	// Clear any existing policies
	proc.policies = make(map[string]interfaces.Policy)
	proc.policyToEngineMapping = make(map[string]interfaces.HeuristicEntity)

	// Start the processor
	err = proc.Start(context.Background(), nil)
	require.NoError(t, err)

	// Get the routing entity
	routingEntity, exists := proc.activeEntities[constants.RoutingEntity]
	require.True(t, exists, "Routing entity should exist")

	// Create policy builder
	policyBuilder := policy.NewPolicyBuilder()

	// Create notification interface
	notificationInterfaceBuilder := policyBuilder.NotificationInterfaceBuilder()

	// Create alert notifier
	alertNotifier := notificationInterfaceBuilder.
		WithCapability(constants.NotificationInterfaceCapability_Alert).
		WithHost(alertHost).
		WithPort(alertPort).
		WithEndpoint("/alert").
		Build()

	// Create a test policy that will trigger an alert
	alertPolicy := policyBuilder.
		WithName("test-alert-policy").
		WithPreEvaluationCondition("true").
		WithEvaluationCondition("true").
		WithHeuristicEngine(routingEntity).
		WithAlert(alertNotifier).
		WithAlertCondition("true"). // Always trigger alert
		Build()

	// Register the policy
	proc.policies[alertPolicy.Name()] = alertPolicy
	proc.policyToEngineMapping[alertPolicy.Name()] = routingEntity

	// We expect an alert notification
	alertWg.Add(1)

	// Create and send test metrics
	metrics := createTestMetrics()
	addRelevantMetrics(metrics)

	t.Logf("Sending metrics to trigger alert policy evaluation")
	err = proc.ConsumeMetrics(context.Background(), metrics)
	require.NoError(t, err)

	// Wait for notification with timeout
	waitTimeout := 5 * time.Second
	alertDone := waitWithTimeout(t, &alertWg, waitTimeout)

	// Verify alert notification was received
	assert.True(t, alertDone, "Alert notification should have been received")

	if alertDone && len(receivedAlertNotification) > 0 {
		var alertData map[string]interface{}
		err = json.Unmarshal(receivedAlertNotification, &alertData)
		require.NoError(t, err, "Alert notification should be valid JSON")

		// Adjust assertions based on actual format
		assert.Contains(t, alertData, "alert")
		assert.Equal(t, "true", alertData["alert"])
		assert.Contains(t, alertData, "message")
	}

	// Clean up
	err = proc.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestHeuristicEngineProcessorWithRouteNotification(t *testing.T) {
	// Setup test notification server for routes
	var routeWg sync.WaitGroup
	var receivedRouteNotification []byte

	routeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Route notification received: %s %s", r.Method, r.URL.Path)

		// Read and store the request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		t.Logf("Route notification body: %s", string(body))
		receivedRouteNotification = body

		// Send a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))

		routeWg.Done()
	}))
	defer routeServer.Close()

	// Extract host and port from test server
	routeHost, routePort := extractHostPort(t, routeServer.URL)
	t.Logf("Route server at %s:%d", routeHost, routePort)

	// Create processor config
	cfg := &Config{}

	// Create the processor
	nextConsumer := consumertest.NewNop()
	settings := processortest.NewNopSettings()
	settings.Logger = zap.NewExample()

	proc, err := newProcessor(cfg, settings, nextConsumer)
	require.NoError(t, err)
	require.NotNil(t, proc)

	// Clear any existing policies
	proc.policies = make(map[string]interfaces.Policy)
	proc.policyToEngineMapping = make(map[string]interfaces.HeuristicEntity)

	// Start the processor
	err = proc.Start(context.Background(), nil)
	require.NoError(t, err)

	// Get the routing entity
	routingEntity, exists := proc.activeEntities[constants.RoutingEntity]
	require.True(t, exists, "Routing entity should exist")
	// Create policy builder
	policyBuilder := policy.NewPolicyBuilder()

	// Create notification interface
	notificationInterfaceBuilder := policyBuilder.NotificationInterfaceBuilder()

	// Create route notifier
	routeNotifier := notificationInterfaceBuilder.
		WithCapability(constants.NotificationInterfaceCapability_Route).
		WithHost(routeHost).
		WithPort(routePort).
		WithEndpoint("/route").
		Build()

	// Create a test policy that will trigger a route notification
	routePolicy := policyBuilder.
		WithName("test-route-policy").
		WithPreEvaluationCondition("true").
		WithEvaluationCondition("true").
		WithHeuristicEngine(routingEntity).
		WithRoute(routeNotifier).
		WithRouteCondition("true"). // Always trigger route
		Build()

	// Register the policy
	proc.policies[routePolicy.Name()] = routePolicy
	proc.policyToEngineMapping[routePolicy.Name()] = routingEntity

	// We expect a route notification
	routeWg.Add(1)

	// Create and send test metrics
	metrics := createTestMetrics()
	addRelevantMetrics(metrics)

	t.Logf("Sending metrics to trigger route policy evaluation")
	err = proc.ConsumeMetrics(context.Background(), metrics)
	require.NoError(t, err)

	// Wait for notification with timeout
	waitTimeout := 5 * time.Second
	routeDone := waitWithTimeout(t, &routeWg, waitTimeout)

	// Verify route notification was received
	assert.True(t, routeDone, "Route notification should have been received")

	if routeDone && len(receivedRouteNotification) > 0 {
		var routeData map[string]interface{}
		err = json.Unmarshal(receivedRouteNotification, &routeData)
		require.NoError(t, err, "Route notification should be valid JSON")

		// Log the route data for inspection
		t.Logf("Route notification data: %v", routeData)

		// Add assertions based on expected format
		// These will depend on your actual implementation
	}

	// Clean up
	err = proc.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestHeuristicEngineProcessorWithFallbackToRouteNotification(t *testing.T) {
	// Setup test notification servers for both alert and route
	var alertWg sync.WaitGroup
	var routeWg sync.WaitGroup

	alertReceived := false
	routeReceived := false

	// Alert server - should not receive anything in this test
	alertServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("UNEXPECTED: Alert notification received: %s %s", r.Method, r.URL.Path)
		alertReceived = true
		alertWg.Done()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer alertServer.Close()

	// Route server - should receive notification
	var receivedRouteNotification []byte
	routeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Route notification received: %s %s", r.Method, r.URL.Path)

		// Read and store the request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		t.Logf("Route notification body: %s", string(body))
		receivedRouteNotification = body

		routeReceived = true
		routeWg.Done()

		// Send a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer routeServer.Close()

	// Extract host and port from test servers
	alertHost, alertPort := extractHostPort(t, alertServer.URL)
	routeHost, routePort := extractHostPort(t, routeServer.URL)

	t.Logf("Alert server at %s:%d", alertHost, alertPort)
	t.Logf("Route server at %s:%d", routeHost, routePort)

	// Create processor config
	cfg := &Config{}

	// Create the processor
	nextConsumer := consumertest.NewNop()
	settings := processortest.NewNopSettings()
	settings.Logger = zap.NewExample()

	proc, err := newProcessor(cfg, settings, nextConsumer)
	require.NoError(t, err)
	require.NotNil(t, proc)

	// Clear any existing policies
	proc.policies = make(map[string]interfaces.Policy)
	proc.policyToEngineMapping = make(map[string]interfaces.HeuristicEntity)

	// Start the processor
	err = proc.Start(context.Background(), nil)
	require.NoError(t, err)

	// Get the routing entity
	routingEntity, exists := proc.activeEntities[constants.RoutingEntity]
	require.True(t, exists, "Routing entity should exist")
	// Create policy builder
	policyBuilder := policy.NewPolicyBuilder()

	// Create notification interface
	notificationInterfaceBuilder := policyBuilder.NotificationInterfaceBuilder()
	// Create alert notifier
	alertNotifier := notificationInterfaceBuilder.
		WithCapability(constants.NotificationInterfaceCapability_Alert).
		WithHost(alertHost).
		WithPort(alertPort).
		WithEndpoint("/alert").
		Build()

	// Create route notifier
	routeNotifier := notificationInterfaceBuilder.
		WithCapability(constants.NotificationInterfaceCapability_Route).
		WithHost(routeHost).
		WithPort(routePort).
		WithEndpoint("/route").
		Build()

	// Create a test policy that will NOT trigger an alert but WILL trigger a route
	fallbackPolicy := policyBuilder.
		WithName("test-fallback-policy").
		WithPreEvaluationCondition("true").
		WithEvaluationCondition("true").
		WithHeuristicEngine(routingEntity).
		WithAlert(alertNotifier).
		WithAlertCondition("false"). // Alert condition is NOT met
		WithRoute(routeNotifier).
		WithRouteCondition("true"). // Route condition IS met
		Build()

	// Register the policy
	proc.policies[fallbackPolicy.Name()] = fallbackPolicy
	proc.policyToEngineMapping[fallbackPolicy.Name()] = routingEntity

	// We expect only a route notification, not an alert
	routeWg.Add(1)

	// Create and send test metrics
	metrics := createTestMetrics()
	addRelevantMetrics(metrics)

	t.Logf("Sending metrics to trigger fallback policy evaluation")
	err = proc.ConsumeMetrics(context.Background(), metrics)
	require.NoError(t, err)

	// Wait for route notification with timeout
	waitTimeout := 5 * time.Second
	routeDone := waitWithTimeout(t, &routeWg, waitTimeout)

	// Since we didn't add to alertWg, we don't need to wait on it
	// Just check if alertReceived is false
	assert.False(t, alertReceived, "No alert should have been received")
	assert.True(t, routeReceived, "Route notification should have been received")

	if routeDone && len(receivedRouteNotification) > 0 {
		var routeData map[string]interface{}
		err = json.Unmarshal(receivedRouteNotification, &routeData)
		require.NoError(t, err, "Route notification should be valid JSON")

		// Log the route data for inspection
		t.Logf("Route notification data: %v", routeData)

		// Add assertions based on expected format
		// These will depend on your actual implementation
	}

	// Clean up
	err = proc.Shutdown(context.Background())
	require.NoError(t, err)
}

// Helper functions remain the same
func extractHostPort(t *testing.T, url string) (string, int) {
	url = strings.TrimPrefix(url, "http://")
	hostPort := strings.Split(url, ":")
	require.Len(t, hostPort, 2, "URL should contain host and port")
	port, err := strconv.Atoi(hostPort[1])
	require.NoError(t, err, "Port should be a number")
	return hostPort[0], port
}

func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return true // completed normally
	case <-timer.C:
		t.Logf("Timed out waiting after %v", timeout)
		return false // timed out
	}
}

func createTestMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "test-service")
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test-scope")
	m := sm.Metrics().AppendEmpty()
	m.SetName("test.metric")
	m.SetDescription("Test metric for policy evaluation")
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetDoubleValue(42.0)
	return metrics
}

func addRelevantMetrics(metrics pmetric.Metrics) {
	rm := metrics.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	m := sm.Metrics().AppendEmpty()
	m.SetName("system.cpu.utilization")
	m.SetDescription("CPU utilization")
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetDoubleValue(0.95)
}
