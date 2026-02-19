# Tracks Database Schema

## Database File

**Location:** `go-backend/data/tracks/tracks.db`

**Configuration:**
```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=-64000;  -- 64MB cache
PRAGMA temp_store=MEMORY;
PRAGMA foreign_keys=ON;
```

## Tables

### 1. 一生足迹 (Life Footprints) - Raw GPS Points

**Status:** ✅ Implemented

```sql
CREATE TABLE "一生足迹" (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  dataTime INTEGER NOT NULL,        -- Unix timestamp (seconds)
  longitude REAL NOT NULL,
  latitude REAL NOT NULL,
  heading REAL,
  accuracy REAL,
  speed REAL,
  distance REAL,
  altitude REAL,
  time_visually TEXT,               -- Format: 2025/01/22 21:42:18.000
  time TEXT,                        -- Format: 20250122214218

  -- Administrative divisions (TO ADD)
  province TEXT,                    -- 省级: e.g., "广东省"
  city TEXT,                        -- 市级: e.g., "深圳市"
  county TEXT,                      -- 区县级: e.g., "南山区"
  town TEXT,                        -- 乡镇级: e.g., "南头街道"
  village TEXT,                     -- 村级/街道级: e.g., "南头社区"

  -- Data quality flags (TO ADD)
  outlier_flag INTEGER DEFAULT 0,   -- 0=normal, 1=outlier
  outlier_score REAL,               -- Confidence score (0-1)
  outlier_reasons TEXT,             -- JSON array of reason codes
  qa_status TEXT DEFAULT 'pending', -- 'pending', 'approved', 'rejected'

  -- Metadata (TO ADD)
  created_at INTEGER,               -- Unix timestamp
  updated_at INTEGER,               -- Unix timestamp
  algo_version TEXT                 -- Algorithm version
);
```

**Indexes:**
```sql
CREATE INDEX idx_datatime ON "一生足迹"(dataTime);
CREATE INDEX idx_admin ON "一生足迹"(province, city, county);
CREATE INDEX idx_time ON "一生足迹"(time);
CREATE INDEX idx_coords ON "一生足迹"(longitude, latitude);
CREATE INDEX idx_qa ON "一生足迹"(qa_status, outlier_flag);
```

**Sample Data:**
```sql
INSERT INTO "一生足迹" VALUES (
  1,
  1737552138,
  114.0579,
  22.5431,
  90.0,
  10.0,
  5.5,
  100.0,
  500.0,
  '2025/01/22 21:42:18.000',
  '20250122214218',
  '广东省',
  '深圳市',
  '南山区',
  '南头街道',
  '南头社区',
  0,
  NULL,
  NULL,
  'approved',
  1737552138,
  1737552138,
  'v1.0'
);
```

### 2. stays - Detected Stays

**Status:** ❌ Not Implemented (Data currently in CSV files)

```sql
CREATE TABLE stays (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  start_time INTEGER NOT NULL,      -- Unix timestamp
  end_time INTEGER NOT NULL,        -- Unix timestamp
  duration INTEGER NOT NULL,        -- Duration in seconds
  center_lon REAL NOT NULL,
  center_lat REAL NOT NULL,
  radius REAL,                      -- Radius in meters
  point_count INTEGER,              -- Number of points in stay

  -- Administrative divisions
  province TEXT,
  city TEXT,
  county TEXT,
  town TEXT,
  village TEXT,

  -- Stay classification
  stay_type TEXT,                   -- 'home', 'work', 'transit', 'other'
  confidence REAL,                  -- Confidence score (0-1)

  -- Metadata
  created_at INTEGER,
  algo_version TEXT,

  FOREIGN KEY (start_time) REFERENCES "一生足迹"(dataTime),
  FOREIGN KEY (end_time) REFERENCES "一生足迹"(dataTime)
);
```

**Indexes:**
```sql
CREATE INDEX idx_stays_time ON stays(start_time, end_time);
CREATE INDEX idx_stays_admin ON stays(province, city, county);
CREATE INDEX idx_stays_type ON stays(stay_type);
CREATE INDEX idx_stays_duration ON stays(duration);
```

### 3. trips - Constructed Trips

**Status:** ❌ Not Implemented

```sql
CREATE TABLE trips (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  start_time INTEGER NOT NULL,
  end_time INTEGER NOT NULL,
  duration INTEGER NOT NULL,
  distance REAL,                    -- Total distance in meters

  -- Origin
  origin_lon REAL,
  origin_lat REAL,
  origin_province TEXT,
  origin_city TEXT,
  origin_county TEXT,

  -- Destination
  dest_lon REAL,
  dest_lat REAL,
  dest_province TEXT,
  dest_city TEXT,
  dest_county TEXT,

  -- Trip classification
  transport_mode TEXT,              -- 'walk', 'bike', 'car', 'train', 'plane'
  mode_confidence REAL,

  -- Metadata
  created_at INTEGER,
  algo_version TEXT,

  FOREIGN KEY (start_time) REFERENCES "一生足迹"(dataTime),
  FOREIGN KEY (end_time) REFERENCES "一生足迹"(dataTime)
);
```

**Indexes:**
```sql
CREATE INDEX idx_trips_time ON trips(start_time, end_time);
CREATE INDEX idx_trips_origin ON trips(origin_province, origin_city);
CREATE INDEX idx_trips_dest ON trips(dest_province, dest_city);
CREATE INDEX idx_trips_mode ON trips(transport_mode);
```

### 4. footprint_stats - Pre-computed Statistics

**Status:** ❌ Not Implemented (Data currently in CSV files)

```sql
CREATE TABLE footprint_stats (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  stat_type TEXT NOT NULL,          -- 'province', 'city', 'county', 'town'
  stat_key TEXT NOT NULL,           -- Admin division name

  -- Time period
  year INTEGER,
  month INTEGER,
  day INTEGER,

  -- Statistics
  visit_count INTEGER,              -- Number of visits
  point_count INTEGER,              -- Number of GPS points
  total_duration INTEGER,           -- Total time in seconds
  first_visit INTEGER,              -- Unix timestamp
  last_visit INTEGER,               -- Unix timestamp

  -- Metadata
  created_at INTEGER,
  updated_at INTEGER,
  algo_version TEXT,

  UNIQUE(stat_type, stat_key, year, month, day)
);
```

**Indexes:**
```sql
CREATE INDEX idx_stats_type ON footprint_stats(stat_type, stat_key);
CREATE INDEX idx_stats_time ON footprint_stats(year, month, day);
```

### 5. admin_divisions - Administrative Division Lookup

**Status:** ❌ Not Implemented

```sql
CREATE TABLE admin_divisions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  level TEXT NOT NULL,              -- 'province', 'city', 'county', 'town', 'village'
  name TEXT NOT NULL,
  parent_id INTEGER,                -- Reference to parent division
  code TEXT,                        -- Administrative code

  -- Bounding box
  min_lon REAL,
  min_lat REAL,
  max_lon REAL,
  max_lat REAL,

  -- Metadata
  created_at INTEGER,

  FOREIGN KEY (parent_id) REFERENCES admin_divisions(id)
);
```

**Indexes:**
```sql
CREATE INDEX idx_admin_level ON admin_divisions(level, name);
CREATE INDEX idx_admin_parent ON admin_divisions(parent_id);
CREATE INDEX idx_admin_bbox ON admin_divisions(min_lon, min_lat, max_lon, max_lat);
```

### 6. schema_version - Schema Migration Tracking

**Status:** ❌ Not Implemented

```sql
CREATE TABLE schema_version (
  version INTEGER PRIMARY KEY,
  applied_at INTEGER NOT NULL,
  description TEXT
);
```

## Migration Scripts

### Migration 001: Add Administrative Divisions

```sql
-- migrations/001_add_admin_columns.sql
ALTER TABLE "一生足迹" ADD COLUMN province TEXT;
ALTER TABLE "一生足迹" ADD COLUMN city TEXT;
ALTER TABLE "一生足迹" ADD COLUMN county TEXT;
ALTER TABLE "一生足迹" ADD COLUMN town TEXT;
ALTER TABLE "一生足迹" ADD COLUMN village TEXT;

CREATE INDEX idx_admin ON "一生足迹"(province, city, county);

INSERT INTO schema_version (version, applied_at, description)
VALUES (1, strftime('%s', 'now'), 'Add administrative division columns');
```

### Migration 002: Add Data Quality Flags

```sql
-- migrations/002_add_qa_flags.sql
ALTER TABLE "一生足迹" ADD COLUMN outlier_flag INTEGER DEFAULT 0;
ALTER TABLE "一生足迹" ADD COLUMN outlier_score REAL;
ALTER TABLE "一生足迹" ADD COLUMN outlier_reasons TEXT;
ALTER TABLE "一生足迹" ADD COLUMN qa_status TEXT DEFAULT 'pending';

CREATE INDEX idx_qa ON "一生足迹"(qa_status, outlier_flag);

INSERT INTO schema_version (version, applied_at, description)
VALUES (2, strftime('%s', 'now'), 'Add data quality flags');
```

### Migration 003: Add Metadata Fields

```sql
-- migrations/003_add_metadata.sql
ALTER TABLE "一生足迹" ADD COLUMN created_at INTEGER;
ALTER TABLE "一生足迹" ADD COLUMN updated_at INTEGER;
ALTER TABLE "一生足迹" ADD COLUMN algo_version TEXT;

-- Backfill created_at with dataTime
UPDATE "一生足迹" SET created_at = dataTime WHERE created_at IS NULL;

INSERT INTO schema_version (version, applied_at, description)
VALUES (3, strftime('%s', 'now'), 'Add metadata fields');
```

### Migration 004: Create Derived Tables

```sql
-- migrations/004_create_derived_tables.sql
CREATE TABLE stays (...);  -- See table definition above
CREATE TABLE trips (...);
CREATE TABLE footprint_stats (...);
CREATE TABLE admin_divisions (...);

INSERT INTO schema_version (version, applied_at, description)
VALUES (4, strftime('%s', 'now'), 'Create derived tables');
```

## Query Examples

### Get Points in Time Range

```sql
SELECT *
FROM "一生足迹"
WHERE dataTime BETWEEN 1737552138 AND 1737638538
ORDER BY dataTime;
```

### Get Points by Administrative Division

```sql
SELECT *
FROM "一生足迹"
WHERE province = '广东省'
  AND city = '深圳市'
  AND county = '南山区'
ORDER BY dataTime;
```

### Get Footprint Statistics

```sql
SELECT
  province,
  city,
  COUNT(*) as point_count,
  MIN(dataTime) as first_visit,
  MAX(dataTime) as last_visit
FROM "一生足迹"
WHERE qa_status = 'approved'
  AND outlier_flag = 0
GROUP BY province, city
ORDER BY point_count DESC;
```

### Get Stays Longer Than 2 Hours

```sql
SELECT *
FROM stays
WHERE duration >= 7200  -- 2 hours in seconds
ORDER BY duration DESC;
```

### Get Trips by Transport Mode

```sql
SELECT
  transport_mode,
  COUNT(*) as trip_count,
  SUM(distance) as total_distance,
  AVG(duration) as avg_duration
FROM trips
GROUP BY transport_mode
ORDER BY trip_count DESC;
```

## Performance Optimization

### Analyze Query Plans

```sql
EXPLAIN QUERY PLAN
SELECT * FROM "一生足迹"
WHERE dataTime BETWEEN 1737552138 AND 1737638538;
```

### Update Statistics

```sql
ANALYZE;
```

### Vacuum Database

```sql
VACUUM;
```

### Check Database Size

```sql
SELECT
  page_count * page_size / 1024 / 1024 as size_mb
FROM pragma_page_count(), pragma_page_size();
```

## Backup and Restore

### Backup

```bash
# WAL checkpoint before backup
sqlite3 tracks.db "PRAGMA wal_checkpoint(TRUNCATE);"

# Copy database file
cp tracks.db tracks_backup_$(date +%Y%m%d).db
```

### Restore

```bash
# Stop application
# Replace database file
cp tracks_backup_20250122.db tracks.db
# Restart application
```

### Export to CSV

```sql
.mode csv
.output tracks_export.csv
SELECT * FROM "一生足迹";
.output stdout
```

## Related Documentation

- `data-pipeline.md` - Complete data pipeline
- `processing-scripts.md` - Python script documentation
- `api-endpoints.md` - API specification
- `implementation-guide.md` - Implementation guide
