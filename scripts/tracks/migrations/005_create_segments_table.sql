-- Migration 005: Create segments table for behavior classification
-- Purpose: Store transport mode segments (WALK/CAR/TRAIN/FLIGHT/STAY)

CREATE TABLE IF NOT EXISTS segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mode TEXT NOT NULL,  -- WALK, CAR, TRAIN, FLIGHT, STAY, UNKNOWN
    start_time INTEGER NOT NULL,  -- Unix timestamp
    end_time INTEGER NOT NULL,
    start_point_id INTEGER,
    end_point_id INTEGER,
    point_count INTEGER DEFAULT 0,
    distance_m REAL DEFAULT 0,
    duration_s INTEGER DEFAULT 0,
    avg_speed_kmh REAL DEFAULT 0,
    max_speed_kmh REAL DEFAULT 0,
    confidence REAL DEFAULT 0,  -- 0~1
    reason_codes TEXT,  -- JSON array of reason codes
    metadata TEXT,  -- JSON metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT '1.0',
    FOREIGN KEY (start_point_id) REFERENCES "一生足迹"(id),
    FOREIGN KEY (end_point_id) REFERENCES "一生足迹"(id)
);

CREATE INDEX IF NOT EXISTS idx_segments_mode ON segments(mode);
CREATE INDEX IF NOT EXISTS idx_segments_start_time ON segments(start_time);
CREATE INDEX IF NOT EXISTS idx_segments_end_time ON segments(end_time);
CREATE INDEX IF NOT EXISTS idx_segments_confidence ON segments(confidence);
