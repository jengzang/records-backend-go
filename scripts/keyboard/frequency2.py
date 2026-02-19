import os
import numpy as np
import matplotlib.pyplot as plt
from matplotlib import rcParams

# 设置Matplotlib以支持中文字符
rcParams['font.sans-serif'] = ['SimHei']  # 使用黑体
rcParams['axes.unicode_minus'] = False  # 解决负号显示问题


def read_data(file_path):
    time_s = []
    f0_hz = []
    curves = []
    current_curve_length = 0

    with open(file_path, 'r', encoding='utf-8') as file:
        for line in file:
            if line.strip():
                try:
                    time, f0 = line.split()
                    if f0 != '--undefined--':
                        time_s.append(float(time))
                        f0_hz.append(float(f0))
                        current_curve_length += 1
                    else:
                        if time_s and f0_hz:
                            curves.append((time_s, f0_hz, current_curve_length))
                            time_s = []
                            f0_hz = []
                            current_curve_length = 0
                except ValueError:
                    continue
    if time_s and f0_hz:
        curves.append((time_s, f0_hz, current_curve_length))

    return curves


def filter_short_curves(curves, min_length=0.05):
    filtered_curves = []
    for time_s, f0_hz, length in curves:
        if time_s and (time_s[-1] - time_s[0]) >= min_length:
            filtered_curves.append((time_s, f0_hz, length))
    return filtered_curves


def convert_to_t_values(frequencies):
    min_freq = min(frequencies)
    max_freq = max(frequencies)
    t_values = [(np.log10(f) - np.log10(min_freq)) / (np.log10(max_freq) - np.log10(min_freq)) * 5 for f in frequencies]
    return t_values


def interpolate_and_filter_t_values(curves, num_points=20, deviation_threshold=0.25):
    interpolated_t_values = []
    total_times = [time_s[-1] for time_s, _, _ in curves]
    avg_length = np.mean(total_times)
    common_time_s = np.linspace(0, avg_length, num_points)

    for time_s, f0_hz, _ in curves:
        adjusted_time_s = [t - time_s[0] for t in time_s]  # 将起始时间归一化为0
        interpolated_freqs = np.interp(common_time_s, adjusted_time_s, f0_hz)
        t_values = convert_to_t_values(interpolated_freqs)
        interpolated_t_values.append(t_values)

    interpolated_t_values = np.array(interpolated_t_values)
    mean_t_values = []

    for i in range(num_points):
        current_points = interpolated_t_values[:, i]
        valid_points = []
        for j, point in enumerate(current_points):
            other_points = np.delete(current_points, j)
            mean_of_rest = np.mean(other_points)
            if abs(point - mean_of_rest) <= deviation_threshold * mean_of_rest:
                valid_points.append(point)
        if valid_points:
            mean_t_values.append(np.mean(valid_points))
        else:
            mean_t_values.append(np.mean(current_points))  # Fallback to overall mean if all values are filtered out

    return common_time_s, mean_t_values


def plot_tone_curve(file_name, curves, output_folder, avg_length):
    plt.figure(figsize=(12, 8))
    for i, (time_s, f0_hz, _) in enumerate(curves):
        adjusted_time_s = [t - time_s[0] for t in time_s]  # 将起始时间归一化为0
        t_values = convert_to_t_values(f0_hz)
        plt.plot(adjusted_time_s, t_values, label=f'Curve {i + 1}', linewidth=1)

    plt.xlabel('Time (s)')
    plt.ylabel('T value')
    plt.title(f'{file_name} - Tone Curve')
    plt.legend()
    plt.grid(True)
    plt.xlim(0, avg_length)

    output_file_name = f'{file_name[:-4]}音调曲线.png'
    output_file_path = os.path.join(output_folder, output_file_name)
    plt.savefig(output_file_path)
    print(f'Successfully saved: {output_file_path}')
    plt.close()


def plot_all_tone_curves(mean_curves, output_folder, title_suffix, avg_length):
    plt.figure(figsize=(12, 8))
    for file_name, mean_curve in mean_curves.items():
        mean_time_s, mean_t_values = mean_curve
        base_name = os.path.splitext(file_name)[0]
        plt.plot(mean_time_s, mean_t_values, label=base_name, linewidth=2)

    plt.xlabel('Time (s)')
    plt.ylabel('T value')
    plt.title(f'All Files Mean Tone Curves {title_suffix}')
    plt.legend()
    plt.grid(True)
    plt.xlim(0, avg_length)

    if mean_curves:
        first_file_name = list(mean_curves.keys())[0]
        output_file_name = f'{first_file_name[:2]}{title_suffix}音调汇总.png'
        output_file_path = os.path.join(output_folder, output_file_name)
        plt.savefig(output_file_path)
        print(f'Successfully saved: {output_file_path}')
    plt.close()


def process_folder_for_tone(folder_path):
    output_folder2 = os.path.join(folder_path, '音调图2')
    if not os.path.exists(output_folder2):
        os.makedirs(output_folder2)

    mean_curves_regular = {}
    mean_curves_ru = {}
    all_tone_curves = {}

    for file_name in os.listdir(folder_path):
        if file_name.endswith('.txt'):
            file_path = os.path.join(folder_path, file_name)
            curves = read_data(file_path)
            filtered_curves = filter_short_curves(curves)
            avg_length = np.mean([time_s[-1] for time_s, _, _ in filtered_curves])  # 计算平均长度
            if "入" in file_name:
                mean_t_curve = interpolate_and_filter_t_values(filtered_curves, num_points=20)
                mean_curves_ru[file_name] = mean_t_curve
            else:
                mean_t_curve = interpolate_and_filter_t_values(filtered_curves, num_points=20)
                mean_curves_regular[file_name] = mean_t_curve
            all_tone_curves[file_name] = mean_t_curve
            plot_tone_curve(file_name, filtered_curves, output_folder2, avg_length)

    if mean_curves_regular:
        plot_all_tone_curves(mean_curves_regular, output_folder2, title_suffix="舒",
                             avg_length=np.mean([curve[0][-1] for curve in mean_curves_regular.values()]))
    if mean_curves_ru:
        plot_all_tone_curves(mean_curves_ru, output_folder2, title_suffix="入",
                             avg_length=np.mean([curve[0][-1] for curve in mean_curves_ru.values()]))
    if all_tone_curves:
        plot_all_tone_curves(all_tone_curves, output_folder2, title_suffix="全部",
                             avg_length=np.mean([curve[0][-1] for curve in all_tone_curves.values()]))


folder_path = r'C:\Users\joengzaang\myfiles\阳春方言调查\语料\岗美'  # 使用您的文件夹路径
process_folder_for_tone(folder_path)
