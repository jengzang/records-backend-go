-- Migration 024: Create altitude_stats_bucketed table
-- Skill: 28_altitude_dimension (Altitude Dimension Analysis)
-- Purpose: Store altitude-based analysis results with time bucketing

CREATE TABLE IF NOT EXISTS altitude_stats_bucketed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Time bucketing
    bucket_type TEXT NOT NULL,  -- 'year', 'month', 'all'
    bucket_key TEXT,            -- 'YYYY', 'YYYY-MM', NULL

    -- Spatial bucketing
    area_type TEXT NOT NULL,    -- 'PROVINCE', 'CITY', 'COUNTY', 'ALL'
    area_key TEXT,              -- Province/city/county name, or NULL for ALL

    -- Altitude statistics
    min_altitude REAL,
    max_altitude REAL,
    avg_altitude REAL,
    altitude_span REAL,         -- max - min

    -- Distribution percentiles
    p25_altitude REAL,
    p50_altitude REAL,
    p75_altitude REAL,
    p90_altitude REAL,

    -- Vertical movement metrics
    total_ascent REAL DEFAULT 0,      -- Cumulative elevation gain (meters)
    total_descent REAL DEFAULT 0,     -- Cumulative elevation loss (meters)
    vertical_intensity REAL DEFAULT 0, -- (ascent + descent) / distance

    -- Supporting data
    point_count INTEGER DEFAULT 0,
    segment_count INTEGER DEFAULT 0,
    total_distance REAL DEFAULT 0,

    -- Metadata
    created_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    updated_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    algo_version TEXT DEFAULT 'v1',

    UNIQUE(bucket_type, bucket_key, area_type, area_key)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_altitude_bucket ON altitude_stats_bucketed(bucket_type, bucket_key);
CREATE INDEX IF NOT EXISTS idx_altitude_area ON altitude_stats_bucketed(area_type, area_key);
CREATE INDEX IF NOT EXISTS idx_altitude_span ON altitude_stats_bucketed(altitude_span DESC);
CREATE INDEX IF NOT EXISTS idx_altitude_intensity ON altitude_stats_bucketed(vertical_intensity DESC);
