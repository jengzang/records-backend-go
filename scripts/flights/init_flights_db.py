#!/usr/bin/env python3
"""
Initialize flights database with schema

Usage:
    python init_flights_db.py
"""

import sqlite3
from pathlib import Path


def init_database():
    """Initialize flights database with schema"""

    # Database path
    db_path = Path(__file__).parent.parent.parent / 'data' / 'flights' / 'flights.db'
    db_path.parent.mkdir(parents=True, exist_ok=True)

    print(f"Initializing flights database: {db_path}")

    # Read schema
    schema_path = Path(__file__).parent / 'migrations' / '001_create_flight_tables.sql'
    with open(schema_path, 'r', encoding='utf-8') as f:
        schema = f.read()

    # Create database and execute schema
    conn = sqlite3.connect(str(db_path))
    try:
        conn.executescript(schema)
        conn.commit()
        print("[OK] Database schema created successfully")

        # Verify tables
        cursor = conn.cursor()
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
        tables = cursor.fetchall()

        print(f"\nCreated tables:")
        for table in tables:
            print(f"  - {table[0]}")

        print(f"\n[OK] Flights database initialized at: {db_path}")

    except Exception as e:
        print(f"[ERROR] Error initializing database: {e}")
        import traceback
        traceback.print_exc()
    finally:
        conn.close()


if __name__ == '__main__':
    init_database()
