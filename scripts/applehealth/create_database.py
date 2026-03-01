#!/usr/bin/env python3
"""
Create Apple Health database and apply schema
"""

import sqlite3
from pathlib import Path

def create_database():
    # Paths
    db_path = Path('go-backend/data/applehealth/health.db')
    schema_path = Path('go-backend/scripts/applehealth/migrations/001_create_health_tables.sql')

    # Create directory if not exists
    db_path.parent.mkdir(parents=True, exist_ok=True)

    # Read schema
    with open(schema_path, 'r', encoding='utf-8') as f:
        schema_sql = f.read()

    # Create database and apply schema
    print(f"Creating database: {db_path}")
    conn = sqlite3.connect(str(db_path))
    cursor = conn.cursor()

    # Execute schema as a script
    try:
        cursor.executescript(schema_sql)
        conn.commit()
    except sqlite3.Error as e:
        print(f"Error executing schema: {e}")
        conn.close()
        return

    # Verify tables created
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
    tables = cursor.fetchall()

    print(f"\nDatabase created successfully!")
    print(f"Tables created: {len(tables)}")
    for table in tables:
        print(f"   - {table[0]}")

    # Get database size
    import os
    db_size = os.path.getsize(db_path)
    print(f"Database size: {db_size:,} bytes")

    conn.close()
    print(f"\nDatabase ready at: {db_path}")

if __name__ == '__main__':
    create_database()
