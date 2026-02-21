#!/usr/bin/env python3
"""
Analyze current outlier detection results to understand what's being flagged.
"""

import sqlite3
import json

DB_PATH = 'C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks/tracks.db'

def main():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Query 1: Outlier Distribution by Speed
    print("=" * 80)
    print("速度分布与异常率")
    print("=" * 80)
    cursor.execute('''
        SELECT
            CASE
                WHEN speed IS NULL THEN 'NULL'
                WHEN speed = 0 THEN '0 (静止)'
                WHEN speed < 1.39 THEN '0-5 km/h (步行)'
                WHEN speed < 8.33 THEN '5-30 km/h (慢速)'
                WHEN speed < 33.33 THEN '30-120 km/h (驾车)'
                WHEN speed < 83.33 THEN '120-300 km/h (高铁)'
                WHEN speed < 120 THEN '300-432 km/h (飞机)'
                ELSE '>432 km/h (异常)'
            END as speed_range,
            COUNT(*) as total,
            SUM(CASE WHEN outlier_flag = 1 THEN 1 ELSE 0 END) as outliers,
            ROUND(SUM(CASE WHEN outlier_flag = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as outlier_pct
        FROM "一生足迹"
        GROUP BY speed_range
        ORDER BY MIN(speed)
    ''')

    print(f"{'速度范围':<25} | {'总数':>10} | {'异常':>10} | {'异常率':>8}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:<25} | {row[1]:>10,} | {row[2]:>10,} | {row[3]:>7}%")

    # Query 2: Outlier Distribution by Accuracy
    print("\n" + "=" * 80)
    print("精度分布与异常率")
    print("=" * 80)
    cursor.execute('''
        SELECT
            CASE
                WHEN accuracy IS NULL THEN 'NULL'
                WHEN accuracy < 10 THEN '<10m (优秀)'
                WHEN accuracy < 50 THEN '10-50m (良好)'
                WHEN accuracy < 100 THEN '50-100m (一般)'
                WHEN accuracy < 500 THEN '100-500m (较差)'
                WHEN accuracy < 1000 THEN '500-1000m (很差)'
                ELSE '>1000m (极差)'
            END as accuracy_range,
            COUNT(*) as total,
            SUM(CASE WHEN outlier_flag = 1 THEN 1 ELSE 0 END) as outliers,
            ROUND(SUM(CASE WHEN outlier_flag = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as outlier_pct
        FROM "一生足迹"
        GROUP BY accuracy_range
        ORDER BY MIN(accuracy)
    ''')

    print(f"{'精度范围':<25} | {'总数':>10} | {'异常':>10} | {'异常率':>8}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:<25} | {row[1]:>10,} | {row[2]:>10,} | {row[3]:>7}%")

    # Query 3: Sample Outlier Points
    print("\n" + "=" * 80)
    print("异常点样本 (随机20个)")
    print("=" * 80)
    cursor.execute('''
        SELECT
            id,
            dataTime,
            speed,
            accuracy,
            latitude,
            longitude,
            province,
            city
        FROM "一生足迹"
        WHERE outlier_flag = 1
        ORDER BY RANDOM()
        LIMIT 20
    ''')

    print(f"{'ID':>8} | {'速度(m/s)':>12} | {'精度(m)':>12} | 位置")
    print("-" * 80)
    for row in cursor.fetchall():
        speed = f"{row[2]:.2f}" if row[2] else "NULL"
        accuracy = f"{row[3]:.1f}" if row[3] else "NULL"
        location = f"{row[6]}-{row[7]}" if row[6] and row[7] else "未知"
        print(f"{row[0]:>8} | {speed:>12} | {accuracy:>12} | {location}")

    # Query 4: Check if Reason Codes are Populated
    print("\n" + "=" * 80)
    print("异常原因分布")
    print("=" * 80)
    cursor.execute('''
        SELECT
            outlier_reason_codes,
            COUNT(*) as count
        FROM "一生足迹"
        WHERE outlier_flag = 1
        GROUP BY outlier_reason_codes
        ORDER BY count DESC
        LIMIT 10
    ''')

    print(f"{'原因代码':<40} | {'数量':>10}")
    print("-" * 80)
    for row in cursor.fetchall():
        reason = row[0] if row[0] else "NULL (未填充)"
        print(f"{reason:<40} | {row[1]:>10,}")

    # Query 5: Overall Statistics
    print("\n" + "=" * 80)
    print("总体统计")
    print("=" * 80)
    cursor.execute('''
        SELECT
            COUNT(*) as total,
            SUM(CASE WHEN outlier_flag = 1 THEN 1 ELSE 0 END) as outliers,
            ROUND(SUM(CASE WHEN outlier_flag = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as pct
        FROM "一生足迹"
    ''')
    row = cursor.fetchone()
    print(f"总点数: {row[0]:,}")
    print(f"异常点数: {row[1]:,}")
    print(f"异常率: {row[2]}%")

    conn.close()

if __name__ == '__main__':
    main()
