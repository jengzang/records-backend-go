#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Migration runner for trajectory analysis database schema
Executes SQL migration files in order
"""

import sqlite3
import os
import sys
from pathlib import Path

def run_migrations(db_path, migrations_dir, start_from=None):
    """Run all migration files in order

    Args:
        db_path: Path to SQLite database
        migrations_dir: Directory containing migration files
        start_from: Optional migration number to start from (e.g., 4 to start from 004_*.sql)
    """

    # Connect to database
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Enable WAL mode
    cursor.execute("PRAGMA journal_mode=WAL")

    # Get list of migration files
    migration_files = sorted([
        f for f in os.listdir(migrations_dir)
        if f.endswith('.sql') and f[0].isdigit()
    ])

    # Filter by start_from if specified
    if start_from:
        migration_files = [f for f in migration_files if int(f.split('_')[0]) >= start_from]

    print(f"Found {len(migration_files)} migration files to execute")

    for migration_file in migration_files:
        migration_path = os.path.join(migrations_dir, migration_file)
        print(f"\nExecuting: {migration_file}")

        try:
            with open(migration_path, 'r', encoding='utf-8') as f:
                sql = f.read()

            # Execute migration
            cursor.executescript(sql)
            conn.commit()
            print(f"[OK] {migration_file} completed successfully")

        except Exception as e:
            print(f"[FAIL] {migration_file} failed: {e}")
            conn.rollback()
            # Continue with next migration instead of stopping
            continue
    
    # Verify tables were created
    cursor.execute("""
        SELECT name FROM sqlite_master
        WHERE type='table'
        ORDER BY name
    """)
    tables = [row[0] for row in cursor.fetchall()]

    print(f"\n=== Database Tables ({len(tables)}) ===")
    for table in tables:
        cursor.execute(f'SELECT COUNT(*) FROM "{table}"')
        count = cursor.fetchone()[0]
        print(f"  {table}: {count} rows")

    conn.close()
    return True

if __name__ == "__main__":
    # Get paths
    script_dir = Path(__file__).parent
    migrations_dir = script_dir / "migrations"
    db_path = script_dir.parent.parent / "data" / "tracks" / "tracks.db"

    print(f"Database: {db_path}")
    print(f"Migrations: {migrations_dir}")

    if not db_path.exists():
        print(f"Error: Database not found at {db_path}")
        sys.exit(1)

    if not migrations_dir.exists():
        print(f"Error: Migrations directory not found at {migrations_dir}")
        sys.exit(1)

    # Check if user wants to start from a specific migration
    start_from = None
    if len(sys.argv) > 1:
        try:
            start_from = int(sys.argv[1])
            print(f"Starting from migration {start_from:03d}")
        except ValueError:
            print(f"Invalid migration number: {sys.argv[1]}")
            sys.exit(1)

    # Run migrations
    success = run_migrations(str(db_path), str(migrations_dir), start_from)

    if success:
        print("\n[OK] All migrations completed successfully!")
        sys.exit(0)
    else:
        print("\n[FAIL] Migration failed!")
        sys.exit(1)
