import pandas as pd
from io import StringIO
from configparser import ConfigParser
import os

# 获取当前脚本所在目录
current_directory = os.path.dirname(os.path.abspath(__file__))
file_path = os.path.join(current_directory, 'KMCounter.ini')

# 读取二进制数据并解码
with open(file_path, 'rb') as file:
    binary_data = file.read()

decoded_data = binary_data.decode('utf-16')

# 使用StringIO模拟文件以供ConfigParser读取
config_file = StringIO(decoded_data)

# 加载配置数据
config = ConfigParser()
config.read_file(config_file)

# 从配置文件中提取数据
data = {}
for section in config.sections():
    if section.isdigit():  # 只处理日期部分
        for key, value in config.items(section):
            if key not in data:
                data[key] = []
            try:
                value = int(value)
            except ValueError:
                value = float(value)
            data[key].append((section, value))

# 将数据转换为DataFrame
columns = list(data.keys())
dates = sorted(set(date for date, _ in sum(data.values(), [])))

# 确保所有日期在范围内
date_range = pd.date_range(start=dates[0], end=dates[-1])
df = pd.DataFrame(index=date_range, columns=columns)

for key, values in data.items():
    for date, count in values:
        df.loc[pd.to_datetime(date, format='%Y%m%d'), key] = count

df.fillna(0, inplace=True)  # 填充缺失值为0
df = df.infer_objects(copy=False)  # 转换对象类型

# 将日期格式化为纯数字格式
df.index = df.index.strftime('%Y%m%d')

# 计算每列的累加和并添加到DataFrame
total_row = df.sum().to_frame().T
total_row.index = ['Total']
df = pd.concat([total_row, df])

# 列名称和排序顺序
sorted_columns = [
    '左键', '右键', '中键', '侧键', '滚轮', '横滚轮', '移动距离', '键盘敲击总数', 'esc', 'tab', '退格', '回车', '数字键盘回车', 'rctrl', 'lCtrl',
    'rshift', 'ralt', 'lshift', 'lalt', 'lwin', 'rwin', 'CapsLock', '空格', 'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's', 'd',
    'f', 'g', 'h', 'j', 'k', 'l', 'z', 'x', 'c', 'v', 'b', 'n', 'm', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'numpad0', 'numpad1',
    'numpad2', 'numpad3', 'numpad4', 'numpad5', 'numpad6', 'numpad7', 'numpad8', 'numpad9', 'numpad/', 'numpad*', 'numpad-', 'numpad+',
    'numpad.', 'F1', 'F2', 'F3', 'F4', 'F5', 'F6', 'F7', 'F8', 'F9', 'F10', 'F11', 'F12', '[', ']', '-', '=', ';', '\'', '~', '\\', ',', '.', '/'
]

# 原始列名称
original_columns = df.columns.tolist()

# 创建原始列名到新列名的映射
column_mapping = {original: new for original, new in zip(original_columns, sorted_columns)}

# 排序DataFrame的列并保留所有列
existing_sorted_columns = [col for col in sorted_columns if col in df.columns]
missing_columns = [col for col in df.columns if col not in existing_sorted_columns]
df = df[existing_sorted_columns + missing_columns]

# 添加新的列名称（多重索引）
new_column_names = [column_mapping.get(col, col) for col in df.columns]
df.columns = pd.MultiIndex.from_arrays([df.columns, new_column_names], names=['Original', 'New'])

# 保存 DataFrame 为 Excel 文件
output_path = os.path.join(current_directory, 'keys_data.xlsx')
df.to_excel(output_path, sheet_name='Data', index=True)

print(f"数据已保存到 {output_path}")