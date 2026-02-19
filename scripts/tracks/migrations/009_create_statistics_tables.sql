-- Migration 009: Create statistics tables for aggregation
-- Purpose: Store footprint, stay, and extreme event statistics

-- Footprint Statistics
CREATE TABLE IF NOT EXISTS footprint_statistics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    stat_type TEXT NOT NULL,  -- PROVINCE, CITY, COUNTY, TOWN, GRID
    stat_key TEXT NOT NULL,  -- Province name, city name, grid_id, etc.
    time_range TEXT,  -- YYYY, YYYY-MM, YYYY-MM-DD, or "all"
    point_count INTEGER DEFAULT 0,
    visit_count INTEGER DEFAULT 0,
    first_visit INTEGER,
    last_visit INTEGER,
    total_distance_m REAL DEFAULT 0,
    total_duration_s INTEGER DEFAULT 0,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stat_type, stat_key, time_range)
);

CREATE INDEX IF NOT EXISTS idx_footprint_type ON footprint_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_footprint_key ON footprint_statistics(stat_key);
CREATE INDEX IF NOT EXISTS idx_footprint_time_range ON footprint_statistics(time_range);
CREATE INDEX IF NOT EXISTS idx_footprint_point_count ON footprint_statistics(point_count DESC);

-- Stay Statistics
CREATE TABLE IF NOT EXISTS stay_statistics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    stat_type TEXT NOT NULL,  -- PROVINCE, CITY, COUNTY, ACTIVITY_TYPE
    stat_key TEXT NOT NULL,
    time_range TEXT,
    stay_count INTEGER DEFAULT 0,
    total_duration_s INTEGER DEFAULT 0,
    avg_duration_s REAL DEFAULT 0,
    max_duration_s INTEGER DEFAULT 0,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stat_type, stat_key, time_range)
);

CREATE INDEX IF NOT EXISTS idx_stay_stat_type ON stay_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_stay_stat_key ON stay_statistics(stat_key);
CREATE INDEX IF NOT EXISTS idx_stay_time_range ON stay_statistics(time_range);
CREATE INDEX IF NOT EXISTS idx_stay_duration ON stay_statistics(total_duration_s DESC);

-- Extreme Events
CREATE TABLE IF NOT EXISTS extreme_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,  -- MAX_ALTITUDE, MAX_SPEED, NORTHMOST, SOUTHMOST, EASTMOST, WESTMOST
    point_id INTEGER NOT NULL,
    value REAL NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    timestamp INTEGER NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (point_id) REFERENCES "一生足迹"(id)
);

CREATE INDEX IF NOT EXISTS idx_extreme_type ON extreme_events(event_type);
CREATE INDEX IF NOT EXISTS idx_extreme_value ON extreme_events(value DESC);
