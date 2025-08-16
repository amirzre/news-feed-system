package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	tcred "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testSuite struct {
	db          *pgxpool.Pool
	redisClient *redis.Client
	repo        PostRepository
	logger      *logger.Logger

	// Containers for cleanup
	pgContainer    testcontainers.Container
	redisContainer testcontainers.Container
}

func setupTestSuite(t *testing.T) *testSuite {
	ctx := context.Background()

	// Setup PostgreSQL container
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	// Get PostgreSQL connection string
	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Setup Redis container
	redisContainer, err := tcred.Run(
		ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	require.NoError(t, err)

	// Get Redis connection details
	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Create database pool
	db, err := pgxpool.New(ctx, pgConnStr)
	require.NoError(t, err)

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
	})

	// Test connections
	require.NoError(t, db.Ping(ctx))
	require.NoError(t, redisClient.Ping(ctx).Err())

	// Create tables
	err = createTestTables(ctx, db)
	require.NoError(t, err)

	cfg := &config.Config{App: config.AppConfig{LogLevel: "debug"}}
	testLogger := logger.New(cfg)

	repo := NewPostRepository(db, redisClient, testLogger, time.Minute)

	return &testSuite{
		db:             db,
		redisClient:    redisClient,
		repo:           repo,
		logger:         testLogger,
		pgContainer:    pgContainer,
		redisContainer: redisContainer,
	}
}

func (ts *testSuite) teardown(t *testing.T) {
	ctx := context.Background()

	if ts.db != nil {
		ts.db.Close()
	}
	if ts.redisClient != nil {
		ts.redisClient.Close()
	}
	if ts.pgContainer != nil {
		require.NoError(t, ts.pgContainer.Terminate(ctx))
	}
	if ts.redisContainer != nil {
		require.NoError(t, ts.redisContainer.Terminate(ctx))
	}
}

func createTestTables(ctx context.Context, db *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			content TEXT,
			url VARCHAR(1000) UNIQUE NOT NULL,
			source VARCHAR(100) NOT NULL,
			category VARCHAR(50),
			image_url VARCHAR(1000),
			published_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
		
		CREATE INDEX idx_posts_published_at ON posts(published_at DESC);
		CREATE INDEX idx_posts_source ON posts(source);
		CREATE INDEX idx_posts_category ON posts(category);
		CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
		CREATE INDEX idx_posts_category_published ON posts(category, published_at DESC);
	`
	_, err := db.Exec(ctx, query)
	return err
}

func (ts *testSuite) cleanupData(ctx context.Context) {
	ts.db.Exec(ctx, "TRUNCATE posts RESTART IDENTITY CASCADE")
	ts.redisClient.FlushAll(ctx)
}

func createSamplePost() *model.CreatePostParams {
	publishedAt := time.Now().UTC().Add(-1 * time.Hour)
	desc := "Test post description"
	content := "Test post content"
	category := "Technology"
	imageURL := "https://example.com/image.jpg"

	return &model.CreatePostParams{
		Title:       "Test Post Title",
		Description: &desc,
		Content:     &content,
		URL:         "https://example.com/test-post",
		Source:      "Test Source",
		Category:    &category,
		ImageURL:    &imageURL,
		PublishedAt: &publishedAt,
	}
}

func TestPostRepositoryCreatePost(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	params := createSamplePost()

	post, err := ts.repo.CreatePost(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, post)
	assert.Greater(t, post.ID, int64(0))
	assert.Equal(t, params.Title, post.Title)
	assert.Equal(t, params.Description, post.Description)
	assert.Equal(t, params.Content, post.Content)
	assert.Equal(t, params.URL, post.URL)
	assert.Equal(t, params.Source, post.Source)
	assert.Equal(t, params.Category, post.Category)
	assert.Equal(t, params.ImageURL, post.ImageURL)
	assert.NotZero(t, post.CreatedAt)
	assert.NotZero(t, post.UpdatedAt)
	assert.NotNil(t, post.PublishedAt)
}

func TestPostRepositoryCreatePostDuplicateURL(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	params := createSamplePost()

	_, err := ts.repo.CreatePost(ctx, params)
	require.NoError(t, err)

	_, err = ts.repo.CreatePost(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create post")
}

func TestPostRepositoryGetPostByURL(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	params := createSamplePost()
	createdPost, err := ts.repo.CreatePost(ctx, params)
	require.NoError(t, err)

	post, err := ts.repo.GetPostByURL(ctx, createdPost.URL)

	require.NoError(t, err)
	require.NotNil(t, post)
	assert.Equal(t, createdPost.ID, post.ID)
	assert.Equal(t, createdPost.Title, post.Title)
	assert.Equal(t, createdPost.URL, post.URL)
}

func TestPostRepositoryGetPostByID(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	params := createSamplePost()
	createdPost, err := ts.repo.CreatePost(ctx, params)
	require.NoError(t, err)

	post, err := ts.repo.GetPostByID(ctx, createdPost.ID)
	require.NoError(t, err)
	assert.Equal(t, createdPost.ID, post.ID)

	post2, err := ts.repo.GetPostByID(ctx, createdPost.ID)
	require.NoError(t, err)
	assert.Equal(t, post.ID, post2.ID)
	assert.Equal(t, post.Title, post2.Title)
}

func TestPostRepositoryGetPostByIDNotFound(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	_, err := ts.repo.GetPostByID(ctx, 99999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get post by id")
}

func TestPostRepositoryUpdatePost(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	params := createSamplePost()
	createdPost, err := ts.repo.CreatePost(ctx, params)
	require.NoError(t, err)

	desc := "Updated description"
	content := "Updated content"
	category := "Updated Category"
	imageURL := "https://example.com/image.jpg"
	updateParams := &model.UpdatePostParams{
		Title:       "Updated Title",
		Description: &desc,
		Content:     &content,
		Category:    &category,
		ImageURL:    &imageURL,
	}

	updatedPost, err := ts.repo.UpdatePost(ctx, createdPost.ID, updateParams)

	require.NoError(t, err)
	assert.Equal(t, createdPost.ID, updatedPost.ID)
	assert.Equal(t, updateParams.Title, updatedPost.Title)
	assert.Equal(t, updateParams.Description, updatedPost.Description)
	assert.Equal(t, updateParams.Content, updatedPost.Content)
	assert.Equal(t, updateParams.Category, updatedPost.Category)
	assert.Equal(t, updateParams.ImageURL, updatedPost.ImageURL)
	assert.True(t, updatedPost.UpdatedAt.After(createdPost.UpdatedAt))
}

func TestPostRepositoryUpdatePostNotFound(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	desc := "Updated description"
	content := "Updated content"
	category := "Updated Category"
	imageURL := "https://example.com/image.jpg"
	updateParams := &model.UpdatePostParams{
		Title:       "Updated Title",
		Description: &desc,
		Content:     &content,
		Category:    &category,
		ImageURL:    &imageURL,
	}

	_, err := ts.repo.UpdatePost(ctx, 99999, updateParams)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update post")
}

func TestPostRepositoryDeletePost(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	params := createSamplePost()
	createdPost, err := ts.repo.CreatePost(ctx, params)
	require.NoError(t, err)

	err = ts.repo.DeletePost(ctx, createdPost.ID)
	require.NoError(t, err)

	_, err = ts.repo.GetPostByID(ctx, createdPost.ID)
	assert.Error(t, err)
}

func TestPostRepositoryDeletePostNotFound(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	err := ts.repo.DeletePost(ctx, 99999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPostRepositoryCountPosts(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	count, err := ts.repo.CountPosts(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	for i := 0; i < 3; i++ {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	count, err = ts.repo.CountPosts(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count2, err := ts.repo.CountPosts(ctx)
	require.NoError(t, err)
	assert.Equal(t, count, count2)
}

func TestPostRepositoryCountPostsByCategory(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	categories := []string{"Technology", "Sports", "Technology", "Health"}
	for i, category := range categories {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		params.Category = &category
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	techCount, err := ts.repo.CountPostsByCategory(ctx, "Technology")
	require.NoError(t, err)
	assert.Equal(t, int64(2), techCount)

	sportsCount, err := ts.repo.CountPostsByCategory(ctx, "Sports")
	require.NoError(t, err)
	assert.Equal(t, int64(1), sportsCount)

	nonExistentCount, err := ts.repo.CountPostsByCategory(ctx, "NonExistent")
	require.NoError(t, err)
	assert.Equal(t, int64(0), nonExistentCount)
}

func TestPostRepositoryListPosts(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	for i := 0; i < 5; i++ {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		params.Title = fmt.Sprintf("Post %d", i)
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	listParams := &model.PostListParams{
		Page:  1,
		Limit: 3,
	}

	posts, err := ts.repo.ListPosts(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts, 3)

	listParams.Page = 2
	posts2, err := ts.repo.ListPosts(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts2, 2)

	listParams.Page = 1
	posts3, err := ts.repo.ListPosts(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts3, 3)
	assert.Equal(t, posts[0].ID, posts3[0].ID)
}

func TestPostRepositoryListPostsByCategory(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	categories := []string{"Technology", "Sports", "Technology", "Health", "Technology"}
	for i, category := range categories {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		params.Category = &category
		params.Title = fmt.Sprintf("Post %d - %s", i, category)
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	listParams := &model.ListPostsByCategoryParams{
		BasePostListParams: model.BasePostListParams{
			Limit:  10,
			Offset: 0,
		},
		Category: "Technology",
	}

	posts, err := ts.repo.ListPostsByCategory(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts, 3)

	for _, post := range posts {
		require.NotNil(t, post.Category, "post category should not be nil")
		assert.Equal(t, "Technology", *post.Category)
	}
}

func TestPostRepositoryListPostsBySource(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	sources := []string{"Source A", "Source B", "Source A", "Source C", "Source A"}
	for i, source := range sources {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		params.Source = source
		params.Title = fmt.Sprintf("Post %d - %s", i, source)
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	listParams := &model.ListPostsBySourceParams{
		BasePostListParams: model.BasePostListParams{
			Limit:  10,
			Offset: 0,
		},
		Source: "Source A",
	}

	posts, err := ts.repo.ListPostsBySource(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts, 3)

	for _, post := range posts {
		assert.Equal(t, "Source A", post.Source)
	}
}

func TestPostRepositorySearchPosts(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	testData := []struct {
		title       string
		description string
	}{
		{"Go Programming Tutorial", "Learn Go programming language"},
		{"JavaScript Basics", "Introduction to JavaScript development"},
		{"Advanced Go Concepts", "Deep dive into Go advanced topics"},
		{"Python for Beginners", "Start learning Python programming"},
	}

	for i, data := range testData {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		params.Title = data.title
		params.Description = &data.description
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	searchParams := &model.SearchPostsParams{
		BasePostListParams: model.BasePostListParams{
			Limit:  10,
			Offset: 0,
		},
		Query: "Go",
	}

	posts, err := ts.repo.SearchPosts(ctx, searchParams)
	require.NoError(t, err)
	assert.Len(t, posts, 2)

	for _, post := range posts {
		titleMatch := contains(post.Title, "Go")
		descMatch := contains(*post.Description, "Go")
		assert.True(t, titleMatch || descMatch, "Post should contain search term")
	}
}

func TestPostRepositoryListPostsWithFilters(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown(t)

	ctx := context.Background()
	defer ts.cleanupData(ctx)

	category := "Technology"
	for i := 0; i < 3; i++ {
		params := createSamplePost()
		params.URL = fmt.Sprintf("https://example.com/post-%d", i)
		params.Category = &category
		params.Source = "Test Source"
		params.Title = fmt.Sprintf("Tech Post %d", i)
		_, err := ts.repo.CreatePost(ctx, params)
		require.NoError(t, err)
	}

	listParams := &model.PostListParams{
		Page:     1,
		Limit:    10,
		Category: &category,
	}

	posts, err := ts.repo.ListPosts(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts, 3)

	source := "Test Source"
	listParams = &model.PostListParams{
		Page:   1,
		Limit:  10,
		Source: &source,
	}

	posts, err = ts.repo.ListPosts(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts, 3)

	search := "Tech"
	listParams = &model.PostListParams{
		Page:   1,
		Limit:  10,
		Search: &search,
	}

	posts, err = ts.repo.ListPosts(ctx, listParams)
	require.NoError(t, err)
	assert.Len(t, posts, 3)
}

// Helper function
func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	return strings.Contains(s, substr)
}
