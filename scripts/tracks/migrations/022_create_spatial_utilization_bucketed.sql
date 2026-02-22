-- Migration 022: Create spatial_utilization_bucketed table
-- Skill: 16_utilization_efficiency (Spatial Utilization Efficiency)
-- Purpose: Distinguish destinations (high stay) from transit corridors (high pass-through)

CREATE TABLE IF NOT EXISTS spatial_utilization_bucketed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Time bucketing
    bucket_type TEXT NOT NULL,  -- 'month', 'year', 'all'
    bucket_key TEXT,            -- 'YYYY-MM', 'YYYY', NULL

    -- Spatial bucketing
    area_type TEXT NOT NULL,    -- 'province', 'city', 'county', 'town', 'grid'
    area_key TEXT NOT NULL,     -- Province name, city name, or grid_id

    -- Core metrics from skill spec
    transit_intensity INTEGER DEFAULT 0,      -- Distinct trip count passing through
    stay_duration_s INTEGER DEFAULT 0,        -- Total stay duration in seconds
    utilization_efficiency REAL DEFAULT 0,    -- stay / (transit + ε)
    transit_dominance REAL DEFAULT 0,         -- transit / (transit + stay)
    area_depth REAL DEFAULT 0,                -- log(1+stay) * log(1+days)
    coverage_efficiency REAL DEFAULT 0,       -- Distinct grids / Total grids

    -- Supporting data
    distinct_visit_days INTEGER DEFAULT 0,
    distinct_grids INTEGER DEFAULT 0,
    total_grids INTEGER DEFAULT 0,
    first_visit INTEGER,
    last_visit INTEGER,

    -- Metadata
    created_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    updated_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    algo_version TEXT DEFAULT 'v1',

    UNIQUE(bucket_type, bucket_key, area_type, area_key)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_util_bucket ON spatial_utilization_bucketed(bucket_type, bucket_key);
CREATE INDEX IF NOT EXISTS idx_util_area ON spatial_utilization_bucketed(area_type, area_key);
CREATE INDEX IF NOT EXISTS idx_util_efficiency ON spatial_utilization_bucketed(utilization_efficiency DESC);
CREATE INDEX IF NOT EXISTS idx_util_dominance ON spatial_utilization_bucketed(transit_dominance DESC);
CREATE INDEX IF NOT EXISTS idx_util_depth ON spatial_utilization_bucketed(area_depth DESC);
