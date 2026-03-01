#!/usr/bin/env python3
"""
Apple Health XML Streaming Parser
解析Apple Health导出的XML文件，使用流式处理避免内存溢出
"""

import xml.etree.ElementTree as ET
import sqlite3
from datetime import datetime
from pathlib import Path
import sys

class HealthDataParser:
    """Apple Health数据流式解析器"""

    def __init__(self, db_path):
        self.db_path = db_path
        self.conn = None
        self.cursor = None

        # 统计计数器
        self.stats = {
            'records': 0,
            'workouts': 0,
            'sleep_records': 0,
            'activity_summaries': 0,
            'errors': 0
        }

    def connect_db(self):
        """连接数据库"""
        self.conn = sqlite3.connect(self.db_path)
        self.cursor = self.conn.cursor()

        # 启用WAL模式
        self.cursor.execute('PRAGMA journal_mode=WAL')
        self.cursor.execute('PRAGMA synchronous=NORMAL')

    def close_db(self):
        """关闭数据库连接"""
        if self.conn:
            self.conn.commit()
            self.conn.close()

    def parse_datetime(self, date_str):
        """解析Apple Health日期时间格式"""
        if not date_str:
            return None
        try:
            # 格式: "2023-12-20 00:46:03 +0800"
            return datetime.strptime(date_str[:19], '%Y-%m-%d %H:%M:%S')
        except:
            return None

    def parse_record(self, elem):
        """解析Record元素"""
        try:
            record_type = elem.get('type')
            value = elem.get('value')

            # 跳过没有value的记录
            if not value:
                return None

            return {
                'type': record_type,
                'value': float(value),
                'unit': elem.get('unit'),
                'start_date': self.parse_datetime(elem.get('startDate')),
                'end_date': self.parse_datetime(elem.get('endDate')),
                'source_name': elem.get('sourceName'),
                'source_version': elem.get('sourceVersion'),
                'device': elem.get('device'),
                'creation_date': self.parse_datetime(elem.get('creationDate'))
            }
        except Exception as e:
            self.stats['errors'] += 1
            return None

    def parse_workout(self, elem):
        """解析Workout元素"""
        try:
            workout_type = elem.get('workoutActivityType')
            duration = elem.get('duration')
            distance = elem.get('totalDistance')
            calories = elem.get('totalEnergyBurned')

            # 查找WorkoutRoute
            route_file = None
            for child in elem:
                if child.tag == 'WorkoutRoute':
                    file_ref = child.find('FileReference')
                    if file_ref is not None:
                        route_file = file_ref.get('path')
                        break

            return {
                'workout_type': workout_type,
                'duration_seconds': float(duration) * 60 if duration else None,  # 转换为秒
                'duration_unit': elem.get('durationUnit'),
                'distance_meters': float(distance) * 1000 if distance else None,  # 转换为米
                'distance_unit': elem.get('totalDistanceUnit'),
                'calories': float(calories) if calories else None,
                'calories_unit': elem.get('totalEnergyBurnedUnit'),
                'start_date': self.parse_datetime(elem.get('startDate')),
                'end_date': self.parse_datetime(elem.get('endDate')),
                'source_name': elem.get('sourceName'),
                'source_version': elem.get('sourceVersion'),
                'device': elem.get('device'),
                'creation_date': self.parse_datetime(elem.get('creationDate')),
                'route_file': route_file
            }
        except Exception as e:
            self.stats['errors'] += 1
            return None

    def parse_activity_summary(self, elem):
        """解析ActivitySummary元素"""
        try:
            return {
                'date': elem.get('dateComponents'),
                'active_energy_burned': float(elem.get('activeEnergyBurned')) if elem.get('activeEnergyBurned') else None,
                'active_energy_burned_unit': elem.get('activeEnergyBurnedUnit'),
                'apple_exercise_time': float(elem.get('appleExerciseTime')) if elem.get('appleExerciseTime') else None,
                'apple_exercise_time_unit': elem.get('appleExerciseTimeUnit'),
                'apple_stand_hours': int(elem.get('appleStandHours')) if elem.get('appleStandHours') else None,
                'apple_stand_hours_goal': int(elem.get('appleStandHoursGoal')) if elem.get('appleStandHoursGoal') else None
            }
        except Exception as e:
            self.stats['errors'] += 1
            return None

    def insert_record_batch(self, records):
        """批量插入健康记录"""
        if not records:
            return

        query = """
        INSERT INTO health_records
        (type, value, unit, start_date, end_date, source_name, source_version, device, creation_date)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        """

        data = [
            (r['type'], r['value'], r['unit'], r['start_date'], r['end_date'],
             r['source_name'], r['source_version'], r['device'], r['creation_date'])
            for r in records
        ]

        self.cursor.executemany(query, data)
        self.stats['records'] += len(records)

    def insert_workout_batch(self, workouts):
        """批量插入运动记录"""
        if not workouts:
            return

        query = """
        INSERT INTO workouts
        (workout_type, duration_seconds, duration_unit, distance_meters, distance_unit,
         calories, calories_unit, start_date, end_date, source_name, source_version,
         device, creation_date, route_file)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """

        data = [
            (w['workout_type'], w['duration_seconds'], w['duration_unit'],
             w['distance_meters'], w['distance_unit'], w['calories'], w['calories_unit'],
             w['start_date'], w['end_date'], w['source_name'], w['source_version'],
             w['device'], w['creation_date'], w['route_file'])
            for w in workouts
        ]

        self.cursor.executemany(query, data)
        self.stats['workouts'] += len(workouts)

    def insert_activity_summary_batch(self, summaries):
        """批量插入活动摘要"""
        if not summaries:
            return

        query = """
        INSERT OR REPLACE INTO activity_summary
        (date, active_energy_burned, active_energy_burned_unit,
         apple_exercise_time, apple_exercise_time_unit,
         apple_stand_hours, apple_stand_hours_goal)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        """

        data = [
            (s['date'], s['active_energy_burned'], s['active_energy_burned_unit'],
             s['apple_exercise_time'], s['apple_exercise_time_unit'],
             s['apple_stand_hours'], s['apple_stand_hours_goal'])
            for s in summaries
        ]

        self.cursor.executemany(query, data)
        self.stats['activity_summaries'] += len(summaries)

    def parse_xml_streaming(self, xml_file, batch_size=1000):
        """流式解析XML文件"""
        print(f"开始解析: {xml_file}")
        print(f"批次大小: {batch_size}")

        records_batch = []
        workouts_batch = []
        summaries_batch = []

        # 使用iterparse进行流式解析
        context = ET.iterparse(xml_file, events=('end',))

        for event, elem in context:
            # 解析Record
            if elem.tag == 'Record':
                record = self.parse_record(elem)
                if record:
                    # 检查是否是睡眠记录
                    if 'Sleep' in record['type']:
                        self.stats['sleep_records'] += 1
                    records_batch.append(record)

                    if len(records_batch) >= batch_size:
                        self.insert_record_batch(records_batch)
                        records_batch = []
                        self.conn.commit()

                        # 显示进度
                        if self.stats['records'] % 10000 == 0:
                            print(f"已处理 {self.stats['records']:,} 条记录...")

            # 解析Workout
            elif elem.tag == 'Workout':
                workout = self.parse_workout(elem)
                if workout:
                    workouts_batch.append(workout)

                    if len(workouts_batch) >= batch_size:
                        self.insert_workout_batch(workouts_batch)
                        workouts_batch = []
                        self.conn.commit()

            # 解析ActivitySummary
            elif elem.tag == 'ActivitySummary':
                summary = self.parse_activity_summary(elem)
                if summary:
                    summaries_batch.append(summary)

                    if len(summaries_batch) >= batch_size:
                        self.insert_activity_summary_batch(summaries_batch)
                        summaries_batch = []
                        self.conn.commit()

            # 清理已处理的元素，释放内存
            elem.clear()

            # 注意: ElementTree不支持getprevious()，使用其他方式管理内存
            # 定期清理根元素的子元素
            if self.stats['records'] % 10000 == 0:
                # 强制垃圾回收
                import gc
                gc.collect()

        # 插入剩余数据
        if records_batch:
            self.insert_record_batch(records_batch)
        if workouts_batch:
            self.insert_workout_batch(workouts_batch)
        if summaries_batch:
            self.insert_activity_summary_batch(summaries_batch)

        self.conn.commit()

        print(f"\n解析完成!")
        print(f"健康记录: {self.stats['records']:,}")
        print(f"运动记录: {self.stats['workouts']:,}")
        print(f"睡眠记录: {self.stats['sleep_records']:,}")
        print(f"活动摘要: {self.stats['activity_summaries']:,}")
        print(f"错误数量: {self.stats['errors']:,}")

def main():
    """主函数"""
    import time

    # 文件路径
    xml_file = Path('go-backend/data/applehealth/apple_health_export/导出.xml')
    db_path = Path('go-backend/data/applehealth/health.db')

    if not xml_file.exists():
        print(f"错误: XML文件不存在: {xml_file}")
        sys.exit(1)

    if not db_path.exists():
        print(f"错误: 数据库不存在: {db_path}")
        print("请先运行 create_database.py 创建数据库")
        sys.exit(1)

    # 获取文件大小
    file_size = xml_file.stat().st_size / (1024 * 1024)  # MB
    print(f"XML文件大小: {file_size:.1f} MB")

    # 创建解析器
    parser = HealthDataParser(str(db_path))
    parser.connect_db()

    # 记录开始时间
    start_time = time.time()

    try:
        # 解析XML
        parser.parse_xml_streaming(str(xml_file), batch_size=1000)

        # 记录导入元数据
        elapsed_time = time.time() - start_time
        parser.cursor.execute("""
            INSERT INTO import_metadata
            (import_date, total_records, total_workouts, file_size_bytes,
             processing_time_seconds, status)
            VALUES (?, ?, ?, ?, ?, ?)
        """, (
            datetime.now(),
            parser.stats['records'],
            parser.stats['workouts'],
            int(file_size * 1024 * 1024),
            elapsed_time,
            'success'
        ))
        parser.conn.commit()

        print(f"\n处理时间: {elapsed_time:.1f} 秒")
        print(f"处理速度: {parser.stats['records'] / elapsed_time:.0f} 记录/秒")

    except Exception as e:
        print(f"\n错误: {e}")
        import traceback
        traceback.print_exc()

        # 记录失败
        parser.cursor.execute("""
            INSERT INTO import_metadata
            (import_date, status, error_message)
            VALUES (?, ?, ?)
        """, (datetime.now(), 'failed', str(e)))
        parser.conn.commit()

    finally:
        parser.close_db()

    print(f"\n数据已导入到: {db_path}")

if __name__ == '__main__':
    main()
