package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// MockAggregatorService is a mock implementation of AggregatorService
type MockAggregatorService struct {
	mock.Mock
}

func (m *MockAggregatorService) AggregateTopHeadlines(ctx context.Context) (*model.AggregationResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AggregationResponse), args.Error(1)
}

func (m *MockAggregatorService) AggregateByCategories(ctx context.Context, categories []string) (*model.AggregationResponse, error) {
	args := m.Called(ctx, categories)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AggregationResponse), args.Error(1)
}

func (m *MockAggregatorService) AggregateBySources(ctx context.Context, sources []string) (*model.AggregationResponse, error) {
	args := m.Called(ctx, sources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AggregationResponse), args.Error(1)
}

func (m *MockAggregatorService) AggregateAll(ctx context.Context) (*model.AggregationResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AggregationResponse), args.Error(1)
}

// AggregatorHandlerTestSuite defines the test suite for AggregatorHandler
type AggregatorHandlerTestSuite struct {
	suite.Suite
	mockService *MockAggregatorService
	logger      *logger.Logger
	handler     AggregatorHandler
	echo        *echo.Echo
}

func (suite *AggregatorHandlerTestSuite) SetupTest() {
	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}

	suite.mockService = new(MockAggregatorService)
	suite.logger = logger.New(cfg)
	suite.handler = NewAggregatorHandler(suite.mockService, suite.logger)
	suite.echo = echo.New()
	suite.echo.Validator = &MockValidator{}
}

func (suite *AggregatorHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// Helper functions
func (suite *AggregatorHandlerTestSuite) createMockAggregationResponse() *model.AggregationResponse {
	catStats := map[string]model.CategoryStats{
		"technology": {},
		"business":   {},
	}
	srcStats := map[string]model.SourceStats{
		"bbc-news":   {},
		"techcrunch": {},
	}

	return &model.AggregationResponse{
		TotalFetched:    100,
		TotalCreated:    85,
		TotalDuplicates: 15,
		TotalErrors:     0,
		Duration:        2 * time.Minute,
		Categories:      catStats,
		Sources:         srcStats,
		Errors:          []string{},
	}
}

func (suite *AggregatorHandlerTestSuite) createEchoContext(method, target string, body interface{}) (echo.Context, *httptest.ResponseRecorder) {
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

func (suite *AggregatorHandlerTestSuite) TestTriggerAggregationSuccess() {
	expectedResult := suite.createMockAggregationResponse()

	suite.mockService.On("AggregateAll", mock.AnythingOfType("*context.timerCtx")).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger", nil)

	err := suite.handler.TriggerAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Aggregation completed successfully", response.Message)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerAggregationError() {
	suite.mockService.On("AggregateAll", mock.AnythingOfType("*context.timerCtx")).Return(nil, errors.New("aggregation service error"))

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger", nil)

	err := suite.handler.TriggerAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerTopHeadlinesSuccess() {
	expectedResult := suite.createMockAggregationResponse()

	suite.mockService.On("AggregateTopHeadlines", mock.AnythingOfType("*context.timerCtx")).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/headlines", nil)

	err := suite.handler.TriggerTopHeadlines(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Top headlines aggregation completed successfully", response.Message)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerTopHeadlinesError() {
	suite.mockService.On("AggregateTopHeadlines", mock.AnythingOfType("*context.timerCtx")).Return(nil, errors.New("headlines service error"))

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/headlines", nil)

	err := suite.handler.TriggerTopHeadlines(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerCategoryAggregationWithValidCategories() {
	expectedResult := suite.createMockAggregationResponse()
	requestBody := model.CategoryAggregationRequest{
		Categories: []string{"technology", "business"},
	}

	suite.mockService.On("AggregateByCategories", mock.AnythingOfType("*context.timerCtx"), []string{"technology", "business"}).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/categories", requestBody)

	err := suite.handler.TriggerCategoryAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Category aggregation completed successfully", response.Message)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerCategoryAggregationWithInvalidJSON() {
	expectedResult := suite.createMockAggregationResponse()

	// When binding fails, should use default categories
	suite.mockService.On("AggregateByCategories", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("[]string")).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/categories", "invalid-json")

	err := suite.handler.TriggerCategoryAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerCategoryAggregationWithEmptyCategories() {
	expectedResult := suite.createMockAggregationResponse()
	requestBody := model.CategoryAggregationRequest{
		Categories: []string{},
	}

	// Should use default categories when empty
	suite.mockService.On("AggregateByCategories", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("[]string")).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/categories", requestBody)

	err := suite.handler.TriggerCategoryAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerCategoryAggregationWithInvalidCategories() {
	requestBody := model.CategoryAggregationRequest{
		Categories: []string{"invalid-category", "another-invalid"},
	}

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/categories", requestBody)

	err := suite.handler.TriggerCategoryAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerCategoryAggregationError() {
	requestBody := model.CategoryAggregationRequest{
		Categories: []string{"technology"},
	}

	suite.mockService.On("AggregateByCategories", mock.AnythingOfType("*context.timerCtx"), []string{"technology"}).Return(nil, errors.New("category service error"))

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/categories", requestBody)

	err := suite.handler.TriggerCategoryAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerSourceAggregationWithValidSources() {
	expectedResult := suite.createMockAggregationResponse()
	requestBody := model.SourceAggregationRequest{
		Sources: []string{"bbc-news", "techcrunch"},
	}

	suite.mockService.On("AggregateBySources", mock.AnythingOfType("*context.timerCtx"), []string{"bbc-news", "techcrunch"}).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/sources", requestBody)

	err := suite.handler.TriggerSourceAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Source aggregation completed successfully", response.Message)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerSourceAggregationWithInvalidJSON() {
	expectedResult := suite.createMockAggregationResponse()

	// When binding fails, should use default sources
	suite.mockService.On("AggregateBySources", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("[]string")).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/sources", "invalid-json")

	err := suite.handler.TriggerSourceAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerSourceAggregationWithEmptySources() {
	requestBody := model.SourceAggregationRequest{
		Sources: []string{"", "   ", ""},
	}

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/sources", requestBody)

	err := suite.handler.TriggerSourceAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AggregatorHandlerTestSuite) TestTriggerSourceAggregationError() {
	requestBody := model.SourceAggregationRequest{
		Sources: []string{"bbc-news"},
	}

	suite.mockService.On("AggregateBySources", mock.AnythingOfType("*context.timerCtx"), []string{"bbc-news"}).Return(nil, errors.New("source service error"))

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/sources", requestBody)

	err := suite.handler.TriggerSourceAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *AggregatorHandlerTestSuite) TestCategoryAggregationResponseFormat() {
	expectedResult := suite.createMockAggregationResponse()
	requestBody := model.CategoryAggregationRequest{
		Categories: []string{"technology"},
	}

	suite.mockService.On("AggregateByCategories", mock.AnythingOfType("*context.timerCtx"), []string{"technology"}).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/categories", requestBody)

	err := suite.handler.TriggerCategoryAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response struct {
		Success bool                              `json:"success"`
		Message string                            `json:"message"`
		Data    model.CategoryAggregationResponse `json:"data"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), []string{"technology"}, response.Data.Categories)
	assert.Equal(suite.T(), expectedResult.TotalFetched, response.Data.Result.TotalFetched)
}

func (suite *AggregatorHandlerTestSuite) TestSourceAggregationResponseFormat() {
	expectedResult := suite.createMockAggregationResponse()
	requestBody := model.SourceAggregationRequest{
		Sources: []string{"bbc-news"},
	}

	suite.mockService.On("AggregateBySources", mock.AnythingOfType("*context.timerCtx"), []string{"bbc-news"}).Return(expectedResult, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/aggregation/trigger/sources", requestBody)

	err := suite.handler.TriggerSourceAggregation(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response struct {
		Success bool                            `json:"success"`
		Message string                          `json:"message"`
		Data    model.SourceAggregationResponse `json:"data"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), []string{"bbc-news"}, response.Data.Sources)
	assert.Equal(suite.T(), expectedResult.TotalCreated, response.Data.Result.TotalCreated)
}

// Run the test suite
func TestAggregatorHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AggregatorHandlerTestSuite))
}
