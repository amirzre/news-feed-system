package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	NewsAPI  NewsAPIConfig
	App      AppConfig
	Cache    CacheConfig
	CORS     CORSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
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
