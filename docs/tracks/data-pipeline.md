# Tracks Data Pipeline

## Overview

This document describes the complete data pipeline for GPS trajectory processing, including the current implementation status and critical gaps.

## Pipeline Stages

### Stage 1: Data Import ✅ IMPLEMENTED

**Script:** `write2sql.py`

**Input:** Excel file with GPS trajectory data

**Process:**
1. User selects Excel file via tkinter file dialog
2. Read Excel file with pandas
3. Filter records where `stepType=0`
4. Extract columns: dataTime, longitude, latitude, heading, accuracy, speed, distance, altitude
5. Convert Unix timestamps to two formats:
   - `time_visually`: YYYY/MM/DD HH:MM:SS.000
   - `time`: YYYYMMDDHHMMSS
6. Sanitize column names for SQL compatibility
7. Create table "一生足迹" with auto-increment ID
8. Insert data using pandas `to_sql()` with `if_exists='replace'`
9. Enable WAL mode: `PRAGMA journal_mode=WAL`

**Output:** SQLite database `data/tracks/tracks.db` with table "一生足迹"

**Performance:**
- Batch insert with pandas (fast)
- Replaces entire table (no incremental updates)
- Typical import time: ~10 seconds for 100k records

**Limitations:**
- No deduplication (relies on source data quality)
- No staging table (can't rollback failed imports)
- No progress tracking
- Replaces all data (can't append)

**Code Location:** `go-backend/scripts/tracks/write2sql.py`

### Stage 2: Geocoding ❌ NOT IMPLEMENTED (CRITICAL GAP)

**Status:** This stage is completely missing. It's a blocker for automated processing.

**Required Functionality:**

**Input:** GPS coordinates (longitude, latitude)

**Process:** Reverse geocoding to 5-level administrative divisions

**Output:** Administrative division names
- Province (省级): e.g., "广东省"
- City (市级): e.g., "深圳市"
- County (区县级): e.g., "南山区"
- Town (乡镇级): e.g., "南头街道"
- Village (村级/街道级): e.g., "南头社区"

**Current Workaround:**
1. Export data from database to CSV
2. Manually geocode using external tools
3. Import geocoded data back
4. Run processing scripts on geocoded CSV

**Implementation Options:**

#### Option 1: Windows Service with Shapefile (Recommended)

**Pros:**
- Offline processing (no API calls)
- Fast (no network latency)
- No rate limits
- Privacy-preserving
- Free

**Cons:**
- Requires shapefile data (~500MB)
- Need spatial query library (e.g., Shapely, GEOS)
- Initial setup complexity

**Implementation:**
```python
import geopandas as gpd
from shapely.geometry import Point

# Load shapefiles once at startup
provinces = gpd.read_file('shapefiles/provinces.shp')
cities = gpd.read_file('shapefiles/cities.shp')
counties = gpd.read_file('shapefiles/counties.shp')
towns = gpd.read_file('shapefiles/towns.shp')

def geocode_point(lon, lat):
    point = Point(lon, lat)

    # Query each level
    province = provinces[provinces.contains(point)]['name'].iloc[0]
    city = cities[cities.contains(point)]['name'].iloc[0]
    county = counties[counties.contains(point)]['name'].iloc[0]
    town = towns[towns.contains(point)]['name'].iloc[0]

    return {
        'province': province,
        'city': city,
        'county': county,
        'town': town,
        'village': None  # Village level may not be available
    }
```

**Data Sources:**
- National Geomatics Center of China (NGCC)
- OpenStreetMap (OSM) with administrative boundaries
- Gaode/Baidu shapefile exports

#### Option 2: External API (Gaode/Baidu)

**Pros:**
- Easy to implement
- Accurate and up-to-date
- Includes village-level data

**Cons:**
- Rate limits (e.g., 300 req/day for free tier)
- API costs for high volume
- Network dependency
- Privacy concerns (sending location data)
- Requires API key management

**Implementation:**
```python
import requests
import time

def geocode_with_gaode(lon, lat, api_key):
    url = f"https://restapi.amap.com/v3/geocode/regeo"
    params = {
        'key': api_key,
        'location': f'{lon},{lat}',
        'extensions': 'all'
    }

    response = requests.get(url, params=params)
    data = response.json()

    if data['status'] == '1':
        addr = data['regeocode']['addressComponent']
        return {
            'province': addr['province'],
            'city': addr['city'],
            'county': addr['district'],
            'town': addr['township'],
            'village': addr.get('neighborhood', {}).get('name')
        }

    time.sleep(0.1)  # Rate limiting
    return None
```

**Rate Limits:**
- Gaode: 300 req/day (free), 300k req/day (paid)
- Baidu: 6k req/day (free), unlimited (paid)

**Cost Estimate:**
- 100k points × ¥0.001/req = ¥100
- One-time cost for historical data
- Incremental cost for new data

#### Option 3: Offline Database (GeoNames)

**Pros:**
- Free and open source
- Offline processing
- No rate limits

**Cons:**
- Less accurate for county/town/village levels
- Data may be outdated
- Requires database setup

**Implementation:**
```python
import sqlite3
from math import radians, cos, sin, asin, sqrt

def haversine(lon1, lat1, lon2, lat2):
    lon1, lat1, lon2, lat2 = map(radians, [lon1, lat1, lon2, lat2])
    dlon = lon2 - lon1
    dlat = lat2 - lat1
    a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
    c = 2 * asin(sqrt(a))
    km = 6371 * c
    return km

def geocode_with_geonames(lon, lat, db_path):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Find nearest city within 50km
    cursor.execute("""
        SELECT name, admin1, admin2, admin3
        FROM geonames
        WHERE latitude BETWEEN ? AND ?
          AND longitude BETWEEN ? AND ?
        ORDER BY
          (latitude - ?) * (latitude - ?) +
          (longitude - ?) * (longitude - ?)
        LIMIT 1
    """, (lat-0.5, lat+0.5, lon-0.5, lon+0.5, lat, lat, lon, lon))

    result = cursor.fetchone()
    conn.close()

    if result:
        return {
            'province': result[1],
            'city': result[0],
            'county': result[2],
            'town': result[3],
            'village': None
        }
    return None
```

**Recommendation:** Use Option 1 (Shapefile) for best balance of accuracy, performance, and cost.

### Stage 3: Database Update ❌ NOT IMPLEMENTED

**Required Changes:**

1. **Add Administrative Division Columns**

```sql
ALTER TABLE "一生足迹" ADD COLUMN province TEXT;
ALTER TABLE "一生足迹" ADD COLUMN city TEXT;
ALTER TABLE "一生足迹" ADD COLUMN county TEXT;
ALTER TABLE "一生足迹" ADD COLUMN town TEXT;
ALTER TABLE "一生足迹" ADD COLUMN village TEXT;
```

2. **Create Indexes**

```sql
CREATE INDEX idx_datatime ON "一生足迹"(dataTime);
CREATE INDEX idx_admin ON "一生足迹"(province, city, county);
CREATE INDEX idx_time ON "一生足迹"(time);
CREATE INDEX idx_coords ON "一生足迹"(longitude, latitude);
```

3. **Update Records with Geocoded Data**

```python
import sqlite3

def update_admin_divisions(db_path, geocode_func):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Get all points without admin data
    cursor.execute("""
        SELECT id, longitude, latitude
        FROM "一生足迹"
        WHERE province IS NULL
    """)

    points = cursor.fetchall()

    # Batch update
    batch = []
    for point_id, lon, lat in points:
        admin = geocode_func(lon, lat)
        if admin:
            batch.append((
                admin['province'],
                admin['city'],
                admin['county'],
                admin['town'],
                admin['village'],
                point_id
            ))

        if len(batch) >= 1000:
            cursor.executemany("""
                UPDATE "一生足迹"
                SET province=?, city=?, county=?, town=?, village=?
                WHERE id=?
            """, batch)
            conn.commit()
            batch = []

    # Update remaining
    if batch:
        cursor.executemany("""
            UPDATE "一生足迹"
            SET province=?, city=?, county=?, town=?, village=?
            WHERE id=?
        """, batch)
        conn.commit()

    conn.close()
```

### Stage 4: Processing Scripts ⚠️ PARTIALLY IMPLEMENTED

**Current Status:** Scripts exist but require manual CSV export with geocoded data.

#### Script 1: tracks.py (Location/Time Statistics)

**Input:** CSV file with geocoded data

**Process:**
- Group by administrative divisions (province, city, county, town)
- Count unique locations visited
- Calculate time spent in each location
- Generate statistics by year, month, day

**Output:** CSV files with statistics

**Limitations:**
- Requires manual CSV export
- No database integration
- No incremental updates
- No caching

**Code Location:** `go-backend/scripts/tracks/process_tracks/tracks.py`

#### Script 2: stop.py (Stay Detection)

**Input:** CSV file with geocoded data

**Process:**
- Detect stays >2 hours at same location
- Group by administrative levels (city, county, town)
- Calculate stay duration and frequency

**Output:** CSV files with stay statistics

**Limitations:**
- Requires manual CSV export
- Fixed threshold (2 hours)
- No parameterization
- No confidence scores

**Code Location:** `go-backend/scripts/tracks/process_tracks/stop.py`

**Required Refactoring:**
1. Query database directly instead of CSV
2. Make thresholds configurable
3. Add confidence scores and reason codes
4. Store results in database
5. Support incremental updates

### Stage 5: API Layer ❌ NOT IMPLEMENTED

**Required Endpoints:**

```
GET  /api/v1/tracks/points
GET  /api/v1/tracks/statistics/footprint
GET  /api/v1/tracks/statistics/stays
GET  /api/v1/tracks/trips
POST /api/v1/tracks/import
```

See `api-endpoints.md` for complete specification.

### Stage 6: Frontend ❌ NOT IMPLEMENTED

**Required Components:**
- Map visualization (Leaflet/Mapbox)
- Statistics dashboard
- Timeline view
- Stay annotations

See `tracks/docs/` for frontend documentation.

## Data Flow Diagram

```
┌─────────────────┐
│  Excel File     │
│  (GPS Data)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  write2sql.py   │  ✅ IMPLEMENTED
│  (Import)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  tracks.db      │
│  (Raw Points)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Geocoding      │  ❌ NOT IMPLEMENTED (BLOCKER)
│  Service        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  tracks.db      │
│  (+ Admin Data) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Processing     │  ⚠️ PARTIALLY IMPLEMENTED
│  Scripts        │  (Requires manual CSV export)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Derived Data   │  ❌ NOT IMPLEMENTED
│  (Stays, Trips) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Go Backend     │  ❌ NOT IMPLEMENTED
│  API            │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  React          │  ❌ NOT IMPLEMENTED
│  Frontend       │
└─────────────────┘
```

## Implementation Roadmap

### Phase 1: Foundation (P0)
1. Implement geocoding service (Option 1: Shapefile)
2. Add administrative division columns to database
3. Create indexes
4. Update existing records with geocoded data

### Phase 2: Processing (P1)
5. Refactor `tracks.py` to query database directly
6. Refactor `stop.py` to query database directly
7. Add parameterization and confidence scores
8. Store results in database

### Phase 3: API (P1)
9. Implement Go backend API
10. Add authentication
11. Add caching
12. Add rate limiting

### Phase 4: Frontend (P2)
13. Implement React frontend
14. Add map visualization
15. Add statistics dashboard
16. Add timeline view

## Performance Considerations

### Geocoding Performance

**Shapefile Approach:**
- Initial load: ~5 seconds (load shapefiles into memory)
- Per-point query: ~1ms (spatial index)
- 100k points: ~100 seconds (~1.5 minutes)

**Optimization:**
- Cache geocoding results by rounded coordinates
- Batch process in chunks of 10k points
- Use spatial index (R-tree)
- Parallelize with multiprocessing

**API Approach:**
- Per-point query: ~100ms (network + API)
- Rate limit: 300 req/day (free) or 300k req/day (paid)
- 100k points: ~3 hours (with rate limiting)

### Database Performance

**Query Optimization:**
- Use indexes on frequently queried fields
- Paginate large result sets
- Cache aggregated statistics
- Use prepared statements

**Storage:**
- Current: ~500MB (raw points)
- With admin data: ~600MB (+20%)
- With derived data: ~800MB (+60%)
- Projected 5 years: ~2GB

## Testing Strategy

### Unit Tests
- Test geocoding accuracy with known coordinates
- Test database operations (insert, update, query)
- Test processing algorithms with sample data

### Integration Tests
- Test complete pipeline end-to-end
- Test with realistic data volumes (100k+ points)
- Test error handling and edge cases

### Performance Tests
- Benchmark geocoding speed
- Benchmark database query performance
- Test with maximum data volume (1M+ points)

## Monitoring

### Metrics to Track
- Import success rate
- Geocoding accuracy
- Processing time per stage
- API response time
- Database size growth

### Logging
- Log all import operations
- Log geocoding failures
- Log processing errors
- Log API requests

## Troubleshooting

### Common Issues

**Issue 1: Geocoding Fails**
- Check shapefile data integrity
- Verify coordinates are within China
- Check for invalid coordinates (0, 0)

**Issue 2: Processing Scripts Fail**
- Verify admin division data exists
- Check for NULL values
- Verify data types

**Issue 3: Database Locked**
- Check for long-running transactions
- Force WAL checkpoint
- Restart application

## Next Steps

1. **Immediate:** Implement geocoding service (Option 1: Shapefile)
2. **Short-term:** Refactor processing scripts to use database
3. **Medium-term:** Implement Go backend API
4. **Long-term:** Implement React frontend

See `implementation-guide.md` for detailed step-by-step instructions.
