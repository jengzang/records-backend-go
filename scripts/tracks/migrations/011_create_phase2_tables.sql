-- Migration 011: Create Phase 2 analysis tables
-- Purpose: Support speed_events, rendering_metadata, and stay_annotation skills

-- Speed Events Table
-- Stores detected high-speed events from CAR segments
CREATE TABLE IF NOT EXISTS speed_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    segment_id INTEGER,  -- FK to segments
    start_ts INTEGER NOT NULL,
    end_ts INTEGER NOT NULL,
    duration_s INTEGER NOT NULL,
    max_speed_mps REAL NOT NULL,
    avg_speed_mps REAL NOT NULL,
    peak_ts INTEGER NOT NULL,
    peak_lat REAL NOT NULL,
    peak_lon REAL NOT NULL,
    province TEXT,
    city TEXT,
    county TEXT,
    town TEXT,
    grid_id TEXT,
    confidence REAL,  -- 0-1
    reason_codes TEXT,  -- JSON array
    profile_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT,
    FOREIGN KEY (segment_id) REFERENCES segments(id)
);

CREATE INDEX IF NOT EXISTS idx_speed_events_max_speed ON speed_events(max_speed_mps DESC);
CREATE INDEX IF NOT EXISTS idx_speed_events_ts ON speed_events(start_ts);
CREATE INDEX IF NOT EXISTS idx_speed_events_segment ON speed_events(segment_id);

-- Stay Annotations Table
-- Stores user-confirmed or suggested labels for stay segments
CREATE TABLE IF NOT EXISTS stay_annotations (
    stay_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL,  -- HOME, WORK, EAT, SLEEP, etc.
    sub_label TEXT,
    note TEXT,
    confirmed INTEGER DEFAULT 0,  -- 0=suggested, 1=confirmed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    label_version TEXT DEFAULT 'v1',
    FOREIGN KEY (stay_id) REFERENCES stay_segments(id)
);

CREATE INDEX IF NOT EXISTS idx_stay_annotations_label ON stay_annotations(label);
CREATE INDEX IF NOT EXISTS idx_stay_annotations_confirmed ON stay_annotations(confirmed);

-- Stay Context Cache Table
-- Stores computed context cards and label suggestions for stays
CREATE TABLE IF NOT EXISTS stay_context_cache (
    stay_id INTEGER PRIMARY KEY,
    context_json TEXT NOT NULL,  -- Compressed JSON with time/location/arrival/departure context
    suggestions_json TEXT NOT NULL,  -- Candidate labels with confidence scores
    computed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT,
    FOREIGN KEY (stay_id) REFERENCES stay_segments(id)
);

CREATE INDEX IF NOT EXISTS idx_stay_context_computed ON stay_context_cache(computed_at);

-- Place Anchors Table
-- Stores known important places (HOME, WORK, etc.) with spatial boundaries
CREATE TABLE IF NOT EXISTS place_anchors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,  -- HOME, WORK, etc.
    grid_id TEXT NOT NULL,
    center_lat REAL,
    center_lon REAL,
    radius_m REAL DEFAULT 500,
    active_from_ts INTEGER,
    active_to_ts INTEGER,  -- NULL = still active
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_place_anchors_type ON place_anchors(type);
CREATE INDEX IF NOT EXISTS idx_place_anchors_grid ON place_anchors(grid_id);
CREATE INDEX IF NOT EXISTS idx_place_anchors_active ON place_anchors(active_from_ts, active_to_ts);

-- Render Segments Cache Table (Optional)
-- Stores pre-computed rendering metadata for map visualization
CREATE TABLE IF NOT EXISTS render_segments_cache (
    segment_id INTEGER NOT NULL,
    lod INTEGER NOT NULL,  -- 0=low, 1=medium, 2=high
    geojson_blob TEXT,  -- Compressed GeoJSON
    speed_bucket INTEGER,
    overlap_rank REAL,
    line_weight_hint REAL,
    alpha_hint REAL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (segment_id, lod),
    FOREIGN KEY (segment_id) REFERENCES segments(id)
);

CREATE INDEX IF NOT EXISTS idx_render_cache_updated ON render_segments_cache(updated_at);
