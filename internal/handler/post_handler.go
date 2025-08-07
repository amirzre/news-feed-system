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
