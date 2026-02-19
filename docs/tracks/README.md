# Tracks Module Documentation

## Overview

The tracks module is the core component of the personal data analysis platform. It processes GPS trajectory data to extract insights about movement patterns, visited locations, and spatial behavior.

**Current Status (Updated 2026-02-19):**
- ✅ Go backend API implemented with complete architecture
- ✅ Database schema updated with administrative divisions
- ✅ Geocoding service implemented (shapefile-based)
- ✅ Python processing scripts refactored to use database directly
- ✅ RESTful API endpoints for tracks and statistics

## Module Structure

```
go-backend/
├── data/tracks/
│   └── tracks.db                    # SQLite database (19 columns)
├── scripts/tracks/
│   ├── write2sql.py                 # Data import script
│   ├── geocode.py                   # Geocoding service (NEW)
│   ├── run_migration.py             # Database migration runner (NEW)
│   ├── migrations/                  # SQL migration files (NEW)
│   ├── process_tracks/
│   │   ├── tracks.py                # Location/time statistics (original)
│   │   ├── tracks_v2.py             # Refactored version (NEW)
│   │   ├── stop.py                  # Stay detection (original)
│   │   └── stop_v2.py               # Refactored version (NEW)
│   └── requirements.txt
├── internal/                        # Go backend implementation (NEW)
│   ├── models/                      # Data models
│   ├── database/                    # Database layer
│   ├── repository/                  # Data access layer
│   ├── service/                     # Business logic layer
│   ├── handler/                     # HTTP handlers
│   ├── middleware/                  # Middleware (CORS, auth, rate limit)
│   └── api/                         # Router configuration
├── pkg/                             # Shared packages (NEW)
│   └── response/                    # Standard response format
└── docs/tracks/
    ├── README.md                    # This file
    ├── implementation-summary.md    # 2026-02-19 implementation summary (NEW)
    ├── geocoding-guide.md           # Geocoding service guide (NEW)
    ├── data-pipeline.md             # Complete data pipeline
    ├── database-schema.md           # Database schema details
    ├── processing-scripts.md        # Python script documentation
    ├── api-endpoints.md             # RESTful API specification
    ├── analysis-skills.md           # 30 skills implementation mapping
    └── implementation-guide.md      # Step-by-step implementation guide
```

## Quick Start

### 1. Import GPS Data

```bash
cd go-backend/scripts/tracks
python write2sql.py
```

- Opens file dialog to select Excel file
- Filters `stepType=0` records
- Imports to `data/tracks/tracks.db`
- Replaces existing table data

### 2. Database Migration

```bash
cd go-backend/scripts/tracks
python run_migration.py
```

- Adds administrative division columns (province, city, county, town, village)
- Adds metadata columns (created_at, updated_at, algo_version)
- Creates performance indexes

### 3. Geocoding (Automated)

```bash
# Download shapefile data to data/shapefiles/
# See geocoding-guide.md for data sources

cd go-backend/scripts/tracks
python geocode.py --batch-size 1000
```

- Automated reverse geocoding using shapefiles
- Processes 100k points in ~2 minutes
- Updates database with administrative divisions

### 4. Process Data

```bash
cd go-backend/scripts/tracks/process_tracks

# Stay detection (refactored version)
python stop_v2.py --threshold 2 --admin-level town

# Footprint statistics (refactored version)
python tracks_v2.py --output stats.json
```

### 5. Start Go Backend

```bash
cd go-backend
go mod tidy
go run cmd/server/main.go
```

### 6. Test API

```bash
# Health check
curl http://localhost:8080/health

# Get track points
curl "http://localhost:8080/api/v1/tracks/points?page=1&pageSize=10"

# Get footprint statistics
curl "http://localhost:8080/api/v1/tracks/statistics/footprint"
```

## Implementation Status

**Geocoding is now implemented!** The critical blocker has been resolved.

**Current Workflow (2026-02-19):**
```
Import → Automated Geocoding → Database → API → Frontend (pending)
```

**Completed Components:**
- ✅ Data import with filtering
- ✅ Database schema with admin divisions
- ✅ Automated geocoding service
- ✅ Go backend API with full architecture
- ✅ Refactored Python processing scripts
- ✅ RESTful API endpoints

See `implementation-summary.md` for complete details.

## Analysis Capabilities

The tracks module implements 30 processing skills organized into 7 categories:

### 1. Data Foundation & Quality Control (4 skills)
- Data import with filtering
- Outlier detection
- Trajectory completion
- Administrative division attribution

### 2. Behavior & Segmentation (5 skills)
- Transport mode classification
- Stay detection
- Trip construction
- Streak detection
- Speed events

### 3. Spatial Analysis (8 skills)
- Grid system
- Road overlap analysis
- Density structure
- Speed-space coupling
- Revisit patterns
- Utilization efficiency
- Spatial complexity
- Directional bias

### 4. Statistical Aggregation (5 skills)
- Footprint statistics
- Stay statistics
- Extreme events
- Administrative crossings
- Administrative view engine

### 5. Advanced Analysis (3 skills)
- Time-space slicing
- Time-space compression
- Altitude dimension

### 6. Visualization & Rendering (3 skills)
- Rendering metadata
- Time axis map
- Stay annotation

### 7. Integration & Persona (1 skill)
- Spatial persona engine

See `analysis-skills.md` for detailed mapping of skills to implementation.

## Database Schema

### Main Table: 一生足迹 (Life Footprints)

```sql
CREATE TABLE "一生足迹" (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  dataTime INTEGER,           -- Unix timestamp (seconds)
  longitude REAL,
  latitude REAL,
  heading REAL,
  accuracy REAL,
  speed REAL,
  distance REAL,
  altitude REAL,
  time_visually TEXT,         -- Format: 2025/01/22 21:42:18.000
  time TEXT,                  -- Format: 20250122214218
  -- Administrative divisions (added 2026-02-19)
  province TEXT,              -- 省级
  city TEXT,                  -- 市级
  county TEXT,                -- 区县级
  town TEXT,                  -- 乡镇级
  village TEXT,               -- 村级/街道级
  -- Metadata (added 2026-02-19)
  created_at TEXT,
  updated_at TEXT,
  algo_version TEXT
);
```

**Indexes (added 2026-02-19):**
- `idx_datatime` on dataTime
- `idx_time` on time
- `idx_admin_province` on province
- `idx_admin_city` on city
- `idx_admin_county` on county
- `idx_admin_full` on (province, city, county)
- `idx_coordinates` on (longitude, latitude)

See `database-schema.md` for complete schema with indexes and derived tables.

## API Endpoints

**Status:** ✅ Implemented (2026-02-19)

**Available Endpoints:**

```
# Track Points
GET  /api/v1/tracks/points              # Get GPS points with filtering & pagination
GET  /api/v1/tracks/points/:id          # Get single track point
GET  /api/v1/tracks/ungeocoded          # Get points without admin divisions

# Statistics
GET  /api/v1/tracks/statistics/footprint          # Footprint statistics
GET  /api/v1/tracks/statistics/time-distribution  # Time distribution
GET  /api/v1/tracks/statistics/speed-distribution # Speed distribution

# Health Check
GET  /health                            # API health status
```

See `api-endpoints.md` for complete API specification.

## Implementation Priority

**Phase 1: Foundation (P0 - Required)** ✅ COMPLETED 2026-02-19
1. ✅ Data import (`write2sql.py`)
2. ✅ Geocoding service (shapefile-based)
3. ✅ Database schema updates (admin columns + indexes)
4. ✅ Basic API endpoints (points, statistics)
5. ✅ Go backend architecture (models, repository, service, handler)
6. ✅ Middleware (CORS, logging, rate limiting, auth)

**Phase 2: Core Analysis (P1)** 🚧 IN PROGRESS
5. ✅ Stay detection (refactored in `stop_v2.py`)
6. ✅ Footprint statistics (refactored in `tracks_v2.py`)
7. ❌ Trip construction
8. ❌ Transport mode classification

**Phase 3: Advanced Analysis (P2)** 📋 PLANNED
9. ❌ Spatial analysis (grid, density, revisit)
10. ❌ Time-space analysis
11. ❌ Visualization rendering

See `implementation-guide.md` for step-by-step implementation instructions.

## Performance Considerations

### Server Constraints
- CPU: 2 cores
- RAM: ≤1GB for Go backend
- Request rate: ≤3 req/s
- Database size: ~500MB current, ~2GB projected

### Design Principles
- Batch processing for heavy computations
- Incremental updates instead of full recalculation
- Aggressive caching of computed results
- Pagination for large result sets (100k+ points)

## Documentation Index

- `README.md` - This file (module overview and quick start)
- `implementation-summary.md` - Complete 2026-02-19 implementation summary
- `geocoding-guide.md` - Geocoding service detailed guide
- `data-pipeline.md` - Complete data pipeline documentation
- `database-schema.md` - Detailed database schema with indexes
- `processing-scripts.md` - Python script documentation
- `api-endpoints.md` - RESTful API specification
- `analysis-skills.md` - Map 30 skills to implementation
- `implementation-guide.md` - Step-by-step implementation guide

## Related Documentation

- Root: `.claude/skills/tracks/` - 30 skill definitions
- Frontend: `tracks/docs/` - Frontend integration documentation
- Backend: `go-backend/docs/` - Overall backend architecture
