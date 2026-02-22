-- Migration 020: Create revisit_patterns table
-- Tracks repeated visit patterns to locations (geohash6 precision)

CREATE TABLE IF NOT EXISTS revisit_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Location identification
    geohash6 TEXT NOT NULL UNIQUE,
    center_lat REAL,
    center_lon REAL,

    -- Administrative context
    province TEXT,
    city TEXT,
    county TEXT,

    -- Visit statistics
    visit_count INTEGER NOT NULL,
    first_visit INTEGER NOT NULL,
    last_visit INTEGER NOT NULL,
    total_duration_seconds INTEGER DEFAULT 0,

    -- Interval analysis
    avg_interval_days REAL,
    std_interval_days REAL,
    min_interval_days REAL,
    max_interval_days REAL,

    -- Pattern metrics
    regularity_score REAL,
    is_periodic INTEGER DEFAULT 0,
    is_habitual INTEGER DEFAULT 0,
    revisit_strength REAL,

    -- Metadata
    created_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    updated_at INTEGER DEFAULT (CAST(strftime('%s', 'now') AS INTEGER)),
    algo_version TEXT DEFAULT 'v1'
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_revisit_geohash ON revisit_patterns(geohash6);
CREATE INDEX IF NOT EXISTS idx_revisit_visits ON revisit_patterns(visit_count DESC);
CREATE INDEX IF NOT EXISTS idx_revisit_strength ON revisit_patterns(revisit_strength DESC);
CREATE INDEX IF NOT EXISTS idx_revisit_periodic ON revisit_patterns(is_periodic);
CREATE INDEX IF NOT EXISTS idx_revisit_habitual ON revisit_patterns(is_habitual);
CREATE INDEX IF NOT EXISTS idx_revisit_admin ON revisit_patterns(province, city, county);
