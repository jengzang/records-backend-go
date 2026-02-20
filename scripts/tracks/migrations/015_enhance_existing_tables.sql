-- Migration 015: Enhance existing tables for Phase 5
-- Purpose: Add columns for advanced analysis features

-- 1. Enhance stay_segments table with DBSCAN clustering results
ALTER TABLE stay_segments ADD COLUMN cluster_id INTEGER;
ALTER TABLE stay_segments ADD COLUMN cluster_confidence REAL;
ALTER TABLE stay_segments ADD COLUMN cluster_type TEXT; -- HOME/WORK/FREQUENT/OCCASIONAL

CREATE INDEX IF NOT EXISTS idx_stay_segments_cluster ON stay_segments(cluster_id);
CREATE INDEX IF NOT EXISTS idx_stay_segments_cluster_type ON stay_segments(cluster_type);

-- 2. Enhance trips table with ML-based purpose classification
ALTER TABLE trips ADD COLUMN purpose_ml TEXT; -- COMMUTE/WORK/LEISURE/SHOPPING/TRAVEL/OTHER
ALTER TABLE trips ADD COLUMN confidence_ml REAL;
ALTER TABLE trips ADD COLUMN features_json TEXT; -- JSON of ML features used

CREATE INDEX IF NOT EXISTS idx_trips_purpose_ml ON trips(purpose_ml);
CREATE INDEX IF NOT EXISTS idx_trips_confidence_ml ON trips(confidence_ml DESC);
