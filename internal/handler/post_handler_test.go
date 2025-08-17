package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

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

// MockValidator is a mock implementation for Echo's validator
type MockValidator struct{}

func (cv *MockValidator) Validate(i any) error {
	return nil
}

// PostHandlerTestSuite defines the test suite for PostHandler
type PostHandlerTestSuite struct {
	suite.Suite
	mockService *MockPostService
	logger      *logger.Logger
	handler     PostHandler
	echo        *echo.Echo
}

// SetupTest prepares each test
func (suite *PostHandlerTestSuite) SetupTest() {
	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}

	suite.mockService = new(MockPostService)
	suite.logger = logger.New(cfg)
	suite.handler = NewPostHandler(suite.mockService, suite.logger)
	suite.echo = echo.New()
	suite.echo.Validator = &MockValidator{}
}

// Helper functions
func (suite *PostHandlerTestSuite) createMockPost() *model.Post {
	now := time.Now()
	description := "Test description"
	content := "Test content"
	category := "technology"
	imageURL := "https://example.com/image.jpg"
	publishedAt := time.Now().UTC().Add(-24 * time.Hour)

	return &model.Post{
		ID:          1,
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

func (suite *PostHandlerTestSuite) createMockCreateParams() *model.CreatePostParams {
	description := "Test description"
	content := "Test content"
	category := "technology"
	imageURL := "https://example.com/image.jpg"
	publishedAt := time.Now().UTC().Add(-24 * time.Hour)

	return &model.CreatePostParams{
		Title:       "Test Post",
		Description: &description,
		Content:     &content,
		URL:         "https://example.com/test",
		Source:      "Test Source",
		Category:    &category,
		ImageURL:    &imageURL,
		PublishedAt: &publishedAt,
	}
}

func (suite *PostHandlerTestSuite) createMockUpdateParams() *model.UpdatePostParams {
	description := "Updated description"
	content := "Updated content"
	category := "business"
	imageURL := "https://example.com/updated.jpg"

	return &model.UpdatePostParams{
		Title:       "Updated Post",
		Description: &description,
		Content:     &content,
		Category:    &category,
		ImageURL:    &imageURL,
	}
}

func (suite *PostHandlerTestSuite) createMockPostListResponse(posts []model.Post, total int64) *model.PostListResponse {
	return &model.PostListResponse{
		Posts: posts,
		Pagination: model.PaginationMeta{
			Page:       1,
			Limit:      10,
			Total:      total,
			TotalPages: int(total+9) / 10,
		},
	}
}

func (suite *PostHandlerTestSuite) createEchoContext(method, target string, body any) (echo.Context, *httptest.ResponseRecorder) {
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

func (suite *PostHandlerTestSuite) TestCreatePostSuccess() {
	req := suite.createMockCreateParams()
	expectedPost := suite.createMockPost()

	suite.mockService.On("CreatePost", mock.Anything, req).Return(expectedPost, nil)

	c, rec := suite.createEchoContext(http.MethodPost, "/posts", req)

	err := suite.handler.CreatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Post created successfully", response.Message)
}

func (suite *PostHandlerTestSuite) TestCreatePostInvalidJSON() {
	c, rec := suite.createEchoContext(http.MethodPost, "/posts", "invalid-json")

	err := suite.handler.CreatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestCreatePostConflict() {
	req := suite.createMockCreateParams()

	suite.mockService.On("CreatePost", mock.Anything, req).Return(nil, service.ErrPostExists)

	c, rec := suite.createEchoContext(http.MethodPost, "/posts", req)

	err := suite.handler.CreatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusConflict, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestCreatePostInternalError() {
	req := suite.createMockCreateParams()

	suite.mockService.On("CreatePost", mock.Anything, req).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodPost, "/posts", req)

	err := suite.handler.CreatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostByIDSuccess() {
	expectedPost := suite.createMockPost()

	suite.mockService.On("GetPostByID", mock.Anything, int64(1)).Return(expectedPost, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/1", nil)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.GetPostByID(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostByIDInvalidID() {
	c, rec := suite.createEchoContext(http.MethodGet, "/posts/invalid", nil)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := suite.handler.GetPostByID(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostByIDZeroID() {
	c, rec := suite.createEchoContext(http.MethodGet, "/posts/0", nil)
	c.SetParamNames("id")
	c.SetParamValues("0")

	err := suite.handler.GetPostByID(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)
}

func (suite *PostHandlerTestSuite) TestGetPostByIDNotFound() {
	suite.mockService.On("GetPostByID", mock.Anything, int64(999)).Return(nil, service.ErrPostNotFound)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/999", nil)
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := suite.handler.GetPostByID(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostByIDInternalError() {
	suite.mockService.On("GetPostByID", mock.Anything, int64(1)).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/1", nil)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.GetPostByID(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestListPostsSuccess() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.AnythingOfType("*model.PostListParams")).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts", nil)

	err := suite.handler.ListPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestListPostsWithPagination() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Page == 2 && req.Limit == 5
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts?page=2&limit=5", nil)

	err := suite.handler.ListPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestListPostsWithFilters() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Category != nil && *req.Category == "technology" &&
			req.Source != nil && *req.Source == "test-source" &&
			req.Search != nil && *req.Search == "test-query"
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts?category=technology&source=test-source&search=test-query", nil)

	err := suite.handler.ListPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestListPostsInternalError() {
	suite.mockService.On("ListPosts", mock.Anything, mock.AnythingOfType("*model.PostListParams")).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodGet, "/posts", nil)

	err := suite.handler.ListPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestUpdatePostSuccess() {
	req := suite.createMockUpdateParams()
	expectedPost := suite.createMockPost()
	expectedPost.Title = req.Title

	suite.mockService.On("UpdatePost", mock.Anything, int64(1), req).Return(expectedPost, nil)

	c, rec := suite.createEchoContext(http.MethodPut, "/posts/1", req)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.UpdatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Post updated successfully", response.Message)
}

func (suite *PostHandlerTestSuite) TestUpdatePostInvalidID() {
	req := suite.createMockUpdateParams()

	c, rec := suite.createEchoContext(http.MethodPut, "/posts/invalid", req)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := suite.handler.UpdatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestUpdatePostInvalidJSON() {
	c, rec := suite.createEchoContext(http.MethodPut, "/posts/1", "invalid-json")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.UpdatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestUpdatePostNotFound() {
	req := suite.createMockUpdateParams()

	suite.mockService.On("UpdatePost", mock.Anything, int64(999), req).Return(nil, service.ErrPostNotFound)

	c, rec := suite.createEchoContext(http.MethodPut, "/posts/999", req)
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := suite.handler.UpdatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestUpdatePostInternalError() {
	req := suite.createMockUpdateParams()

	suite.mockService.On("UpdatePost", mock.Anything, int64(1), req).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodPut, "/posts/1", req)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.UpdatePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestDeletePostSuccess() {
	suite.mockService.On("DeletePost", mock.Anything, int64(1)).Return(nil)

	c, rec := suite.createEchoContext(http.MethodDelete, "/posts/1", nil)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.DeletePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Post deleted successfully", response.Message)
}

func (suite *PostHandlerTestSuite) TestDeletePostInvalidID() {
	c, rec := suite.createEchoContext(http.MethodDelete, "/posts/invalid", nil)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := suite.handler.DeletePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestDeletePostZeroID() {
	c, rec := suite.createEchoContext(http.MethodDelete, "/posts/0", nil)
	c.SetParamNames("id")
	c.SetParamValues("0")

	err := suite.handler.DeletePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)
}

func (suite *PostHandlerTestSuite) TestDeletePostNotFound() {
	suite.mockService.On("DeletePost", mock.Anything, int64(999)).Return(service.ErrPostNotFound)

	c, rec := suite.createEchoContext(http.MethodDelete, "/posts/999", nil)
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := suite.handler.DeletePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestDeletePostInternalError() {
	suite.mockService.On("DeletePost", mock.Anything, int64(1)).Return(errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodDelete, "/posts/1", nil)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := suite.handler.DeletePost(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostsByCategorySuccess() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Category != nil && *req.Category == "technology"
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/category/technology", nil)
	c.SetParamNames("category")
	c.SetParamValues("technology")

	err := suite.handler.GetPostsByCategory(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostsByCategoryEmptyCategory() {
	c, rec := suite.createEchoContext(http.MethodGet, "/posts/category/", nil)
	c.SetParamNames("category")
	c.SetParamValues("")

	err := suite.handler.GetPostsByCategory(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostsByCategoryWithPagination() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Category != nil && *req.Category == "technology" &&
			req.Page == 2 && req.Limit == 5
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/category/technology?page=2&limit=5", nil)
	c.SetParamNames("category")
	c.SetParamValues("technology")

	err := suite.handler.GetPostsByCategory(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestGetPostsByCategoryInternalError() {
	suite.mockService.On("ListPosts", mock.Anything, mock.AnythingOfType("*model.PostListParams")).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/category/technology", nil)
	c.SetParamNames("category")
	c.SetParamValues("technology")

	err := suite.handler.GetPostsByCategory(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostsBySourceSuccess() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Source != nil && *req.Source == "test-source"
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/source/test-source", nil)
	c.SetParamNames("source")
	c.SetParamValues("test-source")

	err := suite.handler.GetPostsBySource(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostsBySourceEmptySource() {
	c, rec := suite.createEchoContext(http.MethodGet, "/posts/source/", nil)
	c.SetParamNames("source")
	c.SetParamValues("")

	err := suite.handler.GetPostsBySource(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestGetPostsBySourceWithPagination() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Source != nil && *req.Source == "test-source" &&
			req.Page == 3 && req.Limit == 15
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/source/test-source?page=3&limit=15", nil)
	c.SetParamNames("source")
	c.SetParamValues("test-source")

	err := suite.handler.GetPostsBySource(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestGetPostsBySourceInternalError() {
	suite.mockService.On("ListPosts", mock.Anything, mock.AnythingOfType("*model.PostListParams")).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/source/test-source", nil)
	c.SetParamNames("source")
	c.SetParamValues("test-source")

	err := suite.handler.GetPostsBySource(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestSearchPostsSuccess() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Search != nil && *req.Search == "test query"
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/search?q=test%20query", nil)

	err := suite.handler.SearchPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestSearchPostsEmptyQuery() {
	c, rec := suite.createEchoContext(http.MethodGet, "/posts/search", nil)

	err := suite.handler.SearchPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestSearchPostsWithFiltersAndPagination() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Search != nil && *req.Search == "golang" &&
			req.Category != nil && *req.Category == "technology" &&
			req.Source != nil && *req.Source == "tech-news" &&
			req.Page == 2 && req.Limit == 20
	})).Return(mockResponse, nil)

	queryParams := url.Values{}
	queryParams.Set("q", "golang")
	queryParams.Set("category", "technology")
	queryParams.Set("source", "tech-news")
	queryParams.Set("page", "2")
	queryParams.Set("limit", "20")

	c, rec := suite.createEchoContext(http.MethodGet, fmt.Sprintf("/posts/search?%s", queryParams.Encode()), nil)

	err := suite.handler.SearchPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestSearchPostsInternalError() {
	suite.mockService.On("ListPosts", mock.Anything, mock.AnythingOfType("*model.PostListParams")).Return(nil, errors.New("database error"))

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/search?q=test", nil)

	err := suite.handler.SearchPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, rec.Code)

	var response response.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
}

func (suite *PostHandlerTestSuite) TestListPostsIgnoreInvalidPaginationParams() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Page == 1 && req.Limit == 20
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts?page=invalid&limit=-5", nil)

	err := suite.handler.ListPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestGetPostsByCategoryIgnoreInvalidLimit() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Limit == 20
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/category/tech?limit=150", nil)
	c.SetParamNames("category")
	c.SetParamValues("tech")

	err := suite.handler.GetPostsByCategory(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestSearchPostsSpecialCharacters() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	searchQuery := "test & special characters"
	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Search != nil && *req.Search == searchQuery
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/search?q="+url.QueryEscape(searchQuery), nil)

	err := suite.handler.SearchPosts(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestMultipleConcurrentRequests() {
	expectedPost := suite.createMockPost()

	suite.mockService.On("GetPostByID", mock.Anything, int64(1)).Return(expectedPost, nil).Times(3)

	for i := 0; i < 3; i++ {
		c, rec := suite.createEchoContext(http.MethodGet, "/posts/1", nil)
		c.SetParamNames("id")
		c.SetParamValues("1")

		err := suite.handler.GetPostByID(c)

		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}
}

// Test URL encoding/decoding
func (suite *PostHandlerTestSuite) TestGetPostsByCategoryWithSpecialCharacters() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	categoryWithSpaces := "tech & science"
	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Category != nil && *req.Category == categoryWithSpaces
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/category/"+url.PathEscape(categoryWithSpaces), nil)
	c.SetParamNames("category")
	c.SetParamValues(categoryWithSpaces)

	err := suite.handler.GetPostsByCategory(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

func (suite *PostHandlerTestSuite) TestGetPostsBySourceWithSpecialCharacters() {
	posts := []model.Post{*suite.createMockPost()}
	mockResponse := suite.createMockPostListResponse(posts, 1)

	sourceWithSpaces := "BBC News & Reports"
	suite.mockService.On("ListPosts", mock.Anything, mock.MatchedBy(func(req *model.PostListParams) bool {
		return req.Source != nil && *req.Source == sourceWithSpaces
	})).Return(mockResponse, nil)

	c, rec := suite.createEchoContext(http.MethodGet, "/posts/source/"+url.PathEscape(sourceWithSpaces), nil)
	c.SetParamNames("source")
	c.SetParamValues(sourceWithSpaces)

	err := suite.handler.GetPostsBySource(c)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

// Run the test suite
func TestPostHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PostHandlerTestSuite))
}
