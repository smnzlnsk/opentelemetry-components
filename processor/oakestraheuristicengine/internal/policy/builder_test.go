package policy

import (
	"testing"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestPolicyBuilder(t *testing.T) {
	t.Run("empty builder creates nil policy", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()

		// When
		policy := builder.Build()

		// Then
		assert.Nil(t, policy)
	})

	t.Run("builder with route capability creates policy with route", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockNotificationInterface := &mockNotificationInterface{}
		mockHeuristicEntity := &mockHeuristicEntity{}

		// When
		policy := builder.
			WithName("test-policy").
			WithPreEvaluationCondition("true").
			WithEvaluationCondition("true").
			WithRoute(mockNotificationInterface).
			WithRouteCondition("true").
			WithHeuristicEntity(mockHeuristicEntity).
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "test-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Route)
		assert.Equal(t, mockNotificationInterface, policy.GetNotificationInterface(constants.NotificationInterfaceCapability_Route))
	})

	t.Run("builder with multiple capabilities creates policy with all capabilities", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockRouteNotificationInterface := &mockNotificationInterface{}
		mockAlertNotificationInterface := &mockNotificationInterface{}
		mockHeuristicEntity := &mockHeuristicEntity{}
		// When
		policy := builder.
			WithName("multi-cap-policy").
			WithPreEvaluationCondition("true").
			WithEvaluationCondition("true").
			WithHeuristicEntity(mockHeuristicEntity).
			WithRoute(mockRouteNotificationInterface).
			WithRouteCondition("true").
			WithAlert(mockAlertNotificationInterface).
			WithAlertCondition("true").
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "multi-cap-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Route)
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Alert)
		assert.Equal(t, mockRouteNotificationInterface, policy.GetNotificationInterface(constants.NotificationInterfaceCapability_Route))
		assert.Equal(t, mockAlertNotificationInterface, policy.GetNotificationInterface(constants.NotificationInterfaceCapability_Alert))
	})

	t.Run("builder resets internal state after build", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockRouteNotificationInterface := &mockNotificationInterface{}

		// When
		firstPolicy := builder.
			WithName("first-policy").
			WithHeuristicEntity(&mockHeuristicEntity{}).
			WithPreEvaluationCondition("true").
			WithEvaluationCondition("true").
			WithRoute(mockRouteNotificationInterface).
			WithRouteCondition("true").
			Build()

		secondPolicy := builder.Build()

		// Then
		assert.NotNil(t, firstPolicy)
		assert.Equal(t, "first-policy", firstPolicy.Name())
		assert.Contains(t, firstPolicy.Capabilities(), constants.NotificationInterfaceCapability_Route)

		// expect to be nil, because builder is reset
		assert.Nil(t, secondPolicy)
	})

	t.Run("builder with measure factory creates policy with measures", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()

		// When
		route := builder.NotificationInterfaceFactory().CreateNotificationInterfaceBuilder(constants.NotificationInterfaceCapability_Route).Build()
		schedule := builder.NotificationInterfaceFactory().CreateNotificationInterfaceBuilder(constants.NotificationInterfaceCapability_Schedule).Build()

		policy := builder.
			WithName("test-policy").
			WithHeuristicEntity(&mockHeuristicEntity{}).
			WithPreEvaluationCondition("true").
			WithEvaluationCondition("true").
			WithRoute(route).
			WithRouteCondition("true").
			WithSchedule(schedule).
			WithScheduleCondition("true").
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "test-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Route)
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Schedule)
		assert.Equal(t, route, policy.GetNotificationInterface(constants.NotificationInterfaceCapability_Route))
		assert.Equal(t, schedule, policy.GetNotificationInterface(constants.NotificationInterfaceCapability_Schedule))
	})

	t.Run("builder requires pre-evaluation conditions", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockNotificationInterface := &mockNotificationInterface{}
		mockHeuristicEntity := &mockHeuristicEntity{}

		// When
		policy := builder.
			WithName("test-policy").
			WithRoute(mockNotificationInterface).
			WithHeuristicEntity(mockHeuristicEntity).
			WithEvaluationCondition("value > 0").
			Build()

		// Then
		assert.Nil(t, policy, "policy should be nil without pre-evaluation conditions")
	})

	t.Run("builder requires evaluation conditions", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockNotificationInterface := &mockNotificationInterface{}
		mockHeuristicEntity := &mockHeuristicEntity{}

		// When
		policy := builder.
			WithName("test-policy").
			WithRoute(mockNotificationInterface).
			WithHeuristicEntity(mockHeuristicEntity).
			WithPreEvaluationCondition("enabled == true").
			Build()

		// Then
		assert.Nil(t, policy, "policy should be nil without evaluation conditions")
	})

	t.Run("builder requires conditions for each notification interface", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockNotificationInterface := &mockNotificationInterface{}
		mockHeuristicEntity := &mockHeuristicEntity{}

		// When
		policy := builder.
			WithName("test-policy").
			WithRoute(mockNotificationInterface).
			WithAlert(mockNotificationInterface).
			WithHeuristicEntity(mockHeuristicEntity).
			WithPreEvaluationCondition("enabled == true").
			WithEvaluationCondition("value > 0").
			WithRouteCondition("route == true").
			// Missing alert condition
			Build()

		// Then
		assert.Nil(t, policy, "policy should be nil when notification interface lacks conditions")
	})

	t.Run("builder creates valid policy with all required conditions", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockNotificationInterface := &mockNotificationInterface{}
		mockHeuristicEntity := &mockHeuristicEntity{}

		// When
		policy := builder.
			WithName("test-policy").
			WithRoute(mockNotificationInterface).
			WithAlert(mockNotificationInterface).
			WithHeuristicEntity(mockHeuristicEntity).
			WithPreEvaluationCondition("enabled == true").
			WithEvaluationCondition("value > 0").
			WithRouteCondition("route == true").
			WithAlertCondition("alert == true").
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "test-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Route)
		assert.Contains(t, policy.Capabilities(), constants.NotificationInterfaceCapability_Alert)
	})
}

// Mock implementation of MeasureNotifier for testing
type mockNotificationInterface struct {
	interfaces.NotificationInterface
}

// Add mock for HeuristicEntity
type mockHeuristicEntity struct {
	interfaces.HeuristicEntity
}
