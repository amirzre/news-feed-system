package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockNewsService is a mock implementation of NewsService
type MockNewsService struct {
	mock.Mock
}

func (m *MockNewsService) GetTopHeadlines(ctx context.Context, req *model.NewsParams) (*model.NewsAPIResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NewsAPIResponse), args.Error(1)
}

func (m *MockNewsService) GetEverything(ctx context.Context, req *model.NewsParams) (*model.NewsAPIResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NewsAPIResponse), args.Error(1)
}

func (m *MockNewsService) GetNewsByCategory(ctx context.Context, category string, pageSize int) (*model.NewsAPIResponse, error) {
	args := m.Called(ctx, category, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NewsAPIResponse), args.Error(1)
}

func (m *MockNewsService) GetNewsBySources(ctx context.Context, sources []string, pageSize int) (*model.NewsAPIResponse, error) {
	args := m.Called(ctx, sources, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.NewsAPIResponse), args.Error(1)
}

// MockPostService is a mock implementation of PostService
type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostService) PostExists(ctx context.Context, url string) (bool, error) {
	args := m.Called(ctx, url)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostService) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostService) ListPosts(ctx context.Context, req *model.PostListParams) (*model.PostListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PostListResponse), args.Error(1)
}

func (m *MockPostService) UpdatePost(ctx context.Context, id int64, req *model.UpdatePostParams) (*model.Post, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostService) DeletePost(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostService) CreatePostFromNewsAPI(ctx context.Context, article *model.NewsAPIArticleParams) (*model.Post, error) {
	args := m.Called(ctx, article)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

// AggregatorServiceTestSuite defines the test suite for AggregatorService
type AggregatorServiceTestSuite struct {
	suite.Suite
	mockNewsService *MockNewsService
	mockPostService *MockPostService
	logger          *logger.Logger
	service         AggregatorService
	ctx             context.Context
}

func (suite *AggregatorServiceTestSuite) SetupTest() {
	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}

	suite.mockNewsService = new(MockNewsService)
	suite.mockPostService = new(MockPostService)
	suite.logger = logger.New(cfg)
	suite.service = NewAggregatorService(suite.mockNewsService, suite.mockPostService, suite.logger)
	suite.ctx = context.Background()
}

func (suite *AggregatorServiceTestSuite) TearDownTest() {
	suite.mockNewsService.AssertExpectations(suite.T())
	suite.mockPostService.AssertExpectations(suite.T())
}

func (suite *AggregatorServiceTestSuite) createMockNewsAPIResponse(articleCount int) *model.NewsAPIResponse {
	articles := make([]model.NewsAPIArticleParams, articleCount)
	for i := 0; i < articleCount; i++ {
		articles[i] = model.NewsAPIArticleParams{
			Source: struct {
				ID   *string `json:"id" example:"techcrunch"`
				Name string  `json:"name" example:"TechCrunch"`
			}{
				Name: "TechCrunch",
			},
			Title:       "Test Article " + string(rune(i+1)),
			Description: stringPtr("Test description"),
			URL:         "https://example.com/article-" + string(rune(i+1)),
			PublishedAt: time.Now().Format(time.RFC3339),
			Content:     stringPtr("Test content"),
		}
	}

	return &model.NewsAPIResponse{
		Status:       "ok",
		TotalResults: articleCount,
		Articles:     articles,
	}
}

func stringPtr(s string) *string {
	return &s
}

func (suite *AggregatorServiceTestSuite) createMockPost(id int64) *model.Post {
	now := time.Now()
	description := "Test description"
	content := "Test content"
	category := "technology"
	imageURL := "https://example.com/image.jpg"
	publishedAt := time.Now().Add(-24 * time.Hour)

	return &model.Post{
		ID:          id,
		Title:       "Test Post",
		Description: &description,
		Content:     &content,
		URL:         "https://example.com/test",
		Source:      "Test Source",
		Category:    &category,
		ImageURL:    &imageURL,
		PublishedAt: &publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (suite *AggregatorServiceTestSuite) TestAggregateTopHeadlinesSuccess() {
	categories := GetDefaultCategories()
	for _, category := range categories {
		mockResponse := suite.createMockNewsAPIResponse(3)
		suite.mockNewsService.On("GetNewsByCategory", suite.ctx, category, 50).Return(mockResponse, nil)

		for _, article := range mockResponse.Articles {
			mockPost := suite.createMockPost(1)
			suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &article).Return(mockPost, nil)
		}
	}

	result, err := suite.service.AggregateTopHeadlines(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Greater(suite.T(), result.TotalFetched, 0)
	assert.Greater(suite.T(), result.TotalCreated, 0)
	assert.Equal(suite.T(), 0, result.TotalErrors)
	assert.NotZero(suite.T(), result.Duration)
	assert.Len(suite.T(), result.Categories, len(categories))
}

func (suite *AggregatorServiceTestSuite) TestAggregateTopHeadlinesWithDuplicates() {
	categories := []string{"technology"}

	mockResponse := suite.createMockNewsAPIResponse(2)
	suite.mockNewsService.On("GetNewsByCategory", suite.ctx, "technology", 50).Return(mockResponse, nil)

	mockPost := suite.createMockPost(1)
	suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &mockResponse.Articles[0]).Return(mockPost, nil)
	suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &mockResponse.Articles[1]).Return(nil, ErrPostExists)

	service := &aggregatorService{
		newsService: suite.mockNewsService,
		postService: suite.mockPostService,
		logger:      suite.logger,
		maxWorkers:  5,
	}

	result := service.aggregateByCategories(suite.ctx, categories, true)

	assert.Equal(suite.T(), 2, result.TotalFetched)
	assert.Equal(suite.T(), 1, result.TotalCreated)
	assert.Equal(suite.T(), 1, result.TotalDuplicates)
	assert.Equal(suite.T(), 0, result.TotalErrors)
}

func (suite *AggregatorServiceTestSuite) TestAggregateTopHeadlinesWithErrors() {
	categories := []string{"technology"}

	suite.mockNewsService.On("GetNewsByCategory", suite.ctx, "technology", 50).Return(nil, errors.New("API error"))

	service := &aggregatorService{
		newsService: suite.mockNewsService,
		postService: suite.mockPostService,
		logger:      suite.logger,
		maxWorkers:  5,
	}

	result := service.aggregateByCategories(suite.ctx, categories, true)

	assert.Equal(suite.T(), 0, result.TotalFetched)
	assert.Equal(suite.T(), 0, result.TotalCreated)
	assert.Equal(suite.T(), 0, result.TotalDuplicates)
	assert.Equal(suite.T(), 1, result.TotalErrors)
}

func (suite *AggregatorServiceTestSuite) TestAggregateByCategoriesSuccess() {
	categories := []string{"technology", "business"}

	for _, category := range categories {
		mockResponse := suite.createMockNewsAPIResponse(2)
		suite.mockNewsService.On("GetNewsByCategory", suite.ctx, category, 50).Return(mockResponse, nil)

		for _, article := range mockResponse.Articles {
			mockPost := suite.createMockPost(1)
			suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &article).Return(mockPost, nil)
		}
	}

	result, err := suite.service.AggregateByCategories(suite.ctx, categories)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 4, result.TotalFetched)
	assert.Equal(suite.T(), 4, result.TotalCreated)
	assert.Equal(suite.T(), 0, result.TotalErrors)
	assert.Len(suite.T(), result.Categories, 2)
}

func (suite *AggregatorServiceTestSuite) TestAggregateByCategoriesEmptyList() {
	categories := []string{}

	result, err := suite.service.AggregateByCategories(suite.ctx, categories)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 0, result.TotalFetched)
	assert.Equal(suite.T(), 0, result.TotalCreated)
	assert.Equal(suite.T(), 0, result.TotalErrors)
	assert.Len(suite.T(), result.Categories, 0)
}

func (suite *AggregatorServiceTestSuite) TestAggregateBySourcesSuccess() {
	sources := []string{"techcrunch", "bbc-news"}

	mockResponse := suite.createMockNewsAPIResponse(3)
	suite.mockNewsService.On("GetNewsBySources", suite.ctx, []string{"techcrunch", "bbc-news"}, 100).Return(mockResponse, nil)

	for _, article := range mockResponse.Articles {
		mockPost := suite.createMockPost(1)
		suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &article).Return(mockPost, nil)
	}

	result, err := suite.service.AggregateBySources(suite.ctx, sources)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 3, result.TotalFetched)
	assert.Equal(suite.T(), 3, result.TotalCreated)
	assert.Equal(suite.T(), 0, result.TotalErrors)
	assert.Len(suite.T(), result.Sources, 2)
}

func (suite *AggregatorServiceTestSuite) TestAggregateBySourcesWithNewsServiceError() {
	sources := []string{"techcrunch"}

	suite.mockNewsService.On("GetNewsBySources", suite.ctx, sources, 100).Return(nil, errors.New("API error"))

	result, err := suite.service.AggregateBySources(suite.ctx, sources)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 0, result.TotalFetched)
	assert.Equal(suite.T(), 0, result.TotalCreated)
	assert.Equal(suite.T(), 1, result.TotalErrors)
	assert.NotEmpty(suite.T(), result.Errors)
}

func (suite *AggregatorServiceTestSuite) TestAggregateBySourcesWithPostServiceError() {
	sources := []string{"techcrunch"}

	mockResponse := suite.createMockNewsAPIResponse(2)
	suite.mockNewsService.On("GetNewsBySources", suite.ctx, sources, 100).Return(mockResponse, nil)

	mockPost := suite.createMockPost(1)
	suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &mockResponse.Articles[0]).Return(mockPost, nil)
	suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &mockResponse.Articles[1]).Return(nil, errors.New("database error"))

	result, err := suite.service.AggregateBySources(suite.ctx, sources)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 2, result.TotalFetched)
	assert.Equal(suite.T(), 1, result.TotalCreated)
	assert.Equal(suite.T(), 1, result.TotalErrors)
}

func (suite *AggregatorServiceTestSuite) TestAggregateBySourcesWithDuplicateByErrorMessage() {
	sources := []string{"techcrunch"}

	mockResponse := suite.createMockNewsAPIResponse(1)
	suite.mockNewsService.On("GetNewsBySources", suite.ctx, sources, 100).Return(mockResponse, nil)

	suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &mockResponse.Articles[0]).Return(nil, errors.New("post with this URL already exists"))

	result, err := suite.service.AggregateBySources(suite.ctx, sources)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 1, result.TotalFetched)
	assert.Equal(suite.T(), 0, result.TotalCreated)
	assert.Equal(suite.T(), 1, result.TotalDuplicates)
}

func (suite *AggregatorServiceTestSuite) TestAggregateAllSuccess() {
	categories := GetDefaultCategories()
	for _, category := range categories {
		mockResponse := suite.createMockNewsAPIResponse(1)
		suite.mockNewsService.On("GetNewsByCategory", suite.ctx, category, 50).Return(mockResponse, nil)

		for _, article := range mockResponse.Articles {
			mockPost := suite.createMockPost(1)
			suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &article).Return(mockPost, nil)
		}
	}

	sources := GetDefaultSources()
	for i := 0; i < len(sources); i += 3 {
		end := i + 3
		if end > len(sources) {
			end = len(sources)
		}
		batch := sources[i:end]

		mockResponse := suite.createMockNewsAPIResponse(1)
		suite.mockNewsService.On("GetNewsBySources", suite.ctx, batch, 100).Return(mockResponse, nil)

		for _, article := range mockResponse.Articles {
			mockPost := suite.createMockPost(1)
			suite.mockPostService.On("CreatePostFromNewsAPI", suite.ctx, &article).Return(mockPost, nil)
		}
	}

	result, err := suite.service.AggregateAll(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Greater(suite.T(), result.TotalFetched, 0)
	assert.Greater(suite.T(), result.TotalCreated, 0)
	assert.NotZero(suite.T(), result.Duration)
	assert.NotEmpty(suite.T(), result.Categories)
	assert.NotEmpty(suite.T(), result.Sources)
}

func (suite *AggregatorServiceTestSuite) TestNewAggregatorService() {
	service := NewAggregatorService(suite.mockNewsService, suite.mockPostService, suite.logger)

	assert.NotNil(suite.T(), service)

	aggregatorServiceImpl, ok := service.(*aggregatorService)
	assert.True(suite.T(), ok)

	assert.Equal(suite.T(), suite.mockNewsService, aggregatorServiceImpl.newsService)
	assert.Equal(suite.T(), suite.mockPostService, aggregatorServiceImpl.postService)
	assert.Equal(suite.T(), suite.logger, aggregatorServiceImpl.logger)
	assert.Equal(suite.T(), 5, aggregatorServiceImpl.maxWorkers)
}

func (suite *AggregatorServiceTestSuite) TestAggregateTopHeadlinesContextCanceled() {
	canceledCtx, cancel := context.WithCancel(suite.ctx)
	cancel()

	categories := []string{"technology"}
	service := &aggregatorService{
		newsService: suite.mockNewsService,
		postService: suite.mockPostService,
		logger:      suite.logger,
		maxWorkers:  5,
	}

	suite.mockNewsService.On("GetNewsByCategory", canceledCtx, "technology", 50).Return(nil, context.Canceled).Maybe()

	result := service.aggregateByCategories(canceledCtx, categories, true)

	assert.Equal(suite.T(), 0, result.TotalFetched)
	assert.Equal(suite.T(), 0, result.TotalCreated)
	assert.Greater(suite.T(), result.TotalErrors, 0)
}

// Run the test suite
func TestAggregatorServiceSuite(t *testing.T) {
	suite.Run(t, new(AggregatorServiceTestSuite))
}
