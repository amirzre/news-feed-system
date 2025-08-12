package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
	"github.com/amirzre/news-feed-system/internal/model"
	"github.com/amirzre/news-feed-system/pkg/logger"
)

// newsService implements NewsService interface
type newsService struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	logger     *logger.Logger
}

// NewNewsService creates a new news service
func NewNewsService(cfg *config.Config, logger *logger.Logger) NewsService {
	return &newsService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  cfg.NewsAPI.APIKey,
		baseURL: cfg.NewsAPI.BaseURL,
		logger:  logger,
	}
}

// GetTopHeadlines fetches top headlines from NewsAPI
func (n *newsService) GetTopHeadlines(ctx context.Context, params *model.NewsParams) (*model.NewsAPIResponse, error) {
	start := time.Now()

	endpoint := fmt.Sprintf("%s/top-headlines", n.baseURL)

	param := url.Values{}
	param.Set("apiKey", n.apiKey)

	if params.Query != "" {
		param.Set("q", params.Query)
	}
	if params.Category != "" {
		param.Set("category", params.Category)
	}
	if params.Country != "" {
		param.Set("country", params.Country)
	}
	if params.Language != "" {
		param.Set("language", params.Language)
	}
	if params.PageSize > 0 {
		param.Set("pageSize", strconv.Itoa(params.PageSize))
	}
	if params.Page > 0 {
		param.Set("page", strconv.Itoa(params.Page))
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, param.Encode())

	response, err := n.makeRequest(ctx, fullURL)
	if err != nil {
		n.logger.LogServiceOperation("news_client", "get_top_headlines", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("failed to get top headlines: %w", err)
	}

	n.logger.LogServiceOperation("news_client", "get_top_headlines", true, time.Since(start).Milliseconds())
	n.logger.Debug("Fetched top headlines",
		"articles_count", len(response.Articles),
		"total_results", response.TotalResults,
	)

	return response, nil
}

// makeRequest makes an HTTP request to NewsAPI and handles the response
func (n *newsService) makeRequest(ctx context.Context, url string) (*model.NewsAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "news-feed-system/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, n.handleAPIError(resp.StatusCode, body)
	}

	var newsResponse model.NewsAPIResponse
	if err := json.Unmarshal(body, &newsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if newsResponse.Status != "ok" {
		return nil, fmt.Errorf("API returened error status: %s", newsResponse.Status)
	}

	return &newsResponse, nil
}

// handleAPIError handles different NewsAPI error responses
func (n *newsService) handleAPIError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid API key")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded")
	case http.StatusBadRequest:
		var errorResp struct {
			Status  string `json:"status"`
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return fmt.Errorf("API error: %s - %s", errorResp.Code, errorResp.Message)
		}
		return fmt.Errorf("bad request")
	case http.StatusInternalServerError:
		return fmt.Errorf("NewsAPI server error")
	default:
		return fmt.Errorf("unexpected HTTP status: %d", statusCode)
	}
}
