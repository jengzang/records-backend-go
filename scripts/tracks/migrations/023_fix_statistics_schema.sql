-- Migration 023: Fix footprint_statistics and stay_statistics schema
-- Purpose: Add missing columns to match code expectations
-- Date: 2026-02-24

-- Add missing columns to footprint_statistics
ALTER TABLE footprint_statistics ADD COLUMN province TEXT;
ALTER TABLE footprint_statistics ADD COLUMN city TEXT;
ALTER TABLE footprint_statistics ADD COLUMN county TEXT;
ALTER TABLE footprint_statistics ADD COLUMN town TEXT;
ALTER TABLE footprint_statistics ADD COLUMN total_distance_meters REAL DEFAULT 0;
ALTER TABLE footprint_statistics ADD COLUMN total_duration_seconds INTEGER DEFAULT 0;
ALTER TABLE footprint_statistics ADD COLUMN first_visit_time INTEGER;
ALTER TABLE footprint_statistics ADD COLUMN last_visit_time INTEGER;
ALTER TABLE footprint_statistics ADD COLUMN rank_by_points INTEGER;
ALTER TABLE footprint_statistics ADD COLUMN rank_by_visits INTEGER;
ALTER TABLE footprint_statistics ADD COLUMN rank_by_duration INTEGER;
ALTER TABLE footprint_statistics ADD COLUMN algo_version TEXT DEFAULT 'v1';

-- Add missing columns to stay_statistics
ALTER TABLE stay_statistics ADD COLUMN province TEXT;
ALTER TABLE stay_statistics ADD COLUMN city TEXT;
ALTER TABLE stay_statistics ADD COLUMN county TEXT;
ALTER TABLE stay_statistics ADD COLUMN stay_category TEXT;
ALTER TABLE stay_statistics ADD COLUMN rank_by_count INTEGER;
ALTER TABLE stay_statistics ADD COLUMN rank_by_duration INTEGER;
ALTER TABLE stay_statistics ADD COLUMN algo_version TEXT DEFAULT 'v1';
ALTER TABLE stay_statistics ADD COLUMN stay_type TEXT; -- SPATIAL, ADMIN_AREA

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_footprint_province ON footprint_statistics(province);
CREATE INDEX IF NOT EXISTS idx_footprint_city ON footprint_statistics(city);
CREATE INDEX IF NOT EXISTS idx_footprint_county ON footprint_statistics(county);

CREATE INDEX IF NOT EXISTS idx_stay_province ON stay_statistics(province);
CREATE INDEX IF NOT EXISTS idx_stay_city ON stay_statistics(city);
CREATE INDEX IF NOT EXISTS idx_stay_county ON stay_statistics(county);
CREATE INDEX IF NOT EXISTS idx_stay_type ON stay_statistics(stay_type);

-- Note: After running this migration, you need to regenerate statistics data
-- by running the footprint_stats and stay_stats analysis tasks
