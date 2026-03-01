# Flight Module Implementation Plan

**Date**: 2026-03-01
**Status**: Phase 1 - Backend Foundation Complete
**Priority**: High (Feature Development)

---

## Overview

The Flight module adds flight route visualization and tracking capabilities to the Records platform. It imports flight data from Variflight/FlightAware JSON exports, stores tracking points, calculates statistics, and provides APIs for visualization.

---

## Implementation Status

### ✅ Phase 1: Backend Foundation (Complete)

**Database Schema** (`scripts/flights/migrations/001_create_flight_tables.sql`)
- ✅ `flights` table - Flight metadata
- ✅ `flight_points` table - GPS tracking points
- ✅ `flight_statistics` table - Pre-computed statistics
- ✅ `flight_track_matches` table - Links to GPS tracks
- ✅ Indexes for performance

**Data Models** (`internal/flights/models.go`)
- ✅ Flight, FlightPoint, FlightStatistics
- ✅ FlightTrackMatch, FlightWithStats
- ✅ FlightSummary, VariflightData

**Repository Layer** (`internal/flights/repository.go`)
- ✅ CreateFlight, CreateFlightPoint, CreateFlightPointsBatch
- ✅ GetFlightByID, ListFlights, GetFlightPoints
- ✅ GetFlightSummary, UpdateFlightStatistics
- ✅ Transaction support for batch operations

**Service Layer** (`internal/flights/service.go`)
- ✅ ImportFlight - Import flight with points
- ✅ GetFlight, ListFlights, GetSummary
- ✅ calculateStatistics - Distance, speed, altitude
- ✅ haversineDistance - GPS distance calculation
- ✅ MatchFlightToTracks - Match flights to GPS tracks

**API Handlers** (`internal/flights/handlers.go`)
- ✅ GET /api/v1/flights - List all flights
- ✅ GET /api/v1/flights/:id - Get flight details
- ✅ GET /api/v1/flights/summary - Summary statistics
- ✅ GET /api/v1/flights/:id/route - Route visualization data
- ✅ Structured logging integration

**Data Import** (`scripts/flights/import_variflight.py`)
- ✅ Parse Variflight JSON format
- ✅ Extract flight number and date from filename
- ✅ Import flight metadata and tracking points
- ✅ Calculate statistics (distance, speed, altitude, duration)
- ✅ Airline code mapping (CA, CZ, MU, HU, KN, etc.)

---

## Features

### Core Features
1. **Flight Data Import**
   - Import from Variflight JSON exports
   - Automatic flight number and date extraction
   - Batch point insertion with transactions
   - Duplicate handling (upsert)

2. **Flight Statistics**
   - Total distance (km)
   - Max/avg altitude (meters)
   - Max/avg speed (km/h)
   - Flight duration (minutes)
   - Point count

3. **Flight Visualization**
   - Route points with coordinates
   - Altitude profile
   - Speed profile
   - Timestamp for animation

4. **Flight-Track Matching**
   - Match flight points to GPS tracks
   - Confidence scoring (0.0-1.0)
   - Match types: exact, interpolated, estimated
   - Time and spatial proximity matching

5. **Summary Statistics**
   - Total flights count
   - Total distance traveled
   - Total flight time
   - Average distance/duration
   - Airlines list

---

## API Endpoints

### GET /api/v1/flights
**Description**: List all flights with pagination
**Query Parameters**:
- `page` (default: 1)
- `pageSize` (default: 20, max: 100)

**Response**:
```json
{
  "flights": [
    {
      "id": 1,
      "flightNumber": "CA1332",
      "airline": "Air China",
      "flightDate": "20250104",
      "dataSource": "variflight",
      "statistics": {
        "totalDistance": 1234.5,
        "maxAltitude": 10500,
        "maxSpeed": 850,
        "avgSpeed": 780,
        "durationMinutes": 120,
        "pointCount": 450
      }
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 15
  }
}
```

### GET /api/v1/flights/:id
**Description**: Get flight details with all tracking points

**Response**:
```json
{
  "flight": {
    "id": 1,
    "flightNumber": "CA1332",
    "aircraftNumber": "B-1234",
    "airline": "Air China",
    "flightDate": "20250104",
    "statistics": {...}
  },
  "points": [
    {
      "id": 1,
      "longitude": 116.4074,
      "latitude": 39.9042,
      "altitude": 10500,
      "speed": 850,
      "heading": 45,
      "updateTime": 1704326400
    }
  ]
}
```

### GET /api/v1/flights/summary
**Description**: Get summary statistics for all flights

**Response**:
```json
{
  "totalFlights": 15,
  "totalDistance": 18500.5,
  "totalDuration": 1800,
  "averageDistance": 1233.4,
  "averageDuration": 120,
  "airlines": ["Air China", "China Southern", "China Eastern"],
  "mostFrequentRoute": "PEK-SHA"
}
```

### GET /api/v1/flights/:id/route
**Description**: Get route visualization data

**Response**:
```json
{
  "flightNumber": "CA1332",
  "flightDate": "20250104",
  "route": [
    {
      "longitude": 116.4074,
      "latitude": 39.9042,
      "altitude": 10500,
      "speed": 850,
      "timestamp": 1704326400
    }
  ],
  "statistics": {...}
}
```

---

## Data Import Workflow

### 1. Prepare Data
Place Variflight JSON files in a directory:
```
data/flights/raw/
├── Variflight_CA1332_20250104.json
├── Variflight_KN5290_20250109.json
└── ...
```

### 2. Run Import Script
```bash
cd go-backend/scripts/flights
python import_variflight.py ../../data/flights/raw/Variflight_CA1332_20250104.json
```

### 3. Verify Import
```bash
# Check database
sqlite3 ../../data/flights/flights.db "SELECT * FROM flights;"
sqlite3 ../../data/flights/flights.db "SELECT COUNT(*) FROM flight_points;"

# Test API
curl http://localhost:8080/api/v1/flights/summary
```

---

## Next Steps

### ⏳ Phase 2: API Integration (2-3 hours)
1. **Update main.go**
   - Initialize flights database connection
   - Register flight routes
   - Add to router setup

2. **Create config entries**
   - Add FlightsDBPath to config
   - Environment variable support

3. **Test APIs**
   - Import sample flight data
   - Test all endpoints
   - Verify statistics calculation

### ⏳ Phase 3: Frontend Visualization (4-5 hours)
1. **Create React components**
   - FlightList - List all flights
   - FlightDetail - Flight details page
   - FlightMap - Map visualization with route
   - FlightStats - Statistics dashboard

2. **Map Integration**
   - Use Leaflet or Mapbox
   - Draw flight route polyline
   - Altitude/speed profile charts
   - Animation support

3. **API Integration**
   - Create flightsApi service
   - TypeScript types
   - Data fetching hooks

### ⏳ Phase 4: Advanced Features (3-4 hours)
1. **Flight-Track Matching**
   - Implement matching algorithm
   - Display matched GPS tracks
   - Confidence visualization

2. **Flight Search**
   - Search by flight number
   - Filter by date range
   - Filter by airline

3. **Statistics Dashboard**
   - Total flights by airline
   - Distance/time charts
   - Frequent routes

### ⏳ Phase 5: Railway Integration (5-6 hours)
1. **Import railway KML data**
   - Parse OpenRailMap KML
   - Store railway line segments
   - Match to GPS tracks

2. **Railway visualization**
   - Display railway lines on map
   - Highlight traveled segments
   - Train journey reconstruction

---

## Database Schema Details

### flights table
```sql
CREATE TABLE flights (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_number TEXT NOT NULL,
  aircraft_number TEXT,
  airline TEXT,
  departure_airport TEXT,
  arrival_airport TEXT,
  departure_time TIMESTAMP,
  arrival_time TIMESTAMP,
  actual_departure TIMESTAMP,
  actual_arrival TIMESTAMP,
  flight_date TEXT NOT NULL,
  data_source TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(flight_number, flight_date)
);
```

### flight_points table
```sql
CREATE TABLE flight_points (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_id INTEGER NOT NULL,
  update_time INTEGER NOT NULL,
  utc_time TEXT,
  longitude REAL NOT NULL,
  latitude REAL NOT NULL,
  altitude REAL,
  speed REAL,
  heading REAL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE
);
```

### flight_statistics table
```sql
CREATE TABLE flight_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  flight_id INTEGER NOT NULL UNIQUE,
  total_distance REAL,
  max_altitude REAL,
  max_speed REAL,
  avg_speed REAL,
  duration_minutes INTEGER,
  point_count INTEGER,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE
);
```

---

## Testing Plan

### Unit Tests
- [ ] Repository CRUD operations
- [ ] Service business logic
- [ ] Statistics calculation
- [ ] Distance calculation (haversine)
- [ ] Flight-track matching algorithm

### Integration Tests
- [ ] Data import from JSON
- [ ] API endpoints
- [ ] Database transactions
- [ ] Error handling

### Manual Tests
- [ ] Import real Variflight data
- [ ] Verify statistics accuracy
- [ ] Test pagination
- [ ] Test route visualization data

---

## Performance Considerations

1. **Batch Inserts**
   - Use transactions for point insertion
   - Insert 1000+ points efficiently

2. **Indexes**
   - flight_id for point queries
   - update_time for temporal queries
   - longitude/latitude for spatial queries

3. **Statistics Caching**
   - Pre-compute statistics on import
   - Update only when data changes

4. **Pagination**
   - Limit results to 20-100 per page
   - Use offset for large datasets

---

## File Structure

```
go-backend/
├── internal/flights/
│   ├── models.go          # Data models
│   ├── repository.go      # Database access
│   ├── service.go         # Business logic
│   └── handlers.go        # HTTP handlers
├── scripts/flights/
│   ├── migrations/
│   │   └── 001_create_flight_tables.sql
│   └── import_variflight.py
└── data/flights/
    ├── flights.db         # SQLite database
    └── raw/               # Raw JSON files
```

---

## Success Criteria

- ✅ Database schema created
- ✅ Models and repository implemented
- ✅ Service layer with statistics
- ✅ API handlers with logging
- ✅ Import script functional
- ⏳ APIs integrated into main server
- ⏳ Sample data imported
- ⏳ Frontend visualization
- ⏳ Flight-track matching working

---

**Next Action**: Integrate flight APIs into main server and test with sample data.
