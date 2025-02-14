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
		mockMeasure := &mockMeasureNotifier{}

		// When
		policy := builder.
			WithName("test-policy").
			WithRoute(mockMeasure).
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "test-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Route)
		assert.Equal(t, mockMeasure, policy.GetMeasure(constants.PolicyNotificationCapability_Route))
	})

	t.Run("builder with multiple capabilities creates policy with all capabilities", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockRouteMeasure := &mockMeasureNotifier{}
		mockAlertMeasure := &mockMeasureNotifier{}

		// When
		policy := builder.
			WithName("multi-cap-policy").
			WithRoute(mockRouteMeasure).
			WithAlert(mockAlertMeasure).
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "multi-cap-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Route)
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Alert)
		assert.Equal(t, mockRouteMeasure, policy.GetMeasure(constants.PolicyNotificationCapability_Route))
		assert.Equal(t, mockAlertMeasure, policy.GetMeasure(constants.PolicyNotificationCapability_Alert))
	})

	t.Run("builder resets internal state after build", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()
		mockRouteMeasure := &mockMeasureNotifier{}

		// When
		firstPolicy := builder.
			WithName("first-policy").
			WithRoute(mockRouteMeasure).
			Build()

		secondPolicy := builder.Build()

		// Then
		assert.NotNil(t, firstPolicy)
		assert.Equal(t, "first-policy", firstPolicy.Name())
		assert.Contains(t, firstPolicy.Capabilities(), constants.PolicyNotificationCapability_Route)

		// expect to be nil, because builder is reset
		assert.Nil(t, secondPolicy)
	})

	t.Run("builder with measure factory creates policy with measures", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()

		// When
		route := builder.MeasureFactory().CreateMeasure(constants.MeasureTypeRoute)
		schedule := builder.MeasureFactory().CreateMeasure(constants.MeasureTypeSchedule)

		policy := builder.
			WithName("test-policy").
			WithRoute(route).
			WithSchedule(schedule).
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "test-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Route)
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Schedule)
		assert.Equal(t, route, policy.GetMeasure(constants.PolicyNotificationCapability_Route))
		assert.Equal(t, schedule, policy.GetMeasure(constants.PolicyNotificationCapability_Schedule))
	})

	t.Run("builder with measure factory does not create measures for unknown types", func(t *testing.T) {
		// Given
		builder := NewPolicyBuilder()

		// When
		route := builder.MeasureFactory().CreateMeasure(constants.MeasureTypeRoute)
		unknown := builder.MeasureFactory().CreateMeasure("unknown")

		policy := builder.
			WithName("test-policy").
			WithRoute(route).
			WithSchedule(unknown).
			Build()

		// Then
		assert.NotNil(t, policy)
		assert.Equal(t, "test-policy", policy.Name())
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Route)
		assert.Contains(t, policy.Capabilities(), constants.PolicyNotificationCapability_Schedule)
		assert.Equal(t, route, policy.GetMeasure(constants.PolicyNotificationCapability_Route))
		assert.Equal(t, unknown, nil)
	})
}

// Mock implementation of MeasureNotifier for testing
type mockMeasureNotifier struct {
	interfaces.MeasureNotifier
}
