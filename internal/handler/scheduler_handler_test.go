package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockSchedulerService is a mock implementation of SchedulerService
type MockSchedulerService struct {
	mock.Mock
}

func (m *MockSchedulerService) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSchedulerService) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSchedulerService) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSchedulerService) AddJob(name string, interval time.Duration, job func(context.Context) error) {
	m.Called(name, interval, job)
}

func (m *MockSchedulerService) RemoveJob(name string) {
	m.Called(name)
}

func (m *MockSchedulerService) GetJobStatus() map[string]model.JobStatus {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]model.JobStatus)
}

// SchedulerHandlerTestSuite defines the test suite for SchedulerHandler
type SchedulerHandlerTestSuite struct {
	suite.Suite
	mockService *MockSchedulerService
	logger      *logger.Logger
	handler     SchedulerHandler
	echo        *echo.Echo
}

func (suite *SchedulerHandlerTestSuite) SetupTest() {
	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}

	suite.mockService = new(MockSchedulerService)
	suite.logger = logger.New(cfg)
	suite.handler = NewSchedulerHandler(suite.mockService, suite.logger)
	suite.echo = echo.New()
	suite.echo.Validator = &MockValidator{}
}

func (suite *SchedulerHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// Helper functions
func (suite *SchedulerHandlerTestSuite) createMockJobStatus() map[string]model.JobStatus {
	now := time.Now()
	lastRun1 := now.Add(-1 * time.Hour)
	nextRun1 := now.Add(1 * time.Hour)
	lastRun2 := now.Add(-30 * time.Minute)
	nextRun2 := now.Add(30 * time.Minute)
	lastRun3 := now.Add(-2 * time.Hour)
	nextRun3 := now.Add(2 * time.Hour)

	return map[string]model.JobStatus{
		"headlines_aggregator": {
			Name:           "headlines_aggregator",
			Interval:       1 * time.Hour,
			LastRun:        &lastRun1,
			NextRun:        &nextRun1,
			RunCount:       42,
			ErrorCount:     0,
			LastError:      "",
			IsRunning:      false,
			AverageRunTime: 30 * time.Second,
		},
		"category_aggregator": {
			Name:           "category_aggregator",
			Interval:       2 * time.Hour,
			LastRun:        &lastRun2,
			NextRun:        &nextRun2,
			RunCount:       25,
			ErrorCount:     1,
			LastError:      "",
			IsRunning:      true,
			AverageRunTime: 45 * time.Second,
		},
		"source_aggregator": {
			Name:           "source_aggregator",
			Interval:       3 * time.Hour,
			LastRun:        &lastRun3,
			NextRun:        &nextRun3,
			RunCount:       15,
			ErrorCount:     3,
			LastError:      "connection timeout",
			IsRunning:      false,
			AverageRunTime: 1 * time.Minute,
		},
	}
}

func (suite *SchedulerHandlerTestSuite) createEchoContext(method, target string, body interface{}) (echo.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, target, bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	return c, rec
}

func (suite *SchedulerHandlerTestSuite) createEchoContextWithParam(method, target, paramName, paramValue string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	c := suite.echo.NewContext(req, rec)
	c.SetParamNames(paramName)
	c.SetParamValues(paramValue)
	return c, rec
}

func (suite *SchedulerHandlerTestSuite) TestGetStatusWhenSchedulerRunning() {
	mockJobStatus := suite.createMockJobStatus()

	suite.mockService.On("IsRunning").Return(true)
	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/status", nil)

	err := suite.handler.GetStatus(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Scheduler status retrieved successfully", response.Message)

	dataBytes, _ := json.Marshal(response.Data)
	var statusData model.SchedulerStatusResponse
	err = json.Unmarshal(dataBytes, &statusData)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), statusData.SchedulerRunning)
	assert.Equal(suite.T(), len(mockJobStatus), statusData.JobsCount)
	assert.WithinDuration(suite.T(), time.Now(), statusData.Timestamp, 5*time.Second)
}

func (suite *SchedulerHandlerTestSuite) TestGetStatusWhenSchedulerNotRunning() {
	mockJobStatus := suite.createMockJobStatus()

	suite.mockService.On("IsRunning").Return(false)
	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/status", nil)

	err := suite.handler.GetStatus(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	dataBytes, _ := json.Marshal(response.Data)
	var statusData model.SchedulerStatusResponse
	err = json.Unmarshal(dataBytes, &statusData)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), statusData.SchedulerRunning)
	assert.Equal(suite.T(), len(mockJobStatus), statusData.JobsCount)
}

func (suite *SchedulerHandlerTestSuite) TestGetStatusWithEmptyJobStatus() {
	emptyJobStatus := make(map[string]model.JobStatus)

	suite.mockService.On("IsRunning").Return(true)
	suite.mockService.On("GetJobStatus").Return(emptyJobStatus)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/status", nil)

	err := suite.handler.GetStatus(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	dataBytes, _ := json.Marshal(response.Data)
	var statusData model.SchedulerStatusResponse
	err = json.Unmarshal(dataBytes, &statusData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, statusData.JobsCount)
	assert.Equal(suite.T(), emptyJobStatus, statusData.Jobs)
}

func (suite *SchedulerHandlerTestSuite) TestGetJobsSuccess() {
	mockJobStatus := suite.createMockJobStatus()

	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/jobs", nil)

	err := suite.handler.GetJobs(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Jobs retrieved successfully", response.Message)

	dataBytes, _ := json.Marshal(response.Data)
	var jobsData model.JobsResponse
	err = json.Unmarshal(dataBytes, &jobsData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(mockJobStatus), jobsData.Count)
	assert.WithinDuration(suite.T(), time.Now(), jobsData.Timestamp, 5*time.Second)
}

func (suite *SchedulerHandlerTestSuite) TestGetJobsWithEmptyJobs() {
	emptyJobStatus := make(map[string]model.JobStatus)

	suite.mockService.On("GetJobStatus").Return(emptyJobStatus)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/jobs", nil)

	err := suite.handler.GetJobs(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	dataBytes, _ := json.Marshal(response.Data)
	var jobsData model.JobsResponse
	err = json.Unmarshal(dataBytes, &jobsData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, jobsData.Count)
	assert.Equal(suite.T(), emptyJobStatus, jobsData.Jobs)
}

func (suite *SchedulerHandlerTestSuite) TestTriggerJobSuccess() {
	mockJobStatus := suite.createMockJobStatus()
	jobName := "headlines_aggregator"

	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContextWithParam(http.MethodPost, "/scheduler/jobs/"+jobName+"/trigger", "name", jobName)

	err := suite.handler.TriggerJob(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Job trigger acknowledged", response.Message)

	dataBytes, _ := json.Marshal(response.Data)
	var triggerData model.JobTriggerResponse
	err = json.Unmarshal(dataBytes, &triggerData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), jobName, triggerData.JobName)
	assert.Contains(suite.T(), triggerData.Note, "Job will run according to its schedule")
	assert.WithinDuration(suite.T(), time.Now(), triggerData.Timestamp, 5*time.Second)
}

func (suite *SchedulerHandlerTestSuite) TestTriggerJobWithEmptyJobName() {
	c, rec := suite.createEchoContextWithParam(http.MethodPost, "/scheduler/jobs//trigger", "name", "")

	err := suite.handler.TriggerJob(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "Job name is required")
}

func (suite *SchedulerHandlerTestSuite) TestTriggerJobWithNonExistentJob() {
	mockJobStatus := suite.createMockJobStatus()
	jobName := "non_existent_job"

	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContextWithParam(http.MethodPost, "/scheduler/jobs/"+jobName+"/trigger", "name", jobName)

	err := suite.handler.TriggerJob(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "Job not found")
	assert.Contains(suite.T(), response.Error.Details, "Available jobs:")

	for jobName := range mockJobStatus {
		assert.Contains(suite.T(), response.Error.Details, jobName)
	}
}

func (suite *SchedulerHandlerTestSuite) TestTriggerJobWhenJobAlreadyRunning() {
	mockJobStatus := suite.createMockJobStatus()
	jobName := "category_aggregator"

	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContextWithParam(http.MethodPost, "/scheduler/jobs/"+jobName+"/trigger", "name", jobName)

	err := suite.handler.TriggerJob(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusConflict, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "Job is already running")
}

func (suite *SchedulerHandlerTestSuite) TestTriggerJobWithFailedJob() {
	mockJobStatus := suite.createMockJobStatus()
	jobName := "source_aggregator"

	suite.mockService.On("GetJobStatus").Return(mockJobStatus)

	c, rec := suite.createEchoContextWithParam(http.MethodPost, "/scheduler/jobs/"+jobName+"/trigger", "name", jobName)

	err := suite.handler.TriggerJob(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Job trigger acknowledged", response.Message)
}

func (suite *SchedulerHandlerTestSuite) TestGetJobNamesHelper() {
	mockJobStatus := suite.createMockJobStatus()

	jobNames := getJobNames(mockJobStatus)

	assert.Equal(suite.T(), len(mockJobStatus), len(jobNames))

	jobNameSet := make(map[string]bool)
	for _, name := range jobNames {
		jobNameSet[name] = true
	}

	for expectedName := range mockJobStatus {
		assert.True(suite.T(), jobNameSet[expectedName], "Job name %s should be present", expectedName)
	}
}

func (suite *SchedulerHandlerTestSuite) TestGetJobNamesWithEmptyMap() {
	emptyJobStatus := make(map[string]model.JobStatus)

	jobNames := getJobNames(emptyJobStatus)

	assert.Equal(suite.T(), 0, len(jobNames))
	assert.NotNil(suite.T(), jobNames)
}

func (suite *SchedulerHandlerTestSuite) TestJoinStringsHelper() {
	testCases := []struct {
		name      string
		strings   []string
		separator string
		expected  string
	}{
		{
			name:      "join multiple strings with comma",
			strings:   []string{"job1", "job2", "job3"},
			separator: ", ",
			expected:  "job1, job2, job3",
		},
		{
			name:      "join single string",
			strings:   []string{"job1"},
			separator: ", ",
			expected:  "job1",
		},
		{
			name:      "join empty slice",
			strings:   []string{},
			separator: ", ",
			expected:  "",
		},
		{
			name:      "join with different separator",
			strings:   []string{"job1", "job2"},
			separator: " | ",
			expected:  "job1 | job2",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := joinStrings(tc.strings, tc.separator)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func (suite *SchedulerHandlerTestSuite) TestTriggerJobWithNilJobStatus() {
	jobName := "test_job"

	suite.mockService.On("GetJobStatus").Return(nil)

	c, rec := suite.createEchoContextWithParam(http.MethodPost, "/scheduler/jobs/"+jobName+"/trigger", "name", jobName)

	err := suite.handler.TriggerJob(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "Job not found")
}

func (suite *SchedulerHandlerTestSuite) TestGetStatusWithNilJobStatus() {
	suite.mockService.On("IsRunning").Return(true)
	suite.mockService.On("GetJobStatus").Return(nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/status", nil)

	err := suite.handler.GetStatus(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	dataBytes, _ := json.Marshal(response.Data)
	var statusData model.SchedulerStatusResponse
	err = json.Unmarshal(dataBytes, &statusData)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), statusData.SchedulerRunning)
	assert.Equal(suite.T(), 0, statusData.JobsCount)
	assert.Nil(suite.T(), statusData.Jobs)
}

func (suite *SchedulerHandlerTestSuite) TestGetJobsWithNilJobStatus() {
	suite.mockService.On("GetJobStatus").Return(nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/scheduler/jobs", nil)

	err := suite.handler.GetJobs(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	dataBytes, _ := json.Marshal(response.Data)
	var jobsData model.JobsResponse
	err = json.Unmarshal(dataBytes, &jobsData)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, jobsData.Count)
	assert.Nil(suite.T(), jobsData.Jobs)
}

// Run the test suite
func TestSchedulerHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerHandlerTestSuite))
}
