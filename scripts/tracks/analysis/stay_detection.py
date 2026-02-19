import pandas as pd
from datetime import datetime

# 读取CSV文件
file_path = "C:/Users/joengzaang/Desktop/处理后的一生足迹数据.csv"
df = pd.read_csv(file_path)

# 处理时间列
df['time'] = pd.to_datetime(df['time'], format='%Y%m%d%H%M%S')

# 提取日期（年月日）
df['date'] = df['time'].dt.date

# # 处理缺失值：忽略市级为NaN的行
# df = df.dropna(subset=['市级'])

# 初始化停留数据
stay_data = []
stay_duration = 0  # 累计的停留时间（秒）
stay_group = 0     # 当前停留段的编号

# 记录前一行的时间和地点
previous_time = None
previous_location = None
previous_time0 = None

# 循环遍历数据
for idx, row in df.iterrows():
    # 获取当前的时间、地点（市级、区县级、乡镇级）
    current_time = row['time']
    current_location = (row['市级'], row['区县级'], row['乡镇级'])
    # current_location = (row['市级'], row['区县级'])
    # current_location = (row['市级'])
    # print(row['市级'], row['区县级'], row['乡镇级'])

    # 检查该行中市级、区县级、乡镇级是否有任何一个为空
    if pd.isnull(row['市级']) or pd.isnull(row['区县级']) or pd.isnull(row['乡镇级']) or \
            row['市级'] == "" or row['区县级'] == "" or row['乡镇级'] == "" or \
            row['市级'] is None or row['区县级'] is None or row['乡镇级'] is None:
        print(row)
        # stay_group += 1  # 切换到下一个停留段
        previous_time0 = current_time
        previous_time = None
        # 重置停留时间
        stay_duration = 0  # 重新开始计时
        previous_location = None
        # exit("程序因某种原因终止。")
        continue


    # 计算与上一条记录的时间差
    if previous_time is not None:
        time_diff = (current_time - previous_time).total_seconds()  # 时间差，单位是秒
    else:
        time_diff = 0

    # 判断是否是同一停留段：如果地点一致，则累积停留时间
    if previous_location == current_location:
        stay_duration += time_diff  # 累计停留时间
        # 如果上一段停留时间超过4小时，记录上一段停留
    else:
        if stay_duration > 2 * 3600:  # 停留超过4小时，记录停留段
            stay_data.append({
                'stay_group': stay_group,
                'start_time': previous_time0,
                'end_time': previous_time,
                'stay_duration': stay_duration / 3600,  # 转化为小时
                'location': previous_location,
            })
        stay_group += 1  # 切换到下一个停留段
        previous_time0 = current_time
        # 重置停留时间
        stay_duration = 0  # 重新开始计时

    # 更新上一行的数据
    previous_time = current_time
    previous_location = current_location

# 记录最后一个停留段（如果它超过了4小时）
if stay_duration > 2 * 3600:
    stay_data.append({
        'stay_group': stay_group,
        'start_time': previous_time0,
        'end_time': previous_time,
        'stay_duration': stay_duration / 3600,  # 转化为小时
        'location': previous_location,
    })

# 将停留段数据转换为DataFrame
stay_df = pd.DataFrame(stay_data)

# 保存结果到CSV
stay_df.to_csv("停留段.csv", index=False, encoding='utf-8-sig')

print("处理完成，停留段已保存到文件！")
