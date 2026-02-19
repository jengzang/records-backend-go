import os
import numpy as np
import matplotlib.pyplot as plt
from matplotlib import rcParams
from adjustText import adjust_text  # 引入adjustText库来避免重叠

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


def normalize_time(curves):
    normalized_curves = []
    lengths = [time_s[-1] - time_s[0] for time_s, _, _ in curves]
    avg_length = np.mean(lengths)

    for time_s, f0_hz, length in curves:
        min_time = time_s[0]
        max_time = time_s[-1]
        normalized_time_s = [(t - min_time) / (max_time - min_time) * avg_length for t in time_s]
        normalized_curves.append((normalized_time_s, f0_hz, length))

    return normalized_curves, avg_length


def interpolate_and_average(curves, avg_length, num_points=1000, deviation_threshold=0.25):
    all_interpolated_freqs = []
    common_time_s = np.linspace(0, avg_length, num_points)

    for time_s, f0_hz, _ in curves:
        interpolated_freqs = np.interp(common_time_s, time_s, f0_hz)
        all_interpolated_freqs.append(interpolated_freqs)

    all_interpolated_freqs = np.array(all_interpolated_freqs)
    mean_freqs = []

    for i in range(num_points):
        current_points = all_interpolated_freqs[:, i]
        valid_points = []
        for j, point in enumerate(current_points):
            other_points = np.delete(current_points, j)
            mean_of_rest = np.mean(other_points)
            if abs(point - mean_of_rest) <= deviation_threshold * mean_of_rest:
                valid_points.append(point)
        if valid_points:
            mean_freqs.append(np.mean(valid_points))
        else:
            mean_freqs.append(np.mean(current_points))  # Fallback to overall mean if all values are filtered out

    return common_time_s, mean_freqs


def plot_individual_curves(file_name, normalized_curves, mean_curve, output_folder, avg_length):
    plt.figure(figsize=(12, 8))
    for i, (time_s, f0_hz, _) in enumerate(normalized_curves):
        plt.plot(time_s, f0_hz, label=f'Curve {i + 1}', linewidth=1)

    mean_time_s, mean_freqs = mean_curve
    plt.plot(mean_time_s, mean_freqs, label='Mean Curve', linewidth=3, color='black')
    plt.xlabel('Normalized Time')
    plt.ylabel('Frequency (Hz)')
    plt.title(f'{file_name} - Individual and Mean Curves')
    plt.legend()
    plt.grid(True)
    plt.xlim(0, avg_length)

    output_file_name = f'{file_name[:-4]}声调曲线.png'
    output_file_path = os.path.join(output_folder, output_file_name)
    plt.savefig(output_file_path)
    print(f'Successfully saved: {output_file_path}')
    plt.close()


def plot_all_mean_curves(mean_curves, output_folder, title_suffix, avg_length, add_guidelines=False):
    plt.figure(figsize=(12, 8))
    min_freq = float('inf')
    max_freq = float('-inf')
    texts = []

    for file_name, mean_curve in mean_curves.items():
        mean_time_s, mean_freqs = mean_curve
        min_freq = min(min_freq, min(mean_freqs))
        max_freq = max(max_freq, max(mean_freqs))
        base_name = os.path.splitext(file_name)[0]
        plt.plot(mean_time_s, mean_freqs, label=base_name, linewidth=2)
        if add_guidelines:
            texts.append(plt.text(mean_time_s[-1], mean_freqs[-1], base_name, fontsize=9))

    plt.xlabel('Normalized Time')
    plt.ylabel('Frequency (Hz)')
    plt.title(f'All Files Mean Frequency Curves {title_suffix}')
    plt.grid(True)
    plt.xlim(0, avg_length)

    # 调整文本位置以避免重叠
    if add_guidelines:
        adjust_text(texts, arrowprops=dict(arrowstyle="->", color='r', lw=0.5))

        for i in range(1, 6):
            plt.axhline(y=min_freq + (max_freq - min_freq) * (i - 1) / 4, color='grey', linestyle='--')
            plt.text(0, min_freq + (max_freq - min_freq) * (i - 1) / 4, str(i), va='center', ha='right', fontsize=10,
                     color='grey')

    if mean_curves:
        first_file_name = list(mean_curves.keys())[0]
        output_file_name = f'{first_file_name[:2]}{title_suffix}声调汇总.png'
        output_file_path = os.path.join(output_folder, output_file_name)
        plt.savefig(output_file_path)
        print(f'Successfully saved: {output_file_path}')
    plt.close()


def process_folder(folder_path):
    output_folder1 = os.path.join(folder_path, '声调图1')
    if not os.path.exists(output_folder1):
        os.makedirs(output_folder1)

    mean_curves_regular = {}
    mean_curves_ru = {}
    all_curves = {}
    avg_lengths = []

    for file_name in os.listdir(folder_path):
        if file_name.endswith('.txt'):
            file_path = os.path.join(folder_path, file_name)
            curves = read_data(file_path)
            filtered_curves = filter_short_curves(curves)
            normalized_curves, avg_length = normalize_time(filtered_curves)
            mean_curve = interpolate_and_average(normalized_curves, avg_length)
            avg_lengths.append(avg_length)

            if "入" in file_name:
                mean_curves_ru[file_name] = mean_curve
            else:
                mean_curves_regular[file_name] = mean_curve

            all_curves[file_name] = mean_curve
            plot_individual_curves(file_name, normalized_curves, mean_curve, output_folder1, avg_length)

    max_avg_length = max(avg_lengths) if avg_lengths else 1  # 如果没有长度数据，默认1

    if mean_curves_regular:
        plot_all_mean_curves(mean_curves_regular, output_folder1, title_suffix="舒", avg_length=max_avg_length)
    if mean_curves_ru:
        plot_all_mean_curves(mean_curves_ru, output_folder1, title_suffix="入", avg_length=max_avg_length)
    if all_curves:
        plot_all_mean_curves(all_curves, output_folder1, title_suffix="全部", avg_length=max_avg_length,
                             add_guidelines=True)


folder_path = r'C:\Users\joengzaang\myfiles\阳春方言调查\语料\岗美'  # 使用您的文件夹路径
process_folder(folder_path)