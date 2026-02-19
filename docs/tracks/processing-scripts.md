# Tracks Processing Scripts

## Overview

Python scripts for GPS trajectory data import and analysis. Located in `go-backend/scripts/tracks/`.

## Scripts

### 1. write2sql.py - Data Import

**Status:** ✅ Implemented

**Purpose:** Import GPS trajectory data from Excel to SQLite database.

**Usage:**
```bash
cd go-backend/scripts/tracks
python write2sql.py
```

**Process:**
1. Opens tkinter file dialog for Excel file selection
2. Reads Excel with pandas
3. Filters `stepType=0` records
4. Extracts columns: dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude
5. Converts timestamps to two formats
6. Sanitizes column names
7. Creates/replaces table "一生足迹"
8. Enables WAL mode

**Input:** Excel file with GPS data

**Output:** SQLite database `data/tracks/tracks.db`

**Performance:** ~10 seconds for 100k records

**Limitations:**
- No deduplication
- No staging table
- Replaces all data (no append mode)
- No progress tracking

**Dependencies:**
- pandas
- sqlite3
- tkinter
- openpyxl

### 2. tracks.py - Location/Time Statistics

**Status:** ⚠️ Partially Implemented (Requires manual CSV export)

**Purpose:** Generate location and time-based statistics.

**Usage:**
```bash
cd go-backend/scripts/tracks/process_tracks
python tracks.py
```

**Process:**
1. Reads CSV file with geocoded data
2. Groups by administrative divisions (province, city, county, town)
3. Counts unique locations
4. Calculates time spent
5. Generates statistics by year, month, day

**Input:** CSV file with geocoded data (requires manual export)

**Output:** CSV files with statistics

**Limitations:**
- Requires manual CSV export with admin divisions
- No database integration
- No incremental updates
- No caching

**Required Refactoring:**
- Query database directly
- Store results in database
- Support incremental updates
- Add progress tracking

### 3. stop.py - Stay Detection

**Status:** ⚠️ Partially Implemented (Requires manual CSV export)

**Purpose:** Detect stays longer than 2 hours.

**Usage:**
```bash
cd go-backend/scripts/tracks/process_tracks
python stop.py
```

**Process:**
1. Reads CSV file with geocoded data
2. Detects consecutive points at same location
3. Filters stays >2 hours
4. Groups by administrative levels
5. Calculates stay duration and frequency

**Input:** CSV file with geocoded data

**Output:** CSV files with stay statistics

**Limitations:**
- Fixed threshold (2 hours)
- No parameterization
- No confidence scores
- No stay classification (home/work/transit)

**Required Refactoring:**
- Make threshold configurable
- Add confidence scores
- Classify stay types
- Store results in database
- Support incremental updates

## Planned Scripts

### 4. geocode.py - Geocoding Service

**Status:** ❌ Not Implemented (BLOCKER)

**Purpose:** Reverse geocode GPS coordinates to administrative divisions.

**Planned Usage:**
```bash
python geocode.py --input tracks.db --output tracks.db --method shapefile
```

**Options:**
- `--method shapefile` - Use shapefile-based geocoding (recommended)
- `--method gaode` - Use Gaode API
- `--method baidu` - Use Baidu API
- `--batch-size 1000` - Batch size for processing
- `--cache` - Enable caching

**Implementation Priority:** P0 (Critical)

### 5. outlier_detection.py - Outlier Detection

**Status:** ❌ Not Implemented

**Purpose:** Detect and flag outlier GPS points.

**Planned Usage:**
```bash
python outlier_detection.py --input tracks.db --output tracks.db
```

**Detection Rules:**
- Speed spike (>200 km/h for non-plane)
- Bounce back (A→B→A in short time)
- Long jump (>500km in <1 hour)
- Accuracy threshold (>100m)

**Implementation Priority:** P1

### 6. trip_construction.py - Trip Construction

**Status:** ❌ Not Implemented

**Purpose:** Construct trips from GPS points and stays.

**Planned Usage:**
```bash
python trip_construction.py --input tracks.db --output tracks.db
```

**Process:**
1. Identify stays as trip endpoints
2. Group points between stays as trips
3. Classify transport mode
4. Calculate trip statistics

**Implementation Priority:** P1

### 7. transport_mode.py - Transport Mode Classification

**Status:** ❌ Not Implemented

**Purpose:** Classify transport mode for each trip.

**Planned Usage:**
```bash
python transport_mode.py --input tracks.db --output tracks.db
```

**Classification:**
- Walk: <5 km/h
- Bike: 5-20 km/h
- Car: 20-120 km/h
- Train: 80-350 km/h
- Plane: >350 km/h

**Implementation Priority:** P1

## Dependencies

**Current:**
```
pandas>=2.0.0
openpyxl>=3.1.0
tkinter (built-in)
sqlite3 (built-in)
```

**Planned:**
```
geopandas>=0.14.0  # For shapefile geocoding
shapely>=2.0.0     # For spatial operations
requests>=2.31.0   # For API geocoding
numpy>=1.24.0      # For numerical operations
scipy>=1.11.0      # For statistical analysis
```

## Installation

```bash
cd go-backend/scripts/tracks
pip install -r requirements.txt
```

## Testing

```bash
# Unit tests
pytest tests/

# Integration tests
pytest tests/integration/

# Performance tests
pytest tests/performance/ --benchmark
```

## Related Documentation

- `data-pipeline.md` - Complete data pipeline
- `database-schema.md` - Database schema
- `api-endpoints.md` - API specification
- `implementation-guide.md` - Implementation guide
