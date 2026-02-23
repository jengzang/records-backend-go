#!/usr/bin/env python3
"""
Initialize keyboard database with schema
"""

import sqlite3
import sys

def init_database(db_path, schema_path):
    """Initialize database with schema"""
    print(f"Initializing database: {db_path}")
    print(f"Using schema: {schema_path}")

    # Read schema file
    with open(schema_path, 'r', encoding='utf-8') as f:
        schema_sql = f.read()

    # Connect to database
    conn = sqlite3.connect(db_path)

    # Execute schema
    conn.executescript(schema_sql)
    conn.commit()
    conn.close()

    print("Database initialized successfully!")

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage: python init_db.py <db_path> <schema_path>")
        sys.exit(1)

    init_database(sys.argv[1], sys.argv[2])
