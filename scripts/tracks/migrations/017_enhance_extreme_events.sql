-- Migration 017: Enhance extreme_events table
-- Purpose: Add missing columns for detailed extreme event tracking

-- Add event_category column
ALTER TABLE extreme_events ADD COLUMN event_category TEXT;

-- Add administrative division columns
ALTER TABLE extreme_events ADD COLUMN province TEXT;
ALTER TABLE extreme_events ADD COLUMN city TEXT;
ALTER TABLE extreme_events ADD COLUMN county TEXT;

-- Add context columns
ALTER TABLE extreme_events ADD COLUMN mode TEXT;
ALTER TABLE extreme_events ADD COLUMN segment_id INTEGER;

-- Add ranking column
ALTER TABLE extreme_events ADD COLUMN rank INTEGER;

-- Add algorithm version column
ALTER TABLE extreme_events ADD COLUMN algo_version TEXT DEFAULT 'v1';

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_extreme_category ON extreme_events(event_category);
CREATE INDEX IF NOT EXISTS idx_extreme_rank ON extreme_events(rank);
CREATE INDEX IF NOT EXISTS idx_extreme_admin ON extreme_events(province, city, county);

-- Update existing records to set event_category based on event_type
UPDATE extreme_events SET event_category =
    CASE
        WHEN event_type IN ('NORTHMOST', 'SOUTHMOST', 'EASTMOST', 'WESTMOST') THEN 'SPATIAL'
        WHEN event_type = 'MAX_SPEED' THEN 'SPEED'
        WHEN event_type = 'MAX_ALTITUDE' THEN 'ALTITUDE'
        ELSE 'OTHER'
    END
WHERE event_category IS NULL;
