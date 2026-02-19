import pandas as pd
from datetime import datetime

# 读取Excel文件
file_path = r"C:\Users\joengzaang\myfiles\统计数据\轨迹数据\apple一生足迹\停留.xlsx"
xls = pd.ExcelFile(file_path)


# 定义处理单个工作表的函数
def process_sheet(df):
    # 转换时间列为datetime格式
    df['start_time'] = pd.to_datetime(df['start_time'], format='%Y/%m/%d %H:%M')
    df['end_time'] = pd.to_datetime(df['end_time'], format='%Y/%m/%d %H:%M')

    # 1. 计算每个地点的总停留时长
    total_stay_duration = df.groupby('location')['stay_duration'].sum().sort_values(ascending=False)

    # 2. 计算每个日期是否有超过四小时的停留
    def find_days_with_stay_above_4_hours(row):
        start_date = row['start_time'].date()
        end_date = row['end_time'].date()
        days = pd.date_range(start=start_date, end=end_date).date

        days_with_stay = []
        for day in days:
            start_of_day = datetime.combine(day, datetime.min.time())
            end_of_day = datetime.combine(day, datetime.max.time())

            overlap_start = max(start_of_day, row['start_time'])
            overlap_end = min(end_of_day, row['end_time'])

            overlap_duration = (overlap_end - overlap_start).total_seconds() / 3600
            if overlap_duration > 4:
                days_with_stay.append(day)

        return days_with_stay

    df['days_with_stay_above_4_hours'] = df.apply(find_days_with_stay_above_4_hours, axis=1)

    # 3. 计算每个地点的停留天数及具体日期
    location_days = df.explode('days_with_stay_above_4_hours').groupby('location')[
        'days_with_stay_above_4_hours'].unique()

    # 4. 按月分析
    df['month'] = df['start_time'].dt.to_period('M')
    monthly_location = df.explode('days_with_stay_above_4_hours').groupby(['month', 'location']).size().reset_index(
        name='stay_count')

    return total_stay_duration, location_days, monthly_location


# 逐个处理所有工作表
results = {}
for sheet_name in xls.sheet_names:
    df = pd.read_excel(xls, sheet_name=sheet_name)
    total_stay_duration, location_days, monthly_location = process_sheet(df)

    # 将每个工作表的结果保存
    results[sheet_name] = {
        "total_stay_duration": total_stay_duration,
        "location_days": location_days,
        "monthly_location": monthly_location
    }

# 保存结果为新的Excel文件
with pd.ExcelWriter(r"C:\Users\joengzaang\myfiles\统计数据\轨迹数据\apple一生足迹\处理后的停留数据.xlsx") as writer:
    for sheet_name, result in results.items():
        result['total_stay_duration'].to_excel(writer, sheet_name=f"{sheet_name}_总停留时长")
        result['location_days'].to_excel(writer, sheet_name=f"{sheet_name}_停留日期")
        result['monthly_location'].to_excel(writer, sheet_name=f"{sheet_name}_月度分析")

print("处理完成，结果已保存到新的Excel文件。")
