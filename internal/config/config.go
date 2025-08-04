package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database     DatabaseConfig
	DatabasePool DatabasePoolConfig
	Redis        RedisConfig
	Server       ServerConfig
	NewsAPI      NewsAPIConfig
	App          AppConfig
	Cache        CacheConfig
	CORS         CORSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type DatabasePoolConfig struct {
	MaxConns          int
	MinConns          int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type ServerConfig struct {
	Host string
	Port int
}

type NewsAPIConfig struct {
	APIKey  string
	BaseURL string
}

type AppConfig struct {
	Environment string
	LogLevel    string
}

type CacheConfig struct {
	TTL time.Duration
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// loads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		DatabasePool: DatabasePoolConfig{
			MaxConns:          getEnvInt("DB_MAX_CONNS", 25),
			MinConns:          getEnvInt("DB_MIN_CONNS", 5),
			MaxConnLifetime:   getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:   getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvDuration("DB_HEALTH_CHECK_PERIOD", time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		NewsAPI: NewsAPIConfig{
			APIKey:  getEnv("NEWS_API_KEY", ""),
			BaseURL: getEnv("NEWS_API_BASE_URL", "https://newsapi.org/v2"),
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		Cache: CacheConfig{
			TTL: time.Duration(getEnvInt("CACHE_TTL", 3600)) * time.Second,
		},
		CORS: CORSConfig{
			AllowOrigins: getEnvStringSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
			AllowMethods: getEnvStringSlice("CORS_ALLOW_METHODS", []string{
				"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
			}),
			AllowHeaders: getEnvStringSlice("CORS_ALLOW_HEADERS", []string{
				"Origin", "Content-Type", "Accept",
			}),
			ExposeHeaders:    getEnvStringSlice("CORS_EXPOSE_HEADERS", []string{}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvInt("CORS_MAX_AGE", 86400),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate checks if required configuration values are present
func (c *Config) validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	if c.NewsAPI.APIKey == "" {
		return fmt.Errorf("news API key is required")
	}

	return nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err != nil {
			return intValue
		}
	}

	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err != nil {
			return boolValue
		}
	}

	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}

func getEnvStringSlice(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	sliceValue := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			sliceValue = append(sliceValue, trimmed)
		}
	}

	if len(sliceValue) == 0 {
		return fallback
	}

	return sliceValue
}
