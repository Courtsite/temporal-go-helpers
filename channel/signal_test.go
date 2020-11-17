package channel

import (
	"testing"
	"time"

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

func (s *UnitTestSuite) Test_BasicReceiveWithTimeoutWorkflow__WithPayload__Success() {
	env := s.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(
		func() {
			message := "testing"
			env.SignalWorkflow("signal-receive-with-timeout", &message)
		},
		time.Minute,
	)
	env.ExecuteWorkflow(BasicReceiveWithTimeoutWorkflow__WithPayload)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result BasicReceiveWithTimeoutWorkflowResult
	s.NoError(env.GetWorkflowResult(&result))
	s.False(result.HasTimedOut)
	s.Equal("testing", result.Message)

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_BasicReceiveWithTimeoutWorkflow__NoPayload__Success() {
	env := s.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(
		func() {
			env.SignalWorkflow("signal-receive-with-timeout", nil)
		},
		time.Minute,
	)
	env.ExecuteWorkflow(BasicReceiveWithTimeoutWorkflow__NoPayload)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result BasicReceiveWithTimeoutWorkflowResult
	s.NoError(env.GetWorkflowResult(&result))
	s.False(result.HasTimedOut)
	s.Equal("", result.Message)

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_BasicReceiveWithTimeoutWorkflow__WithPayload__TimedOut() {
	env := s.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(BasicReceiveWithTimeoutWorkflow__WithPayload)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result BasicReceiveWithTimeoutWorkflowResult
	s.NoError(env.GetWorkflowResult(&result))
	s.True(result.HasTimedOut)
	s.Equal("", result.Message)

	env.AssertExpectations(s.T())
}
