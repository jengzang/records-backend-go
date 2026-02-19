-- Migration 010: Create system management tables
-- Purpose: Store threshold profiles and analysis task management

-- Threshold Profiles
CREATE TABLE IF NOT EXISTS threshold_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    params_json TEXT NOT NULL,  -- JSON object with all thresholds
    is_default BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default profile
INSERT OR IGNORE INTO threshold_profiles (name, description, params_json, is_default) VALUES (
    'default',
    'Default threshold profile for trajectory analysis',
    '{
        "outlier_detection": {
            "max_speed_kmh": 500,
            "max_acceleration_ms2": 10,
            "max_jump_distance_m": 5000
        },
        "transport_mode": {
            "walk_max_speed": 8,
            "car_min_speed": 15,
            "train_min_speed": 60,
            "flight_min_speed": 200
        },
        "stay_detection": {
            "spatial_radius_m": 200,
            "min_duration_s": 7200,
            "admin_min_duration_s": 3600
        },
        "grid_system": {
            "min_level": 8,
            "max_level": 15
        }
    }',
    1
);

-- Analysis Tasks
CREATE TABLE IF NOT EXISTS analysis_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    skill_name TEXT NOT NULL,  -- e.g., "outlier_detection", "stay_detection"
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, running, completed, failed
    mode TEXT NOT NULL DEFAULT 'incremental',  -- incremental, full_recompute
    total_points INTEGER DEFAULT 0,
    processed_points INTEGER DEFAULT 0,
    failed_points INTEGER DEFAULT 0,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    eta_seconds INTEGER,
    error_message TEXT,
    params_json TEXT,  -- Task-specific parameters
    result_summary TEXT,  -- JSON summary of results
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_analysis_tasks_skill ON analysis_tasks(skill_name);
CREATE INDEX IF NOT EXISTS idx_analysis_tasks_status ON analysis_tasks(status);
CREATE INDEX IF NOT EXISTS idx_analysis_tasks_created_at ON analysis_tasks(created_at DESC);

-- Spatial Analysis (generic table for advanced analysis results)
CREATE TABLE IF NOT EXISTS spatial_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    analysis_type TEXT NOT NULL,  -- TIME_SPACE_SLICING, DENSITY_STRUCTURE, etc.
    analysis_key TEXT NOT NULL,  -- Unique identifier for this analysis
    result_json TEXT NOT NULL,  -- JSON result data
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(analysis_type, analysis_key)
);

CREATE INDEX IF NOT EXISTS idx_spatial_analysis_type ON spatial_analysis(analysis_type);
CREATE INDEX IF NOT EXISTS idx_spatial_analysis_key ON spatial_analysis(analysis_key);
