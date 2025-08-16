package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// NewsServiceTestSuite defines the test suite for NewsService
type NewsServiceTestSuite struct {
	suite.Suite
	service    service.NewsService
	logger     *logger.Logger
	httpServer *httptest.Server
	ctx        context.Context
}

func (suite *NewsServiceTestSuite) SetupTest() {
	suite.ctx = context.Background()

	suite.httpServer = httptest.NewServer(http.HandlerFunc(suite.mockNewsAPIHandler))

	cfg := &config.Config{
		App: config.AppConfig{LogLevel: "debug"},
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: suite.httpServer.URL,
		},
	}

	suite.logger = logger.New(cfg)

	suite.service = service.NewNewsService(cfg, suite.logger)
}

func (suite *NewsServiceTestSuite) TearDownTest() {
	if suite.httpServer != nil {
		suite.httpServer.Close()
	}
}

// mockNewsAPIHandler handles HTTP requests for testing
func (suite *NewsServiceTestSuite) mockNewsAPIHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apiKey")
	if apiKey != "test-api-key" {
		suite.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
		return
	}

	path := r.URL.Path

	switch {
	case strings.Contains(path, "/error/rate-limit"):
		w.WriteHeader(http.StatusTooManyRequests)
		suite.writeErrorResponse(w, http.StatusTooManyRequests, "rateLimited", "Rate limit exceeded")
		return
	case strings.Contains(path, "/error/server"):
		w.WriteHeader(http.StatusInternalServerError)
		suite.writeErrorResponse(w, http.StatusInternalServerError, "serverError", "Internal server error")
		return
	case strings.Contains(path, "/error/bad-request"):
		w.WriteHeader(http.StatusBadRequest)
		suite.writeErrorResponse(w, http.StatusBadRequest, "badRequest", "Bad request parameters")
		return
	case strings.Contains(path, "/error/invalid-json"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
		return
	case strings.Contains(path, "/error/api-error-status"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		suite.writeNewsAPIResponse(w, &model.NewsAPIResponse{Status: "error"})
		return

	case strings.Contains(path, "/top-headlines"):
		suite.handleTopHeadlines(w, r)
		return
	case strings.Contains(path, "/everything"):
		suite.handleEverything(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (suite *NewsServiceTestSuite) handleTopHeadlines(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	articles := suite.createMockArticles()

	category := query.Get("category")
	if category == "technology" {
		articles = []model.NewsAPIArticleParams{articles[0]}
	}

	response := &model.NewsAPIResponse{
		Status:       "ok",
		TotalResults: len(articles),
		Articles:     articles,
	}

	suite.writeNewsAPIResponse(w, response)
}

func (suite *NewsServiceTestSuite) handleEverything(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	articles := suite.createMockArticles()

	sources := query.Get("sources")
	if sources == "techcrunch" {
		articles = []model.NewsAPIArticleParams{articles[0]} // Return only TechCrunch article
	}

	response := &model.NewsAPIResponse{
		Status:       "ok",
		TotalResults: len(articles),
		Articles:     articles,
	}

	suite.writeNewsAPIResponse(w, response)
}

func (suite *NewsServiceTestSuite) writeNewsAPIResponse(w http.ResponseWriter, response *model.NewsAPIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (suite *NewsServiceTestSuite) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	errorResp := map[string]any{
		"status":     "error",
		"statusCode": statusCode,
		"code":       code,
		"message":    message,
	}
	json.NewEncoder(w).Encode(errorResp)
}

func (suite *NewsServiceTestSuite) createMockArticles() []model.NewsAPIArticleParams {
	publishedAt := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

	return []model.NewsAPIArticleParams{
		{
			Source: struct {
				ID   *string `json:"id" example:"techcrunch"`
				Name string  `json:"name" example:"TechCrunch"`
			}{
				ID:   stringPtr("techcrunch"),
				Name: "TechCrunch",
			},
			Author:      stringPtr("John Doe"),
			Title:       "New AI Breakthrough Announced",
			Description: stringPtr("A groundbreaking AI development has been announced"),
			URL:         "https://techcrunch.com/ai-breakthrough",
			URLToImage:  stringPtr("https://techcrunch.com/image.jpg"),
			PublishedAt: publishedAt,
			Content:     stringPtr("Full article content about AI breakthrough..."),
		},
		{
			Source: struct {
				ID   *string `json:"id" example:"techcrunch"`
				Name string  `json:"name" example:"TechCrunch"`
			}{
				ID:   stringPtr("bbc-news"),
				Name: "BBC News",
			},
			Author:      stringPtr("Jane Smith"),
			Title:       "Global Economic Update",
			Description: stringPtr("Latest updates on global economy"),
			URL:         "https://bbc.com/economy-update",
			URLToImage:  stringPtr("https://bbc.com/image.jpg"),
			PublishedAt: publishedAt,
			Content:     stringPtr("Full article content about economy..."),
		},
	}
}

func (suite *NewsServiceTestSuite) TestGetTopHeadlines_Success() {
	req := &model.NewsParams{
		Query:    "AI",
		Category: "technology",
		Country:  "us",
		Language: "en",
		PageSize: 10,
		Page:     1,
	}

	result, err := suite.service.GetTopHeadlines(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
	assert.Greater(suite.T(), result.TotalResults, 0)
	assert.NotEmpty(suite.T(), result.Articles)
}

func (suite *NewsServiceTestSuite) TestGetTopHeadlines_EmptyParams() {
	req := &model.NewsParams{}

	result, err := suite.service.GetTopHeadlines(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetTopHeadlines_WithCategory() {
	req := &model.NewsParams{
		Category: "technology",
		PageSize: 5,
	}

	result, err := suite.service.GetTopHeadlines(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
	assert.Equal(suite.T(), 1, result.TotalResults)
}

func (suite *NewsServiceTestSuite) TestGetTopHeadlinesInvalidAPIKey() {
	cfg := &config.Config{
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "invalid-key",
			BaseURL: suite.httpServer.URL,
		},
	}
	invalidService := service.NewNewsService(cfg, suite.logger)

	req := &model.NewsParams{
		Query: "test",
	}

	result, err := invalidService.GetTopHeadlines(suite.ctx, req)
	require.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to get top headlines")

	result, err = suite.service.GetTopHeadlines(suite.ctx, req)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestGetTopHeadlinesContextCanceled() {
	canceledCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	req := &model.NewsParams{
		Query: "test",
	}

	result, err := suite.service.GetTopHeadlines(canceledCtx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get top headlines")
	assert.Nil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestGetEverythingSuccess() {
	req := &model.NewsParams{
		Query:    "technology",
		Sources:  []string{"techcrunch", "bbc-news"},
		Language: "en",
		PageSize: 20,
		Page:     1,
	}

	result, err := suite.service.GetEverything(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
	assert.Greater(suite.T(), result.TotalResults, 0)
	assert.NotEmpty(suite.T(), result.Articles)
}

func (suite *NewsServiceTestSuite) TestGetEverythingWithSources() {
	req := &model.NewsParams{
		Sources:  []string{"techcrunch"},
		Language: "en",
		PageSize: 10,
	}

	result, err := suite.service.GetEverything(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
	assert.Equal(suite.T(), 1, result.TotalResults) // Filtered to TechCrunch
}

func (suite *NewsServiceTestSuite) TestGetEverythingEmptyParams() {
	req := &model.NewsParams{}

	result, err := suite.service.GetEverything(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetEverythingMultipleSources() {
	req := &model.NewsParams{
		Sources:  []string{"techcrunch", "bbc-news", "reuters"},
		Language: "en",
		PageSize: 50,
	}

	result, err := suite.service.GetEverything(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsByCategorySuccess() {
	category := "technology"
	pageSize := 15

	result, err := suite.service.GetNewsByCategory(suite.ctx, category, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsByCategoryEmptyCategory() {
	category := ""
	pageSize := 10

	result, err := suite.service.GetNewsByCategory(suite.ctx, category, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsByCategoryZeroPageSize() {
	category := "business"
	pageSize := 0

	result, err := suite.service.GetNewsByCategory(suite.ctx, category, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsBySourcesSuccess() {
	sources := []string{"techcrunch", "bbc-news"}
	pageSize := 25

	result, err := suite.service.GetNewsBySources(suite.ctx, sources, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsBySourcesEmptySources() {
	sources := []string{}
	pageSize := 10

	result, err := suite.service.GetNewsBySources(suite.ctx, sources, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsBySourcesNilSources() {
	var sources []string = nil
	pageSize := 10

	result, err := suite.service.GetNewsBySources(suite.ctx, sources, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestGetNewsBySourcesSingleSource() {
	sources := []string{"techcrunch"}
	pageSize := 5

	result, err := suite.service.GetNewsBySources(suite.ctx, sources, pageSize)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestHTTPErrors_RateLimit() {
	cfg := &config.Config{
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: suite.httpServer.URL + "/error/rate-limit",
		},
	}
	errorService := service.NewNewsService(cfg, suite.logger)

	req := &model.NewsParams{Query: "test"}

	result, err := errorService.GetTopHeadlines(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "rate limit exceeded")
	assert.Nil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestHTTPErrors_ServerError() {
	cfg := &config.Config{
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: suite.httpServer.URL + "/error/server",
		},
	}
	errorService := service.NewNewsService(cfg, suite.logger)

	req := &model.NewsParams{Query: "test"}

	result, err := errorService.GetTopHeadlines(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "NewsAPI server error")
	assert.Nil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestHTTPErrors_BadRequest() {
	cfg := &config.Config{
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: suite.httpServer.URL + "/error/bad-request",
		},
	}
	errorService := service.NewNewsService(cfg, suite.logger)

	req := &model.NewsParams{Query: "test"}

	result, err := errorService.GetTopHeadlines(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "API error")
	assert.Nil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestHTTPErrors_InvalidJSON() {
	cfg := &config.Config{
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: suite.httpServer.URL + "/error/invalid-json",
		},
	}
	errorService := service.NewNewsService(cfg, suite.logger)

	req := &model.NewsParams{Query: "test"}

	result, err := errorService.GetTopHeadlines(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to parse JSON response")
	assert.Nil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestHTTPErrors_APIErrorStatus() {
	cfg := &config.Config{
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: suite.httpServer.URL + "/error/api-error-status",
		},
	}
	errorService := service.NewNewsService(cfg, suite.logger)

	req := &model.NewsParams{Query: "test"}

	result, err := errorService.GetTopHeadlines(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "API returened error status")
	assert.Nil(suite.T(), result)
}

func (suite *NewsServiceTestSuite) TestGetDefaultSources() {
	sources := service.GetDefaultSources()

	assert.NotEmpty(suite.T(), sources)
	assert.Contains(suite.T(), sources, "bbc-news")
	assert.Contains(suite.T(), sources, "techcrunch")
	assert.Contains(suite.T(), sources, "reuters")
	assert.Contains(suite.T(), sources, "cnn")
	assert.Equal(suite.T(), 10, len(sources))
}

func (suite *NewsServiceTestSuite) TestGetDefaultCategories() {
	categories := service.GetDefaultCategories()

	assert.NotEmpty(suite.T(), categories)
	assert.Contains(suite.T(), categories, "general")
	assert.Contains(suite.T(), categories, "technology")
	assert.Contains(suite.T(), categories, "business")
	assert.Contains(suite.T(), categories, "science")
	assert.Equal(suite.T(), 7, len(categories))
}

func (suite *NewsServiceTestSuite) TestParameterEncoding() {
	req := &model.NewsParams{
		Query:    "test query with spaces",
		Category: "technology",
		Country:  "us",
		Language: "en",
		PageSize: 10,
		Page:     2,
	}

	result, err := suite.service.GetTopHeadlines(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestSourcesParameterEncoding() {
	req := &model.NewsParams{
		Query:    "technology",
		Sources:  []string{"techcrunch", "bbc-news", "reuters"},
		Language: "en",
		PageSize: 15,
		Page:     1,
	}

	result, err := suite.service.GetEverything(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestEdgeCases_NegativePageSize() {
	req := &model.NewsParams{
		Query:    "test",
		PageSize: -5,
	}

	result, err := suite.service.GetTopHeadlines(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

func (suite *NewsServiceTestSuite) TestEdgeCases_NegativePage() {
	req := &model.NewsParams{
		Query: "test",
		Page:  -1,
	}

	result, err := suite.service.GetTopHeadlines(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "ok", result.Status)
}

// Run the test suite
func TestNewsServiceSuite(t *testing.T) {
	suite.Run(t, new(NewsServiceTestSuite))
}
