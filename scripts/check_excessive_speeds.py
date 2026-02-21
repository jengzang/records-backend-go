#!/usr/bin/env python3
"""
Check the actual speed values of points flagged as EXCESSIVE_SPEED.
"""

import sqlite3

DB_PATH = 'C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks/tracks.db'

def main():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Check speed distribution of EXCESSIVE_SPEED outliers
    print("=" * 80)
    print("EXCESSIVE_SPEED 异常点速度分布")
    print("=" * 80)
    cursor.execute('''
        SELECT
            CASE
                WHEN speed < 120 THEN '120-432 km/h (正常高速)'
                WHEN speed < 150 THEN '432-540 km/h (轻微异常)'
                WHEN speed < 200 THEN '540-720 km/h (中度异常)'
                WHEN speed < 300 THEN '720-1080 km/h (严重异常)'
                ELSE '>1080 km/h (极端异常)'
            END as speed_range,
            COUNT(*) as count,
            MIN(speed) as min_speed,
            MAX(speed) as max_speed,
            AVG(speed) as avg_speed
        FROM "一生足迹"
        WHERE outlier_reason_codes LIKE '%EXCESSIVE_SPEED%'
        GROUP BY speed_range
        ORDER BY MIN(speed)
    ''')

    print(f"{'速度范围':<30} | {'数量':>10} | {'最小(m/s)':>12} | {'最大(m/s)':>12} | {'平均(m/s)':>12}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:<30} | {row[1]:>10,} | {row[2]:>12.2f} | {row[3]:>12.2f} | {row[4]:>12.2f}")

    # Sample some extreme speed points
    print("\n" + "=" * 80)
    print("极端速度样本 (>1000 km/h, 随机10个)")
    print("=" * 80)
    cursor.execute('''
        SELECT
            id,
            dataTime,
            speed,
            accuracy,
            province,
            city,
            mode
        FROM "一生足迹"
        WHERE speed > 277.78  -- >1000 km/h
        ORDER BY RANDOM()
        LIMIT 10
    ''')

    print(f"{'ID':>8} | {'速度(m/s)':>12} | {'速度(km/h)':>12} | {'精度(m)':>10} | {'模式':>10} | 位置")
    print("-" * 80)
    for row in cursor.fetchall():
        speed_ms = row[2]
        speed_kmh = speed_ms * 3.6
        accuracy = f"{row[3]:.1f}" if row[3] else "NULL"
        mode = row[6] if row[6] else "NULL"
        location = f"{row[4]}-{row[5]}" if row[4] and row[5] else "未知"
        print(f"{row[0]:>8} | {speed_ms:>12.2f} | {speed_kmh:>12.2f} | {accuracy:>10} | {mode:>10} | {location}")

    conn.close()

if __name__ == '__main__':
    main()
