package handler

import "github.com/labstack/echo/v4"

// SetupRoutes configures all API routes
func SetupRoutes(e *echo.Echo, h *Handler) {
	// API v1 routes
	api := e.Group("/api/v1")

	// Post routes
	posts := api.Group("/posts")
	posts.GET("", h.Post.ListPosts)
	posts.POST("", h.Post.CreatePost)
	posts.GET("/:id", h.Post.GetPostByID)
	posts.PUT("/:id", h.Post.UpdatePost)
	posts.DELETE("/:id", h.Post.DeletePost)

	posts.GET("/category/:category", h.Post.GetPostsByCategory)
	posts.GET("/source/:source", h.Post.GetPostsBySource)
	posts.GET("/search", h.Post.SearchPosts)

	// Aggregation routes
	aggregation := api.Group("/aggregation")
	aggregation.POST("/trigger", h.Aggregator.TriggerAggregation)
	aggregation.POST("/trigger/headlines", h.Aggregator.TriggerTopHeadlines)
	aggregation.POST("/trigger/categories", h.Aggregator.TriggerCategoryAggregation)
	aggregation.POST("/trigger/sources", h.Aggregator.TriggerSourceAggregation)

	// Scheduler routes
	scheduler := api.Group("/scheduler")
	scheduler.GET("/status", h.Scheduler.GetStatus)
	scheduler.GET("/jobs", h.Scheduler.GetJobs)
	scheduler.POST("/jobs/:name/trigger", h.Scheduler.TriggerJob)
}
