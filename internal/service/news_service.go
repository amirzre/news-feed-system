package service

import (
	"net/http"
	"time"

	"github.com/amirzre/news-feed-system/internal/config"
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
