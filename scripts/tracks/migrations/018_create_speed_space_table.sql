-- Migration 018: Create speed_space_stats_bucketed table
-- Purpose: Store speed-space coupling analysis results
-- Skill: 14_speed_space_coupling

CREATE TABLE IF NOT EXISTS speed_space_stats_bucketed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bucket_type TEXT NOT NULL,  -- 'year', 'month', 'all'
    bucket_key TEXT NOT NULL,   -- '2024', '2024-01', 'all'
    area_type TEXT NOT NULL,    -- 'PROVINCE', 'CITY', 'COUNTY'
    area_key TEXT NOT NULL,     -- Area name
    avg_speed REAL NOT NULL,    -- Distance-weighted average speed (km/h)
    speed_variance REAL,        -- Distance-weighted speed variance
    speed_entropy REAL,         -- Shannon entropy of speed distribution
    total_distance REAL,        -- Total distance in this area (meters)
    segment_count INTEGER,      -- Number of segments
    is_high_speed_zone BOOLEAN DEFAULT 0,  -- >90th percentile
    is_slow_life_zone BOOLEAN DEFAULT 0,   -- <25th percentile
    stay_intensity REAL,        -- Stay intensity (for correlation analysis)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version INTEGER DEFAULT 1,
    UNIQUE(bucket_type, bucket_key, area_type, area_key)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_speed_space_bucket ON speed_space_stats_bucketed(bucket_type, bucket_key);
CREATE INDEX IF NOT EXISTS idx_speed_space_area ON speed_space_stats_bucketed(area_type, area_key);
CREATE INDEX IF NOT EXISTS idx_speed_space_high_speed ON speed_space_stats_bucketed(is_high_speed_zone) WHERE is_high_speed_zone = 1;
CREATE INDEX IF NOT EXISTS idx_speed_space_slow_life ON speed_space_stats_bucketed(is_slow_life_zone) WHERE is_slow_life_zone = 1;
CREATE INDEX IF NOT EXISTS idx_speed_space_avg_speed ON speed_space_stats_bucketed(avg_speed DESC);
