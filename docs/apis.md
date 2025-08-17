# News Feed System API Documentation

## Overview
This document describes the REST API endpoints for the News Feed System. The system provides endpoints for managing news posts, triggering aggregations, and monitoring scheduled jobs.

## Base URL
```
http://localhost:8080/api/v1
```

## Response Format
All API responses follow a consistent format:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {...},
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {...}
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

## Authentication
Currently, no authentication is required. This will be added in future versions.

## Endpoints

### Health Check

#### GET /health
Check the health status of the service and database connections.

**Response:**
```json
{
  "status": "healthy",
  "service": "news-feed-system",
  "version": "1.0.0"
}
```

---

## Post Management

### Create Post

#### POST /api/v1/posts
Create a new news post.

**Request Body:**
```json
{
  "title": "Breaking News: Tech Innovation",
  "description": "A brief description of the news article",
  "content": "Full content of the article...",
  "url": "https://example.com/article",
  "source": "TechCrunch",
  "category": "technology",
  "image_url": "https://example.com/image.jpg",
  "published_at": "2024-01-20T10:00:00Z"
}
```

**Validation Rules:**
- `title`: Required, 1-500 characters
- `url`: Required, valid URL, max 1000 characters
- `source`: Required, 1-100 characters
- `category`: Optional, max 50 characters
- `image_url`: Optional, valid URL, max 1000 characters

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Post created successfully",
  "data": {
    "id": 123,
    "title": "Breaking News: Tech Innovation",
    "description": "A brief description...",
    "url": "https://example.com/article",
    "source": "TechCrunch",
    "category": "technology",
    "image_url": "https://example.com/image.jpg",
    "published_at": "2024-01-20T10:00:00Z",
    "created_at": "2024-01-20T10:30:00Z",
    "updated_at": "2024-01-20T10:30:00Z"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Get Post

#### GET /api/v1/posts/{id}
Retrieve a specific post by ID.

**Parameters:**
- `id` (path): Post ID (integer)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "title": "Breaking News: Tech Innovation",
    // ... other post fields
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### List Posts

#### GET /api/v1/posts
Retrieve a paginated list of posts with optional filtering and search.

**Query Parameters:**
- `page` (optional): Page number (default: 1, min: 1)
- `limit` (optional): Items per page (default: 20, min: 1, max: 100)
- `category` (optional): Filter by category
- `source` (optional): Filter by source
- `search` (optional): Search in title, description, and content

**Examples:**
```
GET /api/v1/posts?page=1&limit=10
GET /api/v1/posts?category=technology&page=2
GET /api/v1/posts?source=CNN&limit=5
GET /api/v1/posts?search=artificial%20intelligence
GET /api/v1/posts?search=AI&category=technology&page=1&limit=20
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": 123,
        "title": "Breaking News: Tech Innovation",
        // ... other post fields
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    }
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Update Post

#### PUT /api/v1/posts/{id}
Update an existing post.

**Parameters:**
- `id` (path): Post ID (integer)

**Request Body:**
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "content": "Updated content",
  "category": "updated-category",
  "image_url": "https://example.com/new-image.jpg"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Post updated successfully",
  "data": {
    "id": 123,
    "title": "Updated Title",
    // ... updated post fields
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Delete Post

#### DELETE /api/v1/posts/{id}
Delete a post.

**Parameters:**
- `id` (path): Post ID (integer)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Post deleted successfully",
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Filter by Category

#### GET /api/v1/posts/category/{category}
Get posts from a specific category.

**Parameters:**
- `category` (path): Category name
- `page` (query, optional): Page number
- `limit` (query, optional): Items per page

**Example:**
```
GET /api/v1/posts/category/technology?page=1&limit=10
```

### Filter by Source

#### GET /api/v1/posts/source/{source}
Get posts from a specific source.

**Parameters:**
- `source` (path): Source name
- `page` (query, optional): Page number
- `limit` (query, optional): Items per page

**Example:**
```
GET /api/v1/posts/source/CNN?page=1&limit=10
```

### Search Posts

#### GET /api/v1/posts/search
Search posts by query string.

**Query Parameters:**
- `q` (required): Search query
- `page` (optional): Page number
- `limit` (optional): Items per page
- `category` (optional): Additional category filter
- `source` (optional): Additional source filter

**Examples:**
```
GET /api/v1/posts/search?q=artificial%20intelligence
GET /api/v1/posts/search?q=AI&category=technology&page=2
```

---

## News Aggregation

### Get Aggregation Status

#### GET /api/v1/aggregation/status
Check the status of the aggregation service.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "status": "operational",
    "timestamp": "2024-01-20T10:30:00Z",
    "message": "Aggregation service is running"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Trigger Complete Aggregation

#### POST /api/v1/aggregation/trigger
Trigger a complete news aggregation from all sources and categories.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Aggregation completed successfully",
  "data": {
    "total_fetched": 250,
    "total_created": 45,
    "total_duplicates": 180,
    "total_errors": 25,
    "duration": "2m30s",
    "categories": {
      "technology": {
        "fetched": 50,
        "created": 12,
        "duplicates": 35,
        "errors": 3
      }
    },
            "sources": {
      "cnn": {
        "fetched": 40,
        "created": 8,
        "duplicates": 30,
        "errors": 2
      }
    },
    "errors": [
      "Failed to fetch from source X: rate limit exceeded"
    ]
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Trigger Top Headlines Aggregation

#### POST /api/v1/aggregation/trigger/headlines
Trigger aggregation of top headlines from all categories.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Top headlines aggregation completed successfully",
  "data": {
    "total_fetched": 150,
    "total_created": 25,
    "total_duplicates": 120,
    "total_errors": 5,
    "duration": "1m45s"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Trigger Category Aggregation

#### POST /api/v1/aggregation/trigger/categories
Trigger aggregation from specific categories.

**Request Body (Optional):**
```json
{
  "categories": ["technology", "business", "science"]
}
```

If no body is provided, all default categories will be used.

**Available Categories:**
- general
- business
- entertainment
- health
- science
- sports
- technology

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Category aggregation completed successfully",
  "categories": ["technology", "business", "science"],
  "data": {
    "total_fetched": 120,
    "total_created": 20,
    "total_duplicates": 95,
    "total_errors": 5,
    "duration": "2m10s",
    "categories": {
      "technology": {
        "fetched": 50,
        "created": 8,
        "duplicates": 40,
        "errors": 2
      },
      "business": {
        "fetched": 40,
        "created": 7,
        "duplicates": 30,
        "errors": 3
      },
      "science": {
        "fetched": 30,
        "created": 5,
        "duplicates": 25,
        "errors": 0
      }
    }
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Trigger Source Aggregation

#### POST /api/v1/aggregation/trigger/sources
Trigger aggregation from specific news sources.

**Request Body (Optional):**
```json
{
  "sources": ["cnn", "bbc-news", "reuters"]
}
```

If no body is provided, all default sources will be used.

**Default Sources:**
- bbc-news
- cnn
- reuters
- associated-press
- the-verge
- techcrunch
- ars-technica
- hacker-news
- the-wall-street-journal
- bloomberg

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Source aggregation completed successfully",
  "sources": ["cnn", "bbc-news", "reuters"],
  "data": {
    "total_fetched": 90,
    "total_created": 15,
    "total_duplicates": 70,
    "total_errors": 5,
    "duration": "1m30s",
    "sources": {
      "cnn": {
        "fetched": 30,
        "created": 5,
        "duplicates": 23,
        "errors": 2
      },
      "bbc-news": {
        "fetched": 35,
        "created": 7,
        "duplicates": 27,
        "errors": 1
      },
      "reuters": {
        "fetched": 25,
        "created": 3,
        "duplicates": 20,
        "errors": 2
      }
    }
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

---

## Scheduler Management

### Get Scheduler Status

#### GET /api/v1/scheduler/status
Get the status of the scheduler service and all jobs.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "scheduler_running": true,
    "jobs_count": 3,
    "timestamp": "2024-01-20T10:30:00Z",
    "jobs": {
      "top-headlines": {
        "name": "top-headlines",
        "interval": "30m0s",
        "last_run": "2024-01-20T10:00:00Z",
        "next_run": "2024-01-20T10:30:00Z",
        "run_count": 48,
        "error_count": 2,
        "last_error": "",
        "is_running": false,
        "average_run_time": "45s"
      },
      "category-aggregation": {
        "name": "category-aggregation",
        "interval": "2h0m0s",
        "last_run": "2024-01-20T08:00:00Z",
        "next_run": "2024-01-20T10:00:00Z",
        "run_count": 12,
        "error_count": 0,
        "last_error": "",
        "is_running": false,
        "average_run_time": "2m15s"
      },
      "source-aggregation": {
        "name": "source-aggregation",
        "interval": "4h0m0s",
        "last_run": "2024-01-20T06:00:00Z",
        "next_run": "2024-01-20T10:00:00Z",
        "run_count": 6,
        "error_count": 1,
        "last_error": "timeout exceeded",
        "is_running": false,
        "average_run_time": "3m30s"
      }
    }
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Get Jobs List

#### GET /api/v1/scheduler/jobs
Get a list of all scheduled jobs with their status.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "jobs": {
      "top-headlines": {
        "name": "top-headlines",
        "interval": "30m0s",
        // ... job details
      }
    },
    "count": 3,
    "timestamp": "2024-01-20T10:30:00Z"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

### Trigger Job

#### POST /api/v1/scheduler/jobs/{name}/trigger
Acknowledge a job trigger request. Note: This doesn't immediately execute the job but provides information about when it will run next.

**Parameters:**
- `name` (path): Job name (e.g., "top-headlines", "category-aggregation", "source-aggregation")

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "message": "Job trigger acknowledged",
    "job_name": "top-headlines",
    "note": "Job will run according to its schedule. For immediate execution, use specific aggregation endpoints.",
    "next_run": "2024-01-20T10:30:00Z",
    "timestamp": "2024-01-20T10:30:00Z"
  },
  "timestamp": "2024-01-20T10:30:00Z"
}
```

---

## Pagination

All endpoints that return lists support pagination:

### Default Values
- `page`: 1
- `limit`: 20

### Maximum Values
- `limit`: 100 (requests for higher limits will be capped at 100)

### Pagination Response
```json
{
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Search Functionality

The search endpoint supports full-text search across:
- Post titles
- Post descriptions  
- Post content

### Search Features
- Case-insensitive search
- Partial word matching
- Can be combined with category and source filters
- Supports pagination

### Search Examples
```
# Basic search
GET /api/v1/posts/search?q=artificial intelligence

# Search with filters
GET /api/v1/posts/search?q=AI&category=technology

# Search with pagination
GET /api/v1/posts/search?q=machine learning&page=2&limit=10
```