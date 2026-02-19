-- Migration 004: Extend track points table with analysis fields
-- Purpose: Add fields required for 30 trajectory analysis skills

-- Quality Control fields
ALTER TABLE "一生足迹" ADD COLUMN outlier_flag BOOLEAN DEFAULT 0;
ALTER TABLE "一生足迹" ADD COLUMN outlier_reason_codes TEXT;
ALTER TABLE "一生足迹" ADD COLUMN qa_status TEXT DEFAULT 'pending';

-- Behavior Classification fields
ALTER TABLE "一生足迹" ADD COLUMN mode TEXT;
ALTER TABLE "一生足迹" ADD COLUMN mode_confidence REAL;
ALTER TABLE "一生足迹" ADD COLUMN mode_reason_codes TEXT;
ALTER TABLE "一生足迹" ADD COLUMN segment_id INTEGER;

-- Stay Detection fields
ALTER TABLE "一生足迹" ADD COLUMN stay_id INTEGER;
ALTER TABLE "一生足迹" ADD COLUMN is_stay_point BOOLEAN DEFAULT 0;

-- Trajectory Completion fields
ALTER TABLE "一生足迹" ADD COLUMN is_synthetic BOOLEAN DEFAULT 0;
ALTER TABLE "一生足迹" ADD COLUMN synthetic_source TEXT;
ALTER TABLE "一生足迹" ADD COLUMN synthetic_metadata TEXT;

-- Spatial Analysis fields
ALTER TABLE "一生足迹" ADD COLUMN grid_id TEXT;
ALTER TABLE "一生足迹" ADD COLUMN grid_level INTEGER;
ALTER TABLE "一生足迹" ADD COLUMN revisit_count INTEGER DEFAULT 0;

-- Visualization fields
ALTER TABLE "一生足迹" ADD COLUMN render_color TEXT;
ALTER TABLE "一生足迹" ADD COLUMN render_width REAL;
ALTER TABLE "一生足迹" ADD COLUMN render_opacity REAL;
ALTER TABLE "一生足迹" ADD COLUMN lod_level INTEGER DEFAULT 0;

-- Create indexes for frequently queried fields
CREATE INDEX IF NOT EXISTS idx_mode ON "一生足迹"(mode);
CREATE INDEX IF NOT EXISTS idx_segment_id ON "一生足迹"(segment_id);
CREATE INDEX IF NOT EXISTS idx_stay_id ON "一生足迹"(stay_id);
CREATE INDEX IF NOT EXISTS idx_grid_id ON "一生足迹"(grid_id);
CREATE INDEX IF NOT EXISTS idx_outlier_flag ON "一生足迹"(outlier_flag);
CREATE INDEX IF NOT EXISTS idx_qa_status ON "一生足迹"(qa_status);
