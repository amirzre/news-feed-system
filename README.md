# 📰 News Feed System

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-supported-blue.svg)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/postgresql-15+-blue.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/redis-7+-red.svg)](https://redis.io/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A scalable, high-performance news aggregation and distribution system built with Go. This system automatically collects news from multiple sources, processes them, and provides a clean REST API for accessing curated news content.

## 🚀 Features

- **Multi-Source News Aggregation**: Automatically fetch news from various sources
- **Smart Scheduling**: Configurable job scheduling for different types of content
- **Real-time Processing**: Redis-backed caching and real-time data processing
- **RESTful API**: Clean, well-documented REST API endpoints
- **Database Migrations**: Automated database schema management
- **Docker Support**: Full containerization with Docker and Docker Compose
- **Scalable Architecture**: Clean architecture following Go best practices
- **Hot Reload Development**: Development environment with live reload support

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   External      │    │   News Feed     │    │   Database      │
│   News APIs     │───▶│   System        │───▶│   PostgreSQL    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                         │
                              ▼                         │
                       ┌─────────────────┐              │
                       │     Redis       │◀─────────────┘
                       │    Cache        │
                       └─────────────────┘
```

### Core Components

- **Aggregator Service**: Collects news from external APIs
- **Scheduler Service**: Manages automated news collection jobs
- **Post Service**: Handles news article CRUD operations
- **News Service**: Processes and curates news content
- **Repository Layer**: Data access abstraction
- **Handler Layer**: HTTP request/response handling

## 📋 Prerequisites

- [Go](https://golang.org/doc/install) 1.24 or higher
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/) (optional, for using Makefile commands)

## 🚀 Quick Start

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <your-repository-url>
   cd news-feed-system
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the application**
   ```bash
   # Development environment
   make dev-up

   # Or for production
   make prod-up
   ```

### Manual Setup

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up PostgreSQL and Redis**
   ```bash
   # Using Docker for databases only
   docker run -d --name postgres -p 5432:5432 \
     -e POSTGRES_DB=news_feed \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     postgres:15-alpine

   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

3. **Run migrations**
   ```bash
   make migrate-up
   ```

4. **Start the application**
   ```bash
   go run ./cmd/server
   ```

## 🐳 Docker Commands

| Command | Description |
|---------|-------------|
| `make dev-up` | Start development environment |
| `make dev-down` | Stop development environment |
| `make prod-up` | Start production environment |
| `make prod-down` | Stop production environment |

## 📊 API Documentation

### Base URL
- **Development**: `http://localhost:8080`
- **Production**: `https://your-domain.com`

### Authentication
Currently, the API is open. Authentication can be added by implementing JWT middleware in the handlers.

### Endpoints

#### Posts API

##### Get All Posts
```http
GET /api/v1/posts?limit=10&offset=0&category=technology
```

**Query Parameters:**
- `limit` (optional): Number of posts to return (default: 10, max: 100)
- `offset` (optional): Number of posts to skip (default: 0)
- `category` (optional): Filter by category

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Breaking Tech News",
      "content": "Content here...",
      "category": "technology",
      "source_url": "https://example.com/news",
      "published_at": "2025-08-17T10:00:00Z",
      "created_at": "2025-08-17T10:00:00Z"
    }
  ],
  "message": "Posts retrieved successfully"
}
```

##### Get Post by ID
```http
GET /api/v1/posts/{id}
```

##### Create New Post
```http
POST /api/v1/posts
Content-Type: application/json

{
  "title": "News Title",
  "content": "News content...",
  "category": "technology",
  "source_url": "https://example.com/source",
  "published_at": "2025-08-17T10:00:00Z"
}
```

##### Update Post
```http
PUT /api/v1/posts/{id}
Content-Type: application/json

{
  "title": "Updated Title",
  "content": "Updated content...",
  "category": "technology"
}
```

##### Delete Post
```http
DELETE /api/v1/posts/{id}
```

#### Aggregator API

##### Trigger Manual Aggregation
```http
POST /api/v1/aggregation/trigger
```

##### Get Aggregation Status
```http
GET /api/v1/aggregation/status
```

#### Scheduler API

##### Get Scheduled Jobs
```http
GET /api/v1/scheduler/jobs
```

##### Create Scheduled Job
```http
POST /api/v1/scheduler/jobs
Content-Type: application/json

{
  "name": "hourly_tech_news",
  "cron_expression": "0 * * * *",
  "job_type": "aggregation",
  "enabled": true
}
```

### Error Responses

All endpoints return errors in the following format:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {}
  }
}
```

## 🧪 Testing

### Run All Tests
```bash
make test-unit
```

### Run Specific Package Tests
```bash
go test -v ./internal/service/...
go test -v ./internal/handler/...
go test -v ./internal/repository/...
```

## 📁 Project Structure

```
news-feed-system/
├── cmd/
│   └── server/                 # Application entry point
│       └── main.go
├── internal/                   # Private application code
│   ├── bootstrap/              # Application bootstrap
│   ├── config/                 # Configuration management
│   ├── handler/                # HTTP handlers
│   ├── model/                  # Data models
│   ├── repository/             # Data access layer
│   └── service/                # Business logic
├── pkg/                        # Public packages
│   ├── database/               # Database connection utilities
│   ├── logger/                 # Logging utilities
│   ├── response/               # Response utilities
│   └── validator/              # Validation utilities
├── migrations/                 # Database migrations
├── docs/                       # API documentation (Swagger)
├── docker-compose.yml          # Production Docker setup
├── docker-compose.dev.yml      # Development Docker setup
├── Dockerfile                  # Production Dockerfile
├── Makefile                    # Build and development commands
└── README.md                   # This file
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Application environment | `development` |
| `SERVER_PORT` | Application port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database name | `news_feed` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `NEWS_API_KEY` | News API key | (required) |
| `LOG_LEVEL` | Logging level | `info` |

## 🚀 Deployment

### Docker Production Deployment
1. **Configure production environment**
   ```bash
   cp .env.example .env
   # Edit .env with production values
   ```

2. **Deploy with Docker Compose**
   ```bash
   docker-compose up -d --build
   ```

3. **Check deployment**
   ```bash
   docker-compose ps
   ```

### Logs
```bash
# View application logs
docker-compose logs -f app

# View all service logs
docker-compose logs -f
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
