import sqlite3
from datetime import datetime

print("=" * 60)
print("Screentime Sessions Sample")
print("=" * 60)
conn = sqlite3.connect('data/screentime/screentime.db')
cursor = conn.cursor()
cursor.execute('''
    SELECT date, start_time_ms, end_time_ms, start_time, end_time, app_name, package_id
    FROM screentime_sessions
    ORDER BY date DESC, start_time_ms DESC
    LIMIT 5
''')
for row in cursor.fetchall():
    date, start_ms, end_ms, start_time, end_time, app_name, package_id = row
    # Convert milliseconds to datetime
    start_dt = datetime.fromtimestamp(start_ms / 1000) if start_ms else None
    end_dt = datetime.fromtimestamp(end_ms / 1000) if end_ms else None
    print(f"Date: {date}, Time: {start_time}-{end_time}, App: {app_name}")
    if start_dt:
        print(f"  Start: {start_dt}, Hour: {start_dt.hour}")
conn.close()

print("\n" + "=" * 60)
print("Health Records Sample (Heart Rate)")
print("=" * 60)
conn = sqlite3.connect('data/applehealth/health.db')
cursor = conn.cursor()
cursor.execute('''
    SELECT type, value, unit, start_date, end_date
    FROM health_records
    WHERE type = 'HKQuantityTypeIdentifierHeartRate'
    ORDER BY start_date DESC
    LIMIT 5
''')
for row in cursor.fetchall():
    type_name, value, unit, start_date, end_date = row
    print(f"Type: {type_name}, Value: {value} {unit}")
    print(f"  Start: {start_date}")
conn.close()

print("\n" + "=" * 60)
print("Health Records Sample (Steps)")
print("=" * 60)
conn = sqlite3.connect('data/applehealth/health.db')
cursor = conn.cursor()
cursor.execute('''
    SELECT type, value, unit, start_date, end_date
    FROM health_records
    WHERE type = 'HKQuantityTypeIdentifierStepCount'
    ORDER BY start_date DESC
    LIMIT 5
''')
for row in cursor.fetchall():
    type_name, value, unit, start_date, end_date = row
    print(f"Type: {type_name}, Value: {value} {unit}")
    print(f"  Start: {start_date}")
conn.close()

print("\n" + "=" * 60)
print("Screentime Apps with Categories")
print("=" * 60)
conn = sqlite3.connect('data/screentime/screentime.db')
cursor = conn.cursor()
cursor.execute('''
    SELECT package_id, app_name, category
    FROM screentime_apps
    WHERE category IS NOT NULL
    LIMIT 10
''')
for row in cursor.fetchall():
    package_id, app_name, category = row
    print(f"{app_name} ({package_id}): {category}")
conn.close()
