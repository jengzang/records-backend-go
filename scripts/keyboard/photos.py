import pandas as pd
import matplotlib.pyplot as plt
import os
import matplotlib.dates as mdates
import re

# 获取当前脚本所在目录
current_directory = os.path.dirname(os.path.abspath(__file__))
file_path = os.path.join(current_directory, 'keys_data.xlsx')

# 读取Excel文件
df = pd.read_excel(file_path, header=[0, 1], index_col=0)

# 获取Total行并删除Total行
total_row = df.loc['Total']
df = df.drop('Total')

# 列名称和排序顺序
sorted_columns = [
    '左键', '右键', '中键', '侧键', '滚轮', '横滚轮', '移动距离', '键盘敲击总数', 'esc', 'tab', '退格', '回车',
    '数字键盘回车', 'rctrl', 'lCtrl',
    'rshift', 'ralt', 'lshift', 'lalt', 'lwin', 'rwin', 'CapsLock', '空格', 'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o',
    'p', 'a', 's', 'd',
    'f', 'g', 'h', 'j', 'k', 'l', 'z', 'x', 'c', 'v', 'b', 'n', 'm', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
    'numpad0', 'numpad1',
    'numpad2', 'numpad3', 'numpad4', 'numpad5', 'numpad6', 'numpad7', 'numpad8', 'numpad9', 'numpad/', 'numpad*',
    'numpad-', 'numpad+',
    'numpad.', 'F1', 'F2', 'F3', 'F4', 'F5', 'F6', 'F7', 'F8', 'F9', 'F10', 'F11', 'F12', '[', ']', '-', '=', ';', '\'',
    '~', '\\', ',', '.', '/'
]

# 对应的列名称映射
column_mapping = {key: name for key, name in zip(df.columns.get_level_values(0), sorted_columns)}

# 解决中文显示问题
plt.rcParams['font.sans-serif'] = ['SimHei']  # 使用黑体
plt.rcParams['axes.unicode_minus'] = False  # 解决负号显示问题

# 新建保存图片的文件夹
pictures_directory = os.path.join(current_directory, 'pictures')
os.makedirs(pictures_directory, exist_ok=True)

# 将日期转换为datetime格式，并提取月份
df.index = pd.to_datetime(df.index, format='%Y%m%d')
df.index.name = '时间'

# 生成单独的图表并保存
for key in df.columns.get_level_values(0):
    plt.figure(figsize=(20, 8))  # 更长的图表
    plt.plot(df.index, df[key], label=column_mapping.get(key, key))
    plt.xlabel('时间')
    plt.ylabel('当天按键次数')
    plt.title(f'{column_mapping.get(key, key)} 次数随时间变化图')

    # 设置横坐标格式和显示频率
    plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
    plt.gca().xaxis.set_major_locator(mdates.MonthLocator())

    plt.legend()
    plt.xticks(rotation=45)
    plt.tight_layout()

    # 清理文件名中的特殊字符
    clean_name = re.sub(r'[\\/*?:"<>|]', "_", column_mapping.get(key, key))
    output_chart_path = os.path.join(pictures_directory, f'{clean_name}.png')
    plt.savefig(output_chart_path)
    plt.close()

# 生成数字键盘累加图表
numpad_keys = [col for col in df.columns.get_level_values(0) if 'numpad' in col or col == '数字键盘回车']
numpad_sum = df[numpad_keys].sum(axis=1)

plt.figure(figsize=(20, 8))
plt.plot(df.index, numpad_sum, label='数字键盘总计')
plt.xlabel('时间')
plt.ylabel('次数')
plt.title('数字键盘总计随时间变化图')
plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
plt.gca().xaxis.set_major_locator(mdates.MonthLocator())
plt.legend()
plt.xticks(rotation=45)
plt.tight_layout()

output_chart_path = os.path.join(pictures_directory, '数字键盘总计.png')
plt.savefig(output_chart_path)
plt.close()

# 生成鼠标按键汇总图表
mouse_keys = ['左键', '右键', '中键', '侧键']
mouse_keys = [col for col in df.columns.get_level_values(0) if col in mouse_keys]
mouse_sum = df[mouse_keys].sum(axis=1)

plt.figure(figsize=(20, 8))
plt.plot(df.index, mouse_sum, label='鼠标按键总计')
plt.xlabel('时间')
plt.ylabel('次数')
plt.title('鼠标按键总计随时间变化图')
plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
plt.gca().xaxis.set_major_locator(mdates.MonthLocator())
plt.legend()
plt.xticks(rotation=45)
plt.tight_layout()

output_chart_path = os.path.join(pictures_directory, '鼠标按键总计.png')
plt.savefig(output_chart_path)
plt.close()

# 生成滚轮和横滚轮汇总图表
wheel_keys = ['滚轮', '横滚轮']
wheel_keys = [col for col in df.columns.get_level_values(0) if col in wheel_keys]
wheel_sum = df[wheel_keys].sum(axis=1)

plt.figure(figsize=(20, 8))
plt.plot(df.index, wheel_sum, label='滚轮和横滚轮总计')
plt.xlabel('时间')
plt.ylabel('次数')
plt.title('滚轮和横滚轮总计随时间变化图')
plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
plt.gca().xaxis.set_major_locator(mdates.MonthLocator())
plt.legend()
plt.xticks(rotation=45)
plt.tight_layout()

output_chart_path = os.path.join(pictures_directory, '滚轮和横滚轮总计.png')
plt.savefig(output_chart_path)
plt.close()

print(f"所有图表已保存到 {pictures_directory}")



####此代码汇总图表数字键盘累加图表、鼠标按键汇总图表、滚轮和横滚轮汇总图表，仍有问题