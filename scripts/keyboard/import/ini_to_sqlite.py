import sqlite3
from io import StringIO
from configparser import ConfigParser
import os

# 获取当前脚本所在目录
current_directory = os.path.dirname(os.path.abspath(__file__))
ini_file_path = os.path.join(current_directory, 'KMCounter.ini')
db_file_path = os.path.join(current_directory, 'kmcounter.db')

def create_database():
    """创建 SQLite 数据库和表格"""
    conn = sqlite3.connect(db_file_path)
    cursor = conn.cursor()

    # 创建 mouse_data 表格
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS mouse_data (
            date TEXT PRIMARY KEY,
            lbcount INTEGER,
            rbcount INTEGER,
            mbcount INTEGER,
            xbcount INTEGER,
            wheel INTEGER,
            hwheel INTEGER,
            move REAL
        )
    ''')

    # 创建 keyboard_data 表格
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS keyboard_data (
            date TEXT PRIMARY KEY,
            keystrokes INTEGER
        )
    ''')

    # 创建 scan_codes 表格
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS scan_codes (
            date TEXT,
            scan_code INTEGER,
            count INTEGER,
            PRIMARY KEY (date, scan_code),
            FOREIGN KEY (date) REFERENCES keyboard_data(date)
        )
    ''')

    # 创建索引以优化查询
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_scan_codes_date ON scan_codes(date)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_scan_codes_code ON scan_codes(scan_code)')

    conn.commit()
    return conn

def read_ini_file():
    """读取并解析 INI 文件"""
    print(f"正在读取 INI 文件: {ini_file_path}")

    # 读取二进制数据并解码为 UTF-16
    with open(ini_file_path, 'rb') as file:
        binary_data = file.read()

    decoded_data = binary_data.decode('utf-16')

    # 使用 StringIO 模拟文件以供 ConfigParser 读取
    config_file = StringIO(decoded_data)

    # 加载配置数据
    config = ConfigParser()
    config.read_file(config_file)

    print(f"成功读取 INI 文件，共 {len(config.sections())} 个区段")
    return config

def parse_value(value, value_type='int'):
    """解析并转换值，处理错误"""
    try:
        if value_type == 'int':
            return int(value)
        elif value_type == 'float':
            return float(value)
    except (ValueError, TypeError):
        return 0 if value_type == 'int' else 0.0

def process_date_section(config, section):
    """处理单个日期区段的数据"""
    # 提取滑鼠数据
    mouse_data = {
        'date': section,
        'lbcount': parse_value(config.get(section, 'lbcount', fallback='0')),
        'rbcount': parse_value(config.get(section, 'rbcount', fallback='0')),
        'mbcount': parse_value(config.get(section, 'mbcount', fallback='0')),
        'xbcount': parse_value(config.get(section, 'xbcount', fallback='0')),
        'wheel': parse_value(config.get(section, 'wheel', fallback='0')),
        'hwheel': parse_value(config.get(section, 'hwheel', fallback='0')),
        'move': parse_value(config.get(section, 'move', fallback='0.0'), 'float')
    }

    # 提取键盘数据
    keyboard_data = {
        'date': section,
        'keystrokes': parse_value(config.get(section, 'keystrokes', fallback='0'))
    }

    # 提取扫描码数据
    scan_codes_data = []
    for key, value in config.items(section):
        if key.startswith('sc') and key[2:].isdigit():
            scan_code = int(key[2:])
            count = parse_value(value)
            if count > 0:  # 只存储非零计数
                scan_codes_data.append((section, scan_code, count))

    return mouse_data, keyboard_data, scan_codes_data

def insert_data(conn, mouse_data_list, keyboard_data_list, scan_codes_list):
    """批次插入数据到数据库"""
    cursor = conn.cursor()

    print(f"正在插入 {len(mouse_data_list)} 条滑鼠数据...")
    for mouse_data in mouse_data_list:
        cursor.execute('''
            INSERT OR REPLACE INTO mouse_data
            (date, lbcount, rbcount, mbcount, xbcount, wheel, hwheel, move)
            VALUES (:date, :lbcount, :rbcount, :mbcount, :xbcount, :wheel, :hwheel, :move)
        ''', mouse_data)

    print(f"正在插入 {len(keyboard_data_list)} 条键盘数据...")
    for keyboard_data in keyboard_data_list:
        cursor.execute('''
            INSERT OR REPLACE INTO keyboard_data (date, keystrokes)
            VALUES (:date, :keystrokes)
        ''', keyboard_data)

    print(f"正在插入 {len(scan_codes_list)} 条扫描码数据...")
    cursor.executemany('''
        INSERT OR REPLACE INTO scan_codes (date, scan_code, count)
        VALUES (?, ?, ?)
    ''', scan_codes_list)

    conn.commit()
    print("数据插入完成")

def verify_data(conn):
    """验证数据完整性"""
    cursor = conn.cursor()

    print("\n=== 数据验证 ===")

    # 检查记录数
    cursor.execute("SELECT COUNT(*) FROM mouse_data")
    mouse_count = cursor.fetchone()[0]
    print(f"滑鼠数据记录数: {mouse_count}")

    cursor.execute("SELECT COUNT(*) FROM keyboard_data")
    keyboard_count = cursor.fetchone()[0]
    print(f"键盘数据记录数: {keyboard_count}")

    cursor.execute("SELECT COUNT(*) FROM scan_codes")
    scan_codes_count = cursor.fetchone()[0]
    print(f"扫描码数据记录数: {scan_codes_count}")

    # 检查日期范围
    cursor.execute("SELECT MIN(date), MAX(date) FROM keyboard_data")
    min_date, max_date = cursor.fetchone()
    print(f"日期范围: {min_date} 到 {max_date}")

    # 验证外键关系
    cursor.execute('''
        SELECT COUNT(DISTINCT date) FROM scan_codes
        WHERE date NOT IN (SELECT date FROM keyboard_data)
    ''')
    orphan_count = cursor.fetchone()[0]
    if orphan_count > 0:
        print(f"警告: 发现 {orphan_count} 个没有对应键盘数据的扫描码记录")
    else:
        print("外键关系验证通过")

    print("===================\n")

def main():
    """主函数"""
    print("开始 KMCounter.ini 到 SQLite 的转换")
    print("=" * 50)

    # 创建数据库
    conn = create_database()
    print(f"数据库已创建/打开: {db_file_path}\n")

    # 读取 INI 文件
    config = read_ini_file()

    # 处理所有日期区段
    mouse_data_list = []
    keyboard_data_list = []
    scan_codes_list = []

    date_sections = [section for section in config.sections() if section.isdigit()]
    print(f"\n找到 {len(date_sections)} 个日期区段，开始处理...\n")

    for i, section in enumerate(date_sections, 1):
        if i % 100 == 0:
            print(f"已处理 {i}/{len(date_sections)} 个日期区段...")

        mouse_data, keyboard_data, scan_codes_data = process_date_section(config, section)
        mouse_data_list.append(mouse_data)
        keyboard_data_list.append(keyboard_data)
        scan_codes_list.extend(scan_codes_data)

    print(f"所有日期区段处理完成\n")

    # 插入数据
    insert_data(conn, mouse_data_list, keyboard_data_list, scan_codes_list)

    # 验证数据
    verify_data(conn)

    # 关闭数据库连接
    conn.close()

    print("=" * 50)
    print(f"转换完成! 数据库文件: {db_file_path}")

if __name__ == "__main__":
    main()
