-- Migration 002: Create Personal Efficiency Curve tables
-- Purpose: Store hourly efficiency scores, user profiles, and insights
-- Date: 2026-03-02
-- Feature: Cross-module Personal Efficiency Curve analysis

-- Hourly Efficiency Scores Table (每小时效率评分)
-- Stores efficiency scores for each hour of each day, aggregated from multiple data sources
CREATE TABLE IF NOT EXISTS hourly_efficiency_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,                     -- YYYY-MM-DD format
    hour INTEGER NOT NULL,                  -- 0-23

    -- Keyboard metrics (from keyboard module)
    typing_speed REAL,                      -- keystrokes per hour
    typing_speed_normalized REAL,           -- 0-100 normalized score

    -- ScreenTime metrics (from screentime module)
    work_app_ratio REAL,                    -- percentage of work apps usage (0-1)
    entertainment_app_ratio REAL,           -- percentage of entertainment apps (0-1)
    focus_session_count INTEGER,            -- number of focus sessions (>30min single app)
    app_switch_frequency REAL,              -- switches per hour
    work_app_ratio_normalized REAL,         -- 0-100 normalized score
    focus_normalized REAL,                  -- 0-100 normalized score

    -- Health metrics (from applehealth module)
    avg_heart_rate REAL,                    -- average heart rate (bpm)
    heart_rate_variability REAL,            -- HRV (ms) - higher is better
    step_count INTEGER,                     -- steps in this hour
    hrv_normalized REAL,                    -- 0-100 normalized score
    activity_normalized REAL,               -- 0-100 normalized score

    -- Composite efficiency score
    efficiency_score REAL NOT NULL,         -- 0-100 weighted average
    -- Weights: typing_speed(30%) + work_app_ratio(20%) + hrv(20%) + focus(15%) + activity(15%)

    -- Data quality flags
    has_keyboard_data BOOLEAN DEFAULT 0,
    has_screentime_data BOOLEAN DEFAULT 0,
    has_health_data BOOLEAN DEFAULT 0,
    data_completeness REAL,                 -- 0-1, percentage of available data sources

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(date, hour)
);

CREATE INDEX IF NOT EXISTS idx_hourly_efficiency_date ON hourly_efficiency_scores(date);
CREATE INDEX IF NOT EXISTS idx_hourly_efficiency_hour ON hourly_efficiency_scores(hour);
CREATE INDEX IF NOT EXISTS idx_hourly_efficiency_score ON hourly_efficiency_scores(efficiency_score);
CREATE INDEX IF NOT EXISTS idx_hourly_efficiency_date_hour ON hourly_efficiency_scores(date, hour);

-- Efficiency Curve Profiles Table (用户效率曲线画像)
-- Stores aggregated efficiency patterns for workdays vs weekends
CREATE TABLE IF NOT EXISTS efficiency_curve_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_type TEXT NOT NULL,             -- 'workday' or 'weekend'

    -- 24-hour efficiency curve (one score per hour)
    hour_0 REAL, hour_1 REAL, hour_2 REAL, hour_3 REAL,
    hour_4 REAL, hour_5 REAL, hour_6 REAL, hour_7 REAL,
    hour_8 REAL, hour_9 REAL, hour_10 REAL, hour_11 REAL,
    hour_12 REAL, hour_13 REAL, hour_14 REAL, hour_15 REAL,
    hour_16 REAL, hour_17 REAL, hour_18 REAL, hour_19 REAL,
    hour_20 REAL, hour_21 REAL, hour_22 REAL, hour_23 REAL,

    -- Peak hours analysis
    peak_hour INTEGER,                      -- hour with highest efficiency (0-23)
    peak_score REAL,                        -- efficiency score at peak hour
    peak_start_hour INTEGER,                -- start of peak period
    peak_end_hour INTEGER,                  -- end of peak period

    -- Low hours analysis
    low_hour INTEGER,                       -- hour with lowest efficiency (0-23)
    low_score REAL,                         -- efficiency score at low hour

    -- Chronotype classification
    chronotype TEXT,                        -- 'morning', 'evening', 'intermediate'
    chronotype_confidence REAL,             -- 0-1 confidence score

    -- Statistics
    avg_efficiency REAL,                    -- average efficiency across all hours
    std_efficiency REAL,                    -- standard deviation
    total_samples INTEGER,                  -- number of days used for calculation

    -- Date range
    start_date TEXT,                        -- YYYY-MM-DD
    end_date TEXT,                          -- YYYY-MM-DD

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(profile_type)
);

CREATE INDEX IF NOT EXISTS idx_efficiency_profile_type ON efficiency_curve_profiles(profile_type);
CREATE INDEX IF NOT EXISTS idx_efficiency_profile_chronotype ON efficiency_curve_profiles(chronotype);

-- Efficiency Insights Table (自动生成的洞察和建议)
-- Stores actionable insights derived from efficiency analysis
CREATE TABLE IF NOT EXISTS efficiency_insights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    insight_type TEXT NOT NULL,             -- 'peak_hours', 'chronotype', 'workday_weekend_diff', 'improvement_suggestion'
    priority INTEGER DEFAULT 0,             -- 0=low, 1=medium, 2=high

    -- Insight content
    title TEXT NOT NULL,                    -- Short title (e.g., "Your peak productivity is 9-11am")
    description TEXT NOT NULL,              -- Detailed explanation
    recommendation TEXT,                    -- Actionable recommendation

    -- Supporting data
    data TEXT,                              -- JSON format for charts/details
    confidence REAL,                        -- 0-1 confidence score

    -- Metadata
    is_active BOOLEAN DEFAULT 1,            -- Can be dismissed by user
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_efficiency_insights_type ON efficiency_insights(insight_type);
CREATE INDEX IF NOT EXISTS idx_efficiency_insights_priority ON efficiency_insights(priority);
CREATE INDEX IF NOT EXISTS idx_efficiency_insights_active ON efficiency_insights(is_active);

-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=-64000;  -- 64MB cache
