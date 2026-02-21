#!/usr/bin/env python3
"""
Analyze speed distribution to determine appropriate threshold.
"""

import sqlite3

DB_PATH = 'C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks/tracks.db'

def main():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Check speed distribution of EXCESSIVE_SPEED points
    print("=" * 80)
    print("EXCESSIVE_SPEED 点的速度分布 (更细粒度)")
    print("=" * 80)
    cursor.execute('''
        SELECT
            CASE
                WHEN speed < 120 THEN '< 432 km/h (不应标记)'
                WHEN speed < 150 THEN '432-540 km/h (商业飞行低速)'
                WHEN speed < 200 THEN '540-720 km/h (商业飞行巡航)'
                WHEN speed < 250 THEN '720-900 km/h (商业飞行高速)'
                WHEN speed < 277.78 THEN '900-1000 km/h (接近音速)'
                WHEN speed < 350 THEN '1000-1260 km/h (超音速)'
                ELSE '> 1260 km/h (极端异常)'
            END as speed_range,
            COUNT(*) as count,
            ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM "一生足迹" WHERE outlier_reason_codes LIKE '%EXCESSIVE_SPEED%'), 2) as pct
        FROM "一生足迹"
        WHERE outlier_reason_codes LIKE '%EXCESSIVE_SPEED%'
        GROUP BY speed_range
        ORDER BY MIN(speed)
    ''')

    print(f"{'速度范围':<35} | {'数量':>10} | {'占比':>8}")
    print("-" * 80)
    total = 0
    for row in cursor.fetchall():
        print(f"{row[0]:<35} | {row[1]:>10,} | {row[2]:>7}%")
        total += row[1]
    print("-" * 80)
    print(f"{'总计':<35} | {total:>10,} | {'100.00':>7}%")

    # Compare with PLANE segments
    print("\n" + "=" * 80)
    print("PLANE 段的速度分布")
    print("=" * 80)
    cursor.execute('''
        SELECT
            MIN(avg_speed_kmh) as min_speed,
            MAX(avg_speed_kmh) as max_speed,
            AVG(avg_speed_kmh) as avg_speed,
            MIN(max_speed_kmh) as min_max_speed,
            MAX(max_speed_kmh) as max_max_speed,
            AVG(max_speed_kmh) as avg_max_speed
        FROM segments
        WHERE mode = 'PLANE'
    ''')

    row = cursor.fetchone()
    print(f"平均速度范围: {row[0]:.2f} - {row[1]:.2f} km/h (平均: {row[2]:.2f} km/h)")
    print(f"最大速度范围: {row[3]:.2f} - {row[4]:.2f} km/h (平均: {row[5]:.2f} km/h)")

    # Recommendation
    print("\n" + "=" * 80)
    print("阈值建议")
    print("=" * 80)
    print(f"当前阈值: 432 km/h (120 m/s)")
    print(f"PLANE段最大速度: {row[4]:.2f} km/h ({row[4]/3.6:.2f} m/s)")
    print(f"建议阈值: 1000 km/h (277.78 m/s) - 覆盖所有商业飞行")
    print(f"或者: 1260 km/h (350 m/s) - 仅标记超音速异常")

    # Calculate impact of different thresholds
    print("\n" + "=" * 80)
    print("不同阈值的影响")
    print("=" * 80)
    thresholds = [
        (120, 432),
        (150, 540),
        (200, 720),
        (250, 900),
        (277.78, 1000),
        (350, 1260)
    ]

    for ms, kmh in thresholds:
        cursor.execute('''
            SELECT COUNT(*) FROM "一生足迹" WHERE speed > ?
        ''', (ms,))
        count = cursor.fetchone()[0]
        pct = count * 100.0 / 408184
        print(f"阈值 {kmh:>6.0f} km/h ({ms:>6.2f} m/s): {count:>7,} 点 ({pct:>5.2f}%)")

    conn.close()

if __name__ == '__main__':
    main()
