-- Migration 014: Create Phase 5 analysis tables
-- Purpose: Support 5 complex Python-based trajectory analysis skills
-- Skills: stay_detection (advanced), density_structure_advanced, trip_construction_advanced,
--         spatial_persona, admin_view_engine_advanced

-- 1. Density Clusters Table
-- DBSCAN clustering results for spatial density analysis
CREATE TABLE IF NOT EXISTS density_clusters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cluster_id INTEGER NOT NULL,
    center_lat REAL,
    center_lon REAL,
    point_count INTEGER,
    density_score REAL,
    cluster_type TEXT, -- HOME/WORK/FREQUENT/OCCASIONAL
    radius_m REAL,
    convex_hull_area_km2 REAL,
    province TEXT,
    city TEXT,
    county TEXT,
    confidence REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_density_clusters_id ON density_clusters(cluster_id);
CREATE INDEX IF NOT EXISTS idx_density_clusters_type ON density_clusters(cluster_type);
CREATE INDEX IF NOT EXISTS idx_density_clusters_score ON density_clusters(density_score DESC);
CREATE INDEX IF NOT EXISTS idx_density_clusters_admin ON density_clusters(province, city, county);

-- 2. Spatial Persona Table
-- Spatial behavior persona profiles
CREATE TABLE IF NOT EXISTS spatial_persona (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_date TEXT, -- YYYY-MM-DD or NULL for all-time
    mobility_score REAL, -- 0-100
    exploration_score REAL, -- 0-100
    routine_score REAL, -- 0-100
    diversity_score REAL, -- 0-100
    total_distance_km REAL,
    unique_locations INTEGER,
    revisit_ratio REAL,
    primary_mode TEXT,
    insights_json TEXT, -- JSON array of insights
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_spatial_persona_date ON spatial_persona(persona_date);
CREATE INDEX IF NOT EXISTS idx_spatial_persona_mobility ON spatial_persona(mobility_score DESC);
CREATE INDEX IF NOT EXISTS idx_spatial_persona_exploration ON spatial_persona(exploration_score DESC);

-- 3. Admin Trends Table
-- Administrative analytics with time-series trends
CREATE TABLE IF NOT EXISTS admin_trends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_level TEXT NOT NULL, -- PROVINCE/CITY/COUNTY/TOWN
    admin_name TEXT NOT NULL,
    trend_type TEXT, -- GROWTH/DECLINE/STABLE/SEASONAL
    trend_score REAL, -- -1 to 1
    seasonality_detected INTEGER, -- 0/1
    anomalies_json TEXT, -- JSON array of anomaly timestamps
    prediction_next_month INTEGER, -- Predicted visits
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_admin_trends_level ON admin_trends(admin_level);
CREATE INDEX IF NOT EXISTS idx_admin_trends_name ON admin_trends(admin_name);
CREATE INDEX IF NOT EXISTS idx_admin_trends_type ON admin_trends(trend_type);
CREATE INDEX IF NOT EXISTS idx_admin_trends_score ON admin_trends(trend_score DESC);
