package handler

import (
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/labstack/echo/v4"
)

// PostHandler defines the contract for post HTTP handlers
type PostHandler interface {
	CreatePost(c echo.Context) error
	GetPostByID(c echo.Context) error
	ListPosts(c echo.Context) error
	UpdatePost(c echo.Context) error
	DeletePost(c echo.Context) error
	GetPostsByCategory(c echo.Context) error
	GetPostsBySource(c echo.Context) error
	SearchPosts(c echo.Context) error
}

// AggregatorHandler defines the contract for aggregator HTTP handlers
type AggregatorHandler interface {
	TriggerTopHeadlines(c echo.Context) error
	TriggerCategoryAggregation(c echo.Context) error
	TriggerSourceAggregation(c echo.Context) error
	TriggerAggregation(c echo.Context) error
}

// SchedulerHandler defines the contract for scheduler HTTP handlers
type SchedulerHandler interface {
	GetStatus(c echo.Context) error
	GetJobs(c echo.Context) error
}

// Handler holds all handler implementations
type Handler struct {
	Post       PostHandler
	Aggregator AggregatorHandler
	Scheduler  SchedulerHandler
}

// New creates a new handler instance with all entity handlers
func New(svc *service.Service, logger *logger.Logger) *Handler {
	return &Handler{
		Post:       NewPostHandler(svc.Post, logger),
		Aggregator: NewAggregatorHandler(svc.Aggregator, logger),
		Scheduler:  NewSchedulerHandler(svc.Scheduler, logger),
	}
}
