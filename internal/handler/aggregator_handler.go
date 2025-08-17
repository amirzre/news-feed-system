package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/amirzre/news-feed-system/internal/model"
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

// TriggerAggregation handles POST /api/v1/aggregation/trigger
// @Summary      Trigger full aggregation
// @Description  Trigger a complete aggregation across all categories and sources
// @Tags         aggregation
// @Accept       json
// @Produce      json
// @Success      201  {object}  response.APIResponse{data=model.AggregationResponse}  "Aggregation result"
// @Failure      500  {object}  response.APIResponse{error=response.ErrorInfo}        "Aggregation failed"
// @Router       /aggregation/trigger [post]
func (h *aggregatorHandler) TriggerAggregation(c echo.Context) error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Minute)
	defer cancel()

	h.logger.Info("Manual aggregation triggered via API")

	result, err := h.aggregatorService.AggregateAll(ctx)
	if err != nil {
		h.logger.LogServiceOperation("aggregator_handler", "trigger_aggregation", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Aggregation failed", err.Error())
	}

	h.logger.LogServiceOperation("aggregator_handler", "trigger_aggregation", true, time.Since(start).Milliseconds())
	return response.Success(c, http.StatusCreated, result, "Aggregation completed successfully")
}

// TriggerTopHeadlines handles POST /api/v1/aggregation/trigger/headlines
// @Summary      Trigger top headlines aggregation
// @Description  Trigger aggregation for top headlines (global/top-level headlines)
// @Tags         aggregation
// @Accept       json
// @Produce      json
// @Success      201  {object}  response.APIResponse{data=model.AggregationResponse}  "Top headlines aggregation result"
// @Failure      500  {object}  response.APIResponse{error=response.ErrorInfo}        "Aggregation failed"
// @Router       /aggregation/trigger/headlines [post]
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

	return response.Success(c, http.StatusCreated, result, "Top headlines aggregation completed successfully")
}

// TriggerCategoryAggregation handles POST /api/v1/aggregation/trigger/categories
// @Summary      Trigger category aggregation
// @Description  Trigger aggregation for one or more categories. If no categories provided, defaults are used.
// @Tags         aggregation
// @Accept       json
// @Produce      json
// @Param        body  body      model.CategoryAggregationRequest     false  "Categories payload (optional)"
// @Success      201   {object}  response.APIResponse{data=model.CategoryAggregationResponse}    "Category aggregation result"
// @Failure      400   {object}  response.APIResponse{error=response.ErrorInfo}                   "No valid categories provided"
// @Failure      500   {object}  response.APIResponse{error=response.ErrorInfo}                   "Aggregation failed"
// @Router       /aggregation/trigger/categories [post]
func (h *aggregatorHandler) TriggerCategoryAggregation(c echo.Context) error {
	start := time.Now()

	var req model.CategoryAggregationRequest
	if err := c.Bind(&req); err != nil {
		// If binding fails, use default categories
		req.Categories = service.GetDefaultCategories()
	}

	if len(req.Categories) == 0 {
		req.Categories = service.GetDefaultCategories()
	}

	validCategories := service.GetDefaultCategories()
	var filteredCategories []string
	for _, category := range req.Categories {
		category = strings.ToLower(strings.TrimSpace(category))
		for _, valid := range validCategories {
			if category == valid {
				filteredCategories = append(filteredCategories, category)
				break
			}
		}
	}

	if len(filteredCategories) == 0 {
		h.logger.LogServiceOperation("aggregator_handler", "trigger_category_aggregation", false, time.Since(start).Milliseconds())
		return response.BadRequest(c, "No valid categories provided", "Available categories: "+strings.Join(validCategories, ", "))
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 4*time.Minute)
	defer cancel()

	h.logger.Info("Manual category aggregation triggered via API", "categories", filteredCategories)

	result, err := h.aggregatorService.AggregateByCategories(ctx, filteredCategories)
	if err != nil {
		h.logger.LogServiceOperation("aggregator_handler", "trigger_category_aggregation", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Category aggregation failed", err.Error())
	}

	h.logger.LogServiceOperation("aggregator_handler", "trigger_category_aggregation", true, time.Since(start).Milliseconds())

	responseData := model.CategoryAggregationResponse{
		Categories: filteredCategories,
		Result: model.AggregationResponse{
			TotalFetched:    result.TotalFetched,
			TotalCreated:    result.TotalCreated,
			TotalDuplicates: result.TotalDuplicates,
			TotalErrors:     result.TotalErrors,
			Duration:        result.Duration,
			Categories:      result.Categories,
			Sources:         result.Sources,
			Errors:          result.Errors,
		},
	}

	return response.Success(c, http.StatusCreated, responseData, "Category aggregation completed successfully")
}

// TriggerSourceAggregation handles POST /api/v1/aggregation/trigger/sources
// @Summary      Trigger source aggregation
// @Description  Trigger aggregation for one or more sources. If no sources provided, defaults are used.
// @Tags         aggregation
// @Accept       json
// @Produce      json
// @Param        body  body      model.SourceAggregationRequest       false  "Sources payload (optional)"
// @Success      201   {object}  response.APIResponse{data=model.SourceAggregationResponse}      "Source aggregation result"
// @Failure      400   {object}  response.APIResponse{error=response.ErrorInfo}                   "No valid sources provided"
// @Failure      500   {object}  response.APIResponse{error=response.ErrorInfo}                   "Aggregation failed"
// @Router       /aggregation/trigger/sources [post]
func (h *aggregatorHandler) TriggerSourceAggregation(c echo.Context) error {
	start := time.Now()

	var req model.SourceAggregationRequest
	if err := c.Bind(&req); err != nil {
		req.Sources = service.GetDefaultSources()
	}

	if len(req.Sources) == 0 {
		req.Sources = service.GetDefaultSources()
	}

	var filteredSources []string
	for _, source := range req.Sources {
		source = strings.TrimSpace(source)
		if source != "" {
			filteredSources = append(filteredSources, source)
		}
	}

	if len(filteredSources) == 0 {
		h.logger.LogServiceOperation("aggregator_handler", "trigger_source_aggregation", false, time.Since(start).Milliseconds())

		defaultSources := service.GetDefaultSources()
		return response.BadRequest(c, "No valid sources provided", "Available sources: "+strings.Join(defaultSources, ", "))
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Minute)
	defer cancel()

	h.logger.Info("Manual source aggregation triggered via API", "sources", filteredSources)

	result, err := h.aggregatorService.AggregateBySources(ctx, filteredSources)
	if err != nil {
		h.logger.LogServiceOperation("aggregator_handler", "trigger_source_aggregation", false, time.Since(start).Milliseconds())
		return response.InternalServerError(c, "Source aggregation failed", err.Error())
	}

	h.logger.LogServiceOperation("aggregator_handler", "trigger_source_aggregation", true, time.Since(start).Milliseconds())

	responseData := model.SourceAggregationResponse{
		Sources: filteredSources,
		Result: model.AggregationResponse{
			TotalFetched:    result.TotalFetched,
			TotalCreated:    result.TotalCreated,
			TotalDuplicates: result.TotalDuplicates,
			TotalErrors:     result.TotalErrors,
			Duration:        result.Duration,
			Categories:      result.Categories,
			Sources:         result.Sources,
			Errors:          result.Errors,
		},
	}

	return response.Success(c, http.StatusCreated, responseData, "Source aggregation completed successfully")
}
