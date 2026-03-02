import sqlite3

# Connect to database
conn = sqlite3.connect('data/applehealth/health.db')
cursor = conn.cursor()

# Read and execute migration
with open('scripts/applehealth/migrations/002_create_efficiency_curve_tables.sql', 'r', encoding='utf-8') as f:
    sql = f.read()
    cursor.executescript(sql)

conn.commit()

# Verify
cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%efficiency%'")
print("Created tables:", [row[0] for row in cursor.fetchall()])

conn.close()
print("✅ Migration completed!")
