-- Migration 008: Create grid_cells table for spatial analysis
-- Purpose: Store multi-level grid system for heatmap and spatial aggregation

CREATE TABLE IF NOT EXISTS grid_cells (
    grid_id TEXT PRIMARY KEY,  -- Format: "L{level}_{x}_{y}"
    level INTEGER NOT NULL,  -- 1-15 (zoom level)
    bbox_min_lat REAL NOT NULL,
    bbox_min_lon REAL NOT NULL,
    bbox_max_lat REAL NOT NULL,
    bbox_max_lon REAL NOT NULL,
    center_lat REAL NOT NULL,
    center_lon REAL NOT NULL,
    point_count INTEGER DEFAULT 0,
    visit_count INTEGER DEFAULT 0,
    first_visit INTEGER,  -- Unix timestamp
    last_visit INTEGER,
    total_duration_s INTEGER DEFAULT 0,
    modes TEXT,  -- JSON array of modes
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_grid_level ON grid_cells(level);
CREATE INDEX IF NOT EXISTS idx_grid_point_count ON grid_cells(point_count);
CREATE INDEX IF NOT EXISTS idx_grid_visit_count ON grid_cells(visit_count);
CREATE INDEX IF NOT EXISTS idx_grid_bbox ON grid_cells(bbox_min_lat, bbox_min_lon, bbox_max_lat, bbox_max_lon);
