-- Migration 023: Create spatial_density_grid_stats table
-- Skill: 13_density_structure (Density Structure Analysis)
-- Purpose: Store grid-based density analysis results with time bucketing

CREATE TABLE IF NOT EXISTS spatial_density_grid_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Time bucketing
    bucket_type TEXT NOT NULL,  -- 'year', 'rolling_90d', 'all'
    bucket_key TEXT,            -- 'YYYY', NULL

    -- Grid identification
    grid_id TEXT NOT NULL,
    center_lat REAL,
    center_lon REAL,

    -- Administrative context
    province TEXT,
    city TEXT,
    county TEXT,

    -- Density metrics
    density_score REAL NOT NULL,
    density_level TEXT NOT NULL,  -- 'core', 'secondary', 'active', 'peripheral', 'rare'
    stay_duration_s INTEGER DEFAULT 0,
    stay_count INTEGER DEFAULT 0,
    visit_days INTEGER DEFAULT 0,

    -- Cluster information
    cluster_id INTEGER,
    cluster_area_km2 REAL,

    -- Metadata
    created_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    updated_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    algo_version TEXT DEFAULT 'v1',

    UNIQUE(bucket_type, bucket_key, grid_id)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_density_bucket ON spatial_density_grid_stats(bucket_type, bucket_key);
CREATE INDEX IF NOT EXISTS idx_density_grid ON spatial_density_grid_stats(grid_id);
CREATE INDEX IF NOT EXISTS idx_density_score ON spatial_density_grid_stats(density_score DESC);
CREATE INDEX IF NOT EXISTS idx_density_level ON spatial_density_grid_stats(density_level);
CREATE INDEX IF NOT EXISTS idx_density_cluster ON spatial_density_grid_stats(cluster_id);
