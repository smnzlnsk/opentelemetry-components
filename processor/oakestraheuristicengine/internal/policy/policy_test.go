package policy

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/constants"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/types"
	"github.com/stretchr/testify/assert"
)

// MockHeuristicEntity implements interfaces.HeuristicEntity
type pmockHeuristicEntity struct {
	interfaces.HeuristicEntity
	evaluateFunc func(map[string]interface{}) map[string]interface{}
}

func (m *pmockHeuristicEntity) Evaluate(values map[string]interface{}) map[string]interface{} {
	if m.evaluateFunc != nil {
		return m.evaluateFunc(values)
	}
	return values
}

// MockNotificationInterface implements interfaces.NotificationInterface
type pmockNotificationInterface struct {
	interfaces.NotificationInterface
	notified bool
	err      error
}

func (m *pmockNotificationInterface) Notify() error {
	m.notified = true
	return m.err
}

func TestPolicy(t *testing.T) {
	// Create mock notification interfaces
	mockAlert := &pmockNotificationInterface{}
	mockRoute := &pmockNotificationInterface{}

	// Create mock heuristic entity
	mockHeuristic := &pmockHeuristicEntity{
		evaluateFunc: func(values map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"score": 95,
			}
		},
	}

	// Create test expressions
	preEvalExpr, _ := govaluate.NewEvaluableExpression("true")
	evalExpr, _ := govaluate.NewEvaluableExpression("true")
	alertExpr, _ := govaluate.NewEvaluableExpression("score > 90")
	routeExpr, _ := govaluate.NewEvaluableExpression("score <= 90")

	// Create policy
	p := &policy{
		name: "test_policy",
		capabilities: []types.NotificationInterfaceCapability{
			constants.NotificationInterfaceCapability_Alert,
			constants.NotificationInterfaceCapability_Route,
		},
		notificationInterfaces: map[types.NotificationInterfaceCapability]interfaces.NotificationInterface{
			constants.NotificationInterfaceCapability_Alert: mockAlert,
			constants.NotificationInterfaceCapability_Route: mockRoute,
		},
		heuristicEntity:         mockHeuristic,
		preEvaluationConditions: []*govaluate.EvaluableExpression{preEvalExpr},
		evaluationConditions:    []*govaluate.EvaluableExpression{evalExpr},
		alertConditions:         []*govaluate.EvaluableExpression{alertExpr},
		routeConditions:         []*govaluate.EvaluableExpression{routeExpr},
	}

	t.Run("TestBasicProperties", func(t *testing.T) {
		assert.Equal(t, "test_policy", p.Name())
		assert.Equal(t, 2, len(p.Capabilities()))
	})

	t.Run("TestPreEvaluationCondition", func(t *testing.T) {
		err := p.CheckPreEvaluationCondition(map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("TestEvaluationCondition", func(t *testing.T) {
		err := p.CheckEvaluationCondition(map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("TestEnforceWithAlert", func(t *testing.T) {
		err := p.Enforce(map[string]interface{}{})
		assert.NoError(t, err)
		assert.True(t, mockAlert.notified)
		assert.False(t, mockRoute.notified)
	})

	t.Run("TestEnforceWithRoute", func(t *testing.T) {
		// Reset notification flags
		mockAlert.notified = false
		mockRoute.notified = false

		// Update mock to return lower score
		mockHeuristic.evaluateFunc = func(values map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"score": 85,
			}
		}

		err := p.Enforce(map[string]interface{}{})
		assert.NoError(t, err)
		assert.False(t, mockAlert.notified)
		assert.True(t, mockRoute.notified)
	})
}
