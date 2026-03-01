#!/usr/bin/env python3
"""
Initialize railway database
"""

import sqlite3
import os

def init_railway_db():
    # Database path
    db_dir = "../../data/railway"
    db_path = os.path.join(db_dir, "railway.db")

    # Create directory if not exists
    os.makedirs(db_dir, exist_ok=True)

    # Check if database already exists
    if os.path.exists(db_path):
        print(f"[INFO] Database already exists: {db_path}")
        return

    print(f"[INFO] Creating railway database: {db_path}")

    # Connect to database
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Read and execute migration SQL
    migration_file = "migrations/001_create_railway_tables.sql"
    with open(migration_file, 'r', encoding='utf-8') as f:
        sql = f.read()

    # Execute SQL
    cursor.executescript(sql)
    conn.commit()

    print("[OK] Railway database initialized successfully")
    print(f"[INFO] Database location: {os.path.abspath(db_path)}")

    # Verify tables
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
    tables = cursor.fetchall()
    print(f"[INFO] Created tables: {[t[0] for t in tables]}")

    conn.close()

if __name__ == "__main__":
    init_railway_db()
