#!/usr/bin/env python3
"""
GPX Route Parser for Apple Health Workout Routes
解析运动路线GPX文件并导入数据库
"""

import gpxpy
import gpxpy.gpx
import sqlite3
from pathlib import Path
from datetime import datetime
import sys

class GPXRouteParser:
    """GPX路线解析器"""

    def __init__(self, db_path):
        self.db_path = db_path
        self.conn = None
        self.cursor = None
        self.stats = {
            'files_processed': 0,
            'routes_imported': 0,
            'points_imported': 0,
            'errors': 0
        }

    def connect_db(self):
        """连接数据库"""
        self.conn = sqlite3.connect(self.db_path)
        self.cursor = self.conn.cursor()

    def close_db(self):
        """关闭数据库连接"""
        if self.conn:
            self.conn.commit()
            self.conn.close()

    def parse_gpx_file(self, gpx_file):
        """解析单个GPX文件"""
        try:
            with open(gpx_file, 'r', encoding='utf-8') as f:
                gpx = gpxpy.parse(f)

            route_points = []

            # 遍历所有轨迹
            for track in gpx.tracks:
                for segment in track.segments:
                    for point in segment.points:
                        route_points.append({
                            'timestamp': point.time,
                            'latitude': point.latitude,
                            'longitude': point.longitude,
                            'altitude': point.elevation,
                            'speed': point.speed if hasattr(point, 'speed') else None
                        })

            return route_points

        except Exception as e:
            print(f"解析GPX文件失败 {gpx_file.name}: {e}")
            self.stats['errors'] += 1
            return None

    def find_matching_workout(self, gpx_file, route_points):
        """根据GPX文件名和时间查找匹配的运动记录"""
        if not route_points:
            return None

        # 从文件名提取日期: route_2025-06-19_1.24pm.gpx
        filename = gpx_file.stem  # route_2025-06-19_1.24pm
        parts = filename.split('_')

        if len(parts) < 2:
            return None

        date_str = parts[1]  # 2025-06-19

        # 获取路线的开始时间
        start_time = route_points[0]['timestamp']
        end_time = route_points[-1]['timestamp']

        if not start_time or not end_time:
            return None

        # 查找匹配的运动记录（时间范围内）
        query = """
        SELECT id FROM workouts
        WHERE DATE(start_date) = ?
        AND start_date <= ?
        AND end_date >= ?
        ORDER BY ABS(JULIANDAY(start_date) - JULIANDAY(?))
        LIMIT 1
        """

        self.cursor.execute(query, (date_str, end_time, start_time, start_time))
        result = self.cursor.fetchone()

        return result[0] if result else None

    def import_route_points(self, workout_id, route_points):
        """导入路线点到数据库"""
        if not route_points:
            return

        query = """
        INSERT INTO workout_routes
        (workout_id, timestamp, latitude, longitude, altitude, speed)
        VALUES (?, ?, ?, ?, ?, ?)
        """

        data = [
            (workout_id, p['timestamp'], p['latitude'], p['longitude'],
             p['altitude'], p['speed'])
            for p in route_points
        ]

        self.cursor.executemany(query, data)
        self.stats['points_imported'] += len(route_points)

    def process_gpx_directory(self, gpx_dir):
        """处理GPX目录中的所有文件"""
        gpx_files = list(gpx_dir.glob('*.gpx'))

        if not gpx_files:
            print(f"未找到GPX文件: {gpx_dir}")
            return

        print(f"找到 {len(gpx_files)} 个GPX文件")

        for gpx_file in gpx_files:
            print(f"\n处理: {gpx_file.name}")

            # 解析GPX文件
            route_points = self.parse_gpx_file(gpx_file)

            if not route_points:
                print(f"  跳过: 无有效路线点")
                continue

            print(f"  路线点数: {len(route_points)}")

            # 查找匹配的运动记录
            workout_id = self.find_matching_workout(gpx_file, route_points)

            if workout_id:
                print(f"  匹配到运动记录 ID: {workout_id}")

                # 导入路线点
                self.import_route_points(workout_id, route_points)

                # 更新运动记录的route_file字段
                self.cursor.execute(
                    "UPDATE workouts SET route_file = ? WHERE id = ?",
                    (str(gpx_file.name), workout_id)
                )

                self.conn.commit()
                self.stats['routes_imported'] += 1
                print(f"  ✓ 导入成功")
            else:
                print(f"  ⚠ 未找到匹配的运动记录")

            self.stats['files_processed'] += 1

    def print_summary(self):
        """打印导入摘要"""
        print("\n" + "=" * 60)
        print("GPX路线导入摘要")
        print("=" * 60)
        print(f"处理文件数: {self.stats['files_processed']}")
        print(f"导入路线数: {self.stats['routes_imported']}")
        print(f"导入路线点: {self.stats['points_imported']:,}")
        print(f"错误数量: {self.stats['errors']}")
        print("=" * 60)

def main():
    """主函数"""
    # 路径配置
    gpx_dir = Path('go-backend/data/applehealth/apple_health_export/workout-routes')
    db_path = Path('go-backend/data/applehealth/health.db')

    # 检查路径
    if not gpx_dir.exists():
        print(f"错误: GPX目录不存在: {gpx_dir}")
        sys.exit(1)

    if not db_path.exists():
        print(f"错误: 数据库不存在: {db_path}")
        sys.exit(1)

    # 创建解析器
    parser = GPXRouteParser(str(db_path))
    parser.connect_db()

    try:
        # 处理GPX文件
        parser.process_gpx_directory(gpx_dir)

        # 打印摘要
        parser.print_summary()

    except Exception as e:
        print(f"\n错误: {e}")
        import traceback
        traceback.print_exc()

    finally:
        parser.close_db()

if __name__ == '__main__':
    main()
