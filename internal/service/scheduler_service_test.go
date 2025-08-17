package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SchedulerServiceTestSuite defines the test suite for SchedulerService
type SchedulerServiceTestSuite struct {
	suite.Suite
	logger  *logger.Logger
	service SchedulerService
	ctx     context.Context
	cancel  context.CancelFunc
}

func (suite *SchedulerServiceTestSuite) SetupTest() {
	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}

	suite.logger = logger.New(cfg)
	suite.service = NewSchedulerService(suite.logger)
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
}

func (suite *SchedulerServiceTestSuite) TearDownTest() {
	_ = suite.service.Stop()
	suite.cancel()
}

// Helper functions
func (suite *SchedulerServiceTestSuite) createMockJob(name string, shouldFail bool, executionCount *int32) func(context.Context) error {
	return func(ctx context.Context) error {
		atomic.AddInt32(executionCount, 1)
		if shouldFail {
			return errors.New("mock job error")
		}
		return nil
	}
}

func (suite *SchedulerServiceTestSuite) waitForJobExecution(executionCount *int32, expectedCount int32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(executionCount) >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func (suite *SchedulerServiceTestSuite) TestStartSchedulerSuccess() {
	err := suite.service.Start(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.service.IsRunning())
}

func (suite *SchedulerServiceTestSuite) TestStartSchedulerAlreadyRunning() {
	// Start scheduler first time
	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.service.IsRunning())

	// Try to start again
	err = suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.service.IsRunning())
}

func (suite *SchedulerServiceTestSuite) TestStopSchedulerSuccess() {
	// Start scheduler first
	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.service.IsRunning())

	// Stop scheduler
	err = suite.service.Stop()
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.service.IsRunning())
}

func (suite *SchedulerServiceTestSuite) TestStopSchedulerNotRunning() {
	err := suite.service.Stop()
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.service.IsRunning())
}

func (suite *SchedulerServiceTestSuite) TestIsRunningInitialState() {
	assert.False(suite.T(), suite.service.IsRunning())
}

func (suite *SchedulerServiceTestSuite) TestAddJobWhenSchedulerStopped() {
	var executionCount int32
	job := suite.createMockJob("test-job", false, &executionCount)

	suite.service.AddJob("test-job", 100*time.Millisecond, job)

	status := suite.service.GetJobStatus()
	assert.Contains(suite.T(), status, "test-job")
	assert.Equal(suite.T(), "test-job", status["test-job"].Name)
	assert.Equal(suite.T(), 100*time.Millisecond, status["test-job"].Interval)
	assert.False(suite.T(), status["test-job"].IsRunning)
	assert.Equal(suite.T(), int64(0), status["test-job"].RunCount)
}

func (suite *SchedulerServiceTestSuite) TestAddJobWhenSchedulerRunning() {
	var executionCount int32
	job := suite.createMockJob("test-job", false, &executionCount)

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("test-job", 50*time.Millisecond, job)

	success := suite.waitForJobExecution(&executionCount, 1, 200*time.Millisecond)
	assert.True(suite.T(), success, "Job should have executed at least once")

	status := suite.service.GetJobStatus()
	assert.Contains(suite.T(), status, "test-job")
	assert.True(suite.T(), status["test-job"].RunCount > 0)
}

func (suite *SchedulerServiceTestSuite) TestAddJobReplaceExisting() {
	var executionCount1, executionCount2 int32
	job1 := suite.createMockJob("test-job-1", false, &executionCount1)
	job2 := suite.createMockJob("test-job-2", false, &executionCount2)

	suite.service.AddJob("same-name", 100*time.Millisecond, job1)

	suite.service.AddJob("same-name", 200*time.Millisecond, job2)

	status := suite.service.GetJobStatus()
	assert.Contains(suite.T(), status, "same-name")
	assert.Equal(suite.T(), 200*time.Millisecond, status["same-name"].Interval)
}

func (suite *SchedulerServiceTestSuite) TestRemoveJobExists() {
	var executionCount int32
	job := suite.createMockJob("test-job", false, &executionCount)

	suite.service.AddJob("test-job", 100*time.Millisecond, job)

	status := suite.service.GetJobStatus()
	assert.Contains(suite.T(), status, "test-job")

	suite.service.RemoveJob("test-job")

	status = suite.service.GetJobStatus()
	assert.NotContains(suite.T(), status, "test-job")
}

func (suite *SchedulerServiceTestSuite) TestRemoveJobNotExists() {
	suite.service.RemoveJob("non-existent-job")

	status := suite.service.GetJobStatus()
	assert.NotContains(suite.T(), status, "non-existent-job")
}

func (suite *SchedulerServiceTestSuite) TestJobExecutionSuccess() {
	var executionCount int32
	job := suite.createMockJob("success-job", false, &executionCount)

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("success-job", 50*time.Millisecond, job)

	success := suite.waitForJobExecution(&executionCount, 2, 300*time.Millisecond)
	assert.True(suite.T(), success, "Job should have executed at least twice")

	status := suite.service.GetJobStatus()
	jobStatus := status["success-job"]
	assert.True(suite.T(), jobStatus.RunCount >= 2)
	assert.Equal(suite.T(), int64(0), jobStatus.ErrorCount)
	assert.Empty(suite.T(), jobStatus.LastError)
	assert.NotNil(suite.T(), jobStatus.LastRun)
	assert.NotNil(suite.T(), jobStatus.NextRun)
}

func (suite *SchedulerServiceTestSuite) TestJobExecutionError() {
	var executionCount int32
	job := suite.createMockJob("error-job", true, &executionCount)

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("error-job", 50*time.Millisecond, job)

	success := suite.waitForJobExecution(&executionCount, 2, 300*time.Millisecond)
	assert.True(suite.T(), success, "Job should have executed at least twice despite errors")

	status := suite.service.GetJobStatus()
	jobStatus := status["error-job"]
	assert.True(suite.T(), jobStatus.RunCount >= 2)
	assert.True(suite.T(), jobStatus.ErrorCount >= 2)
	assert.Equal(suite.T(), "mock job error", jobStatus.LastError)
	assert.NotNil(suite.T(), jobStatus.LastRun)
	assert.NotNil(suite.T(), jobStatus.NextRun)
}

func (suite *SchedulerServiceTestSuite) TestJobExecutionStopsOnContextCancellation() {
	var executionCount int32
	job := suite.createMockJob("cancel-job", false, &executionCount)

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("cancel-job", 30*time.Millisecond, job)

	success := suite.waitForJobExecution(&executionCount, 1, 100*time.Millisecond)
	assert.True(suite.T(), success)

	initialCount := atomic.LoadInt32(&executionCount)

	err = suite.service.Stop()
	assert.NoError(suite.T(), err)

	time.Sleep(100 * time.Millisecond)
	finalCount := atomic.LoadInt32(&executionCount)

	assert.True(suite.T(), finalCount <= initialCount+1, "Job should stop executing after scheduler stop")
}

func (suite *SchedulerServiceTestSuite) TestGetJobStatusEmpty() {
	status := suite.service.GetJobStatus()
	assert.Empty(suite.T(), status)
}

func (suite *SchedulerServiceTestSuite) TestGetJobStatusMultipleJobs() {
	var count1, count2 int32
	job1 := suite.createMockJob("job-1", false, &count1)
	job2 := suite.createMockJob("job-2", false, &count2)

	suite.service.AddJob("job-1", 100*time.Millisecond, job1)
	suite.service.AddJob("job-2", 200*time.Millisecond, job2)

	status := suite.service.GetJobStatus()
	assert.Len(suite.T(), status, 2)
	assert.Contains(suite.T(), status, "job-1")
	assert.Contains(suite.T(), status, "job-2")
	assert.Equal(suite.T(), 100*time.Millisecond, status["job-1"].Interval)
	assert.Equal(suite.T(), 200*time.Millisecond, status["job-2"].Interval)
}

func (suite *SchedulerServiceTestSuite) TestJobStatusMetricsUpdate() {
	var executionCount int32
	job := suite.createMockJob("metrics-job", false, &executionCount)

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("metrics-job", 50*time.Millisecond, job)

	success := suite.waitForJobExecution(&executionCount, 3, 300*time.Millisecond)
	assert.True(suite.T(), success)

	status := suite.service.GetJobStatus()
	jobStatus := status["metrics-job"]

	assert.True(suite.T(), jobStatus.RunCount >= 3)
	assert.NotNil(suite.T(), jobStatus.LastRun)
	assert.NotNil(suite.T(), jobStatus.NextRun)
	assert.True(suite.T(), jobStatus.AverageRunTime > 0)
	assert.True(suite.T(), jobStatus.NextRun.After(*jobStatus.LastRun))
}

func (suite *SchedulerServiceTestSuite) TestMultipleJobsConcurrentExecution() {
	var count1, count2, count3 int32
	job1 := suite.createMockJob("concurrent-1", false, &count1)
	job2 := suite.createMockJob("concurrent-2", false, &count2)
	job3 := suite.createMockJob("concurrent-3", false, &count3)

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("concurrent-1", 30*time.Millisecond, job1)
	suite.service.AddJob("concurrent-2", 50*time.Millisecond, job2)
	suite.service.AddJob("concurrent-3", 70*time.Millisecond, job3)

	time.Sleep(200 * time.Millisecond)

	assert.True(suite.T(), atomic.LoadInt32(&count1) > 0)
	assert.True(suite.T(), atomic.LoadInt32(&count2) > 0)
	assert.True(suite.T(), atomic.LoadInt32(&count3) > 0)

	assert.True(suite.T(), atomic.LoadInt32(&count1) >= atomic.LoadInt32(&count2))
	assert.True(suite.T(), atomic.LoadInt32(&count2) >= atomic.LoadInt32(&count3))
}

func (suite *SchedulerServiceTestSuite) TestJobExecutionWithTimeout() {
	var executionCount int32

	longRunningJob := func(ctx context.Context) error {
		atomic.AddInt32(&executionCount, 1)

		select {
		case <-time.After(10 * time.Minute):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := suite.service.Start(suite.ctx)
	assert.NoError(suite.T(), err)

	suite.service.AddJob("timeout-job", 50*time.Millisecond, longRunningJob)

	success := suite.waitForJobExecution(&executionCount, 1, 200*time.Millisecond)
	assert.True(suite.T(), success)

	time.Sleep(100 * time.Millisecond)

	status := suite.service.GetJobStatus()
	jobStatus := status["timeout-job"]

	assert.True(suite.T(), jobStatus.RunCount > 0)
}

func (suite *SchedulerServiceTestSuite) TestAddJobWithZeroInterval() {
	var executionCount int32
	job := suite.createMockJob("zero-interval", false, &executionCount)

	suite.service.AddJob("zero-interval", 0, job)

	status := suite.service.GetJobStatus()
	assert.Contains(suite.T(), status, "zero-interval")
	assert.Equal(suite.T(), time.Duration(0), status["zero-interval"].Interval)
}

func (suite *SchedulerServiceTestSuite) TestAddJobWithNilFunction() {
	// This should not panic
	suite.service.AddJob("nil-job", 100*time.Millisecond, nil)

	status := suite.service.GetJobStatus()
	assert.Contains(suite.T(), status, "nil-job")
}

// Run the test suite
func TestSchedulerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerServiceTestSuite))
}
