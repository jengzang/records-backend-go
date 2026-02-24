-- ManicTime database schema for computer screentime data
-- This schema is designed for ManicTime Excel exports

-- Main activities table
CREATE TABLE IF NOT EXISTS manictime_activities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP NOT NULL,
  duration_seconds INTEGER NOT NULL,
  application TEXT NOT NULL,
  window_title TEXT,
  category TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_manictime_activities_date ON manictime_activities(date);
CREATE INDEX idx_manictime_activities_app ON manictime_activities(application);
CREATE INDEX idx_manictime_activities_start ON manictime_activities(start_time);
CREATE INDEX idx_manictime_activities_category ON manictime_activities(category);

-- Application metadata table
CREATE TABLE IF NOT EXISTS manictime_apps (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  application TEXT NOT NULL UNIQUE,
  category TEXT,
  first_seen TEXT,
  last_seen TEXT,
  total_duration_seconds INTEGER DEFAULT 0,
  total_sessions INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_manictime_apps_application ON manictime_apps(application);
CREATE INDEX idx_manictime_apps_category ON manictime_apps(category);
CREATE INDEX idx_manictime_apps_duration ON manictime_apps(total_duration_seconds DESC);

-- Daily statistics table
CREATE TABLE IF NOT EXISTS manictime_daily (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  application TEXT NOT NULL,
  total_duration_seconds INTEGER NOT NULL,
  session_count INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(date, application)
);

CREATE INDEX idx_manictime_daily_date ON manictime_daily(date);
CREATE INDEX idx_manictime_daily_app ON manictime_daily(application);
CREATE INDEX idx_manictime_daily_duration ON manictime_daily(total_duration_seconds DESC);

-- Category statistics cache
CREATE TABLE IF NOT EXISTS manictime_statistics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  stat_type TEXT NOT NULL,
  stat_date TEXT NOT NULL,
  total_duration_seconds INTEGER,
  total_sessions INTEGER,
  unique_apps INTEGER,
  top_app TEXT,
  top_app_duration_seconds INTEGER,
  data TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(stat_type, stat_date)
);

CREATE INDEX idx_manictime_stats_type ON manictime_statistics(stat_type);
CREATE INDEX idx_manictime_stats_date ON manictime_statistics(stat_date);
