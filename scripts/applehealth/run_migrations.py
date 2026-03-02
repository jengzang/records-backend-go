#!/usr/bin/env python3
"""
Run AppleHealth database migrations
"""
import sqlite3
import os
import sys

def run_migrations():
    # Database path
    db_path = os.path.join('data', 'applehealth', 'health.db')

    # Check if database exists
    if not os.path.exists(db_path):
        print(f"❌ Database not found: {db_path}")
        sys.exit(1)

    # Migrations path
    migrations_path = os.path.join('scripts', 'applehealth', 'migrations')
    migration_file = os.path.join(migrations_path, '002_create_efficiency_curve_tables.sql')

    # Check if migration file exists
    if not os.path.exists(migration_file):
        print(f"❌ Migration file not found: {migration_file}")
        sys.exit(1)

    # Read migration SQL
    with open(migration_file, 'r', encoding='utf-8') as f:
        migration_sql = f.read()

    # Connect to database
    print(f"📂 Connecting to database: {db_path}")
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    try:
        # Execute migration
        print(f"🔄 Running migration: 002_create_efficiency_curve_tables.sql")
        cursor.executescript(migration_sql)
        conn.commit()

        # Verify tables were created
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%efficiency%'")
        tables = cursor.fetchall()

        print(f"\n✅ Migration completed successfully!")
        print(f"📊 Created tables:")
        for table in tables:
            print(f"   - {table[0]}")

    except Exception as e:
        print(f"❌ Migration failed: {e}")
        conn.rollback()
        sys.exit(1)
    finally:
        conn.close()

if __name__ == '__main__':
    os.chdir(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    run_migrations()
