-- Migration: Add geocoding_tasks table for tracking batch geocoding operations
-- Version: 003
-- Date: 2026-02-19

CREATE TABLE IF NOT EXISTS geocoding_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed')),
    total_points INTEGER NOT NULL,
    processed_points INTEGER DEFAULT 0,
    failed_points INTEGER DEFAULT 0,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    eta_seconds INTEGER,
    error_message TEXT,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying tasks by status
CREATE INDEX IF NOT EXISTS idx_geocoding_tasks_status ON geocoding_tasks(status);

-- Index for querying tasks by creation time
CREATE INDEX IF NOT EXISTS idx_geocoding_tasks_created_at ON geocoding_tasks(created_at DESC);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_geocoding_tasks_timestamp
AFTER UPDATE ON geocoding_tasks
FOR EACH ROW
BEGIN
    UPDATE geocoding_tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
