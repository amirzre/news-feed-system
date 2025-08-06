package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/response"
	"github.com/labstack/echo/v4"
)

type PostHandler struct {
	postService service.PostService
	logger      *logger.Logger
}

// NewPostHandler creates a new post handler
func NewPostHandler(postService service.PostService, logger *logger.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		logger:      logger,
	}
}

// CreatePost handles POST /api/v1/posts
func (h *PostHandler) CreatePost(c echo.Context) error {
	start := time.Now()

	var req model.CreatePostParams
	if err := c.Bind(&req); err != nil {
		h.logger.LogServiceOperation("post_handler", "create_post", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "Invalid request body", err.Error())
	}

	// TODO: validate request

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
