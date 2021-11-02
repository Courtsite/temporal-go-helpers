package channel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type DrainUnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestDrainUnitTestSuite(t *testing.T) {
	suite.Run(t, new(DrainUnitTestSuite))
}

func (s *DrainUnitTestSuite) Test_BasicDrainWorkflow_WithoutInflight() {
	env := s.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(
		func() {
			message := "OK"
			env.SignalWorkflow("signal", &message)
		},
		time.Second,
	)
	env.RegisterDelayedCallback(
		func() {
			message := "STOP"
			env.SignalWorkflow("signal", &message)
		},
		time.Second*6,
	)
	env.ExecuteWorkflow(BasicDrainWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result int
	s.NoError(env.GetWorkflowResult(&result))
	s.Equal(0, result)

	env.AssertExpectations(s.T())
}

func (s *DrainUnitTestSuite) Test_BasicDrainWorkflow_WithInflight() {
	env := s.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(
		func() {
			message := "OK"
			env.SignalWorkflow("signal", &message)

			// These in-flight signals should be dropped.
			env.SignalWorkflow("signal", &message)
			env.SignalWorkflow("signal", &message)

			// This should be disregarded too.
			message = "STOP"
			env.SignalWorkflow("signal", &message)
		},
		time.Second,
	)
	env.RegisterDelayedCallback(
		func() {
			// This should be disregarded.
			message := "OK"
			env.SignalWorkflow("signal", &message)
		},
		time.Second*3,
	)
	env.RegisterDelayedCallback(
		func() {
			// This will be handled, resulting in the workflow's completion.
			message := "STOP"
			env.SignalWorkflow("signal", &message)
		},
		time.Second*6,
	)
	env.ExecuteWorkflow(BasicDrainWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result int
	s.NoError(env.GetWorkflowResult(&result))
	s.Equal(4, result)

	env.AssertExpectations(s.T())
}
