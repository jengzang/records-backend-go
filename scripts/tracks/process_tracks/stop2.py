import pandas as pd
from datetime import datetime, timedelta

# 文件路径
input_file_path = r"C:\Users\joengzaang\myfiles\统计数据\轨迹数据\apple一生足迹\停留.xlsx"
output_file_path = r"C:\Users\joengzaang\myfiles\统计数据\轨迹数据\apple一生足迹\停留_结果.xlsx"

# 读取 Excel 文件
xls = pd.ExcelFile(input_file_path)

# 读取所有工作表名称
sheet_names = xls.sheet_names

# 创建一个 ExcelWriter 用于保存输出结果
with pd.ExcelWriter(output_file_path, engine='openpyxl') as writer:
    # 处理每个工作表
    for sheet in sheet_names:
        # 读取工作表数据
        df = pd.read_excel(input_file_path, sheet_name=sheet)

        # 假设每个工作表包含这几列：stay_group, start_time, end_time, stay_duration, location
        # 如果列名不同，请调整下面的列名
        df['start_time'] = pd.to_datetime(df['start_time'], format='%Y/%m/%d %H:%M')
        df['end_time'] = pd.to_datetime(df['end_time'], format='%Y/%m/%d %H:%M')


        # 计算每个地点的停留天数及具体日期
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
                    # 格式化日期为 "YYYY.M.D" 形式
                    formatted_day = day.strftime('%Y.%m.%d')
                    days_with_stay.append(formatted_day)

            return days_with_stay


        # 应用到数据框
        df['days_with_stay_above_4_hours'] = df.apply(find_days_with_stay_above_4_hours, axis=1)

        # 展开列表并按 location 分组
        location_days = df.explode('days_with_stay_above_4_hours').groupby('location')[
            'days_with_stay_above_4_hours'].unique()

        # 计算每个地点停留的天数
        location_days_count = location_days.apply(lambda x: len(set(x)))  # 使用 set 去重日期

        # 将日期和停留天数合并
        location_days_df = location_days.reset_index()
        location_days_count_df = location_days_count.reset_index()

        # 将停留天数列添加到数据框中
        location_days_df['Stay_Days_Count'] = location_days_count.values  # 通过 .values 访问 Series

        # 将处理结果写入新的 Excel 工作表
        location_days_df.columns = ['Location', 'Days_Stayed', 'Stay_Days_Count']

        # 计算按月份分析，每月每个城市停留的天数
        df['Month'] = df['start_time'].dt.to_period('M')  # 提取月份
        month_location_days = df.explode('days_with_stay_above_4_hours').groupby(
            ['Month', 'location']).size().reset_index(name='Stay_Days_Count')

        # 将月份和城市停留天数分析写入工作表
        month_location_days_df = month_location_days.rename(
            columns={'Month': 'Month_Year', 'location': 'City', 'Stay_Days_Count': 'Stay_Days_Per_Month'})

        # 确保数据框非空再写入
        if not location_days_df.empty:
            location_days_df.to_excel(writer, sheet_name=f'{sheet}_location_days', index=False)

        if not month_location_days_df.empty:
            month_location_days_df.to_excel(writer, sheet_name=f'{sheet}_month_location', index=False)

print(f"处理结果已保存到: {output_file_path}")
