# Tracks API Endpoints

## Overview

RESTful API endpoints for GPS trajectory data access and analysis.

**Status:** ❌ Not Implemented

**Base URL:** `https://record.yzup.top/api/v1/tracks`

## Authentication

**Public Endpoints:** GET requests (read-only)
**Protected Endpoints:** POST/PUT/DELETE requests (require JWT)

```
Authorization: Bearer <JWT_TOKEN>
```

## Endpoints

### 1. Get GPS Points

```
GET /api/v1/tracks/points
```

**Description:** Retrieve GPS trajectory points with filtering and pagination.

**Query Parameters:**
- `start_time` (integer) - Start time (Unix timestamp)
- `end_time` (integer) - End time (Unix timestamp)
- `province` (string) - Filter by province
- `city` (string) - Filter by city
- `county` (string) - Filter by county
- `min_accuracy` (float) - Minimum accuracy threshold
- `exclude_outliers` (boolean) - Exclude flagged outliers (default: true)
- `offset` (integer) - Pagination offset (default: 0)
- `limit` (integer) - Page size (default: 100, max: 1000)

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "dataTime": 1737552138,
        "longitude": 114.0579,
        "latitude": 22.5431,
        "heading": 90.0,
        "accuracy": 10.0,
        "speed": 5.5,
        "distance": 100.0,
        "altitude": 500.0,
        "time_visually": "2025/01/22 21:42:18.000",
        "time": "20250122214218",
        "province": "广东省",
        "city": "深圳市",
        "county": "南山区",
        "town": "南头街道",
        "village": "南头社区"
      }
    ],
    "total": 1000,
    "offset": 0,
    "limit": 100,
    "has_more": true
  },
  "timestamp": 1737552138
}
```

### 2. Get Footprint Statistics

```
GET /api/v1/tracks/statistics/footprint
```

**Description:** Get footprint statistics by administrative divisions.

**Query Parameters:**
- `level` (string) - Admin level: 'province', 'city', 'county', 'town' (default: 'city')
- `year` (integer) - Filter by year
- `month` (integer) - Filter by month
- `province` (string) - Filter by province
- `city` (string) - Filter by city

**Response:**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "province": "广东省",
        "city": "深圳市",
        "visit_count": 150,
        "point_count": 50000,
        "total_duration": 2592000,
        "first_visit": 1704067200,
        "last_visit": 1737552138
      }
    ],
    "total": 50
  }
}
```

### 3. Get Stay Statistics

```
GET /api/v1/tracks/statistics/stays
```

**Description:** Get stay detection statistics.

**Query Parameters:**
- `min_duration` (integer) - Minimum stay duration in seconds (default: 7200)
- `province` (string) - Filter by province
- `city` (string) - Filter by city
- `county` (string) - Filter by county
- `start_time` (integer) - Start time filter
- `end_time` (integer) - End time filter

**Response:**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "start_time": 1737552138,
        "end_time": 1737559338,
        "duration": 7200,
        "center_lon": 114.0579,
        "center_lat": 22.5431,
        "radius": 50.0,
        "point_count": 100,
        "province": "广东省",
        "city": "深圳市",
        "county": "南山区",
        "stay_type": "home",
        "confidence": 0.95
      }
    ],
    "total": 200
  }
}
```

### 4. Get Trips

```
GET /api/v1/tracks/trips
```

**Description:** Get constructed trips.

**Query Parameters:**
- `start_time` (integer) - Start time filter
- `end_time` (integer) - End time filter
- `transport_mode` (string) - Filter by mode: 'walk', 'bike', 'car', 'train', 'plane'
- `min_distance` (float) - Minimum trip distance in meters
- `offset` (integer) - Pagination offset
- `limit` (integer) - Page size

**Response:**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "start_time": 1737552138,
        "end_time": 1737555738,
        "duration": 3600,
        "distance": 50000,
        "origin_lon": 114.0579,
        "origin_lat": 22.5431,
        "origin_city": "深圳市",
        "dest_lon": 113.2644,
        "dest_lat": 23.1291,
        "dest_city": "广州市",
        "transport_mode": "train",
        "mode_confidence": 0.92
      }
    ],
    "total": 500
  }
}
```

### 5. Get Extreme Events

```
GET /api/v1/tracks/statistics/extremes
```

**Description:** Get extreme events (fastest speed, longest trip, etc.).

**Query Parameters:**
- `event_type` (string) - 'fastest_speed', 'longest_trip', 'highest_altitude', 'most_visited'
- `year` (integer) - Filter by year
- `limit` (integer) - Number of results (default: 10)

**Response:**
```json
{
  "code": 0,
  "data": {
    "event_type": "fastest_speed",
    "items": [
      {
        "id": 12345,
        "dataTime": 1737552138,
        "speed": 850.5,
        "longitude": 114.0579,
        "latitude": 22.5431,
        "transport_mode": "plane",
        "confidence": 0.98
      }
    ]
  }
}
```

### 6. Get Administrative Crossings

```
GET /api/v1/tracks/statistics/crossings
```

**Description:** Get statistics on administrative boundary crossings.

**Query Parameters:**
- `level` (string) - 'province', 'city', 'county'
- `year` (integer) - Filter by year
- `month` (integer) - Filter by month

**Response:**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "from_province": "广东省",
        "from_city": "深圳市",
        "to_province": "广东省",
        "to_city": "广州市",
        "crossing_count": 25,
        "first_crossing": 1704067200,
        "last_crossing": 1737552138
      }
    ]
  }
}
```

### 7. Import GPS Data (Protected)

```
POST /api/v1/tracks/import
```

**Description:** Import GPS data from Excel file.

**Authentication:** Required

**Request:**
```
Content-Type: multipart/form-data

file: <Excel file>
options: {
  "filter_step_type": 0,
  "deduplicate": true,
  "geocode": true
}
```

**Response:**
```json
{
  "code": 0,
  "message": "Import started",
  "data": {
    "task_id": "uuid-string",
    "status": "processing",
    "estimated_time": 120
  }
}
```

### 8. Get Import Status (Protected)

```
GET /api/v1/tracks/import/{task_id}
```

**Description:** Check import task status.

**Authentication:** Required

**Response:**
```json
{
  "code": 0,
  "data": {
    "task_id": "uuid-string",
    "status": "completed",
    "progress": 100,
    "records_imported": 100000,
    "records_failed": 5,
    "started_at": 1737552138,
    "completed_at": 1737552258
  }
}
```

### 9. Delete Points (Protected)

```
DELETE /api/v1/tracks/points
```

**Description:** Delete GPS points by criteria.

**Authentication:** Required

**Request:**
```json
{
  "start_time": 1737552138,
  "end_time": 1737638538,
  "province": "广东省",
  "city": "深圳市"
}
```

**Response:**
```json
{
  "code": 0,
  "message": "Deleted successfully",
  "data": {
    "deleted_count": 1000
  }
}
```

### 10. Update Point (Protected)

```
PUT /api/v1/tracks/points/{id}
```

**Description:** Update a specific GPS point.

**Authentication:** Required

**Request:**
```json
{
  "outlier_flag": 1,
  "outlier_reasons": ["speed_spike", "accuracy_low"],
  "qa_status": "rejected"
}
```

**Response:**
```json
{
  "code": 0,
  "message": "Updated successfully",
  "data": {
    "id": 1,
    "updated_fields": ["outlier_flag", "outlier_reasons", "qa_status"]
  }
}
```

## Visualization Endpoints

### 11. Get Heatmap Data

```
GET /api/v1/viz/heatmap
```

**Description:** Get heatmap data with normalized intensity scores for visualization.

**Query Parameters:**
- `level` (integer) - Grid level (1-15, default: 3)
  - Level 1: Province level (~100km cells)
  - Level 2: City level (~50km cells)
  - Level 3: District level (~10km cells)
  - Level 4: Town level (~5km cells)
  - Level 5: Road level (~1km cells)
- `metric` (string) - Metric to visualize (default: 'point_count')
  - `point_count` - Number of GPS points in each cell
  - `duration` - Total duration in seconds
  - `visit_count` - Number of distinct visits
- `minLat` (float) - Minimum latitude for bounding box
- `maxLat` (float) - Maximum latitude for bounding box
- `minLon` (float) - Minimum longitude for bounding box
- `maxLon` (float) - Maximum longitude for bounding box

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "points": [
      {
        "lat": 22.5432,
        "lng": 113.9234,
        "intensity": 0.85,
        "value": 1250,
        "metric": "point_count"
      },
      {
        "lat": 22.5567,
        "lng": 113.9456,
        "intensity": 0.42,
        "value": 620,
        "metric": "point_count"
      }
    ],
    "count": 150,
    "max_value": 1500,
    "min_value": 10,
    "metric": "point_count",
    "grid_level": 3
  }
}
```

**Notes:**
- Intensity values are normalized to [0, 1] range for easy visualization
- Maximum 10,000 cells per request for performance
- Empty result if no grid data exists (run grid_system analyzer first)

**Example Usage:**
```bash
# Basic heatmap (point density)
curl "http://localhost:8080/api/v1/viz/heatmap?level=3&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"

# Duration-based heatmap
curl "http://localhost:8080/api/v1/viz/heatmap?level=3&metric=duration&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"

# Visit count heatmap
curl "http://localhost:8080/api/v1/viz/heatmap?level=3&metric=visit_count&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"
```

### 12. Get Grid Cells

```
GET /api/v1/viz/grid-cells
```

**Description:** Get raw grid cell data with detailed statistics.

**Query Parameters:**
- `level` (integer) - Grid level (1-15, default: 3)
- `minLat` (float) - Minimum latitude for bounding box
- `maxLat` (float) - Maximum latitude for bounding box
- `minLon` (float) - Minimum longitude for bounding box
- `maxLon` (float) - Maximum longitude for bounding box
- `minDensity` (integer) - Minimum point count filter

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "grid_id": "L3_1234_5678",
        "level": 3,
        "center_lat": 22.5432,
        "center_lon": 113.9234,
        "min_lat": 22.5400,
        "max_lat": 22.5464,
        "min_lon": 113.9200,
        "max_lon": 113.9268,
        "point_count": 1250,
        "visit_count": 45,
        "total_duration_seconds": 86400,
        "first_visit": 1704067200,
        "last_visit": 1737552138,
        "modes_json": "[\"walk\",\"bike\",\"car\"]",
        "created_at": "2026-02-22T00:00:00Z",
        "updated_at": "2026-02-22T00:00:00Z"
      }
    ],
    "count": 150
  }
}
```

## Error Responses

### 400 Bad Request

```json
{
  "code": 400,
  "message": "Invalid parameter: start_time",
  "error": {
    "field": "start_time",
    "reason": "must be a valid Unix timestamp",
    "value": "invalid"
  }
}
```

### 401 Unauthorized

```json
{
  "code": 401,
  "message": "Unauthorized",
  "error": {
    "reason": "Missing or invalid JWT token"
  }
}
```

### 404 Not Found

```json
{
  "code": 404,
  "message": "Resource not found",
  "error": {
    "resource": "point",
    "id": 12345
  }
}
```

### 429 Too Many Requests

```json
{
  "code": 429,
  "message": "Rate limit exceeded",
  "error": {
    "retry_after": 1,
    "limit": 3,
    "window": 1
  }
}
```

## Rate Limiting

- **Limit:** 3 requests per second per IP
- **Headers:**
  - `X-RateLimit-Limit: 3`
  - `X-RateLimit-Remaining: 2`
  - `X-RateLimit-Reset: 1737552138`

## Caching

- **ETag:** Supported for GET requests
- **Cache-Control:** `max-age=3600` for statistics
- **Last-Modified:** Supported for data endpoints

## Pagination

- **Default:** 100 items per page
- **Maximum:** 1000 items per page
- **Cursor-based:** For time-series data

## Related Documentation

- `data-pipeline.md` - Data pipeline
- `database-schema.md` - Database schema
- `processing-scripts.md` - Processing scripts
- `implementation-guide.md` - Implementation guide
- `tracks/docs/api-integration.md` - Frontend integration guide
