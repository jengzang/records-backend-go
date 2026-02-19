# Phase 5: Go Backend API Endpoints

## Overview

This document describes all REST API endpoints implemented in Phase 5 for querying trajectory analysis data.

**Base URL:** `http://localhost:8080/api/v1`

**Response Format:**
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**Error Response:**
```json
{
  "code": 400,
  "message": "Invalid query parameters",
  "data": {
    "error": "detailed error message"
  }
}
```

## 1. Segment Endpoints

### GET /tracks/segments

Retrieve behavior segments with filtering and pagination.

**Query Parameters:**
- `mode` (string, optional): Filter by transport mode (WALK, CAR, TRAIN, FLIGHT, STAY, UNKNOWN)
- `startTime` (int64, optional): Unix timestamp - filter segments starting after this time
- `endTime` (int64, optional): Unix timestamp - filter segments ending before this time
- `province` (string, optional): Filter by province
- `city` (string, optional): Filter by city
- `county` (string, optional): Filter by county
- `minDistance` (float64, optional): Minimum distance in meters
- `minDuration` (int64, optional): Minimum duration in seconds
- `minConfidence` (float64, optional): Minimum confidence (0-1)
- `page` (int, optional, default: 1): Page number
- `pageSize` (int, optional, default: 100, max: 1000): Results per page

**Example Request:**
```
GET /api/v1/tracks/segments?mode=CAR&minDistance=10000&page=1&pageSize=50
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "mode": "CAR",
        "start_point_id": 1000,
        "end_point_id": 1500,
        "start_time": 1705910400,
        "end_time": 1705914000,
        "duration_seconds": 3600,
        "distance_meters": 45000,
        "start_lat": 39.9042,
        "start_lon": 116.4074,
        "end_lat": 40.0042,
        "end_lon": 116.5074,
        "avg_speed_kmh": 45.5,
        "max_speed_kmh": 80.2,
        "confidence": 0.92,
        "reason_codes": "[\"SPEED_RANGE\", \"DURATION\"]",
        "province": "北京市",
        "city": "北京市",
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "total": 1234,
    "page": 1,
    "pageSize": 50,
    "totalPages": 25
  }
}
```

### GET /tracks/segments/:id

Retrieve a single segment by ID.

**Path Parameters:**
- `id` (int64, required): Segment ID

**Example Request:**
```
GET /api/v1/tracks/segments/1
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "mode": "CAR",
    ...
  }
}
```

## 2. Stay Endpoints

### GET /tracks/stays

Retrieve stay segments with filtering and pagination.

**Query Parameters:**
- `stayType` (string, optional): SPATIAL or ADMIN
- `stayCategory` (string, optional): HOME, WORK, TRANSIT, VISIT, UNKNOWN
- `minDuration` (int64, optional): Minimum duration in seconds
- `province` (string, optional): Filter by province
- `city` (string, optional): Filter by city
- `county` (string, optional): Filter by county
- `startTime` (int64, optional): Unix timestamp
- `endTime` (int64, optional): Unix timestamp
- `minConfidence` (float64, optional): Minimum confidence (0-1)
- `page` (int, optional, default: 1)
- `pageSize` (int, optional, default: 100, max: 1000)

**Example Request:**
```
GET /api/v1/tracks/stays?stayCategory=HOME&minDuration=7200
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "stay_type": "SPATIAL",
        "stay_category": "HOME",
        "start_time": 1705910400,
        "end_time": 1705939200,
        "duration_seconds": 28800,
        "center_lat": 39.9042,
        "center_lon": 116.4074,
        "radius_meters": 50,
        "point_count": 120,
        "province": "北京市",
        "city": "北京市",
        "confidence": 0.95,
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "total": 456,
    "page": 1,
    "pageSize": 100,
    "totalPages": 5
  }
}
```

### GET /tracks/stays/:id

Retrieve a single stay segment by ID.

## 3. Trip Endpoints

### GET /tracks/trips

Retrieve trips with filtering and pagination.

**Query Parameters:**
- `startTime` (int64, optional): Unix timestamp
- `endTime` (int64, optional): Unix timestamp
- `originCity` (string, optional): Origin city name
- `destCity` (string, optional): Destination city name
- `minDistance` (float64, optional): Minimum distance in meters
- `primaryMode` (string, optional): Primary transport mode
- `tripType` (string, optional): COMMUTE, ROUND_TRIP, ONE_WAY, MULTI_STOP
- `page` (int, optional, default: 1)
- `pageSize` (int, optional, default: 100, max: 1000)

**Example Request:**
```
GET /api/v1/tracks/trips?tripType=COMMUTE&primaryMode=CAR
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "date": "2025-01-22",
        "start_time": 1705910400,
        "end_time": 1705914000,
        "duration_seconds": 3600,
        "origin_stay_id": 10,
        "dest_stay_id": 20,
        "distance_meters": 45000,
        "primary_mode": "CAR",
        "modes_json": "[\"CAR\", \"WALK\"]",
        "trip_type": "COMMUTE",
        "origin_city": "北京市",
        "dest_city": "北京市",
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "total": 789,
    "page": 1,
    "pageSize": 100,
    "totalPages": 8
  }
}
```

### GET /tracks/trips/:id

Retrieve a single trip by ID.

## 4. Grid Cell Endpoints

### GET /viz/grid-cells

Retrieve grid cells for heatmap visualization.

**Query Parameters:**
- `level` (int, optional, default: 3): Grid level (1-5)
  - 1: National/Province (~40km x 20km)
  - 2: City (~5km x 5km)
  - 3: District (~1.2km x 0.6km) - default
  - 4: Street (~150m x 150m)
  - 5: Road (~40m x 20m)
- `minLat` (float64, optional): Bounding box minimum latitude
- `maxLat` (float64, optional): Bounding box maximum latitude
- `minLon` (float64, optional): Bounding box minimum longitude
- `maxLon` (float64, optional): Bounding box maximum longitude
- `minDensity` (int, optional): Minimum point count

**Example Request:**
```
GET /api/v1/viz/grid-cells?level=3&minLat=39.8&maxLat=40.0&minLon=116.3&maxLon=116.5
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "grid_id": "L3_wx4g0e",
        "level": 3,
        "center_lat": 39.9042,
        "center_lon": 116.4074,
        "min_lat": 39.8982,
        "max_lat": 39.9102,
        "min_lon": 116.4014,
        "max_lon": 116.4134,
        "point_count": 1234,
        "visit_count": 45,
        "first_visit": 1705910400,
        "last_visit": 1706515200,
        "total_duration_seconds": 86400,
        "modes_json": "[\"WALK\", \"CAR\"]",
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "count": 150
  }
}
```

**Note:** Limited to 10,000 cells per request for performance.

## 5. Statistics Endpoints

### GET /stats/footprint/rankings

Retrieve footprint statistics rankings.

**Query Parameters:**
- `statType` (string, optional, default: "PROVINCE"): PROVINCE, CITY, COUNTY, TOWN, GRID
- `timeRange` (string, optional, default: "all"): all, YYYY, YYYY-MM, YYYY-MM-DD
- `orderBy` (string, optional, default: "points"): points, visits, duration, distance
- `limit` (int, optional, default: 100, max: 1000): Maximum results

**Example Request:**
```
GET /api/v1/stats/footprint/rankings?statType=CITY&orderBy=visits&limit=50
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "stat_type": "CITY",
        "stat_key": "北京市",
        "time_range": "all",
        "province": "北京市",
        "city": "北京市",
        "point_count": 123456,
        "visit_count": 456,
        "total_distance_meters": 5000000,
        "total_duration_seconds": 864000,
        "first_visit_time": 1640995200,
        "last_visit_time": 1706515200,
        "rank_by_points": 1,
        "rank_by_visits": 1,
        "rank_by_duration": 1,
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "count": 50
  }
}
```

### GET /stats/stay/rankings

Retrieve stay statistics rankings.

**Query Parameters:**
- `statType` (string, optional, default: "PROVINCE"): PROVINCE, CITY, COUNTY, CATEGORY
- `timeRange` (string, optional, default: "all"): all, YYYY, YYYY-MM, YYYY-MM-DD
- `orderBy` (string, optional, default: "count"): count, duration
- `limit` (int, optional, default: 100, max: 1000)

**Example Request:**
```
GET /api/v1/stats/stay/rankings?statType=CATEGORY&orderBy=duration
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "stat_type": "CATEGORY",
        "stat_key": "HOME",
        "time_range": "all",
        "stay_count": 456,
        "total_duration_seconds": 3600000,
        "avg_duration_seconds": 7894.7,
        "max_duration_seconds": 86400,
        "stay_category": "HOME",
        "rank_by_count": 1,
        "rank_by_duration": 1,
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "count": 4
  }
}
```

### GET /stats/extreme-events

Retrieve extreme events (max altitude, speed, etc.).

**Query Parameters:**
- `eventType` (string, optional): MAX_ALTITUDE, MAX_SPEED, NORTHMOST, SOUTHMOST, EASTMOST, WESTMOST
- `eventCategory` (string, optional): SPATIAL, SPEED, ALTITUDE
- `limit` (int, optional, default: 100, max: 100)

**Example Request:**
```
GET /api/v1/stats/extreme-events?eventCategory=SPEED&limit=10
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "event_type": "MAX_SPEED",
        "event_category": "SPEED",
        "point_id": 12345,
        "event_time": 1705910400,
        "event_value": 320.5,
        "latitude": 39.9042,
        "longitude": 116.4074,
        "province": "北京市",
        "city": "北京市",
        "mode": "TRAIN",
        "segment_id": 100,
        "rank": 1,
        "created_at": "2025-01-22T10:00:00Z"
      }
    ],
    "count": 10
  }
}
```

## 6. Visualization Endpoints

### GET /viz/rendering

Retrieve track points with rendering metadata for map display.

**Query Parameters:**
- `minLat` (float64, optional): Bounding box minimum latitude
- `maxLat` (float64, optional): Bounding box maximum latitude
- `minLon` (float64, optional): Bounding box minimum longitude
- `maxLon` (float64, optional): Bounding box maximum longitude
- `lodLevel` (int, optional, default: 3): Level of detail (1-5)
- `startTime` (int64, optional): Unix timestamp
- `endTime` (int64, optional): Unix timestamp
- `mode` (string, optional): Filter by transport mode
- `limit` (int, optional, default: 10000, max: 50000): Maximum points

**Example Request:**
```
GET /api/v1/viz/rendering?minLat=39.8&maxLat=40.0&minLon=116.3&maxLon=116.5&lodLevel=3&limit=5000
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "dataTime": 1705910400,
        "longitude": 116.4074,
        "latitude": 39.9042,
        "heading": 45.5,
        "speed": 60.2,
        "altitude": 50.0
      }
    ],
    "count": 5000
  }
}
```

**Note:** Automatically excludes outlier points (outlier_flag = 1).

### GET /viz/time-slices

Retrieve aggregated data for time axis filtering.

**Query Parameters:**
- `startTime` (int64, required): Unix timestamp
- `endTime` (int64, required): Unix timestamp
- `granularity` (string, optional, default: "day"): day, month, year

**Example Request:**
```
GET /api/v1/viz/time-slices?startTime=1704067200&endTime=1706745600&granularity=month
```

**Example Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "granularity": "month",
    "slices": [
      {
        "time_slice": "2025-01",
        "point_count": 12345,
        "mode_count": 4,
        "min_time": 1704067200,
        "max_time": 1706745600
      },
      {
        "time_slice": "2025-02",
        "point_count": 10234,
        "mode_count": 3,
        "min_time": 1706745601,
        "max_time": 1709251200
      }
    ],
    "count": 2
  }
}
```

## Performance Notes

1. **Pagination:** All list endpoints support pagination. Use `page` and `pageSize` parameters.
2. **Limits:** Maximum page size is 1000 for most endpoints, 50000 for rendering metadata.
3. **Indexes:** Database indexes are optimized for common query patterns (time range, admin divisions, mode).
4. **Caching:** Consider implementing client-side caching for frequently accessed data.
5. **Rate Limiting:** API is rate-limited to 3 requests per second.

## Error Codes

- `400` - Bad Request: Invalid query parameters
- `404` - Not Found: Resource not found
- `500` - Internal Server Error: Server-side error

## Next Steps

After Phase 5 completion:
- **Phase 6:** Frontend implementation (6-8 days)
  - Admin task management UI
  - Map visualization with rendering metadata
  - Statistics dashboards with rankings
  - Time axis filtering
  - Heatmap visualization
