-- Migration 013: Create Phase 4 analysis tables
-- Purpose: Support 12 new trajectory analysis skills
-- Skills: admin_crossings, admin_view_engine, utilization_efficiency, altitude_dimension,
--         road_overlap, spatial_complexity, directional_bias, density_structure,
--         time_space_slicing, time_space_compression, trip_construction, time_axis_map

-- 1. Admin Crossings Table
-- Stores administrative boundary crossing events
CREATE TABLE IF NOT EXISTS admin_crossings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    crossing_ts INTEGER NOT NULL,
    from_province TEXT,
    from_city TEXT,
    from_county TEXT,
    from_town TEXT,
    to_province TEXT,
    to_city TEXT,
    to_county TEXT,
    to_town TEXT,
    crossing_type TEXT NOT NULL, -- PROVINCE/CITY/COUNTY/TOWN
    latitude REAL,
    longitude REAL,
    distance_from_prev_m REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_admin_crossings_ts ON admin_crossings(crossing_ts);
CREATE INDEX IF NOT EXISTS idx_admin_crossings_type ON admin_crossings(crossing_type);
CREATE INDEX IF NOT EXISTS idx_admin_crossings_from_province ON admin_crossings(from_province);
CREATE INDEX IF NOT EXISTS idx_admin_crossings_to_province ON admin_crossings(to_province);

-- 2. Admin Stats Table
-- Multi-level administrative view statistics
CREATE TABLE IF NOT EXISTS admin_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_level TEXT NOT NULL, -- PROVINCE/CITY/COUNTY/TOWN
    admin_name TEXT NOT NULL,
    parent_name TEXT,
    visit_count INTEGER DEFAULT 0,
    total_duration_s INTEGER DEFAULT 0,
    unique_days INTEGER DEFAULT 0,
    first_visit_ts INTEGER,
    last_visit_ts INTEGER,
    total_distance_m REAL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1',
    UNIQUE(admin_level, admin_name)
);

CREATE INDEX IF NOT EXISTS idx_admin_stats_level ON admin_stats(admin_level);
CREATE INDEX IF NOT EXISTS idx_admin_stats_name ON admin_stats(admin_name);
CREATE INDEX IF NOT EXISTS idx_admin_stats_parent ON admin_stats(parent_name);
CREATE INDEX IF NOT EXISTS idx_admin_stats_visits ON admin_stats(visit_count DESC);

-- 3. Utilization Metrics Table
-- Spatial utilization efficiency metrics
CREATE TABLE IF NOT EXISTS utilization_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_date TEXT, -- YYYY-MM-DD or NULL for all-time
    total_area_km2 REAL,
    visited_area_km2 REAL,
    utilization_ratio REAL,
    revisit_efficiency REAL,
    unique_grids INTEGER,
    total_visits INTEGER,
    avg_visits_per_grid REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_utilization_metrics_date ON utilization_metrics(metric_date);

-- 4. Altitude Events Table
-- Elevation change events (climbs/descents)
CREATE TABLE IF NOT EXISTS altitude_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL, -- CLIMB/DESCENT/PLATEAU
    start_ts INTEGER NOT NULL,
    end_ts INTEGER NOT NULL,
    start_altitude REAL,
    end_altitude REAL,
    altitude_change REAL,
    duration_s INTEGER,
    avg_grade REAL, -- Percentage grade
    max_grade REAL,
    distance_m REAL,
    province TEXT,
    city TEXT,
    county TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_altitude_events_type ON altitude_events(event_type);
CREATE INDEX IF NOT EXISTS idx_altitude_events_ts ON altitude_events(start_ts);
CREATE INDEX IF NOT EXISTS idx_altitude_events_change ON altitude_events(altitude_change DESC);

-- 5. Road Overlap Stats Table
-- Road network overlap analysis
CREATE TABLE IF NOT EXISTS road_overlap_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    segment_id INTEGER,
    on_road_distance_m REAL DEFAULT 0,
    off_road_distance_m REAL DEFAULT 0,
    overlap_ratio REAL,
    road_type TEXT, -- HIGHWAY/ARTERIAL/LOCAL/UNKNOWN
    confidence REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1',
    FOREIGN KEY (segment_id) REFERENCES segments(id)
);

CREATE INDEX IF NOT EXISTS idx_road_overlap_segment ON road_overlap_stats(segment_id);
CREATE INDEX IF NOT EXISTS idx_road_overlap_ratio ON road_overlap_stats(overlap_ratio);

-- 6. Complexity Metrics Table
-- Spatial complexity scores
CREATE TABLE IF NOT EXISTS complexity_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_date TEXT, -- YYYY-MM-DD or NULL for all-time
    trajectory_complexity REAL, -- 0-1 score
    direction_changes INTEGER,
    avg_turn_angle REAL,
    spatial_entropy REAL,
    path_efficiency REAL, -- Actual distance / straight-line distance
    tortuosity REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_complexity_metrics_date ON complexity_metrics(metric_date);
CREATE INDEX IF NOT EXISTS idx_complexity_metrics_score ON complexity_metrics(trajectory_complexity DESC);

-- 7. Directional Stats Table
-- Direction distribution and bias
CREATE TABLE IF NOT EXISTS directional_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_date TEXT, -- YYYY-MM-DD or NULL for all-time
    direction_bucket INTEGER NOT NULL, -- 0-7 (N, NE, E, SE, S, SW, W, NW)
    distance_m REAL DEFAULT 0,
    duration_s INTEGER DEFAULT 0,
    point_count INTEGER DEFAULT 0,
    percentage REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_directional_stats_date ON directional_stats(metric_date);
CREATE INDEX IF NOT EXISTS idx_directional_stats_bucket ON directional_stats(direction_bucket);

-- 8. Density Zones Table
-- High-density spatial areas
CREATE TABLE IF NOT EXISTS density_zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    grid_id TEXT NOT NULL,
    density_score REAL NOT NULL,
    point_count INTEGER,
    visit_count INTEGER,
    total_duration_s INTEGER,
    zone_type TEXT, -- HOT/WARM/COLD
    center_lat REAL,
    center_lon REAL,
    province TEXT,
    city TEXT,
    county TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_density_zones_grid ON density_zones(grid_id);
CREATE INDEX IF NOT EXISTS idx_density_zones_score ON density_zones(density_score DESC);
CREATE INDEX IF NOT EXISTS idx_density_zones_type ON density_zones(zone_type);

-- 9. Time Space Slices Table
-- Time-space aggregations
CREATE TABLE IF NOT EXISTS time_space_slices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slice_type TEXT NOT NULL, -- HOURLY/DAILY/WEEKLY/MONTHLY
    slice_key TEXT NOT NULL, -- e.g., "2025-01-22", "2025-01-W04", "14" (hour)
    admin_level TEXT, -- PROVINCE/CITY/COUNTY/TOWN or NULL
    admin_name TEXT,
    grid_id TEXT,
    point_count INTEGER DEFAULT 0,
    distance_m REAL DEFAULT 0,
    duration_s INTEGER DEFAULT 0,
    unique_locations INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1',
    UNIQUE(slice_type, slice_key, admin_level, admin_name, grid_id)
);

CREATE INDEX IF NOT EXISTS idx_time_space_slices_type ON time_space_slices(slice_type);
CREATE INDEX IF NOT EXISTS idx_time_space_slices_key ON time_space_slices(slice_key);
CREATE INDEX IF NOT EXISTS idx_time_space_slices_admin ON time_space_slices(admin_level, admin_name);

-- 10. Compressed Trajectories Table
-- Simplified trajectories for efficient storage/transmission
CREATE TABLE IF NOT EXISTS compressed_trajectories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    compression_type TEXT NOT NULL, -- DOUGLAS_PEUCKER/TIME_BASED/ADAPTIVE
    epsilon REAL, -- Simplification tolerance
    original_point_count INTEGER,
    compressed_point_count INTEGER,
    compression_ratio REAL,
    points_json TEXT NOT NULL, -- JSON array of simplified points
    start_ts INTEGER,
    end_ts INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_compressed_trajectories_ts ON compressed_trajectories(start_ts);
CREATE INDEX IF NOT EXISTS idx_compressed_trajectories_ratio ON compressed_trajectories(compression_ratio);

-- 11. Trips Table (Enhanced version from Phase 4)
-- Trip records constructed from segments and stays
CREATE TABLE IF NOT EXISTS trips (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    start_ts INTEGER NOT NULL,
    end_ts INTEGER NOT NULL,
    origin_lat REAL,
    origin_lon REAL,
    origin_province TEXT,
    origin_city TEXT,
    origin_county TEXT,
    dest_lat REAL,
    dest_lon REAL,
    dest_province TEXT,
    dest_city TEXT,
    dest_county TEXT,
    total_distance_m REAL,
    duration_s INTEGER,
    primary_mode TEXT, -- WALK/BIKE/CAR/TRAIN/PLANE
    segment_count INTEGER DEFAULT 0,
    stay_count INTEGER DEFAULT 0,
    purpose TEXT, -- COMMUTE/LEISURE/TRAVEL/UNKNOWN
    confidence REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_trips_start_ts ON trips(start_ts);
CREATE INDEX IF NOT EXISTS idx_trips_end_ts ON trips(end_ts);
CREATE INDEX IF NOT EXISTS idx_trips_origin_city ON trips(origin_city);
CREATE INDEX IF NOT EXISTS idx_trips_dest_city ON trips(dest_city);
CREATE INDEX IF NOT EXISTS idx_trips_mode ON trips(primary_mode);

-- 12. Time Axis Markers Table
-- Timeline visualization metadata
CREATE TABLE IF NOT EXISTS time_axis_markers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    marker_ts INTEGER NOT NULL,
    marker_type TEXT NOT NULL, -- SEGMENT_START/SEGMENT_END/STAY/EVENT
    entity_id INTEGER, -- ID of segment/stay/event
    entity_type TEXT, -- SEGMENT/STAY/SPEED_EVENT/ALTITUDE_EVENT
    latitude REAL,
    longitude REAL,
    label TEXT,
    icon TEXT,
    color TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_time_axis_markers_ts ON time_axis_markers(marker_ts);
CREATE INDEX IF NOT EXISTS idx_time_axis_markers_type ON time_axis_markers(marker_type);
CREATE INDEX IF NOT EXISTS idx_time_axis_markers_entity ON time_axis_markers(entity_type, entity_id);
