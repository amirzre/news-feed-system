package handler

import "github.com/labstack/echo/v4"

// SetupRoutes configures all API routes
func SetupRoutes(e *echo.Echo, h *Handler) {
	// API v1 routes
	api := e.Group("/api/v1")

	// Post routes
	posts := api.Group("/posts")
	posts.POST("", h.Post.CreatePost)
	posts.GET("/:id", h.Post.GetPostByID)
}
