package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_BasicSagaWorkflow_Success() {
	env := s.NewTestWorkflowEnvironment()
	var ca *CalculatorActivities
	env.RegisterActivity(ca)

	env.ExecuteWorkflow(BasicSagaWorkflow, 0)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var finalAmount int
	s.NoError(env.GetWorkflowResult(&finalAmount))
	s.Equal(60, finalAmount)

	result, err := env.QueryWorkflow("currentAmount")
	s.NoError(err)
	var currentAmount int
	err = result.Get(&currentAmount)
	s.NoError(err)
	s.Equal(60, currentAmount)

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_BasicSagaWorkflow_Failure() {
	env := s.NewTestWorkflowEnvironment()
	var ca *CalculatorActivities
	env.RegisterActivity(ca)

	env.OnActivity(ca.Add, mock.Anything, 0, 10).Return(10, nil)
	env.OnActivity(ca.Add, mock.Anything, 10, 20).Return(30, nil)
	env.OnActivity(ca.Add, mock.Anything, 30, 30).Return(0, errors.New("Test Compensation"))

	env.ExecuteWorkflow(BasicSagaWorkflow, 0)

	s.True(env.IsWorkflowCompleted())
	s.EqualError(env.GetWorkflowError(), "workflow execution error (type: BasicSagaWorkflow, workflowID: default-test-workflow-id, runID: default-test-run-id): activity error (type: Add, scheduledEventID: 0, startedEventID: 0, identity: ): Test Compensation")

	result, err := env.QueryWorkflow("currentAmount")
	s.NoError(err)
	var currentAmount int
	err = result.Get(&currentAmount)
	s.NoError(err)
	s.Equal(15, currentAmount)

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_BasicSagaWorkflow_Cancelled() {
	env := s.NewTestWorkflowEnvironment()
	var ca *CalculatorActivities
	env.RegisterActivity(ca)

	env.OnActivity(ca.Add, mock.Anything, 0, 10).Return(10, nil)
	env.OnActivity(ca.Add, mock.Anything, 10, 20).Return(30, nil)
	env.OnActivity(ca.Add, mock.Anything, 30, 30).Return(func(ctx context.Context, a int, b int) (int, error) {
		env.CancelWorkflow()
		return 0, nil
	})

	env.ExecuteWorkflow(BasicSagaWorkflow, 0)

	s.True(env.IsWorkflowCompleted())
	s.EqualError(env.GetWorkflowError(), "workflow execution error (type: BasicSagaWorkflow, workflowID: default-test-workflow-id, runID: default-test-run-id): canceled")

	result, err := env.QueryWorkflow("currentAmount")
	s.NoError(err)
	var currentAmount int
	err = result.Get(&currentAmount)
	s.NoError(err)
	s.Equal(15, currentAmount)

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_MultipleCompensateSagaWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	var ca *CalculatorActivities
	env.RegisterActivity(ca)

	env.ExecuteWorkflow(MultipleCompensateSagaWorkflow, 0)

	s.True(env.IsWorkflowCompleted())
	s.EqualError(env.GetWorkflowError(), "workflow execution error (type: MultipleCompensateSagaWorkflow, workflowID: default-test-workflow-id, runID: default-test-run-id): Compensate: Compensate should only be called once")

	result, err := env.QueryWorkflow("currentAmount")
	s.NoError(err)
	var currentAmount int
	err = result.Get(&currentAmount)
	s.NoError(err)
	s.Equal(5, currentAmount)
}
