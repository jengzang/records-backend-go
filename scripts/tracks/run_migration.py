#!/usr/bin/env python3
"""
Database migration script to add administrative division columns.
Run this script to update the tracks database schema.
"""

import sqlite3
import sys
from pathlib import Path

def run_migration(db_path: str, migration_file: str):
    """Run a SQL migration file against the database."""
    try:
        # Read migration SQL
        with open(migration_file, 'r', encoding='utf-8') as f:
            migration_sql = f.read()

        # Connect to database
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()

        # Enable WAL mode
        cursor.execute("PRAGMA journal_mode=WAL")

        # Execute migration
        print(f"Running migration: {migration_file}")
        cursor.executescript(migration_sql)

        conn.commit()
        print("Migration completed successfully!")

        # Verify columns were added
        cursor.execute("PRAGMA table_info(\"一生足迹\")")
        columns = cursor.fetchall()
        print(f"\nTable now has {len(columns)} columns:")
        for col in columns:
            print(f"  - {col[1]} ({col[2]})")

        conn.close()
        return True

    except Exception as e:
        print(f"Error running migration: {e}", file=sys.stderr)
        return False

if __name__ == "__main__":
    # Paths
    script_dir = Path(__file__).parent
    db_path = script_dir.parent.parent / "data" / "tracks" / "tracks.db"

    # Check if migration file is provided as argument
    if len(sys.argv) > 1:
        migration_file = script_dir / "migrations" / sys.argv[1]
    else:
        migration_file = script_dir / "migrations" / "002_add_metadata_and_indexes.sql"

    print(f"Database: {db_path}")
    print(f"Migration: {migration_file}")

    if not db_path.exists():
        print(f"Error: Database not found at {db_path}", file=sys.stderr)
        sys.exit(1)

    if not migration_file.exists():
        print(f"Error: Migration file not found at {migration_file}", file=sys.stderr)
        sys.exit(1)

    success = run_migration(str(db_path), str(migration_file))
    sys.exit(0 if success else 1)
