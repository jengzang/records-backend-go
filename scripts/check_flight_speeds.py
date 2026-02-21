#!/usr/bin/env python3
"""
Check if EXCESSIVE_SPEED points are classified as FLIGHT mode.
"""

import sqlite3

DB_PATH = 'C:/Users/joengzaang/CodeProject/records/go-backend/data/tracks/tracks.db'

def main():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # Check transport mode distribution of EXCESSIVE_SPEED outliers
    print("=" * 80)
    print("EXCESSIVE_SPEED 异常点的交通模式分布")
    print("=" * 80)
    cursor.execute('''
        SELECT
            COALESCE(mode, 'NULL') as transport_mode,
            COUNT(*) as count,
            ROUND(AVG(speed), 2) as avg_speed_ms,
            ROUND(AVG(speed) * 3.6, 2) as avg_speed_kmh,
            ROUND(AVG(accuracy), 2) as avg_accuracy
        FROM "一生足迹"
        WHERE outlier_reason_codes LIKE '%EXCESSIVE_SPEED%'
        GROUP BY mode
        ORDER BY count DESC
    ''')

    print(f"{'交通模式':<15} | {'数量':>10} | {'平均速度(m/s)':>15} | {'平均速度(km/h)':>15} | {'平均精度(m)':>12}")
    print("-" * 80)
    for row in cursor.fetchall():
        print(f"{row[0]:<15} | {row[1]:>10,} | {row[2]:>15} | {row[3]:>15} | {row[4]:>12}")

    # Check if these points are in flight segments
    print("\n" + "=" * 80)
    print("EXCESSIVE_SPEED 异常点是否在 FLIGHT 段内")
    print("=" * 80)
    cursor.execute('''
        SELECT
            COUNT(DISTINCT t.id) as total_excessive_speed,
            COUNT(DISTINCT CASE WHEN s.mode = 'FLIGHT' THEN t.id END) as in_flight_segment,
            ROUND(COUNT(DISTINCT CASE WHEN s.mode = 'FLIGHT' THEN t.id END) * 100.0 / COUNT(DISTINCT t.id), 2) as pct_in_flight
        FROM "一生足迹" t
        LEFT JOIN segments s ON t.dataTime BETWEEN s.start_time AND s.end_time
        WHERE t.outlier_reason_codes LIKE '%EXCESSIVE_SPEED%'
    ''')

    row = cursor.fetchone()
    print(f"总EXCESSIVE_SPEED点数: {row[0]:,}")
    print(f"在FLIGHT段内的点数: {row[1]:,}")
    print(f"比例: {row[2]}%")

    # Sample some FLIGHT mode excessive speed points
    print("\n" + "=" * 80)
    print("FLIGHT 模式的 EXCESSIVE_SPEED 样本 (随机10个)")
    print("=" * 80)
    cursor.execute('''
        SELECT
            t.id,
            t.speed,
            t.accuracy,
            t.province,
            t.city,
            s.mode,
            s.id
        FROM "一生足迹" t
        LEFT JOIN segments s ON t.dataTime BETWEEN s.start_time AND s.end_time
        WHERE t.outlier_reason_codes LIKE '%EXCESSIVE_SPEED%'
          AND s.mode = 'FLIGHT'
        ORDER BY RANDOM()
        LIMIT 10
    ''')

    print(f"{'ID':>8} | {'速度(m/s)':>12} | {'速度(km/h)':>12} | {'精度(m)':>10} | {'段ID':>8} | 位置")
    print("-" * 80)
    for row in cursor.fetchall():
        speed_ms = row[1]
        speed_kmh = speed_ms * 3.6
        accuracy = f"{row[2]:.1f}" if row[2] else "NULL"
        segment_id = row[6] if row[6] else "NULL"
        location = f"{row[3]}-{row[4]}" if row[3] and row[4] else "未知"
        print(f"{row[0]:>8} | {speed_ms:>12.2f} | {speed_kmh:>12.2f} | {accuracy:>10} | {segment_id:>8} | {location}")

    conn.close()

if __name__ == '__main__':
    main()
