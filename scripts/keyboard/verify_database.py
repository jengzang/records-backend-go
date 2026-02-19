import sqlite3
import os

# 获取数据库文件路径
current_directory = os.path.dirname(os.path.abspath(__file__))
db_file_path = os.path.join(current_directory, 'kmcounter.db')

# 连接数据库
conn = sqlite3.connect(db_file_path)
cursor = conn.cursor()

print("=" * 60)
print("数据库验证报告")
print("=" * 60)

# 1. 检查表格记录数
print("\n1. 表格记录数:")
cursor.execute("SELECT COUNT(*) FROM mouse_data")
print(f"   mouse_data: {cursor.fetchone()[0]} 条记录")

cursor.execute("SELECT COUNT(*) FROM keyboard_data")
print(f"   keyboard_data: {cursor.fetchone()[0]} 条记录")

cursor.execute("SELECT COUNT(*) FROM scan_codes")
print(f"   scan_codes: {cursor.fetchone()[0]} 条记录")

# 2. 检查日期范围
print("\n2. 日期范围:")
cursor.execute("SELECT MIN(date), MAX(date) FROM keyboard_data")
min_date, max_date = cursor.fetchone()
print(f"   最早日期: {min_date}")
print(f"   最晚日期: {max_date}")

# 3. 抽查特定日期 (20221216)
print("\n3. 抽查日期 20221216:")
cursor.execute("SELECT * FROM mouse_data WHERE date = '20221216'")
mouse_row = cursor.fetchone()
if mouse_row:
    print(f"   滑鼠数据: lbcount={mouse_row[1]}, rbcount={mouse_row[2]}, mbcount={mouse_row[3]}, xbcount={mouse_row[4]}")
    print(f"             wheel={mouse_row[5]}, hwheel={mouse_row[6]}, move={mouse_row[7]}")

cursor.execute("SELECT * FROM keyboard_data WHERE date = '20221216'")
keyboard_row = cursor.fetchone()
if keyboard_row:
    print(f"   键盘数据: keystrokes={keyboard_row[1]}")

cursor.execute("SELECT * FROM scan_codes WHERE date = '20221216' AND scan_code = 57")
scan_code_row = cursor.fetchone()
if scan_code_row:
    print(f"   扫描码 57 (空格键): count={scan_code_row[2]}")

# 4. 统计信息
print("\n4. 统计信息:")
cursor.execute("SELECT SUM(keystrokes) FROM keyboard_data")
total_keystrokes = cursor.fetchone()[0]
print(f"   总按键次数: {total_keystrokes:,}")

cursor.execute("SELECT SUM(lbcount) FROM mouse_data")
total_left_clicks = cursor.fetchone()[0]
print(f"   总左键点击次数: {total_left_clicks:,}")

cursor.execute("SELECT COUNT(DISTINCT scan_code) FROM scan_codes")
unique_scan_codes = cursor.fetchone()[0]
print(f"   不同的扫描码数量: {unique_scan_codes}")

# 5. 最活跃的按键
print("\n5. 最常用的 10 个按键 (按总次数):")
cursor.execute("""
    SELECT scan_code, SUM(count) as total_count
    FROM scan_codes
    GROUP BY scan_code
    ORDER BY total_count DESC
    LIMIT 10
""")
for row in cursor.fetchall():
    print(f"   扫描码 {row[0]}: {row[1]:,} 次")

# 6. 最近 5 天的数据
print("\n6. 最近 5 天的按键统计:")
cursor.execute("""
    SELECT date, keystrokes
    FROM keyboard_data
    ORDER BY date DESC
    LIMIT 5
""")
for row in cursor.fetchall():
    print(f"   {row[0]}: {row[1]:,} 次按键")

# 7. 验证外键完整性
print("\n7. 数据完整性检查:")
cursor.execute("""
    SELECT COUNT(DISTINCT date) FROM scan_codes
    WHERE date NOT IN (SELECT date FROM keyboard_data)
""")
orphan_count = cursor.fetchone()[0]
if orphan_count > 0:
    print(f"   ⚠️ 发现 {orphan_count} 个孤立的扫描码记录")
else:
    print(f"   ✓ 所有扫描码记录都有对应的键盘数据")

# 8. 检查是否有缺失值
print("\n8. 缺失值检查:")
cursor.execute("SELECT COUNT(*) FROM mouse_data WHERE lbcount IS NULL")
null_mouse = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(*) FROM keyboard_data WHERE keystrokes IS NULL")
null_keyboard = cursor.fetchone()[0]
if null_mouse == 0 and null_keyboard == 0:
    print("   ✓ 没有发现 NULL 值")
else:
    print(f"   ⚠️ 发现 NULL 值: mouse={null_mouse}, keyboard={null_keyboard}")

print("\n" + "=" * 60)
print("验证完成")
print("=" * 60)

conn.close()
