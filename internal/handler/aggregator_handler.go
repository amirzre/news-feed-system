package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
	"github.com/amirzre/news-feed-system/pkg/response"
	"github.com/labstack/echo/v4"
)

// aggregatorHandler implements AggregatorHandler interface
type aggregatorHandler struct {
	aggregatorService service.AggregatorService
	logger            *logger.Logger
}

// NewAggregatorHandler creates a new aggregator handler
func NewAggregatorHandler(aggregatorService service.AggregatorService, logger *logger.Logger) AggregatorHandler {
	return &aggregatorHandler{
		aggregatorService: aggregatorService,
		logger:            logger,
	}
}

// TriggerTopHeadlines handles POST /api/v1/aggregation/trigger/headlines
func (h *aggregatorHandler) TriggerTopHeadlines(c echo.Context) error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Minute)
	defer cancel()

	h.logger.Info("Manual top headlines aggregation triggered via API")

	result, err := h.aggregatorService.AggregateTopHeadlines(ctx)

	if err != nil {
		h.logger.LogServiceOperation("aggregator_handler", "trigger_top_headlines", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Top headlines aggregation failed", err.Error())
	}

	h.logger.LogServiceOperation("aggregator_handler", "trigger_top_headlines", true, time.Since(start).Milliseconds())

	return response.Success(c, http.StatusOK, result, "Top headlines aggregation completed successfully")
}
