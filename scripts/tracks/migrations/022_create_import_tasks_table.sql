-- Migration: Create import_tasks table
-- Purpose: Track data import tasks with status and statistics
-- Date: 2026-02-23

CREATE TABLE IF NOT EXISTS import_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, running, completed, failed
    file_path TEXT,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mode TEXT NOT NULL DEFAULT 'append',  -- append, replace
    deduplicate INTEGER NOT NULL DEFAULT 1,  -- 0=false, 1=true
    auto_trigger INTEGER NOT NULL DEFAULT 1,  -- 0=false, 1=true (auto trigger geocoding)
    total_records INTEGER DEFAULT 0,
    new_records INTEGER DEFAULT 0,
    duplicate_records INTEGER DEFAULT 0,
    error_message TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    completed_at DATETIME
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_import_tasks_status ON import_tasks(status);
CREATE INDEX IF NOT EXISTS idx_import_tasks_created_at ON import_tasks(created_at DESC);
