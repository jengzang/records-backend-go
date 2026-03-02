import sqlite3
import sys

db_path = 'data/tracks.db'
migration_file = 'scripts/tracks/migrations/023_create_location_behavior_tables.sql'

try:
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    with open(migration_file, 'r', encoding='utf-8') as f:
        sql = f.read()

    cursor.executescript(sql)
    conn.commit()
    print('Migration executed successfully')

    # Verify tables created
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'location%'")
    tables = cursor.fetchall()
    print(f'Created tables: {[t[0] for t in tables]}')

    conn.close()
except Exception as e:
    print(f'Error: {e}', file=sys.stderr)
    sys.exit(1)
