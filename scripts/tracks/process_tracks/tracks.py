import pandas as pd
import matplotlib.pyplot as plt
import matplotlib
from matplotlib import font_manager
pd.set_option('display.max_rows', None)

from matplotlib.ticker import FuncFormatter, MaxNLocator
import numpy as np
matplotlib.use('TkAgg')  # 设置后端为 TkAgg，以避免在显示时出错
# 设置matplotlib使用中文字体
font = font_manager.FontProperties(fname='C:/Windows/Fonts/simsun.ttc')  # Windows系统的微软雅黑路径
plt.rcParams['font.family'] = font.get_name()

# 读取 CSV 文件
# file_path = r"C:\Users\joengzaang\Desktop\Appletotal.csv"
file_path = r"C:\Users\joengzaang\Desktop\处理后的一生足迹数据.csv"
data = pd.read_csv(file_path)

# 只选择需要的列
columns_of_interest = ['time', '省级', '市级', '区县级', '乡镇级']
data = data[columns_of_interest]

# 确保 'time' 列为字符串类型
data['time'] = data['time'].astype(str)

# 统计省、市、县、镇的出现次数，并显示前20个
province_counts = data['省级'].value_counts().head(20)
city_counts = data['市级'].value_counts().head(30)
county_counts = data['区县级'].value_counts().head(20)
town_counts = data['乡镇级'].value_counts().head(20)

# 打印省、市、县、镇的前20个统计结果
print("Top 20 Provinces counts:")
print(province_counts)

print("\nTop 20 Cities counts:")
print(city_counts)

# 获取区县级的前20个并输出它们的上级（对应的市级）
county_with_city = data[['市级', '区县级']].dropna()
county_with_city_counts = county_with_city.groupby(['市级', '区县级']).size().reset_index(name='count')
top_20_counties_with_city = county_with_city_counts.sort_values(by='count', ascending=False).head(30)

print("\nTop 20 Counties and their corresponding Cities:")
print(top_20_counties_with_city)

# 获取乡镇级的前20个并输出它们的上级（对应的区县级）
town_with_county = data[['市级', '区县级', '乡镇级']].dropna()
town_with_county_counts = town_with_county.groupby(['市级', '区县级', '乡镇级']).size().reset_index(name='count')
top_20_towns_with_county = town_with_county_counts.sort_values(by='count', ascending=False).head(50)

print("\nTop 20 Towns and their corresponding Counties:")
print(top_20_towns_with_county)

# print("\nTop 20 Counties counts:")
# print(county_counts)
#
# print("\nTop 20 Towns counts:")
# print(town_counts)



# 确保 'time' 列为字符串类型
data['time'] = data['time'].astype(str)

# 打印调试信息，检查 'time' 列数据
print("调试：读取到的 'time' 列数据（前5行）：")
print(data['time'].head())
# 提取时间中的小时信息（倒数第5和第6位）
data['hour'] = data['time'].str[8:10]  # 提取time中的小时部分（倒数第5和第6位）

# 统计每小时拍摄的照片数量，并按出现次数排序
hourly_counts = data['hour'].value_counts().sort_values(ascending=False)

# 打印每小时的照片数量，按出现次数从高到低排序
print("\n每小时拍摄的照片数量（按频率排序）：")
print(hourly_counts)


# 提取完整年份信息（前4位）
data['year'] = data['time'].str[:4]

# 统计每年的照片数量
yearly_counts = data['year'].value_counts().sort_index()

# 打印每年的照片数量
print("\nYearly Photo Counts:")
print(yearly_counts)

# 提取年月信息（前6位：yyyyMM），保留完整的年月信息
data['year_month'] = data['time'].str[:6]  # 提取年月（格式：yyyyMM）

# 统计每月的照片数量，按数量从高到低排序
monthly_counts = data['year_month'].value_counts().sort_values(ascending=False)
month_counts = data['year_month'].value_counts().sort_index()
# 打印前20个每月的照片数量，按频率从高到低排序
print("\n每月拍摄的照片数量（按频率排序，显示前20个）：")
print(monthly_counts.head(20))
print(month_counts)


# 提取月份信息（前6位：yyyyMM），只保留月份
data['month'] = data['time'].str[4:6]  # 提取月份（格式：MM）

# 统计每月的照片数量，按数量从高到低排序（忽略年份）
monthly_counts = data['month'].value_counts().sort_values(ascending=False)

# 打印每月的照片数量，按频率从高到低排序
print("\n每月拍摄的照片数量（按频率排序，不考虑年份）：")
print(monthly_counts)

import pandas as pd
from datetime import datetime

# 读取 CSV 文件并确保 'time' 列为字符串类型
data = pd.read_csv(file_path, dtype={'time': str})

# 只选择需要的列
columns_of_interest = ['time', '省级', '市级', '区县级', '乡镇级']
data = data[columns_of_interest]

# 确保 'time' 列为字符串类型
data['time'] = data['time'].astype(str)

# 打印前几行，查看 'time' 列的内容
print(data['time'].head(20))  # 调试，确保 'time' 列内容是正确的

# # 使用 `errors='coerce'` 来处理格式不一致的数据，错误的时间会被转换为 NaT
# data['date'] = pd.to_datetime(data['time'], format='%Y%m%d%H%M%S', errors='coerce')
# 将 'time' 列转化为日期格式，只保留年月日部分，忽略时间
data['date'] = pd.to_datetime(data['time'].str[:8], format='%Y%m%d', errors='coerce')

# 检查哪些行是 NaT（即转换失败的行）
invalid_dates = data[data['date'].isna()]
print(f"无效的日期行数：{len(invalid_dates)}")
print(invalid_dates.head())  # 打印出无效的日期数据

# 如果需要，可以删除无效日期的数据
# data = data.dropna(subset=['date'])

# 定义每个农历年开始和结束的日期（大年初一）
lunar_years = {
    '乙巳年 (蛇年)': ('2025-01-29', '2026-02-17'),
    '甲辰年 (龙年)': ('2024-02-10', '2025-01-28'),
    '癸卯年 (兔年)': ('2023-01-22', '2024-02-09'),
    '壬寅年 (虎年)': ('2022-02-01', '2023-01-21'),
    '辛丑年 (牛年)': ('2021-02-12', '2022-01-31'),
    '庚子年 (鼠年)': ('2020-01-25', '2021-02-11'),
    '己亥年 (猪年)': ('2019-02-05', '2020-01-24'),
    '戊戌年 (狗年)': ('2018-02-16', '2019-02-04'),
    '丁酉年 (鸡年)': ('2017-01-28', '2018-02-15'),
    '丙申年 (猴年)': ('2016-02-08', '2017-01-27'),
}

# lunar_years = {
#     '2023冬': ('2023-11-01', '2024-01-31'),
#     '2024春': ('2024-02-01', '2024-04-30'),
#     '2024夏': ('2024-05-01', '2024-07-31'),
#     '2024秋': ('2024-08-01', '2024-10-31'),
#     '2024冬': ('2024-11-01', '2025-01-31'),
# }

# 统计每个农历年内拍照的照片数量
lunar_year_counts = {}

for year, (start_date, end_date) in lunar_years.items():
    # 转换为 datetime 格式
    start_date = pd.to_datetime(start_date)
    end_date = pd.to_datetime(end_date)

    # 过滤出在农历年区间内的照片
    photos_in_year = data[(data['date'] >= start_date) & (data['date'] <= end_date)]

    # 统计数量
    lunar_year_counts[year] = len(photos_in_year)

# 打印每个农历年内的照片数量
print("每个农历年拍照的照片数量：")
for year, count in lunar_year_counts.items():
    print(f"{year}: {count} 张照片")


# 统计每个农历年内到过的市级数量
lunar_year_city_counts = {}
for year, (start_date, end_date) in lunar_years.items():
    # 转换为 datetime 格式
    start_date = pd.to_datetime(start_date)
    end_date = pd.to_datetime(end_date)

    # 过滤出在农历年区间内的照片
    photos_in_year = data[(data['date'] >= start_date) & (data['date'] <= end_date)]

    # 统计每个市级的出现次数
    city_counts = photos_in_year['市级'].value_counts()

    # 保存每个农历年对应的市级数量
    lunar_year_city_counts[year] = city_counts

# 打印每个农历年到过的市级数量
print("每个农历年到过的市级数量：")
for year, city_counts in lunar_year_city_counts.items():
    print(f"\n{year}:")
    print(city_counts)

# 设置日期范围从2016年到2025年
start_date = '2016-08-01'
end_date = '2025-1-26'

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

# 将结果转换为 DataFrame 并打印
weekly_counts_df = pd.DataFrame(weekly_photo_counts)
# 设置 pandas 显示所有行，防止省略
pd.set_option('display.max_rows', None)
# 打印每一周的拍照数量
print("每一周的拍照数量：")
print(weekly_counts_df)

# 可选：如果需要将结果保存为 CSV 文件
# weekly_counts_df.to_csv('weekly_photo_counts.csv', index=False)

# 1. 按月份分组，统计每个月出现的城市数量，并输出每个月的城市名称
data['month'] = data['date'].dt.to_period('M')  # 获取每个时间点的年月

# 统计每个月出现的不同城市数量
monthly_city_count = data.groupby('month')['市级'].nunique().reset_index(name='City Count')

# 输出每个月的城市数量
print("每个月不同城市的数量：")
print(monthly_city_count)

# 输出每个月的具体城市名称
for month, count in monthly_city_count.values:
    print(f"\n{month} 月：")
    # 获取该月所有的城市并去重，同时去除 NaN 和确保城市名称是字符串
    cities_in_month = data[data['month'] == month]['市级'].dropna().unique()
    cities_in_month = [str(city) for city in cities_in_month]  # 确保所有城市名称为字符串
    print(f"城市数量：{count}")
    print(f"出现的城市：{', '.join(cities_in_month)}")

# 2. 按年份统计不同城市数量
data['year'] = data['date'].dt.year  # 获取每个时间点的年份
yearly_city_count = data.groupby('year')['市级'].nunique().reset_index(name='City Count')

# 输出每年的城市数量
print("\n每年不同城市的数量：")
print(yearly_city_count)

# 输出每年的具体城市名称
for year, count in yearly_city_count.values:
    print(f"\n{year} 年：")
    # 获取该年所有的城市并去重，同时去除 NaN 和确保城市名称是字符串
    cities_in_year = data[data['year'] == year]['市级'].dropna().unique()
    cities_in_year = [str(city) for city in cities_in_year]  # 确保所有城市名称为字符串
    print(f"城市数量：{count}")
    print(f"出现的城市：{', '.join(cities_in_year)}")

# 3. 按农历年

# 统计每个农历年内的不同城市数量及城市名称
for lunar_year, (start_date, end_date) in lunar_years.items():
    # 过滤出该农历年内的数据
    year_data = data[(data['date'] >= start_date) & (data['date'] <= end_date)]

    # 统计该农历年内不同城市数量
    city_count = year_data['市级'].nunique()

    # 获取该农历年内的城市名称
    cities_in_year = year_data['市级'].dropna().unique()
    cities_in_year = [str(city) for city in cities_in_year]  # 确保所有城市名称为字符串

    # 输出农历年和相关信息
    print(f"\n{lunar_year} ({start_date} 到 {end_date})")
    print(f"城市数量：{city_count}")
    print(f"出现的城市：{', '.join(cities_in_year)}")

##########################################
    # 按照城市和日期去重，以确保每个城市每天只算一次
    data_unique = data.drop_duplicates(subset=['市级', 'date'])
    # 1. 统计每个城市出现的总天数
    city_day_count = data_unique.groupby('市级')['date'].nunique().reset_index(name='Total Days')

    # 2. 按照总天数从高到低排序
    city_day_count_sorted = city_day_count.sort_values(by='Total Days', ascending=False).reset_index(drop=True)

    # 输出每个城市出现的总天数，按总天数从高到低排序
    print("\n每个城市出现的总天数（按从高到低排序）：")
    print(city_day_count_sorted)

    # 3. 统计每个城市每个月出现的天数
    # 提取年-月信息
    data_unique['year_month'] = data_unique['date'].dt.to_period('M')

    # 按照城市和年-月分组统计
    city_month_day_count = data_unique.groupby(['市级', 'year_month'])['date'].nunique().reset_index(
        name='Days in Month')

    # 输出每个城市每个月出现的天数
    print("\n每个城市每个月出现的天数：")
    for city in city_month_day_count['市级'].unique():
        print(f"\n{city}的每月出现天数：")
        city_data = city_month_day_count[city_month_day_count['市级'] == city]
        for _, row in city_data.iterrows():
            print(f"{row['year_month']} 月: {row['Days in Month']} 天")

###################################
# 按日期统计每一天拍照数量
date_photo_count = data.groupby('date')['市级'].count().reset_index(name='Photo Count')

# 排序并找出拍照数量最多的前50天
top_50_days = date_photo_count.sort_values(by='Photo Count', ascending=False).head(50)

# 输出前50天的日期及对应的拍照数量
print("\n拍照数量最多的前50天及拍照数量：")
print(top_50_days)

# 对于这50天，列出每一天在哪些市级拍过照
for _, row in top_50_days.iterrows():
    date = row['date']
    photo_count = row['Photo Count']

    # 获取该日期所有的市级
    cities_on_day = data[data['date'] == date]['市级'].unique()
    cities_on_day = [str(city) for city in cities_on_day]  # 确保城市名称为字符串

    # 输出该天的拍照信息
    print(f"\n{date.strftime('%Y-%m-%d')} 这天拍照数量：{photo_count}")
    print(f"拍照的市级：{', '.join(cities_on_day)}")