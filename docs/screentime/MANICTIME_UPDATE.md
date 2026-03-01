# ManicTime数据导入 - 更新说明

## 更新时间
2026-02-24

## 更新内容

### ✅ 导入脚本增强

已更新 `import_manictime.py` 脚本，新增以下功能:

#### 1. 自动文件格式检测
- ✅ 自动识别CSV和Excel文件
- ✅ 支持多种编码格式 (UTF-8, GBK, GB2312, Latin1, CP1252)
- ✅ 智能回退机制 (CSV失败→Excel，openpyxl失败→pandas)

#### 2. 多文件类型支持
- ✅ 支持 `.xlsx` Excel文件
- ✅ 支持 `.csv` CSV文件
- ✅ 自动处理所有文件（不限数量）

#### 3. 数据结构识别
脚本会自动识别以下列名模式:
- **时间列**: Start time, End time, Start, End, From, To
- **应用列**: Application, App, Program, Process
- **时长列**: Duration, Time, Length
- **标题列**: Title, Window, Document, Window title

#### 4. 保留所有数据
- ✅ 导入所有4个文件的数据
- ✅ 不同统计方式的数据都会保留
- ✅ 自动去重（基于时间+应用）

---

## 关于4个文件

根据文件大小分析，这4个文件可能代表:

1. **ManicTimeData_2026-01-26.xlsx (2.3 KB)**
   - 可能是: 汇总统计或配置信息
   - 建议: 保留，可能包含重要元数据

2. **ManicTimeData_2026-01-26(1).xlsx (211 KB)**
   - 可能是: 每日汇总数据
   - 建议: 保留，用于日统计

3. **ManicTimeData_2026-01-26(2).xlsx (8.6 MB)** - 最大
   - 可能是: 详细活动日志（每个活动记录）
   - 建议: **必须保留**，这是最详细的数据

4. **ManicTimeData_2026-01-26(3).xlsx (744 KB)**
   - 可能是: 应用级汇总或周统计
   - 建议: 保留，补充统计维度

### 我的建议
**保留所有4个文件**，原因:
- 不同文件可能包含不同粒度的数据
- 导入脚本会自动去重
- 更多数据 = 更准确的统计
- 磁盘空间占用不大 (~10MB)

---

## 当前问题

### 文件格式问题
当前的`.csv`文件实际上还是Excel二进制格式，需要真正转换为CSV文本格式。

### 解决方案

#### 方法1: Excel另存为 (最简单)
```
1. 在Excel中打开文件
2. 文件 → 另存为
3. 文件类型: CSV UTF-8 (逗号分隔)(*.csv)
4. 保存
```

#### 方法2: 保持Excel格式
如果转换困难，可以:
1. 将文件重命名回 `.xlsx`
2. 运行导入脚本
3. 脚本会尝试多种方法读取

---

## 使用方法

### 1. 准备文件
确保文件在正确位置:
```
go-backend/data/screentime/ManicTime/
├── ManicTimeData_2026-01-26.csv (或 .xlsx)
├── ManicTimeData_2026-01-26(1).csv
├── ManicTimeData_2026-01-26(2).csv
└── ManicTimeData_2026-01-26(3).csv
```

### 2. 运行导入
```bash
cd go-backend/scripts/screentime
python import_manictime.py
```

### 3. 查看结果
脚本会输出:
- 每个文件的处理状态
- 提取的活动数量
- 应用分类结果
- Top 5应用统计

---

## 预期输出

```
=== ManicTime Data Import ===
Found 4 data files:
  - ManicTimeData_2026-01-26.csv (0.0 MB) [CSV]
  - ManicTimeData_2026-01-26(1).csv (0.2 MB) [CSV]
  - ManicTimeData_2026-01-26(2).csv (8.6 MB) [CSV]
  - ManicTimeData_2026-01-26(3).csv (0.7 MB) [CSV]

=== Processing: ManicTimeData_2026-01-26.csv ===
Attempting to read as CSV...
CSV read successful!
  Rows: 50, Columns: 8
  Extracted 45 activities

=== Processing: ManicTimeData_2026-01-26(1).csv ===
...

=== Import Summary ===
Total activities: 15,234
Total applications: 87
Date range: 20250101 to 20260126
Total duration: 1,234.5 hours

Top 5 applications:
  1. Chrome (Browser): 345.2h
  2. VSCode (Development): 234.1h
  3. Slack (Communication): 123.4h
  4. Excel (Office): 98.7h
  5. Spotify (Entertainment): 67.3h
```

---

## 自动分类

脚本会自动将应用分类为:
- **Development**: VSCode, PyCharm, IntelliJ, etc.
- **Browser**: Chrome, Firefox, Edge, Safari
- **Communication**: Slack, Teams, Zoom, Discord
- **Office**: Word, Excel, PowerPoint, Outlook
- **Entertainment**: Spotify, Music, Video, Games
- **System**: Explorer, Terminal, CMD
- **Other**: 未分类应用

---

## 故障排除

### 问题1: 文件读取失败
```
Error: 'utf-8' codec can't decode...
```
**解决**: 文件还是Excel格式，需要真正转换为CSV

### 问题2: 没有提取到数据
```
Extracted 0 activities
```
**解决**:
1. 检查CSV列名是否正确
2. 查看脚本输出的列名
3. 可能需要调整列名识别逻辑

### 问题3: openpyxl错误
```
expected <class 'openpyxl.styles.fills.Fill'>
```
**解决**:
1. 转换为CSV格式
2. 或使用LibreOffice打开后另存为

---

## 下一步

1. **转换文件格式** (如果还是Excel格式)
2. **运行导入脚本**
3. **验证数据**:
   ```bash
   sqlite3 go-backend/data/screentime/manictime_computer.db "SELECT COUNT(*) FROM manictime_activities;"
   ```
4. **查看统计**:
   ```bash
   sqlite3 go-backend/data/screentime/manictime_computer.db "SELECT * FROM manictime_apps ORDER BY total_duration_seconds DESC LIMIT 10;"
   ```

---

## 相关文档

- [CSV转换指南](./CSV_CONVERSION_GUIDE.md)
- [导入状态报告](./MANICTIME_IMPORT_STATUS.md)
- [数据库Schema](../migrations/003_create_manictime_tables.sql)

---

**更新时间**: 2026-02-24
**状态**: 等待CSV文件转换
**完成度**: 90% (脚本完成，等待数据)
