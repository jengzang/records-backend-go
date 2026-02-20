-- Migration 016: Add missing columns to analysis_tasks table
-- Purpose: Add progress_percent, threshold_profile_id, depends_on_task_ids, blocks_task_ids

ALTER TABLE analysis_tasks ADD COLUMN progress_percent INTEGER DEFAULT 0;
ALTER TABLE analysis_tasks ADD COLUMN threshold_profile_id INTEGER DEFAULT NULL;
ALTER TABLE analysis_tasks ADD COLUMN depends_on_task_ids TEXT DEFAULT NULL;
ALTER TABLE analysis_tasks ADD COLUMN blocks_task_ids TEXT DEFAULT NULL;

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_analysis_tasks_threshold_profile ON analysis_tasks(threshold_profile_id);
