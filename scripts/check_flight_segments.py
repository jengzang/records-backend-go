#!/usr/bin/env python3
"""
Check if there are any FLIGHT segments in the database.
"""

import sqlite3

DB_PATH = 'C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks/tracks.db'

def main():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Check mode distribution in segments
    print("=" * 80)
    print("segments 表中的交通模式分布")
    print("=" * 80)
    cursor.execute('''
        SELECT
            mode,
            COUNT(*) as count,
            ROUND(AVG(avg_speed_kmh), 2) as avg_speed,
            ROUND(AVG(max_speed_kmh), 2) as max_speed_avg
        FROM segments
        GROUP BY mode
        ORDER BY count DESC
    ''')

    print(f"{'模式':<15} | {'段数':>10} | {'平均速度(km/h)':>15} | {'最大速度平均(km/h)':>20}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:<15} | {row[1]:>10,} | {row[2]:>15} | {row[3]:>20}")

    # Check if there are any high-speed segments
    print("\n" + "=" * 80)
    print("高速段统计 (avg_speed_kmh > 100)")
    print("=" * 80)
    cursor.execute('''
        SELECT
            mode,
            COUNT(*) as count,
            MIN(avg_speed_kmh) as min_speed,
            MAX(avg_speed_kmh) as max_speed,
            AVG(avg_speed_kmh) as avg_speed
        FROM segments
        WHERE avg_speed_kmh > 100
        GROUP BY mode
        ORDER BY count DESC
    ''')

    print(f"{'模式':<15} | {'段数':>10} | {'最小速度':>12} | {'最大速度':>12} | {'平均速度':>12}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:<15} | {row[1]:>10,} | {row[2]:>12.2f} | {row[3]:>12.2f} | {row[4]:>12.2f}")

    # Sample some high-speed segments
    print("\n" + "=" * 80)
    print("高速段样本 (avg_speed_kmh > 200, 前10个)")
    print("=" * 80)
    cursor.execute('''
        SELECT
            id,
            mode,
            avg_speed_kmh,
            max_speed_kmh,
            point_count,
            duration_s
        FROM segments
        WHERE avg_speed_kmh > 200
        ORDER BY avg_speed_kmh DESC
        LIMIT 10
    ''')

    print(f"{'段ID':>8} | {'模式':>10} | {'平均速度':>12} | {'最大速度':>12} | {'点数':>8} | {'时长(s)':>10}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:>8} | {row[1]:>10} | {row[2]:>12.2f} | {row[3]:>12.2f} | {row[4]:>8} | {row[5]:>10}")

    conn.close()

if __name__ == '__main__':
    main()
