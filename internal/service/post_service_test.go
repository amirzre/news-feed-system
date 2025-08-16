package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockPostRepository is a mock implementation of PostRepository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(ctx context.Context, req *model.CreatePostParams) (*model.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) GetPostByURL(ctx context.Context, url string) (*model.Post, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) ListPosts(ctx context.Context, req *model.PostListParams) ([]model.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Post), args.Error(1)
}

func (m *MockPostRepository) ListPostsByCategory(ctx context.Context, req *model.ListPostsByCategoryParams) ([]model.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Post), args.Error(1)
}

func (m *MockPostRepository) ListPostsBySource(ctx context.Context, req *model.ListPostsBySourceParams) ([]model.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Post), args.Error(1)
}

func (m *MockPostRepository) UpdatePost(ctx context.Context, id int64, req *model.UpdatePostParams) (*model.Post, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) CountPosts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) CountPostsByCategory(ctx context.Context, category string) (int64, error) {
	args := m.Called(ctx, category)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) SearchPosts(ctx context.Context, req *model.SearchPostsParams) ([]model.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Post), args.Error(1)
}

// PostServiceTestSuite defines the test suite for PostService
type PostServiceTestSuite struct {
	suite.Suite
	mockRepo *MockPostRepository
	logger   *logger.Logger
	service  service.PostService
	ctx      context.Context
}

func (suite *PostServiceTestSuite) SetupTest() {
	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}

	suite.mockRepo = new(MockPostRepository)
	suite.logger = logger.New(cfg)
	suite.service = service.NewPostService(suite.mockRepo, suite.logger)
	suite.ctx = context.Background()
}

func (suite *PostServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// Helper functions
func (suite *PostServiceTestSuite) createMockPost() *model.Post {
	now := time.Now()
	description := "Test description"
	content := "Test content"
	category := "technology"
	imageURL := "https://example.com/image.jpg"
	publishedAt := time.Now().Add(-24 * time.Hour)

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

func (suite *PostServiceTestSuite) createMockCreateParams() *model.CreatePostParams {
	description := "Test description"
	content := "Test content"
	category := "technology"
	imageURL := "https://example.com/image.jpg"
	publishedAt := time.Now().Add(-24 * time.Hour)

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

func (suite *PostServiceTestSuite) createMockUpdateParams() *model.UpdatePostParams {
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

func (suite *PostServiceTestSuite) TestCreatePostSuccess() {
	req := suite.createMockCreateParams()
	expectedPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByURL", suite.ctx, req.URL).Return(nil, pgx.ErrNoRows)

	suite.mockRepo.On("CreatePost", suite.ctx, req).Return(expectedPost, nil)

	result, err := suite.service.CreatePost(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedPost, result)
}

func (suite *PostServiceTestSuite) TestCreatePostPostExists() {
	req := suite.createMockCreateParams()
	existingPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByURL", suite.ctx, req.URL).Return(existingPost, nil)

	result, err := suite.service.CreatePost(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostExists, err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestCreatePostPostExistsCheckError() {
	req := suite.createMockCreateParams()
	dbError := errors.New("database error")

	// Mock PostExists to return error
	suite.mockRepo.On("GetPostByURL", suite.ctx, req.URL).Return(nil, dbError)

	result, err := suite.service.CreatePost(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "post with this URL already exists")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestCreatePostCreateError() {
	req := suite.createMockCreateParams()
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByURL", suite.ctx, req.URL).Return(nil, pgx.ErrNoRows)

	suite.mockRepo.On("CreatePost", suite.ctx, req).Return(nil, dbError)

	result, err := suite.service.CreatePost(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create post service")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestGetPostByIDSuccess() {
	id := int64(1)
	expectedPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(expectedPost, nil)

	result, err := suite.service.GetPostByID(suite.ctx, id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedPost, result)
}

func (suite *PostServiceTestSuite) TestGetPostByIDInvalidID() {
	id := int64(0)

	result, err := suite.service.GetPostByID(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostIDInvalid, err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestGetPostByIDPostNotFound() {
	id := int64(1)

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(nil, pgx.ErrNoRows)

	result, err := suite.service.GetPostByID(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostNotFound, err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestGetPostByIDDatabaseError() {
	id := int64(1)
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(nil, dbError)

	result, err := suite.service.GetPostByID(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get post")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestListPostsSuccess() {
	req := &model.PostListParams{
		Page:  1,
		Limit: 10,
	}
	posts := []model.Post{*suite.createMockPost()}
	totalCount := int64(1)

	suite.mockRepo.On("ListPosts", suite.ctx, req).Return(posts, nil)
	suite.mockRepo.On("CountPosts", suite.ctx).Return(totalCount, nil)

	result, err := suite.service.ListPosts(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), posts, result.Posts)
	assert.Equal(suite.T(), totalCount, result.Pagination.Total)
}

func (suite *PostServiceTestSuite) TestListPostsWithCategory() {
	category := "technology"
	req := &model.PostListParams{
		Page:     1,
		Limit:    10,
		Category: &category,
	}
	posts := []model.Post{*suite.createMockPost()}
	totalCount := int64(1)

	suite.mockRepo.On("ListPosts", suite.ctx, req).Return(posts, nil)
	suite.mockRepo.On("CountPostsByCategory", suite.ctx, category).Return(totalCount, nil)

	result, err := suite.service.ListPosts(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), posts, result.Posts)
	assert.Equal(suite.T(), totalCount, result.Pagination.Total)
}

func (suite *PostServiceTestSuite) TestListPostsDefaultPagination() {
	req := &model.PostListParams{
		Page:  0,
		Limit: 0,
	}
	posts := []model.Post{*suite.createMockPost()}
	totalCount := int64(1)

	expectedReq := &model.PostListParams{
		Page:  1,
		Limit: 20,
	}

	suite.mockRepo.On("ListPosts", suite.ctx, expectedReq).Return(posts, nil)
	suite.mockRepo.On("CountPosts", suite.ctx).Return(totalCount, nil)

	result, err := suite.service.ListPosts(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestListPostsMaxLimitEnforcement() {
	req := &model.PostListParams{
		Page:  1,
		Limit: 200,
	}
	posts := []model.Post{*suite.createMockPost()}
	totalCount := int64(1)

	expectedReq := &model.PostListParams{
		Page:  1,
		Limit: 100,
	}

	suite.mockRepo.On("ListPosts", suite.ctx, expectedReq).Return(posts, nil)
	suite.mockRepo.On("CountPosts", suite.ctx).Return(totalCount, nil)

	result, err := suite.service.ListPosts(suite.ctx, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestListPostsListError() {
	req := &model.PostListParams{
		Page:  1,
		Limit: 10,
	}
	dbError := errors.New("database error")

	suite.mockRepo.On("ListPosts", suite.ctx, req).Return(nil, dbError)

	result, err := suite.service.ListPosts(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to list posts")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestListPostsCountError() {
	req := &model.PostListParams{
		Page:  1,
		Limit: 10,
	}
	posts := []model.Post{*suite.createMockPost()}
	dbError := errors.New("database error")

	suite.mockRepo.On("ListPosts", suite.ctx, req).Return(posts, nil)
	suite.mockRepo.On("CountPosts", suite.ctx).Return(int64(0), dbError)

	result, err := suite.service.ListPosts(suite.ctx, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to count posts")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestUpdatePostSuccess() {
	id := int64(1)
	req := suite.createMockUpdateParams()
	existingPost := suite.createMockPost()
	updatedPost := suite.createMockPost()
	updatedPost.Title = req.Title

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(existingPost, nil)
	suite.mockRepo.On("UpdatePost", suite.ctx, id, req).Return(updatedPost, nil)

	result, err := suite.service.UpdatePost(suite.ctx, id, req)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedPost, result)
}

func (suite *PostServiceTestSuite) TestUpdatePostInvalidID() {
	id := int64(0)
	req := suite.createMockUpdateParams()

	result, err := suite.service.UpdatePost(suite.ctx, id, req)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostIDInvalid, err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestUpdatePostPostNotFound() {
	id := int64(1)
	req := suite.createMockUpdateParams()

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(nil, pgx.ErrNoRows)

	result, err := suite.service.UpdatePost(suite.ctx, id, req)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostNotFound, err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestUpdatePostGetPostError() {
	id := int64(1)
	req := suite.createMockUpdateParams()
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(nil, dbError)

	result, err := suite.service.UpdatePost(suite.ctx, id, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to check post existence")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestUpdatePostUpdateError() {
	id := int64(1)
	req := suite.createMockUpdateParams()
	existingPost := suite.createMockPost()
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(existingPost, nil)
	suite.mockRepo.On("UpdatePost", suite.ctx, id, req).Return(nil, dbError)

	result, err := suite.service.UpdatePost(suite.ctx, id, req)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to update post")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestDeletePostSuccess() {
	id := int64(1)
	existingPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(existingPost, nil)
	suite.mockRepo.On("DeletePost", suite.ctx, id).Return(nil)

	err := suite.service.DeletePost(suite.ctx, id)

	assert.NoError(suite.T(), err)
}

func (suite *PostServiceTestSuite) TestDeletePostInvalidID() {
	id := int64(0)

	err := suite.service.DeletePost(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostIDInvalid, err)
}

func (suite *PostServiceTestSuite) TestDeletePostPostNotFound() {
	id := int64(1)

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(nil, pgx.ErrNoRows)

	err := suite.service.DeletePost(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), service.ErrPostNotFound, err)
}

func (suite *PostServiceTestSuite) TestDeletePostGetPostError() {
	id := int64(1)
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(nil, dbError)

	err := suite.service.DeletePost(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to check post existence")
}

func (suite *PostServiceTestSuite) TestDeletePostDeleteError() {
	id := int64(1)
	existingPost := suite.createMockPost()
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByID", suite.ctx, id).Return(existingPost, nil)
	suite.mockRepo.On("DeletePost", suite.ctx, id).Return(dbError)

	err := suite.service.DeletePost(suite.ctx, id)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to delete post")
}

func (suite *PostServiceTestSuite) TestPostExistsTrue() {
	url := "https://example.com/test"
	existingPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByURL", suite.ctx, url).Return(existingPost, nil)

	result, err := suite.service.PostExists(suite.ctx, url)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestPostExistsFalse() {
	url := "https://example.com/test"

	suite.mockRepo.On("GetPostByURL", suite.ctx, url).Return(nil, pgx.ErrNoRows)

	result, err := suite.service.PostExists(suite.ctx, url)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestPostExistsDatabaseError() {
	url := "https://example.com/test"
	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByURL", suite.ctx, url).Return(nil, dbError)

	result, err := suite.service.PostExists(suite.ctx, url)

	// Note: Current implementation returns true, nil for any non-pgx.ErrNoRows error
	// This might be a bug in the original implementation
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result)
}

// Tests for CreatePostFromNewsAPI
func (suite *PostServiceTestSuite) TestCreatePostFromNewsAPISuccess() {
	article := &model.NewsAPIArticleParams{
		Source: struct {
			ID   *string `json:"id" example:"techcrunch"`
			Name string  `json:"name" example:"TechCrunch"`
		}{
			Name: "TechCrunch",
		},
		Title:       "Test Article",
		Description: stringPtr("Test description"),
		URL:         "https://example.com/test",
		URLToImage:  stringPtr("https://example.com/image.jpg"),
		PublishedAt: "2024-01-20T10:00:00Z",
		Content:     stringPtr("Test content"),
	}

	expectedPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByURL", suite.ctx, article.URL).Return(nil, pgx.ErrNoRows)

	suite.mockRepo.On("CreatePost", suite.ctx, mock.AnythingOfType("*model.CreatePostParams")).Return(expectedPost, nil)

	result, err := suite.service.CreatePostFromNewsAPI(suite.ctx, article)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedPost, result)
}

func (suite *PostServiceTestSuite) TestCreatePostFromNewsAPIDuplicatePost() {
	article := &model.NewsAPIArticleParams{
		Source: struct {
			ID   *string `json:"id" example:"techcrunch"`
			Name string  `json:"name" example:"TechCrunch"`
		}{
			Name: "TechCrunch",
		},
		Title:       "Test Article",
		URL:         "https://example.com/test",
		PublishedAt: "2024-01-20T10:00:00Z",
	}

	existingPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByURL", suite.ctx, article.URL).Return(existingPost, nil)

	result, err := suite.service.CreatePostFromNewsAPI(suite.ctx, article)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestCreatePostFromNewsAPIPostExistsError() {
	article := &model.NewsAPIArticleParams{
		Source: struct {
			ID   *string `json:"id" example:"techcrunch"`
			Name string  `json:"name" example:"TechCrunch"`
		}{
			Name: "TechCrunch",
		},
		Title:       "Test Article",
		URL:         "https://example.com/test",
		PublishedAt: "2024-01-20T10:00:00Z",
	}

	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByURL", suite.ctx, article.URL).Return(nil, dbError)

	result, err := suite.service.CreatePostFromNewsAPI(suite.ctx, article)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestCreatePostFromNewsAPICreateError() {
	article := &model.NewsAPIArticleParams{
		Source: struct {
			ID   *string `json:"id" example:"techcrunch"`
			Name string  `json:"name" example:"TechCrunch"`
		}{
			Name: "TechCrunch",
		},
		Title:       "Test Article",
		URL:         "https://example.com/test",
		PublishedAt: "2024-01-20T10:00:00Z",
	}

	dbError := errors.New("database error")

	suite.mockRepo.On("GetPostByURL", suite.ctx, article.URL).Return(nil, pgx.ErrNoRows)

	suite.mockRepo.On("GetPostByURL", suite.ctx, article.URL).Return(nil, pgx.ErrNoRows)
	suite.mockRepo.On("CreatePost", suite.ctx, mock.AnythingOfType("*model.CreatePostParams")).Return(nil, dbError)

	result, err := suite.service.CreatePostFromNewsAPI(suite.ctx, article)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create post from NewsAPI")
	assert.Nil(suite.T(), result)
}

func (suite *PostServiceTestSuite) TestCreatePostFromNewsAPIInvalidPublishedAt() {
	article := &model.NewsAPIArticleParams{
		Source: struct {
			ID   *string `json:"id" example:"techcrunch"`
			Name string  `json:"name" example:"TechCrunch"`
		}{
			Name: "TechCrunch",
		},
		Title:       "Test Article",
		URL:         "https://example.com/test",
		PublishedAt: "invalid-date",
	}

	expectedPost := suite.createMockPost()

	suite.mockRepo.On("GetPostByURL", suite.ctx, article.URL).Return(nil, pgx.ErrNoRows)

	suite.mockRepo.On("CreatePost", suite.ctx, mock.AnythingOfType("*model.CreatePostParams")).Return(expectedPost, nil)

	result, err := suite.service.CreatePostFromNewsAPI(suite.ctx, article)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedPost, result)
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// Run the test suite
func TestPostServiceSuite(t *testing.T) {
	suite.Run(t, new(PostServiceTestSuite))
}
