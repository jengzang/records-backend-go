#!/usr/bin/env python3
"""
Check geocoded results in the database.
"""
import sqlite3
from pathlib import Path

db_path = Path(__file__).parent.parent.parent / "data" / "tracks" / "tracks.db"

conn = sqlite3.connect(str(db_path))
cursor = conn.cursor()

# Get 10 random geocoded points
cursor.execute('''
    SELECT id, longitude, latitude, province, city, county, town, time_visually
    FROM "一生足迹"
    WHERE province IS NOT NULL
    LIMIT 10
''')

print("Sample geocoded points:\n")
print(f"{'ID':<8} {'Longitude':<12} {'Latitude':<12} {'Province':<10} {'City':<10} {'County':<10} {'Town':<20} {'Time':<20}")
print("-" * 120)

for row in cursor.fetchall():
    print(f"{row[0]:<8} {row[1]:<12.6f} {row[2]:<12.6f} {row[3]:<10} {row[4]:<10} {row[5]:<10} {row[6]:<20} {row[7]:<20}")

# Get statistics
cursor.execute('''
    SELECT
        COUNT(*) as total,
        COUNT(province) as geocoded,
        COUNT(*) - COUNT(province) as not_geocoded
    FROM "一生足迹"
''')

stats = cursor.fetchone()
print(f"\nStatistics:")
print(f"  Total points: {stats[0]}")
print(f"  Geocoded: {stats[1]}")
print(f"  Not geocoded: {stats[2]}")
print(f"  Geocoded percentage: {stats[1] * 100 / stats[0]:.2f}%")

conn.close()
