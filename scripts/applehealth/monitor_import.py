#!/usr/bin/env python3
"""实时监控Apple Health数据导入进度"""

import sqlite3
import time
from pathlib import Path

db_path = Path('go-backend/data/applehealth/health.db')

print("Apple Health 数据导入监控")
print("=" * 60)

last_count = 0
start_time = time.time()

try:
    while True:
        conn = sqlite3.connect(str(db_path))
        cursor = conn.cursor()

        # 获取当前记录数
        cursor.execute("SELECT COUNT(*) FROM health_records")
        current_count = cursor.fetchone()[0]

        # 计算速度
        elapsed = time.time() - start_time
        speed = (current_count - last_count) / 30 if elapsed > 0 else 0

        # 显示进度
        print(f"\r健康记录: {current_count:,} | 速度: {speed:.0f} 记录/秒 | 已用时: {elapsed:.0f}秒", end='', flush=True)

        last_count = current_count
        conn.close()

        time.sleep(30)

except KeyboardInterrupt:
    print("\n\n监控已停止")
except Exception as e:
    print(f"\n错误: {e}")
