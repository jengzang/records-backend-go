-- Migration 025: Create time_space_compression_bucketed table
-- Skill: 27_time_space_compression (Time-Space Compression)
-- Purpose: Store time-space compression analysis results

CREATE TABLE IF NOT EXISTS time_space_compression_bucketed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Time bucketing
    bucket_type TEXT NOT NULL,  -- 'year', 'month', 'all'
    bucket_key TEXT,            -- 'YYYY', 'YYYY-MM', NULL

    -- Spatial bucketing
    area_type TEXT NOT NULL,    -- 'PROVINCE', 'CITY', 'COUNTY', 'ALL'
    area_key TEXT,              -- Province/city/county name, or NULL for ALL

    -- Movement intensity metrics
    movement_intensity REAL DEFAULT 0,     -- Total distance / Total time (km/h)
    burst_intensity REAL DEFAULT 0,        -- Max movement intensity in burst periods
    burst_count INTEGER DEFAULT 0,         -- Number of burst periods detected
    burst_duration_s INTEGER DEFAULT 0,    -- Total duration of burst periods

    -- Activity density metrics
    active_time_s INTEGER DEFAULT 0,       -- Time spent moving (speed > threshold)
    inactive_time_s INTEGER DEFAULT 0,     -- Time spent stationary
    activity_ratio REAL DEFAULT 0,         -- active_time / total_time
    effective_movement_ratio REAL DEFAULT 0, -- Distance in active time / Total distance

    -- Time-space efficiency
    avg_speed_kmh REAL DEFAULT 0,          -- Average speed during active periods
    max_speed_kmh REAL DEFAULT 0,          -- Maximum speed recorded
    distance_per_day REAL DEFAULT 0,       -- Average daily distance
    time_compression_index REAL DEFAULT 0, -- Composite efficiency metric

    -- Supporting data
    total_distance_m REAL DEFAULT 0,
    total_duration_s INTEGER DEFAULT 0,
    trip_count INTEGER DEFAULT 0,
    distinct_days INTEGER DEFAULT 0,

    -- Metadata
    created_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    updated_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    algo_version TEXT DEFAULT 'v1',

    UNIQUE(bucket_type, bucket_key, area_type, area_key)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_tsc_bucket ON time_space_compression_bucketed(bucket_type, bucket_key);
CREATE INDEX IF NOT EXISTS idx_tsc_area ON time_space_compression_bucketed(area_type, area_key);
CREATE INDEX IF NOT EXISTS idx_tsc_intensity ON time_space_compression_bucketed(movement_intensity DESC);
CREATE INDEX IF NOT EXISTS idx_tsc_burst ON time_space_compression_bucketed(burst_intensity DESC);
CREATE INDEX IF NOT EXISTS idx_tsc_efficiency ON time_space_compression_bucketed(time_compression_index DESC);
