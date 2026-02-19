# Scripts Directory Structure

This directory contains all Python scripts for data processing, organized by module and function.

## Directory Structure

```
scripts/
├── common/                      # Shared utility scripts
│   ├── check_dbf.py            # DBF file inspection
│   ├── check_shapefile.py      # Shapefile validation
│   └── check_schema.py         # Database schema verification
│
├── geocoding/                   # Geocoding service (independent module)
│   ├── geocode.py              # Core geocoding logic
│   ├── check_geocoded.py       # Verify geocoding results
│   ├── verify_geocoding.py     # Quality report generation
│   └── inspect_shapefile.py    # Shapefile inspection
│
├── tracks/                      # GPS trajectory processing
│   ├── import/
│   │   └── write2sql.py        # Excel → SQLite import
│   ├── analysis/
│   │   ├── stay_detection.py   # Stay detection algorithm
│   │   ├── stay_detection_v2.py # Alternative implementation
│   │   ├── statistics.py       # Location/time statistics
│   │   ├── statistics_v2.py    # Alternative implementation
│   │   └── photos.py           # Visualization (to be moved)
│   ├── migrations/
│   │   └── (SQL migration files)
│   └── run_migration.py        # Migration runner
│
└── keyboard/                    # Keyboard/mouse usage tracking
    ├── import/
    │   ├── ini_parser.py       # Parse KMCounter.ini
    │   ├── ini_to_sqlite.py    # Import to SQLite
    │   └── verify_database.py  # Database verification
    └── analysis/
        ├── charts.py           # Time-series chart generation
        └── frequency.py        # Frequency analysis
```

## Module Descriptions

### common/
Shared utility scripts used across multiple modules. These are diagnostic and inspection tools.

### geocoding/
**Independent service** for converting GPS coordinates to administrative divisions (province/city/county/town/village). This is NOT part of the tracks module - it's a cross-module service that can be used by any module needing geocoding.

**Key Features:**
- Offline geocoding using shapefiles
- GeoHash-based caching for performance
- Batch processing support
- Quality verification tools

### tracks/
GPS trajectory data processing pipeline.

**import/**: Data ingestion from Excel files
**analysis/**: Statistical analysis and behavior detection
**migrations/**: Database schema evolution

### keyboard/
Keyboard and mouse usage tracking and analysis.

**import/**: Parse and import KMCounter.ini data
**analysis/**: Generate charts and frequency statistics

## Usage Guidelines

1. **Import scripts**: Run these to load raw data into SQLite databases
2. **Analysis scripts**: Run these to generate statistics and insights
3. **Common scripts**: Use these for debugging and data quality checks
4. **Geocoding scripts**: Run these to add administrative division data to GPS points

## Next Steps (TODO)

- [ ] Merge stay_detection.py and stay_detection_v2.py
- [ ] Merge statistics.py and statistics_v2.py
- [ ] Move photos.py to appropriate location
- [ ] Update import paths in all scripts
- [ ] Add __init__.py files for Python package structure
- [ ] Create geocode_worker.py for Docker integration
- [ ] Add configuration files (config.yaml)
