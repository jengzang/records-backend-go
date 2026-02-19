# Database Design

## Overview

The backend uses SQLite databases with WAL mode for each module. This design provides:
- Simple deployment (no separate database server)
- Good performance for read-heavy workloads
- ACID transactions
- Low memory footprint

## Database Files

```
go-backend/data/
├── tracks/
│   └── tracks.db          # GPS trajectory data
├── keyboard/
│   └── kmcounter.db       # Keyboard/mouse usage data
├── screentime/
│   └── screentime.db      # Phone app usage data (planned)
├── applehealth/
│   └── health.db          # Health data (planned)
└── trainplane/
    └── routes.db          # Flight/train routes (planned)
```

## Common Patterns

### 1. WAL Mode

All databases use Write-Ahead Logging for better concurrency:

```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=-64000;  -- 64MB cache
PRAGMA temp_store=MEMORY;
```

### 2. Timestamps

Two formats for flexibility:
- `dataTime` INTEGER - Unix timestamp (seconds) for calculations
- `time_visually` TEXT - Human-readable format (YYYY/MM/DD HH:MM:SS.000)
- `time` TEXT - Compact format (YYYYMMDDHHMMSS) for sorting

### 3. Versioning

Track data lineage and algorithm versions:
- `created_at` INTEGER - Unix timestamp when record was created
- `updated_at` INTEGER - Unix timestamp when record was last updated
- `algo_version` TEXT - Algorithm version used to generate derived data

### 4. Indexes

Create indexes on frequently queried fields:
- Time-based queries: `CREATE INDEX idx_time ON table(dataTime)`
- Administrative divisions: `CREATE INDEX idx_admin ON table(province, city, county)`
- Composite indexes for common query patterns

### 5. Foreign Keys

Enable foreign key constraints:

```sql
PRAGMA foreign_keys=ON;
```

## Schema Design Principles

### 1. Normalization

- Separate tables for different entity types
- Avoid data duplication
- Use foreign keys for relationships

### 2. Denormalization for Performance

- Pre-compute aggregations in separate tables
- Store derived data with `algo_version` for cache invalidation
- Duplicate frequently accessed fields to avoid joins

### 3. Data Quality Flags

Add quality indicators to raw data:
- `outlier_flag` INTEGER - 0=normal, 1=outlier
- `outlier_score` REAL - Confidence score (0-1)
- `outlier_reasons` TEXT - JSON array of reason codes
- `qa_status` TEXT - 'pending', 'approved', 'rejected'

### 4. Soft Deletes

Use soft deletes instead of hard deletes:
- `deleted_at` INTEGER - NULL if not deleted, timestamp if deleted
- Filter with `WHERE deleted_at IS NULL` in queries

## Module-Specific Schemas

### Tracks Module

See `tracks/database-schema.md` for detailed schema.

**Main Tables:**
- `一生足迹` - Raw GPS points
- `stays` - Detected stays (>2 hours)
- `trips` - Constructed trips
- `footprint_stats` - Pre-computed statistics
- `admin_divisions` - Administrative division lookup

### Keyboard Module

See `keyboard/database-schema.md` for detailed schema.

**Main Tables:**
- `keyboard_data` - Daily keystroke counts
- `mouse_data` - Daily mouse action counts
- `scan_codes` - Detailed scan code statistics

### Screen Time Module

See `screentime/database-schema.md` for planned schema.

**Planned Tables:**
- `app_usage` - Per-app usage records
- `unlock_events` - Phone unlock events
- `daily_summary` - Daily aggregations

### Apple Health Module

See `applehealth/database-schema.md` for planned schema.

**Planned Tables:**
- `health_records` - Health metrics
- `workout_routes` - GPX route data
- `workout_summary` - Workout aggregations

### Train/Plane Module

See `trainplane/database-schema.md` for planned schema.

**Planned Tables:**
- `flights` - Flight records
- `trains` - Train journey records
- `route_points` - Route tracking points

## Backup Strategy

### 1. WAL Checkpointing

```sql
PRAGMA wal_checkpoint(TRUNCATE);
```

### 2. File-Based Backup

```bash
# Backup all databases
cp -r data/ backup/data-$(date +%Y%m%d)/
```

### 3. Export to CSV

```sql
.mode csv
.output backup.csv
SELECT * FROM table_name;
.output stdout
```

## Migration Strategy

### 1. Schema Versioning

Track schema version in metadata table:

```sql
CREATE TABLE schema_version (
  version INTEGER PRIMARY KEY,
  applied_at INTEGER NOT NULL,
  description TEXT
);
```

### 2. Migration Scripts

```sql
-- migrations/001_add_admin_columns.sql
ALTER TABLE 一生足迹 ADD COLUMN province TEXT;
ALTER TABLE 一生足迹 ADD COLUMN city TEXT;
ALTER TABLE 一生足迹 ADD COLUMN county TEXT;
ALTER TABLE 一生足迹 ADD COLUMN town TEXT;
ALTER TABLE 一生足迹 ADD COLUMN village TEXT;

INSERT INTO schema_version (version, applied_at, description)
VALUES (1, strftime('%s', 'now'), 'Add administrative division columns');
```

### 3. Rollback Strategy

- Keep backup before migration
- Test migrations on copy of production data
- Document rollback steps for each migration

## Performance Optimization

### 1. Query Optimization

- Use EXPLAIN QUERY PLAN to analyze queries
- Add indexes for slow queries
- Avoid SELECT * in production code
- Use prepared statements

### 2. Batch Operations

- Insert in batches of 1000-10000 records
- Use transactions for multiple operations
- Disable indexes during bulk import, rebuild after

### 3. Memory Management

- Set appropriate cache_size
- Use temp_store=MEMORY for temporary tables
- Monitor database file size growth

### 4. Vacuum

```sql
-- Reclaim unused space
VACUUM;

-- Analyze query planner statistics
ANALYZE;
```

## Monitoring

### 1. Database Size

```sql
SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();
```

### 2. Table Statistics

```sql
SELECT name, SUM(pgsize) as size
FROM dbstat
GROUP BY name
ORDER BY size DESC;
```

### 3. Index Usage

```sql
SELECT * FROM sqlite_stat1;
```

## Security

### 1. SQL Injection Prevention

- Always use parameterized queries
- Never concatenate user input into SQL
- Validate and sanitize all inputs

### 2. File Permissions

```bash
chmod 600 data/**/*.db  # Owner read/write only
```

### 3. Encryption

- Consider SQLCipher for encryption at rest
- Use HTTPS for data in transit
- Encrypt backups

## Troubleshooting

### Database Locked

```sql
-- Check for long-running transactions
SELECT * FROM pragma_wal_checkpoint;

-- Force checkpoint
PRAGMA wal_checkpoint(RESTART);
```

### Corruption

```sql
-- Check integrity
PRAGMA integrity_check;

-- Recover from backup
.recover backup.db
```

### Performance Issues

```sql
-- Analyze query plan
EXPLAIN QUERY PLAN SELECT ...;

-- Rebuild indexes
REINDEX;

-- Update statistics
ANALYZE;
```
