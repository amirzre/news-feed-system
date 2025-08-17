package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/response"
	"github.com/labstack/echo/v4"
)

// postHandler implements PostHandler interface
type postHandler struct {
	postService service.PostService
	logger      *logger.Logger
}

// NewPostHandler creates a new post handler
func NewPostHandler(postService service.PostService, logger *logger.Logger) PostHandler {
	return &postHandler{
		postService: postService,
		logger:      logger,
	}
}

// CreatePost handles POST /api/v1/posts
// @Summary      Create a new post
// @Description  Create a new post with the provided payload
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post  body      model.CreatePostParams  true  "Create Post payload"
// @Success      201   {object}  response.APIResponse{data=model.Post}              "Post created"
// @Failure      400   {object}  response.APIResponse{error=response.ErrorInfo}  "Invalid request"
// @Failure      409   {object}  response.APIResponse{error=response.ErrorInfo}  "Conflict - post exists"
// @Failure      500   {object}  response.APIResponse{error=response.ErrorInfo}  "Internal server error"
// @Router       /posts [post]
func (h *postHandler) CreatePost(c echo.Context) error {
	start := time.Now()

	var req model.CreatePostParams
	if err := c.Bind(&req); err != nil {
		h.logger.LogServiceOperation("post_handler", "create_post", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		h.logger.LogServiceOperation("post_handler", "create_post", false, time.Since(start).Milliseconds())
		return response.ValidationError(c, err)
	}

	post, err := h.postService.CreatePost(c.Request().Context(), &req)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "create_post", false, time.Since(start).Milliseconds())

		if errors.Is(err, service.ErrPostExists) {
			return response.Conflict(c, "Post with this URL already exists")
		}

		return response.InternalServerError(c, "Failed to create post")
	}

	h.logger.LogServiceOperation("post_handler", "create_post", true, time.Since(start).Milliseconds())

	return response.Success(c, http.StatusCreated, post, "Post created successfully")
}

// GetPost handles GET /api/v1/posts/:id
// @Summary      Get a post by ID
// @Description  Retrieve a single post by its ID
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.APIResponse{data=model.Post}              "Post retrieved"
// @Failure      400  {object}  response.APIResponse{error=response.ErrorInfo}     "Invalid ID"
// @Failure      404  {object}  response.APIResponse{error=response.ErrorInfo}     "Post not found"
// @Failure      500  {object}  response.APIResponse{error=response.ErrorInfo}     "Internal server error"
// @Router       /posts/{id} [get]
func (h *postHandler) GetPostByID(c echo.Context) error {
	start := time.Now()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		h.logger.LogServiceOperation("post_handler", "get_post", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Invalid post ID")
	}

	post, err := h.postService.GetPostByID(c.Request().Context(), id)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "get_post", false, time.Since(start).Milliseconds())

		if errors.Is(err, service.ErrPostNotFound) {
			return response.NotFound(c, "Post not found")
		}

		return response.InternalServerError(c, "Failed to retrieve post")
	}

	h.logger.LogServiceOperation("post_handler", "get_post", true, time.Since(start).Milliseconds())

	return response.Success(c, http.StatusOK, post)
}

// ListPosts handles GET /api/v1/posts with pagination, filtering, and search
// @Summary      List posts
// @Description  List posts with pagination, optional filtering by category/source and search
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Results per page"
// @Param        category  query     string  false  "Filter by category"
// @Param        source    query     string  false  "Filter by source"
// @Param        search    query     string  false  "Search term"
// @Success      200       {object}  response.APIResponse{data=response.PaginatedResponse{items=[]model.Post,pagination=response.PaginationInfo}}	"List of posts"
// @Failure      400       {object}  response.APIResponse{error=response.ErrorInfo}  "Validation error"
// @Failure      500       {object}  response.APIResponse{error=response.ErrorInfo}  "Internal server error"
// @Router       /posts [get]
func (h *postHandler) ListPosts(c echo.Context) error {
	start := time.Now()

	req := model.DefaultPostListParams()

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	filters := make(map[string]string)
	if category := c.QueryParam("category"); category != "" {
		req.Category = &category
		filters["category"] = category
	}

	if source := c.QueryParam("source"); source != "" {
		req.Source = &source
		filters["source"] = source
	}

	if search := c.QueryParam("search"); search != "" {
		req.Search = &search
		filters["search"] = search
	}

	if err := c.Validate(&req); err != nil {
		h.logger.LogServiceOperation("post_handler", "list_posts", false, time.Since(start).Milliseconds())
		return response.ValidationError(c, err)
	}

	posts, err := h.postService.ListPosts(c.Request().Context(), &req)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "list_posts", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Failed to retrieve posts")
	}

	h.logger.LogServiceOperation("post_handler", "list_posts", true, time.Since(start).Milliseconds())
	h.logger.Debug("Listed posts",
		"page", req.Page,
		"limit", req.Limit,
		"total", posts.Pagination.Total,
		"returned", len(posts.Posts),
	)

	paginationInfo := response.CreatePaginationInfo(req.Page, req.Limit, int(posts.Pagination.Total))

	return response.SuccessWithPagination(c, posts.Posts, paginationInfo, filters)
}

// UpdatePost handles PUT /api/v1/posts/:id
// @Summary      Update a post
// @Description  Update a post by ID with the provided payload
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path      int                     true  "Post ID"
// @Param        post  body      model.UpdatePostParams  true  "Update Post payload"
// @Success      200   {object}  response.APIResponse{data=model.Post}           "Updated post"
// @Failure      400   {object}  response.APIResponse{error=response.ErrorInfo}  "Invalid request"
// @Failure      404   {object}  response.APIResponse{error=response.ErrorInfo}  "Post not found"
// @Failure      500   {object}  response.APIResponse{error=response.ErrorInfo}  "Internal server error"
// @Router       /posts/{id} [put]
func (h *postHandler) UpdatePost(c echo.Context) error {
	start := time.Now()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "update_post", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Invalid post ID")
	}

	var req model.UpdatePostParams
	if err := c.Bind(&req); err != nil {
		h.logger.LogServiceOperation("post_handler", "update_post", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	if err := c.Validate(&req); err != nil {
		h.logger.LogServiceOperation("post_handler", "update_post", false, time.Since(start).Milliseconds())
		return response.ValidationError(c, err)
	}

	post, err := h.postService.UpdatePost(c.Request().Context(), id, &req)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "update_post", false, time.Since(start).Milliseconds())

		if errors.Is(err, service.ErrPostNotFound) {
			return response.NotFound(c, "Post not found")
		}

		return response.InternalServerError(c, "Failed to update post")
	}

	h.logger.LogServiceOperation("post_handler", "update_post", true, time.Since(start).Milliseconds())

	return response.Success(c, http.StatusOK, post, "Post updated successfully")
}

// DeletePost handles DELETE /api/v1/posts/:id
// @Summary      Delete a post
// @Description  Delete a post by ID
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      204  {string}  string                  "No Content"
// @Failure      400  {object}  response.APIResponse{error=response.ErrorInfo} "Invalid ID"
// @Failure      404  {object}  response.APIResponse{error=response.ErrorInfo} "Post not found"
// @Failure      500  {object}  response.APIResponse{error=response.ErrorInfo} "Internal server error"
// @Router       /posts/{id} [delete]
func (h *postHandler) DeletePost(c echo.Context) error {
	start := time.Now()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		h.logger.LogServiceOperation("post_handler", "delete_post", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Invalid post ID")
	}

	err = h.postService.DeletePost(c.Request().Context(), id)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "delete_post", false, time.Since(start).Milliseconds())

		if errors.Is(err, service.ErrPostNotFound) {
			return response.NotFound(c, "Post not found")
		}

		return response.InternalServerError(c, "Failed to delete post")
	}

	h.logger.LogServiceOperation("post_handler", "delete_post", true, time.Since(start).Milliseconds())

	return response.Success(c, http.StatusNoContent, nil, "Post deleted successfully")
}

// GetPostsByCategory handles GET /api/v1/posts/category/:category
// @Summary      List posts by category
// @Description  List posts filtered by category
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        category  path      string  true   "Category"
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Results per page"
// @Success      200       {object}   response.APIResponse{data=response.PaginatedResponse{items=[]model.Post,pagination=response.PaginationInfo}}	"List of posts"
// @Failure      400       {object}  response.APIResponse{error=response.ErrorInfo}  "Validation error"
// @Failure      500       {object}  response.APIResponse{error=response.ErrorInfo}  "Internal server error"
// @Router       /posts/category/{category} [get]
func (h *postHandler) GetPostsByCategory(c echo.Context) error {
	start := time.Now()

	category := c.Param("category")
	if category == "" {
		h.logger.LogServiceOperation("post_handler", "get_posts_by_category", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Category is required")
	}

	req := model.DefaultPostListParams()
	req.Category = &category

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	posts, err := h.postService.ListPosts(c.Request().Context(), &req)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "get_posts_by_category", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Failed to retrieve posts by category")
	}

	h.logger.LogServiceOperation("post_handler", "get_posts_by_category", true, time.Since(start).Milliseconds())

	paginationInfo := response.CreatePaginationInfo(req.Page, req.Limit, int(posts.Pagination.Total))
	filters := map[string]string{"category": category}

	return response.SuccessWithPagination(c, posts.Posts, paginationInfo, filters)
}

// GetPostsBySource handles GET /api/v1/posts/source/:source
// @Summary      List posts by source
// @Description  List posts filtered by source
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        source    path      string  true   "Source"
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Results per page"
// @Success      200       {object}  response.APIResponse{data=response.PaginatedResponse{items=[]model.Post,pagination=response.PaginationInfo}}	"List of posts"
// @Failure      400       {object}  response.APIResponse{error=response.ErrorInfo}  "Validation error"
// @Failure      500       {object}  response.APIResponse{error=response.ErrorInfo}  "Internal server error"
// @Router       /posts/source/{source} [get]
func (h *postHandler) GetPostsBySource(c echo.Context) error {
	start := time.Now()

	source := c.Param("source")
	if source == "" {
		h.logger.LogServiceOperation("post_handler", "get_posts_by_source", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Source is required")
	}

	req := model.DefaultPostListParams()
	req.Source = &source

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	posts, err := h.postService.ListPosts(c.Request().Context(), &req)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "get_posts_by_source", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Failed to retrieve posts by source")
	}

	h.logger.LogServiceOperation("post_handler", "get_posts_by_source", true, time.Since(start).Milliseconds())

	paginationInfo := response.CreatePaginationInfo(req.Page, req.Limit, int(posts.Pagination.Total))
	filters := map[string]string{"source": source}

	return response.SuccessWithPagination(c, posts.Posts, paginationInfo, filters)
}

// SearchPosts handles GET /api/v1/posts/search
// @Summary      Search posts
// @Description  Search posts by query string with optional filters
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        q         query     string  true   "Search query"
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Results per page"
// @Param        category  query     string  false  "Filter by category"
// @Param        source    query     string  false  "Filter by source"
// @Success      200       {object}   response.APIResponse{data=response.PaginatedResponse{items=[]model.Post,pagination=response.PaginationInfo}}              "Search results"
// @Failure      400       {object}  response.APIResponse{error=response.ErrorInfo}  "Validation error"
// @Failure      500       {object}  response.APIResponse{error=response.ErrorInfo}  "Internal server error"
// @Router       /posts/search [get]
func (h *postHandler) SearchPosts(c echo.Context) error {
	start := time.Now()

	query := c.QueryParam("q")
	if query == "" {
		h.logger.LogServiceOperation("post_handler", "search_posts", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Search query parameter 'q' is required")
	}

	req := model.DefaultPostListParams()
	req.Search = &query

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	filters := map[string]string{"search": query}
	if category := c.QueryParam("category"); category != "" {
		req.Category = &category
		filters["category"] = category
	}

	if source := c.QueryParam("source"); source != "" {
		req.Source = &source
		filters["source"] = source
	}

	posts, err := h.postService.ListPosts(c.Request().Context(), &req)
	if err != nil {
		h.logger.LogServiceOperation("post_handler", "search_posts", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Failed to search posts")
	}

	h.logger.LogServiceOperation("post_handler", "search_posts", true, time.Since(start).Milliseconds())

	paginationInfo := response.CreatePaginationInfo(req.Page, req.Limit, int(posts.Pagination.Total))

	return response.SuccessWithPagination(c, posts.Posts, paginationInfo, filters)
}
