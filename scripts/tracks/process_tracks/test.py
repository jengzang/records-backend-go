import pandas as pd
import matplotlib
import matplotlib.pyplot as plt
from matplotlib import font_manager
from matplotlib.ticker import FuncFormatter
import numpy as np

# 设置matplotlib使用中文字体
font = font_manager.FontProperties(fname='C:/Windows/Fonts/simsun.ttc')  # Windows系统的微软雅黑路径
plt.rcParams['font.family'] = font.get_name()

# 设置后端为TkAgg，避免错误
matplotlib.use('TkAgg')

# 读取 CSV 文件并确保 'time' 列为字符串类型
file_path = r"C:\Users\joengzaang\Desktop\照片汇总.csv"
data = pd.read_csv(file_path, dtype={'time': str})

# 只选择需要的列
columns_of_interest = ['time', '省级', '市级', '区县级', '乡镇级']
data = data[columns_of_interest]

# 确保 'time' 列为字符串类型
data['time'] = data['time'].astype(str)

# 将 'time' 列转化为日期格式，只保留年月日部分，忽略时间（小时、分钟、秒）
data['date'] = pd.to_datetime(data['time'].str[:8], format='%Y%m%d', errors='coerce')

# 设置日期范围从2016年到2025年
start_date = '2016-01-01'
end_date = '2025-12-31'

# 生成每周的周一日期
dates = pd.date_range(start=start_date, end=end_date, freq='W-MON')

# 生成每周的周一到周日日期
weekly_dates = []

for monday in dates:
    week_dates = [monday]
    for i in range(1, 7):
        week_dates.append((monday + pd.Timedelta(days=i)).strftime('%Y-%m-%d'))
    weekly_dates.append(week_dates)

# 统计每周的照片数量
weekly_photo_counts = []

for week in weekly_dates:
    week_start = pd.to_datetime(week[0])
    week_end = pd.to_datetime(week[6])

    # 过滤出在该周内的照片，忽略时间部分，只比较日期
    photos_in_week = data[(data['date'] >= week_start) & (data['date'] <= week_end)]

    # 统计该周的照片数量
    weekly_photo_counts.append({
        'Week Start': week_start.strftime('%Y-%m-%d'),
        'Week End': week_end.strftime('%Y-%m-%d'),
        'Photo Count': len(photos_in_week)
    })

# 将结果转换为 DataFrame
weekly_counts_df = pd.DataFrame(weekly_photo_counts)

# 绘制折线图
plt.figure(figsize=(12, 6))  # 设置图表的大小
plt.plot(weekly_counts_df['Week Start'], weekly_counts_df['Photo Count'], marker='o', linestyle='-', color='b')

# 设置图表标题和标签
plt.title('每周拍照数量', fontsize=16, fontproperties=font)
plt.xlabel('每周', fontsize=12, fontproperties=font)
plt.ylabel('照片数量', fontsize=12, fontproperties=font)

# 横坐标只显示总的开始和结束日期
plt.xticks([weekly_counts_df['Week Start'].iloc[0], weekly_counts_df['Week End'].iloc[-1]],
           [weekly_counts_df['Week Start'].iloc[0], weekly_counts_df['Week End'].iloc[-1]])


# 自定义纵坐标：0到200范围每20一个格，200以上范围每50一个格
def custom_yticks(y, pos):
    if y <= 200:
        return f'{int(y)}'
    else:
        return f'{int(y)}'


formatter = FuncFormatter(custom_yticks)
plt.gca().yaxis.set_major_formatter(formatter)

# 自定义纵坐标刻度，0到200范围每20，200以上范围每50
y_ticks = []
for i in range(0, 201, 20):  # 0-200范围，每隔20
    y_ticks.append(i)
for i in range(200, max(weekly_counts_df['Photo Count']), 50):  # 200以上范围，每隔50
    y_ticks.append(i)

plt.yticks(y_ticks)

# 设置显示网格线
plt.grid(True)

# 显示图表
plt.tight_layout()
plt.show()
