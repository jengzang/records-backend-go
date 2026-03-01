-- Migration 001: Create Apple Health database schema
-- Purpose: Store health records, workouts, sleep data, and statistics
-- Date: 2026-03-01

-- Health Records Table (步数、心率、睡眠等健康记录)
CREATE TABLE IF NOT EXISTS health_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,                 -- HKQuantityTypeIdentifierStepCount, HKQuantityTypeIdentifierHeartRate, etc.
    value REAL NOT NULL,
    unit TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    source_name TEXT,
    source_version TEXT,
    device TEXT,
    creation_date TIMESTAMP,
    metadata TEXT,                      -- JSON format for additional data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_health_records_type ON health_records(type);
CREATE INDEX IF NOT EXISTS idx_health_records_start_date ON health_records(start_date);
CREATE INDEX IF NOT EXISTS idx_health_records_type_date ON health_records(type, start_date);

-- Workouts Table (运动记录)
CREATE TABLE IF NOT EXISTS workouts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_type TEXT NOT NULL,         -- HKWorkoutActivityTypeRunning, HKWorkoutActivityTypeCycling, etc.
    duration_seconds REAL,
    duration_unit TEXT,
    distance_meters REAL,
    distance_unit TEXT,
    calories REAL,
    calories_unit TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    source_name TEXT,
    source_version TEXT,
    device TEXT,
    creation_date TIMESTAMP,
    route_file TEXT,                    -- GPX file path if available
    metadata TEXT,                      -- JSON format for additional data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_workouts_type ON workouts(workout_type);
CREATE INDEX IF NOT EXISTS idx_workouts_start_date ON workouts(start_date);
CREATE INDEX IF NOT EXISTS idx_workouts_type_date ON workouts(workout_type, start_date);

-- Workout Routes Table (运动路线GPS点)
CREATE TABLE IF NOT EXISTS workout_routes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_id INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    altitude REAL,
    speed REAL,
    horizontal_accuracy REAL,
    vertical_accuracy REAL,
    course REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workout_routes_workout_id ON workout_routes(workout_id);
CREATE INDEX IF NOT EXISTS idx_workout_routes_timestamp ON workout_routes(timestamp);

-- Sleep Analysis Table (睡眠分析)
CREATE TABLE IF NOT EXISTS sleep_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sleep_type TEXT,                    -- HKCategoryValueSleepAnalysisInBed, Asleep, Awake
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    duration_seconds INTEGER,
    source_name TEXT,
    device TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sleep_analysis_start_date ON sleep_analysis(start_date);
CREATE INDEX IF NOT EXISTS idx_sleep_analysis_type ON sleep_analysis(sleep_type);

-- Activity Summary Table (每日活动摘要)
CREATE TABLE IF NOT EXISTS activity_summary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL UNIQUE,          -- YYYY-MM-DD format
    active_energy_burned REAL,
    active_energy_burned_unit TEXT,
    apple_exercise_time REAL,
    apple_exercise_time_unit TEXT,
    apple_stand_hours INTEGER,
    apple_stand_hours_goal INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_summary_date ON activity_summary(date);

-- Health Statistics Cache Table (统计数据缓存)
CREATE TABLE IF NOT EXISTS health_statistics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    stat_type TEXT NOT NULL,            -- daily, weekly, monthly, yearly
    stat_date TEXT NOT NULL,            -- YYYY-MM-DD, YYYY-Www, YYYY-MM, YYYY
    metric_type TEXT NOT NULL,          -- steps, heart_rate, sleep, distance, calories, etc.
    total_value REAL,
    avg_value REAL,
    min_value REAL,
    max_value REAL,
    count INTEGER,
    data TEXT,                          -- JSON format for detailed data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stat_type, stat_date, metric_type)
);

CREATE INDEX IF NOT EXISTS idx_health_statistics_type ON health_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_health_statistics_date ON health_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_health_statistics_metric ON health_statistics(metric_type);
CREATE INDEX IF NOT EXISTS idx_health_statistics_type_date_metric ON health_statistics(stat_type, stat_date, metric_type);

-- Metadata Table (元数据，存储导入信息)
CREATE TABLE IF NOT EXISTS import_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    import_date TIMESTAMP NOT NULL,
    export_date TEXT,                   -- From XML ExportDate
    total_records INTEGER DEFAULT 0,
    total_workouts INTEGER DEFAULT 0,
    total_sleep_records INTEGER DEFAULT 0,
    total_route_points INTEGER DEFAULT 0,
    file_size_bytes INTEGER,
    processing_time_seconds REAL,
    status TEXT,                        -- success, failed, partial
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=-64000;  -- 64MB cache
