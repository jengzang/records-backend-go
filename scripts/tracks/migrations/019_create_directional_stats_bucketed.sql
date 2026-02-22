-- Migration 019: Create directional_stats_bucketed table
-- Skill: 18_directional_bias (方向偏好分析)
-- Purpose: Store directional movement pattern statistics with time/mode bucketing

CREATE TABLE IF NOT EXISTS directional_stats_bucketed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Bucketing dimensions
    bucket_type TEXT NOT NULL,           -- 'year', 'month', 'all'
    bucket_key TEXT NOT NULL,            -- '2025', '2025-01', 'all'
    area_type TEXT NOT NULL,             -- 'PROVINCE', 'CITY', 'COUNTY'
    area_key TEXT NOT NULL,              -- Area name (e.g., '广东省', '深圳市')
    mode_filter TEXT DEFAULT 'ALL',      -- 'ALL', 'WALK', 'CAR', 'TRAIN', 'FLIGHT'

    -- Directional histogram (JSON array of 8 bins)
    -- Format: [{"bin": 0, "distance": 1234.5, "count": 10}, ...]
    direction_histogram_json TEXT,
    num_bins INTEGER DEFAULT 8,          -- Number of bins (8 or 16)

    -- Key metrics
    dominant_direction_deg REAL,         -- 0-360 degrees (weighted average of dominant bin)
    directional_concentration REAL,      -- 0-1 (vector synthesis magnitude)
    bidirectional_score REAL,            -- 0-1 (strength of back-and-forth pattern)
    directional_entropy REAL,            -- 0-1 (Shannon entropy, normalized)

    -- Aggregated statistics
    total_distance REAL,                 -- Total distance in meters
    total_duration INTEGER,              -- Total duration in seconds
    segment_count INTEGER,               -- Number of segments

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version INTEGER DEFAULT 1,

    -- Ensure uniqueness per bucket combination
    UNIQUE(bucket_type, bucket_key, area_type, area_key, mode_filter)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_directional_bucketed_bucket
    ON directional_stats_bucketed(bucket_type, bucket_key);

CREATE INDEX IF NOT EXISTS idx_directional_bucketed_area
    ON directional_stats_bucketed(area_type, area_key);

CREATE INDEX IF NOT EXISTS idx_directional_bucketed_mode
    ON directional_stats_bucketed(mode_filter);

CREATE INDEX IF NOT EXISTS idx_directional_bucketed_concentration
    ON directional_stats_bucketed(directional_concentration DESC);

CREATE INDEX IF NOT EXISTS idx_directional_bucketed_bidirectional
    ON directional_stats_bucketed(bidirectional_score DESC);
