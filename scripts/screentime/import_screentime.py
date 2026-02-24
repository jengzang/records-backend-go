#!/usr/bin/env python3
"""
Screentime Data Import Script
Imports phone screen time data from Vivo X90 into SQLite database

Data Sources:
- AppUsage_day_*.txt: Daily app usage summary
- AppUsage_line_*.txt: Detailed session data
- AppUsage_unlock_*.txt: Unlock records
- AppUsage_app_*.txt: Per-app time series (not imported, redundant with daily)

Author: Claude Code
Date: 2026-02-24
"""

import sqlite3
import pandas as pd
import json
import re
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Tuple

# Configuration
DATA_DIR = Path(__file__).parent.parent.parent / "data" / "screentime" / "phone"
DB_PATH = Path(__file__).parent.parent.parent / "data" / "screentime" / "screentime.db"
BATCH_SIZE = 1000

def init_database():
    """Initialize database with schema"""
    print("Initializing database...")

    # Create database directory if not exists
    DB_PATH.parent.mkdir(parents=True, exist_ok=True)

    # Run migration SQL
    migration_file = Path(__file__).parent / "migrations" / "001_create_screentime_tables.sql"

    if not migration_file.exists():
        print(f"Error: Migration file not found: {migration_file}")
        return False

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    with open(migration_file, 'r', encoding='utf-8') as f:
        sql_script = f.read()

    # Execute the entire script at once
    try:
        cursor.executescript(sql_script)
        conn.commit()
        print("Database initialized successfully")
    except sqlite3.Error as e:
        print(f"Error initializing database: {e}")
        conn.close()
        return False

    conn.close()
    return True

def parse_duration_text(duration_text: str) -> int:
    """
    Convert Chinese duration text to milliseconds
    Examples:
    - "1分钟" -> 60000
    - "1小时6分钟" -> 3960000
    - "少于一分钟" -> 30000 (estimate)
    """
    if not duration_text or duration_text == '':
        return 0

    if "少于一分钟" in duration_text:
        return 30000  # Estimate 30 seconds

    total_ms = 0

    # Extract hours
    hour_match = re.search(r'(\d+)小时', duration_text)
    if hour_match:
        total_ms += int(hour_match.group(1)) * 3600000

    # Extract minutes
    min_match = re.search(r'(\d+)分钟', duration_text)
    if min_match:
        total_ms += int(min_match.group(1)) * 60000

    # Extract seconds
    sec_match = re.search(r'(\d+)秒', duration_text)
    if sec_match:
        total_ms += int(sec_match.group(1)) * 1000

    return total_ms

def parse_date(date_str: str) -> str:
    """Convert date string to YYYYMMDD format"""
    if not date_str or date_str.strip() == '':
        return ''

    try:
        # Handle YYYY-MM-DD format
        if '-' in date_str:
            dt = datetime.strptime(date_str.strip(), '%Y-%m-%d')
            return dt.strftime('%Y%m%d')
        # Already in YYYYMMDD format
        return date_str.strip()
    except:
        return ''

def import_daily_usage():
    """Import daily app usage data from AppUsage_day_*.txt"""
    print("\n=== Importing Daily Usage Data ===")

    file_pattern = "AppUsage_day_*.txt"
    files = list(DATA_DIR.glob(file_pattern))

    if not files:
        print(f"No files found matching {file_pattern}")
        return 0

    print(f"Found {len(files)} file(s)")

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    total_records = 0

    for file_path in files:
        print(f"Processing {file_path.name}...")

        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        current_date = None
        records = []

        for line in lines:
            line = line.strip()
            if not line:
                continue

            parts = line.split(',')

            # Check if this is a date header line
            if len(parts) >= 7 and parts[0] and '-' in parts[0]:
                current_date = parse_date(parts[0])
                continue

            # Data line
            if current_date and len(parts) >= 7:
                app_name = parts[1]
                package_id = parts[2]
                formatted_time = parts[3]
                duration_ms = int(parts[4]) if parts[4].isdigit() else parse_duration_text(formatted_time)
                launch_count = int(parts[5]) if parts[5].isdigit() else 0
                notification_count = int(parts[6]) if parts[6].isdigit() else 0

                # Extract split screen duration for "ALL" entry
                split_screen_ms = 0
                if package_id == 'ALL' and len(parts) > 7:
                    split_match = re.search(r'(\d+)', parts[7])
                    if split_match:
                        split_screen_ms = int(split_match.group(1))

                records.append((
                    current_date,
                    app_name,
                    package_id,
                    duration_ms,
                    launch_count,
                    notification_count,
                    split_screen_ms
                ))

                # Batch insert
                if len(records) >= BATCH_SIZE:
                    cursor.executemany("""
                        INSERT OR REPLACE INTO screentime_daily
                        (date, app_name, package_id, duration_ms, launch_count, notification_count, split_screen_ms)
                        VALUES (?, ?, ?, ?, ?, ?, ?)
                    """, records)
                    conn.commit()
                    total_records += len(records)
                    print(f"  Inserted {total_records} records...")
                    records = []

        # Insert remaining records
        if records:
            cursor.executemany("""
                INSERT OR REPLACE INTO screentime_daily
                (date, app_name, package_id, duration_ms, launch_count, notification_count, split_screen_ms)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            """, records)
            conn.commit()
            total_records += len(records)

    conn.close()
    print(f"[OK] Imported {total_records} daily usage records")
    return total_records

def import_sessions():
    """Import session data from AppUsage_line_*.txt"""
    print("\n=== Importing Session Data ===")

    file_pattern = "AppUsage_line_*.txt"
    files = list(DATA_DIR.glob(file_pattern))

    if not files:
        print(f"No files found matching {file_pattern}")
        return 0

    print(f"Found {len(files)} file(s)")

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    total_records = 0

    for file_path in files:
        print(f"Processing {file_path.name}...")

        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        records = []

        for line in lines[1:]:  # Skip header
            line = line.strip()
            if not line:
                continue

            parts = line.split(',')
            if len(parts) < 8:
                continue

            date = parse_date(parts[0])
            if not date:
                continue

            # Parse timestamps
            start_time_str = parts[1]  # T:1771430658600
            start_time_display = parts[2]  # 00:04
            app_name = parts[3]
            package_id = parts[4]
            duration_text = parts[5]
            end_time_str = parts[6]  # T:1771430659746
            end_time_display = parts[7]  # 00:04

            # Extract millisecond timestamps
            start_time_ms = 0
            end_time_ms = 0

            start_match = re.search(r'T:(\d+)', start_time_str)
            if start_match:
                start_time_ms = int(start_match.group(1))

            end_match = re.search(r'T:(\d+)', end_time_str)
            if end_match:
                end_time_ms = int(end_match.group(1))

            # Check for streaming flag
            is_streaming = 1 if len(parts) > 8 and '流媒体' in parts[8] else 0

            records.append((
                date,
                start_time_ms,
                end_time_ms,
                start_time_display,
                end_time_display,
                app_name,
                package_id,
                duration_text,
                is_streaming
            ))

            # Batch insert
            if len(records) >= BATCH_SIZE:
                cursor.executemany("""
                    INSERT INTO screentime_sessions
                    (date, start_time_ms, end_time_ms, start_time, end_time, app_name, package_id, duration_text, is_streaming)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, records)
                conn.commit()
                total_records += len(records)
                print(f"  Inserted {total_records} records...")
                records = []

        # Insert remaining records
        if records:
            cursor.executemany("""
                INSERT INTO screentime_sessions
                (date, start_time_ms, end_time_ms, start_time, end_time, app_name, package_id, duration_text, is_streaming)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, records)
            conn.commit()
            total_records += len(records)

    conn.close()
    print(f"[OK] Imported {total_records} session records")
    return total_records

def import_unlocks():
    """Import unlock data from AppUsage_unlock_*.txt"""
    print("\n=== Importing Unlock Data ===")

    file_pattern = "AppUsage_unlock_*.txt"
    files = list(DATA_DIR.glob(file_pattern))

    if not files:
        print(f"No files found matching {file_pattern}")
        return 0

    print(f"Found {len(files)} file(s)")

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    total_records = 0

    for file_path in files:
        print(f"Processing {file_path.name}...")

        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        records = []
        current_unlock = None
        app_sequence = []

        for line in lines[1:]:  # Skip header
            line = line.strip()
            if not line:
                continue

            parts = line.split(',')

            # Check if this is an unlock line (has date and time)
            if len(parts) >= 3 and parts[0] and parts[1]:
                # Save previous unlock if exists
                if current_unlock:
                    records.append((
                        current_unlock['date'],
                        current_unlock['time'],
                        current_unlock['duration'],
                        json.dumps(app_sequence, ensure_ascii=False)
                    ))
                    app_sequence = []

                # Start new unlock
                date = parse_date(parts[0])
                unlock_time = parts[1]
                duration = parts[2]

                current_unlock = {
                    'date': date,
                    'time': unlock_time,
                    'duration': duration
                }

            # App line (no date/time, just app info)
            elif len(parts) >= 4 and parts[3]:
                app_name = parts[3]
                package_id = parts[4] if len(parts) > 4 else ''
                if app_name:
                    app_sequence.append({
                        'app_name': app_name,
                        'package_id': package_id
                    })

            # Batch insert
            if len(records) >= BATCH_SIZE:
                cursor.executemany("""
                    INSERT INTO screentime_unlocks
                    (date, unlock_time, session_duration, app_sequence)
                    VALUES (?, ?, ?, ?)
                """, records)
                conn.commit()
                total_records += len(records)
                print(f"  Inserted {total_records} records...")
                records = []

        # Save last unlock
        if current_unlock:
            records.append((
                current_unlock['date'],
                current_unlock['time'],
                current_unlock['duration'],
                json.dumps(app_sequence, ensure_ascii=False)
            ))

        # Insert remaining records
        if records:
            cursor.executemany("""
                INSERT INTO screentime_unlocks
                (date, unlock_time, session_duration, app_sequence)
                VALUES (?, ?, ?, ?)
            """, records)
            conn.commit()
            total_records += len(records)

    conn.close()
    print(f"[OK] Imported {total_records} unlock records")
    return total_records

def update_app_metadata():
    """Update app metadata table with aggregated statistics"""
    print("\n=== Updating App Metadata ===")

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Aggregate from daily usage
    cursor.execute("""
        INSERT OR REPLACE INTO screentime_apps
        (package_id, app_name, first_seen, last_seen, total_duration_ms, total_launches, total_notifications)
        SELECT
            package_id,
            MAX(app_name) as app_name,
            MIN(date) as first_seen,
            MAX(date) as last_seen,
            SUM(duration_ms) as total_duration_ms,
            SUM(launch_count) as total_launches,
            SUM(notification_count) as total_notifications
        FROM screentime_daily
        WHERE package_id != 'ALL'
        GROUP BY package_id
    """)

    conn.commit()

    # Get count
    cursor.execute("SELECT COUNT(*) FROM screentime_apps")
    count = cursor.fetchone()[0]

    conn.close()
    print(f"[OK] Updated metadata for {count} apps")
    return count

def categorize_apps():
    """Auto-categorize apps based on package ID and name"""
    print("\n=== Categorizing Apps ===")

    # Category mapping rules
    categories = {
        'Social': ['wechat', 'qq', 'weibo', 'xiaohongshu', 'xingin', 'tencent.mm', 'tencent.mobileqq'],
        'Entertainment': ['douyin', 'bilibili', 'tiktok', 'ugc.aweme', 'danmaku.bili', 'iqiyi', 'youku'],
        'Gaming': ['game', 'sgame', 'tmgp', 'mihoyo', 'tencent.tmgp'],
        'Shopping': ['taobao', 'jd', 'pinduoduo', 'tmall'],
        'Tools': ['map', 'browser', 'camera', 'gallery', 'file', 'autonavi', 'baidu.BaiduMap'],
        'Productivity': ['office', 'wps', 'note', 'calendar', 'email'],
        'News': ['news', 'zhihu', 'toutiao', 'hupu'],
        'System': ['launcher', 'settings', 'systemui', 'bbk.launcher']
    }

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Get all apps
    cursor.execute("SELECT id, package_id, app_name FROM screentime_apps")
    apps = cursor.fetchall()

    updated = 0

    for app_id, package_id, app_name in apps:
        category = 'Other'

        # Check package ID and app name against category rules
        search_text = f"{package_id} {app_name}".lower()

        for cat, keywords in categories.items():
            if any(keyword.lower() in search_text for keyword in keywords):
                category = cat
                break

        cursor.execute("""
            UPDATE screentime_apps
            SET category = ?, updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (category, app_id))
        updated += 1

    conn.commit()
    conn.close()

    print(f"[OK] Categorized {updated} apps")
    return updated

def print_summary():
    """Print import summary statistics"""
    print("\n" + "="*60)
    print("IMPORT SUMMARY")
    print("="*60)

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Daily records
    cursor.execute("SELECT COUNT(*), MIN(date), MAX(date) FROM screentime_daily")
    daily_count, min_date, max_date = cursor.fetchone()
    print(f"Daily Usage Records: {daily_count:,}")
    print(f"  Date Range: {min_date} to {max_date}")

    # Sessions
    cursor.execute("SELECT COUNT(*) FROM screentime_sessions")
    session_count = cursor.fetchone()[0]
    print(f"Session Records: {session_count:,}")

    # Unlocks
    cursor.execute("SELECT COUNT(*) FROM screentime_unlocks")
    unlock_count = cursor.fetchone()[0]
    print(f"Unlock Records: {unlock_count:,}")

    # Apps
    cursor.execute("SELECT COUNT(*), COUNT(DISTINCT category) FROM screentime_apps")
    app_count, category_count = cursor.fetchone()
    print(f"Unique Apps: {app_count}")
    print(f"Categories: {category_count}")

    # Top 5 apps by duration
    print("\nTop 5 Apps by Total Duration:")
    cursor.execute("""
        SELECT app_name, total_duration_ms, category
        FROM screentime_apps
        ORDER BY total_duration_ms DESC
        LIMIT 5
    """)
    for i, (app_name, duration_ms, category) in enumerate(cursor.fetchall(), 1):
        hours = duration_ms / 3600000
        print(f"  {i}. {app_name} ({category}): {hours:.1f} hours")

    # Database size
    db_size_mb = DB_PATH.stat().st_size / (1024 * 1024)
    print(f"\nDatabase Size: {db_size_mb:.2f} MB")

    conn.close()
    print("="*60)

def main():
    """Main import workflow"""
    print("Screentime Data Import")
    print("="*60)
    print(f"Data Directory: {DATA_DIR}")
    print(f"Database Path: {DB_PATH}")
    print("="*60)

    # Check if data directory exists
    if not DATA_DIR.exists():
        print(f"Error: Data directory not found: {DATA_DIR}")
        return

    # Initialize database
    if not init_database():
        print("Failed to initialize database")
        return

    # Import data
    try:
        import_daily_usage()
        import_sessions()
        import_unlocks()
        update_app_metadata()
        categorize_apps()
        print_summary()

        print("\n[OK] Import completed successfully!")

    except Exception as e:
        print(f"\nX Import failed: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()
