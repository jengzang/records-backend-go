-- Migration: Create spatial_persona table
-- Description: Stores comprehensive spatial persona profiles

CREATE TABLE IF NOT EXISTS spatial_persona (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Comprehensive metrics
    psi REAL NOT NULL,                    -- Personal Spatial Index (0-100)
    behavior_type TEXT NOT NULL,          -- EXPLORER, COMMUTER, HOMEBODY, TRAVELER, WANDERER, BALANCED
    stability REAL NOT NULL,              -- Profile stability (0-1)

    -- Feature dimensions
    footprint_diversity REAL DEFAULT 0,   -- Log(unique locations)
    footprint_spread REAL DEFAULT 0,      -- Log(total points)
    movement_intensity REAL DEFAULT 0,    -- Average movement speed
    movement_burst REAL DEFAULT 0,        -- High-speed burst intensity
    spatial_complexity REAL DEFAULT 0,    -- Trajectory complexity
    spatial_entropy REAL DEFAULT 0,       -- Location distribution entropy
    temporal_regularity REAL DEFAULT 0,   -- Weekly pattern regularity
    temporal_coverage REAL DEFAULT 0,     -- Time coverage
    road_overlap REAL DEFAULT 0,          -- On-road ratio

    -- Full feature vector (JSON)
    features_json TEXT,

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT DEFAULT 'v1'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_persona_psi ON spatial_persona(psi DESC);
CREATE INDEX IF NOT EXISTS idx_persona_type ON spatial_persona(behavior_type);
CREATE INDEX IF NOT EXISTS idx_persona_created ON spatial_persona(created_at DESC);
