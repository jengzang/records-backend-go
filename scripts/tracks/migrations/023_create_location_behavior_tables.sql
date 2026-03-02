-- Location Behavior Correlation Analysis Tables
-- Created: 2026-03-02

-- Locations table: stores clustered locations from stay_segments
CREATE TABLE IF NOT EXISTS locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    geohash TEXT NOT NULL,
    center_lat REAL NOT NULL,
    center_lon REAL NOT NULL,
    radius REAL NOT NULL,  -- meters
    visit_count INTEGER DEFAULT 0,
    total_duration INTEGER DEFAULT 0,  -- seconds
    first_visit TEXT,
    last_visit TEXT,
    label TEXT,  -- HOME, OFFICE, CAFE, GYM, TRANSIT, LEISURE, UNKNOWN
    label_confidence REAL DEFAULT 0.0,  -- 0-1
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_locations_geohash ON locations(geohash);
CREATE INDEX IF NOT EXISTS idx_locations_label ON locations(label);

-- Location behaviors table: stores behavior data for each visit
CREATE TABLE IF NOT EXISTS location_behaviors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,
    visit_date TEXT NOT NULL,  -- YYYY-MM-DD
    visit_start TEXT NOT NULL,
    visit_end TEXT NOT NULL,
    duration INTEGER NOT NULL,  -- seconds

    -- Keyboard metrics
    typing_speed REAL DEFAULT 0.0,  -- keystrokes per hour

    -- Screentime metrics
    work_app_ratio REAL DEFAULT 0.0,  -- 0-1
    entertainment_ratio REAL DEFAULT 0.0,  -- 0-1
    focus_duration INTEGER DEFAULT 0,  -- seconds
    app_switch_count INTEGER DEFAULT 0,

    -- Health metrics
    avg_heart_rate REAL DEFAULT 0.0,
    heart_rate_variability REAL DEFAULT 0.0,
    steps INTEGER DEFAULT 0,

    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(id)
);

CREATE INDEX IF NOT EXISTS idx_location_behaviors_location ON location_behaviors(location_id);
CREATE INDEX IF NOT EXISTS idx_location_behaviors_date ON location_behaviors(visit_date);

-- Location efficiency scores table: stores calculated efficiency scores
CREATE TABLE IF NOT EXISTS location_efficiency_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,

    -- Component scores (0-100)
    productivity_score REAL DEFAULT 0.0,
    health_score REAL DEFAULT 0.0,
    focus_score REAL DEFAULT 0.0,

    -- Overall efficiency score (0-100)
    efficiency_score REAL DEFAULT 0.0,

    -- Statistics
    visit_count INTEGER DEFAULT 0,
    avg_duration INTEGER DEFAULT 0,  -- seconds

    calculated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(id)
);

CREATE INDEX IF NOT EXISTS idx_location_efficiency_location ON location_efficiency_scores(location_id);
CREATE INDEX IF NOT EXISTS idx_location_efficiency_score ON location_efficiency_scores(efficiency_score DESC);

-- Location habits table: stores detected habits for each location
CREATE TABLE IF NOT EXISTS location_habits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,
    habit_type TEXT NOT NULL,  -- app_usage, time_pattern, activity_pattern
    habit_description TEXT NOT NULL,
    confidence REAL DEFAULT 0.0,  -- 0-1
    occurrence_count INTEGER DEFAULT 0,
    detected_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(id)
);

CREATE INDEX IF NOT EXISTS idx_location_habits_location ON location_habits(location_id);
CREATE INDEX IF NOT EXISTS idx_location_habits_type ON location_habits(habit_type);

-- Location visualization cache table: stores pre-computed visualization data
CREATE TABLE IF NOT EXISTS location_visualization_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cache_type TEXT NOT NULL,  -- heatmap, markers, clusters
    cache_data TEXT NOT NULL,  -- JSON
    generated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_location_viz_cache_type ON location_visualization_cache(cache_type);
