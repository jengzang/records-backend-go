-- Migration 007: Create trips table for trip construction
-- Purpose: Store constructed trips based on stay segments

CREATE TABLE IF NOT EXISTS trips (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,  -- YYYY-MM-DD
    trip_number INTEGER NOT NULL,  -- 1, 2, 3... for the day
    origin_stay_id INTEGER,
    dest_stay_id INTEGER,
    start_time INTEGER NOT NULL,
    end_time INTEGER NOT NULL,
    duration_s INTEGER NOT NULL,
    distance_m REAL DEFAULT 0,
    segment_count INTEGER DEFAULT 0,
    modes TEXT,  -- JSON array of modes used
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT '1.0',
    FOREIGN KEY (origin_stay_id) REFERENCES stay_segments(id),
    FOREIGN KEY (dest_stay_id) REFERENCES stay_segments(id)
);

CREATE INDEX IF NOT EXISTS idx_trips_date ON trips(date);
CREATE INDEX IF NOT EXISTS idx_trips_start_time ON trips(start_time);
CREATE INDEX IF NOT EXISTS idx_trips_origin ON trips(origin_stay_id);
CREATE INDEX IF NOT EXISTS idx_trips_dest ON trips(dest_stay_id);
