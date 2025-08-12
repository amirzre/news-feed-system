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
func (n *newsService) GetTopHeadlines(ctx context.Context, req *model.NewsParams) (*model.NewsAPIResponse, error) {
	start := time.Now()

	endpoint := fmt.Sprintf("%s/top-headlines", n.baseURL)

	params := url.Values{}
	params.Set("apiKey", n.apiKey)

	if req.Query != "" {
		params.Set("q", req.Query)
	}
	if req.Category != "" {
		params.Set("category", req.Category)
	}
	if req.Country != "" {
		params.Set("country", req.Country)
	}
	if req.Language != "" {
		params.Set("language", req.Language)
	}
	if req.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(req.PageSize))
	}
	if req.Page > 0 {
		params.Set("page", strconv.Itoa(req.Page))
	}

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := n.makeRequest(ctx, fullURL)
	if err != nil {
		n.logger.LogServiceOperation("news_service", "get_top_headlines", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("failed to get top headlines: %w", err)
	}

	n.logger.LogServiceOperation("news_service", "get_top_headlines", true, time.Since(start).Milliseconds())
	n.logger.Debug("Fetched top headlines",
		"articles_count", len(response.Articles),
		"total_results", response.TotalResults,
	)

	return response, nil
}

// GetEverything fetches all articles matching the criteria
func (n *newsService) GetEverything(ctx context.Context, req *model.NewsParams) (*model.NewsAPIResponse, error) {
	start := time.Now()

	endpoint := fmt.Sprintf("%s/everything", n.baseURL)

	params := url.Values{}
	params.Set("apiKey", n.apiKey)

	if req.Query != "" {
		params.Set("q", req.Query)
	}
	if len(req.Sources) > 0 {
		sources := ""
		for i, source := range req.Sources {
			if i > 0 {
				sources += ","
			}
			sources += source
		}
		params.Set("sources", sources)
	}
	if req.Language != "" {
		params.Set("language", req.Language)
	}
	if req.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(req.PageSize))
	}
	if req.Page > 0 {
		params.Set("page", strconv.Itoa(req.Page))
	}

	from := time.Now().AddDate(0, 0, -7).Format(time.DateOnly)
	params.Set("from", from)
	params.Set("sortBy", "publishedAt")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	response, err := n.makeRequest(ctx, fullURL)
	if err != nil {
		n.logger.LogServiceOperation("news_service", "get_everything", false, time.Since(start).Milliseconds())
		return nil, fmt.Errorf("failed to get everything: %w", err)
	}

	n.logger.LogServiceOperation("news_service", "get_everything", true, time.Since(start).Milliseconds())
	n.logger.Debug("Fetched everything articles",
		"articles_count", len(response.Articles),
		"total_results", response.TotalResults,
	)

	return response, nil
}

// GetNewsByCategory fetches news by category
func (n *newsService) GetNewsByCategory(ctx context.Context, category string, pageSize int) (*model.NewsAPIResponse, error) {
	params := &model.NewsParams{
		Category: category,
		Country:  "us",
		Language: "en",
		PageSize: pageSize,
	}

	return n.GetTopHeadlines(ctx, params)
}

// GetNewsBySources fetches news from specific sources
func (n *newsService) GetNewsBySources(ctx context.Context, sources []string, pageSize int) (*model.NewsAPIResponse, error) {
	params := &model.NewsParams{
		Sources:  sources,
		Language: "en",
		PageSize: pageSize,
	}

	return n.GetEverything(ctx, params)
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

// GetDefaultSources returns a list of popular news sources
func GetDefaultSources() []string {
	return []string{
		"bbc-news",
		"cnn",
		"reuters",
		"associated-press",
		"the-verge",
		"techcrunch",
		"ars-technica",
		"hacker-news",
		"the-wall-street-journal",
		"bloomberg",
	}
}

// GetDefaultCategories returns a list of news categories
func GetDefaultCategories() []string {
	return []string{
		"general",
		"business",
		"entertainment",
		"health",
		"science",
		"sports",
		"technology",
	}
}
