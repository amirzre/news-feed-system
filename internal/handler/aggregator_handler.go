package handler

import (
	"github.com/amirzre/news-feed-system/internal/service"
	"github.com/amirzre/news-feed-system/pkg/logger"
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
