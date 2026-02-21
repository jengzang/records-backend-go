#!/usr/bin/env python3
"""
Check segments table schema and content.
"""

import sqlite3

DB_PATH = 'C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks/tracks.db'

def main():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Get segments table schema
    print("=" * 80)
    print("segments 表结构")
    print("=" * 80)
    cursor.execute("PRAGMA table_info(segments)")
    for row in cursor.fetchall():
        print(f"{row[1]:<20} {row[2]:<15} {'NOT NULL' if row[3] else ''} {'PK' if row[5] else ''}")

    # Check if segments table has data
    print("\n" + "=" * 80)
    print("segments 表数据统计")
    print("=" * 80)
    cursor.execute("SELECT COUNT(*) FROM segments")
    count = cursor.fetchone()[0]
    print(f"总段数: {count:,}")

    if count > 0:
        # Sample some segments
        cursor.execute("SELECT * FROM segments LIMIT 5")
        print("\n样本数据 (前5条):")
        for row in cursor.fetchall():
            print(row)

    conn.close()

if __name__ == '__main__':
    main()
