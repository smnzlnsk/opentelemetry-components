package policy

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"github.com/stretchr/testify/assert"
)

// MockHeuristicEntity implements interfaces.HeuristicEntity
type pmockHeuristicEntity struct {
	domain.HeuristicEntity
	evaluateFunc func(processorIdentifier string, jobname string, values map[string]interface{}) domain.Evaluation
}

func (m *pmockHeuristicEntity) Evaluate(processorIdentifier string, options ...interface{}) (domain.EvaluationResult, error) {
	if m.evaluateFunc != nil {
		evaluation := m.evaluateFunc(processorIdentifier, "test_job", map[string]interface{}{})
		return domain.EvaluationResult{
			JobName: evaluation.JobName,
			Results: evaluation.Entries,
		}, nil
	}
	return domain.EvaluationResult{
		JobName: "test_job",
		Results: []domain.EvaluationEntry{
			{InstanceNumber: 1, Priority: 1.0},
		},
	}, nil
}

// MockNotificationInterface implements interfaces.NotificationInterface
type pmockNotificationInterface struct {
	domain.NotificationInterface[any]
	notified bool
	err      error
}

func (m *pmockNotificationInterface) Notify(notification any) error {
	m.notified = true
	return m.err
}

func TestPolicy(t *testing.T) {
	// Create mock notification interfaces
	mockAlert := &pmockNotificationInterface{}
	mockRoute := &pmockNotificationInterface{}

	// Create mock heuristic entity
	mockHeuristic := &pmockHeuristicEntity{
		evaluateFunc: func(processorIdentifier string, jobname string, values map[string]interface{}) domain.Evaluation {
			return domain.Evaluation{
				JobName: jobname,
				Entries: []domain.EvaluationEntry{
					{InstanceNumber: 1, Priority: 95},
				},
			}
		},
	}

	// Create test expressions
	preEvalExpr, _ := govaluate.NewEvaluableExpression("true")
	evalExpr, _ := govaluate.NewEvaluableExpression("true")
	alertExpr, _ := govaluate.NewEvaluableExpression("result > 90")
	routeExpr, _ := govaluate.NewEvaluableExpression("result <= 90")

	// Create policy
	p := &policy{
		name: "test_policy",
		capabilities: []domain.NotificationInterfaceCapability{
			domain.NotificationInterfaceCapability_Alert,
			domain.NotificationInterfaceCapability_Route,
		},
		notificationInterfaces: map[domain.NotificationInterfaceCapability]domain.NotificationInterface[any]{
			domain.NotificationInterfaceCapability_Alert: mockAlert,
			domain.NotificationInterfaceCapability_Route: mockRoute,
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
		err := p.Enforce("test_processor", "test_job")
		assert.NoError(t, err)
		assert.True(t, mockAlert.notified)
		assert.False(t, mockRoute.notified)
	})

	t.Run("TestEnforceWithRoute", func(t *testing.T) {
		// Reset notification flags
		mockAlert.notified = false
		mockRoute.notified = false

		// Update mock to return lower score
		mockHeuristic.evaluateFunc = func(processorIdentifier string, jobname string, values map[string]interface{}) domain.Evaluation {
			return domain.Evaluation{
				JobName: jobname,
				Entries: []domain.EvaluationEntry{
					{InstanceNumber: 1, Priority: 85},
				},
			}
		}

		err := p.Enforce("test_processor", "test_job")
		assert.NoError(t, err)
		assert.False(t, mockAlert.notified)
		assert.True(t, mockRoute.notified)
	})
}
