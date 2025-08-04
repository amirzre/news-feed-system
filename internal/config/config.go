package config

import "time"

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
