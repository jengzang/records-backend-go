-- Migration 006: Create stay_segments table for stay detection
-- Purpose: Store detected stay segments with spatial and administrative criteria

CREATE TABLE IF NOT EXISTS stay_segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    stay_type TEXT NOT NULL,  -- SPATIAL, ADMIN_PROVINCE, ADMIN_CITY, ADMIN_COUNTY, ADMIN_TOWN
    start_time INTEGER NOT NULL,
    end_time INTEGER NOT NULL,
    duration_s INTEGER NOT NULL,
    center_lat REAL,
    center_lon REAL,
    radius_m REAL,
    province TEXT,
    city TEXT,
    county TEXT,
    town TEXT,
    village TEXT,
    point_count INTEGER DEFAULT 0,
    confidence REAL DEFAULT 0,
    reason_codes TEXT,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT '1.0'
);

CREATE INDEX IF NOT EXISTS idx_stay_type ON stay_segments(stay_type);
CREATE INDEX IF NOT EXISTS idx_stay_start_time ON stay_segments(start_time);
CREATE INDEX IF NOT EXISTS idx_stay_duration ON stay_segments(duration_s);
CREATE INDEX IF NOT EXISTS idx_stay_province ON stay_segments(province);
CREATE INDEX IF NOT EXISTS idx_stay_city ON stay_segments(city);
CREATE INDEX IF NOT EXISTS idx_stay_county ON stay_segments(county);
