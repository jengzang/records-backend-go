-- Migration 021: Add geohash6 column to stay_segments
-- This enables efficient spatial grouping for revisit pattern analysis

-- Add geohash6 column
ALTER TABLE stay_segments ADD COLUMN geohash6 TEXT;

-- Create index for efficient querying
CREATE INDEX IF NOT EXISTS idx_stay_segments_geohash6 ON stay_segments(geohash6);
