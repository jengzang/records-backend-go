# ManicTime CSV转换指南

## 问题
当前的.csv文件实际上还是Excel格式（二进制），需要真正转换为CSV文本格式。

## 解决方案

### 方法1: 在Excel中另存为CSV (推荐)

1. **打开每个文件**:
   - ManicTimeData_2026-01-26.xlsx
   - ManicTimeData_2026-01-26(1).xlsx
   - ManicTimeData_2026-01-26(2).xlsx
   - ManicTimeData_2026-01-26(3).xlsx

2. **另存为CSV**:
   - 文件 → 另存为
   - 文件类型选择: **CSV UTF-8 (逗号分隔)(*.csv)**
   - 保存位置: 同一目录
   - 文件名: 保持原名但扩展名为.csv

3. **重要**: 如果Excel提示"某些功能可能会丢失"，选择"是"继续

### 方法2: 使用LibreOffice (如果Excel不可用)

1. 用LibreOffice Calc打开Excel文件
2. 文件 → 另存为
3. 文件类型: Text CSV (.csv)
4. 字符集: Unicode (UTF-8)
5. 字段分隔符: 逗号

### 方法3: 使用Python脚本转换

如果有很多文件需要转换，可以使用以下脚本:

```python
import pandas as pd
from pathlib import Path

# 转换所有Excel文件为CSV
data_dir = Path('go-backend/data/screentime/ManicTime')

for excel_file in data_dir.glob('*.xlsx'):
    csv_file = excel_file.with_suffix('.csv')
    print(f"Converting {excel_file.name} to {csv_file.name}")

    try:
        # 读取Excel
        df = pd.read_excel(excel_file)
        # 保存为CSV
        df.to_csv(csv_file, index=False, encoding='utf-8')
        print(f"  Success! Rows: {len(df)}")
    except Exception as e:
        print(f"  Failed: {e}")
```

## 验证CSV文件

转换后，可以用文本编辑器打开CSV文件，应该看到类似这样的内容:

```
Start time,End time,Duration,Application,Window title
2026-01-26 09:00:00,2026-01-26 09:15:00,00:15:00,Chrome,Gmail
2026-01-26 09:15:00,2026-01-26 09:30:00,00:15:00,VSCode,main.py
```

而不是乱码或二进制数据。

## 下一步

转换完成后，运行导入脚本:

```bash
cd go-backend/scripts/screentime
python import_manictime.py
```

导入脚本会自动:
1. 检测CSV文件结构
2. 识别不同类型的统计数据
3. 导入所有4个文件
4. 生成统计报告
