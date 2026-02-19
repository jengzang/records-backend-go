# API Design Overview

## Design Principles

### 1. RESTful Architecture

- Use standard HTTP methods (GET, POST, PUT, DELETE)
- Resource-based URLs
- Stateless requests
- Standard HTTP status codes

### 2. Consistency

- Consistent naming conventions (snake_case for JSON fields)
- Consistent response format across all endpoints
- Consistent error handling
- Consistent pagination

### 3. Performance

- Pagination for large result sets (default: 100 items, max: 1000)
- Caching headers (ETag, Last-Modified)
- Compression (gzip)
- Rate limiting (3 req/s per IP)

### 4. Security

- JWT authentication for write operations
- Input validation and sanitization
- SQL injection prevention (parameterized queries)
- CORS configuration
- HTTPS only in production

## URL Structure

```
/api/v1/{module}/{resource}/{id}/{action}
```

### Examples

```
GET  /api/v1/tracks/points?start_time=1234567890&end_time=1234567900
GET  /api/v1/tracks/statistics/footprint?year=2025
GET  /api/v1/tracks/stays?province=广东省&city=深圳市
POST /api/v1/tracks/import
GET  /api/v1/keyboard/statistics/daily?start_date=20250101&end_date=20250131
```

## Request Format

### Query Parameters

- Use snake_case for parameter names
- Use ISO 8601 for dates (YYYY-MM-DD)
- Use Unix timestamps for precise times
- Use pagination parameters: `page`, `page_size`, `offset`, `limit`

### Request Body

```json
{
  "data": {
    "field1": "value1",
    "field2": "value2"
  },
  "options": {
    "validate": true,
    "async": false
  }
}
```

## Response Format

### Success Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [...],
    "total": 1000,
    "page": 1,
    "page_size": 100
  },
  "timestamp": 1234567890,
  "request_id": "uuid-string"
}
```

### Error Response

```json
{
  "code": 400,
  "message": "Invalid parameter: start_time",
  "error": {
    "field": "start_time",
    "reason": "must be a valid Unix timestamp",
    "value": "invalid"
  },
  "timestamp": 1234567890,
  "request_id": "uuid-string"
}
```

## Pagination

### Offset-Based Pagination

```
GET /api/v1/tracks/points?offset=0&limit=100
```

Response includes:
```json
{
  "data": {
    "items": [...],
    "total": 1000,
    "offset": 0,
    "limit": 100,
    "has_more": true
  }
}
```

### Cursor-Based Pagination (for time-series data)

```
GET /api/v1/tracks/points?cursor=1234567890&limit=100
```

Response includes:
```json
{
  "data": {
    "items": [...],
    "next_cursor": 1234567990,
    "has_more": true
  }
}
```

## Filtering

### Common Filters

- `start_time`, `end_time` - Time range (Unix timestamp)
- `start_date`, `end_date` - Date range (YYYY-MM-DD)
- `province`, `city`, `county` - Administrative divisions
- `min_value`, `max_value` - Numeric range

### Example

```
GET /api/v1/tracks/points?start_time=1234567890&end_time=1234567900&province=广东省&city=深圳市
```

## Sorting

```
GET /api/v1/tracks/points?sort_by=dataTime&sort_order=desc
```

- `sort_by` - Field name
- `sort_order` - `asc` or `desc`

## Rate Limiting

- **Limit:** 3 requests per second per IP
- **Headers:**
  - `X-RateLimit-Limit: 3`
  - `X-RateLimit-Remaining: 2`
  - `X-RateLimit-Reset: 1234567890`

- **Response when exceeded:**
```json
{
  "code": 429,
  "message": "Rate limit exceeded",
  "error": {
    "retry_after": 1
  }
}
```

## Caching

### ETag

```
GET /api/v1/tracks/statistics/footprint?year=2025
Response Headers:
  ETag: "abc123"
  Cache-Control: max-age=3600

Subsequent Request:
  If-None-Match: "abc123"
Response:
  304 Not Modified
```

### Last-Modified

```
Response Headers:
  Last-Modified: Wed, 21 Oct 2025 07:28:00 GMT
  Cache-Control: max-age=3600

Subsequent Request:
  If-Modified-Since: Wed, 21 Oct 2025 07:28:00 GMT
Response:
  304 Not Modified
```

## Authentication

### JWT Token

```
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "password"
}

Response:
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1234567890
  }
}

Authenticated Request:
GET /api/v1/tracks/import
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Public vs Protected Endpoints

**Public (no auth required):**
- GET endpoints for reading data
- Statistics and aggregations

**Protected (auth required):**
- POST /api/v1/tracks/import
- DELETE /api/v1/tracks/points/{id}
- PUT /api/v1/tracks/points/{id}

## Error Handling

### Error Codes

- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `422` - Unprocessable Entity (validation failed)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error
- `503` - Service Unavailable (maintenance mode)

### Error Response Format

```json
{
  "code": 400,
  "message": "Validation failed",
  "error": {
    "field": "start_time",
    "reason": "must be less than end_time",
    "value": 1234567900
  },
  "timestamp": 1234567890,
  "request_id": "uuid-string"
}
```

## Versioning

- URL-based versioning: `/api/v1/`, `/api/v2/`
- Breaking changes require new version
- Old versions supported for 6 months after new version release

## CORS

```
Access-Control-Allow-Origin: https://record.yzup.top
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

## Compression

- gzip compression for responses > 1KB
- Request header: `Accept-Encoding: gzip`
- Response header: `Content-Encoding: gzip`

## Health Check

```
GET /api/v1/health

Response:
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": 3600,
  "database": "ok",
  "timestamp": 1234567890
}
```
