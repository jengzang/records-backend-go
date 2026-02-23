#!/usr/bin/env python3
"""
Verify imported keyboard/mouse data
"""

import sqlite3
import sys

def verify_data(db_path):
    """Verify imported data"""
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    print("="*60)
    print("DATABASE VERIFICATION")
    print("="*60)

    # Check daily_stats
    cursor.execute("SELECT COUNT(*) FROM daily_stats")
    daily_count = cursor.fetchone()[0]
    print(f"\nDaily stats records: {daily_count}")

    cursor.execute("SELECT SUM(keystrokes), SUM(left_clicks + right_clicks + middle_clicks + extra_clicks), SUM(mouse_distance_m) FROM daily_stats")
    total_keystrokes, total_clicks, total_distance = cursor.fetchone()
    print(f"Total keystrokes: {total_keystrokes:,}")
    print(f"Total clicks: {total_clicks:,}")
    print(f"Total mouse distance: {total_distance:.2f} meters ({total_distance/1000:.2f} km)")

    # Check date range
    cursor.execute("SELECT MIN(date), MAX(date) FROM daily_stats")
    min_date, max_date = cursor.fetchone()
    print(f"Date range: {min_date} to {max_date}")

    # Check scancode_stats
    cursor.execute("SELECT COUNT(*) FROM scancode_stats")
    scancode_count = cursor.fetchone()[0]
    print(f"\nScancode stats records: {scancode_count:,}")

    cursor.execute("SELECT COUNT(DISTINCT scancode) FROM scancode_stats")
    unique_scancodes = cursor.fetchone()[0]
    print(f"Unique scancodes: {unique_scancodes}")

    # Top 10 days by keystrokes
    print("\nTop 10 days by keystrokes:")
    cursor.execute("""
        SELECT date, keystrokes, left_clicks + right_clicks + middle_clicks + extra_clicks as total_clicks
        FROM daily_stats
        ORDER BY keystrokes DESC
        LIMIT 10
    """)
    for i, (date, keystrokes, clicks) in enumerate(cursor.fetchall(), 1):
        print(f"  {i}. {date}: {keystrokes:,} keystrokes, {clicks:,} clicks")

    # Top 10 scancodes
    print("\nTop 10 scancodes by total count:")
    cursor.execute("""
        SELECT s.scancode, m.key_name, SUM(s.count) as total
        FROM scancode_stats s
        LEFT JOIN scancode_mapping m ON s.scancode = m.scancode
        GROUP BY s.scancode
        ORDER BY total DESC
        LIMIT 10
    """)
    for i, (scancode, key_name, total) in enumerate(cursor.fetchall(), 1):
        key_display = key_name if key_name else f"Unknown ({scancode})"
        print(f"  {i}. {key_display}: {total:,}")

    # Average daily usage
    cursor.execute("""
        SELECT
            AVG(keystrokes) as avg_keystrokes,
            AVG(left_clicks + right_clicks + middle_clicks + extra_clicks) as avg_clicks,
            AVG(mouse_distance_m) as avg_distance
        FROM daily_stats
        WHERE keystrokes > 0
    """)
    avg_keystrokes, avg_clicks, avg_distance = cursor.fetchone()
    print(f"\nAverage daily usage (active days):")
    print(f"  Keystrokes: {avg_keystrokes:.0f}")
    print(f"  Clicks: {avg_clicks:.0f}")
    print(f"  Mouse distance: {avg_distance:.2f} meters")

    # Active days (>100 keystrokes)
    cursor.execute("SELECT COUNT(*) FROM daily_stats WHERE keystrokes > 100")
    active_days = cursor.fetchone()[0]
    print(f"\nActive days (>100 keystrokes): {active_days}")

    print("="*60)

    conn.close()

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python verify_data.py <db_path>")
        sys.exit(1)

    verify_data(sys.argv[1])
