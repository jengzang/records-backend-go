#!/usr/bin/env python3
"""
Test script to verify geocoding completion and data quality.
"""
import sqlite3
from pathlib import Path

db_path = Path(__file__).parent.parent.parent / "data" / "tracks" / "tracks.db"

conn = sqlite3.connect(str(db_path))
cursor = conn.cursor()

print("=" * 80)
print("GEOCODING VERIFICATION REPORT")
print("=" * 80)

# Overall statistics
cursor.execute('''
    SELECT
        COUNT(*) as total,
        COUNT(province) as geocoded,
        COUNT(*) - COUNT(province) as not_geocoded
    FROM "一生足迹"
''')

stats = cursor.fetchone()
print(f"\n1. OVERALL STATISTICS")
print(f"   Total points: {stats[0]:,}")
print(f"   Geocoded: {stats[1]:,}")
print(f"   Not geocoded: {stats[2]:,}")
print(f"   Geocoded percentage: {stats[1] * 100 / stats[0]:.2f}%")

# Admin level coverage
print(f"\n2. ADMINISTRATIVE LEVEL COVERAGE")
cursor.execute('SELECT COUNT(DISTINCT province) FROM "一生足迹" WHERE province IS NOT NULL')
print(f"   Unique provinces: {cursor.fetchone()[0]}")

cursor.execute('SELECT COUNT(DISTINCT city) FROM "一生足迹" WHERE city IS NOT NULL')
print(f"   Unique cities: {cursor.fetchone()[0]}")

cursor.execute('SELECT COUNT(DISTINCT county) FROM "一生足迹" WHERE county IS NOT NULL')
print(f"   Unique counties: {cursor.fetchone()[0]}")

cursor.execute('SELECT COUNT(DISTINCT town) FROM "一生足迹" WHERE town IS NOT NULL')
print(f"   Unique towns: {cursor.fetchone()[0]}")

# Top provinces
print(f"\n3. TOP 10 PROVINCES BY POINT COUNT")
cursor.execute('''
    SELECT province, COUNT(*) as count
    FROM "一生足迹"
    WHERE province IS NOT NULL
    GROUP BY province
    ORDER BY count DESC
    LIMIT 10
''')
for row in cursor.fetchall():
    print(f"   {row[0]}: {row[1]:,} points")

# Top cities
print(f"\n4. TOP 10 CITIES BY POINT COUNT")
cursor.execute('''
    SELECT city, COUNT(*) as count
    FROM "一生足迹"
    WHERE city IS NOT NULL
    GROUP BY city
    ORDER BY count DESC
    LIMIT 10
''')
for row in cursor.fetchall():
    print(f"   {row[0]}: {row[1]:,} points")

# Data quality check
print(f"\n5. DATA QUALITY CHECKS")
cursor.execute('''
    SELECT COUNT(*)
    FROM "一生足迹"
    WHERE province IS NOT NULL AND city IS NULL
''')
print(f"   Points with province but no city: {cursor.fetchone()[0]}")

cursor.execute('''
    SELECT COUNT(*)
    FROM "一生足迹"
    WHERE city IS NOT NULL AND county IS NULL
''')
print(f"   Points with city but no county: {cursor.fetchone()[0]}")

cursor.execute('''
    SELECT COUNT(*)
    FROM "一生足迹"
    WHERE county IS NOT NULL AND town IS NULL
''')
print(f"   Points with county but no town: {cursor.fetchone()[0]}")

# Sample geocoded points
print(f"\n6. SAMPLE GEOCODED POINTS (5 random)")
cursor.execute('''
    SELECT id, longitude, latitude, province, city, county, town
    FROM "一生足迹"
    WHERE province IS NOT NULL
    ORDER BY RANDOM()
    LIMIT 5
''')
print(f"   {'ID':<8} {'Lon':<10} {'Lat':<10} {'Province':<10} {'City':<10} {'County':<10} {'Town':<20}")
print("   " + "-" * 90)
for row in cursor.fetchall():
    print(f"   {row[0]:<8} {row[1]:<10.4f} {row[2]:<10.4f} {row[3]:<10} {row[4]:<10} {row[5]:<10} {row[6]:<20}")

print("\n" + "=" * 80)
print("VERIFICATION COMPLETE")
print("=" * 80)

conn.close()
