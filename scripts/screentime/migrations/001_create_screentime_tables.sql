-- Screentime Module Database Schema
-- Created: 2026-02-24
-- Purpose: Store phone screen time usage data from Vivo X90

-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;

-- Daily app usage summary
-- Source: AppUsage_day_*.txt
CREATE TABLE IF NOT EXISTS screentime_daily (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,                    -- YYYYMMDD format
  app_name TEXT NOT NULL,                -- Chinese app name
  package_id TEXT NOT NULL,              -- com.tencent.mm
  duration_ms INTEGER NOT NULL,          -- Usage duration in milliseconds
  launch_count INTEGER DEFAULT 0,        -- Number of times app was launched
  notification_count INTEGER DEFAULT 0,  -- Number of notifications received
  split_screen_ms INTEGER DEFAULT 0,     -- Split screen duration (for ALL entry)
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(date, package_id)
);

CREATE INDEX idx_screentime_daily_date ON screentime_daily(date);
CREATE INDEX idx_screentime_daily_package ON screentime_daily(package_id);
CREATE INDEX idx_screentime_daily_duration ON screentime_daily(duration_ms DESC);
CREATE INDEX idx_screentime_daily_date_duration ON screentime_daily(date, duration_ms DESC);

-- Usage sessions (detailed line data)
-- Source: AppUsage_line_*.txt
CREATE TABLE IF NOT EXISTS screentime_sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,                    -- YYYYMMDD format
  start_time_ms INTEGER NOT NULL,        -- Unix timestamp in milliseconds
  end_time_ms INTEGER NOT NULL,          -- Unix timestamp in milliseconds
  start_time TEXT NOT NULL,              -- HH:MM format
  end_time TEXT NOT NULL,                -- HH:MM format
  app_name TEXT NOT NULL,
  package_id TEXT NOT NULL,
  duration_text TEXT,                    -- "1分钟", "少于一分钟"
  is_streaming BOOLEAN DEFAULT 0,        -- 流媒体 flag
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_screentime_sessions_date ON screentime_sessions(date);
CREATE INDEX idx_screentime_sessions_package ON screentime_sessions(package_id);
CREATE INDEX idx_screentime_sessions_start ON screentime_sessions(start_time_ms);
CREATE INDEX idx_screentime_sessions_date_start ON screentime_sessions(date, start_time_ms);

-- Unlock records
-- Source: AppUsage_unlock_*.txt
CREATE TABLE IF NOT EXISTS screentime_unlocks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,                    -- YYYYMMDD format
  unlock_time TEXT NOT NULL,             -- HH:MM format
  session_duration TEXT,                 -- "1分钟", "少于一分钟"
  app_sequence TEXT,                     -- JSON array of apps used in this session
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_screentime_unlocks_date ON screentime_unlocks(date);
CREATE INDEX idx_screentime_unlocks_time ON screentime_unlocks(unlock_time);

-- App metadata (for caching and categorization)
CREATE TABLE IF NOT EXISTS screentime_apps (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  package_id TEXT NOT NULL UNIQUE,
  app_name TEXT NOT NULL,
  category TEXT,                         -- Social, Entertainment, Productivity, etc.
  first_seen TEXT,                       -- YYYYMMDD
  last_seen TEXT,                        -- YYYYMMDD
  total_duration_ms INTEGER DEFAULT 0,
  total_launches INTEGER DEFAULT 0,
  total_notifications INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_screentime_apps_package ON screentime_apps(package_id);
CREATE INDEX idx_screentime_apps_category ON screentime_apps(category);
CREATE INDEX idx_screentime_apps_duration ON screentime_apps(total_duration_ms DESC);

-- Statistics cache (for performance optimization)
CREATE TABLE IF NOT EXISTS screentime_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  stat_type TEXT NOT NULL,               -- daily, weekly, monthly, hourly
  stat_date TEXT NOT NULL,               -- YYYYMMDD, YYYYWW, YYYYMM, YYYYMMDDHH
  total_duration_ms INTEGER,
  total_unlocks INTEGER,
  total_launches INTEGER,
  total_notifications INTEGER,
  unique_apps INTEGER,
  top_app_package TEXT,
  top_app_duration_ms INTEGER,
  data TEXT,                             -- JSON blob for detailed stats
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(stat_type, stat_date)
);

CREATE INDEX idx_screentime_stats_type ON screentime_statistics(stat_type);
CREATE INDEX idx_screentime_stats_date ON screentime_statistics(stat_date);
CREATE INDEX idx_screentime_stats_type_date ON screentime_statistics(stat_type, stat_date);
