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

func (s *UnitTestSuite) Test_BasicReceiveWithTimeoutWorkflow_Success() {
	env := s.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(
		func() {
			message := "testing"
			env.SignalWorkflow("signal-receive-with-timeout", &message)
		},
		time.Minute,
	)
	env.ExecuteWorkflow(BasicReceiveWithTimeoutWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result BasicReceiveWithTimeoutWorkflowResult
	s.NoError(env.GetWorkflowResult(&result))
	s.False(result.HasTimedOut)
	s.Equal("testing", result.Message)

	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_BasicReceiveWithTimeoutWorkflow_TimedOut() {
	env := s.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(BasicReceiveWithTimeoutWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result BasicReceiveWithTimeoutWorkflowResult
	s.NoError(env.GetWorkflowResult(&result))
	s.True(result.HasTimedOut)
	s.Equal("", result.Message)

	env.AssertExpectations(s.T())
}
