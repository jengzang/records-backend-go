#!/usr/bin/env python3
"""检查数据库导入进度"""

import sqlite3
from pathlib import Path

db_path = Path('go-backend/data/applehealth/health.db')

if not db_path.exists():
    print("数据库不存在")
    exit(1)

conn = sqlite3.connect(str(db_path))
cursor = conn.cursor()

# 检查各表的记录数
tables = ['health_records', 'workouts', 'sleep_analysis', 'activity_summary']

print("数据库导入进度:")
print("-" * 50)

for table in tables:
    try:
        cursor.execute(f"SELECT COUNT(*) FROM {table}")
        count = cursor.fetchone()[0]
        print(f"{table:20s}: {count:,} 条记录")
    except Exception as e:
        print(f"{table:20s}: 错误 - {e}")

# 检查导入元数据
cursor.execute("SELECT * FROM import_metadata ORDER BY import_date DESC LIMIT 1")
metadata = cursor.fetchone()
if metadata:
    print("\n最近导入:")
    print(f"  状态: {metadata[9]}")
    if metadata[9] == 'success':
        print(f"  记录数: {metadata[2]:,}")
        print(f"  运动数: {metadata[3]:,}")
        print(f"  处理时间: {metadata[7]:.1f} 秒")

conn.close()
