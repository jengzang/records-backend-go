# ManicTime Data Import - Status Report

## 实施日期
2026-02-24

## 当前状态
⚠️ **部分完成** - 数据库Schema已创建，但Excel导入遇到技术问题

---

## 已完成工作

### 1. 数据库Schema ✅
**文件**: `go-backend/scripts/screentime/migrations/003_create_manictime_tables.sql`

**表结构**:
- `manictime_activities` - 活动记录表
- `manictime_apps` - 应用元数据表
- `manictime_daily` - 每日统计表
- `manictime_statistics` - 统计缓存表

**数据库文件**: `go-backend/data/screentime/manictime_computer.db` (已创建)

### 2. 导入脚本 ✅
**文件**: `go-backend/scripts/screentime/import_manictime.py`

**功能**:
- 自动读取ManicTime Excel文件
- 智能列识别 (Start time, End time, Application, Duration等)
- 自动应用分类 (Development, Browser, Communication等)
- 批量导入和统计更新

---

## 遇到的问题

### 问题: openpyxl读取Excel文件失败

**错误信息**:
```
Error reading Excel file: expected <class 'openpyxl.styles.fills.Fill'>
```

**原因**:
ManicTime导出的Excel文件包含复杂的样式信息，openpyxl 3.1.5无法正确解析这些样式。这是openpyxl的已知限制。

**影响**:
- 无法直接导入ManicTime Excel文件
- 需要使用替代方案

---

## 解决方案

### 方案1: 转换Excel格式 (推荐)

**步骤**:
1. 在Excel中打开ManicTime导出的文件
2. 另存为 → 选择"Excel工作簿 (*.xlsx)" 或 "CSV (逗号分隔) (*.csv)"
3. 保存时选择"否"(不保留格式)
4. 使用转换后的文件重新运行导入脚本

**优点**:
- 简单快速
- 保留所有数据
- 兼容性好

### 方案2: 使用CSV导出

如果ManicTime支持CSV导出:
1. 在ManicTime中导出为CSV格式
2. 修改导入脚本使用pandas.read_csv()
3. 重新运行导入

### 方案3: 手动数据提取

如果数据量不大:
1. 在Excel中复制数据
2. 粘贴到新的Excel文件
3. 保存为简单格式
4. 运行导入脚本

---

## 数据文件信息

**ManicTime文件**:
```
ManicTimeData_2026-01-26.xlsx      (2.3 KB)
ManicTimeData_2026-01-26(1).xlsx   (211 KB)
ManicTimeData_2026-01-26(2).xlsx   (8.6 MB) - 最大
ManicTimeData_2026-01-26(3).xlsx   (744 KB)
```

**总大小**: ~9.6 MB
**位置**: `go-backend/data/screentime/ManicTime/`

---

## 下一步行动

### 立即行动 (推荐)

1. **转换Excel文件**:
   ```bash
   # 在Excel中打开每个文件
   # 另存为 → Excel工作簿 (不保留格式)
   # 保存到同一目录
   ```

2. **重新运行导入**:
   ```bash
   cd go-backend/scripts/screentime
   python import_manictime.py
   ```

3. **验证导入**:
   ```bash
   # 检查数据库
   sqlite3 go-backend/data/screentime/manictime_computer.db "SELECT COUNT(*) FROM manictime_activities;"
   ```

### 替代方案

如果无法转换Excel文件，可以:
1. 使用其他Excel读取库 (xlrd, pyexcel)
2. 使用LibreOffice Calc转换
3. 编写自定义解析器

---

## 技术细节

### openpyxl版本
- 当前版本: 3.1.5 (最新)
- Python版本: 3.12
- pandas版本: 2.2.3

### 尝试的解决方法
1. ✅ 使用engine='openpyxl'显式指定
2. ✅ 使用read_only=True跳过样式
3. ✅ 使用data_only=True只读数据
4. ✅ 手动迭代行读取
5. ❌ 所有方法都失败

### 根本原因
ManicTime Excel文件的样式信息格式与openpyxl期望的格式不兼容。这是Excel文件格式的兼容性问题，不是代码问题。

---

## 已创建的文件

1. ✅ `go-backend/scripts/screentime/migrations/003_create_manictime_tables.sql`
2. ✅ `go-backend/scripts/screentime/import_manictime.py`
3. ✅ `go-backend/data/screentime/manictime_computer.db` (空数据库)
4. ✅ `go-backend/docs/screentime/MANICTIME_IMPORT_STATUS.md` (本文件)

---

## 预期结果 (转换后)

一旦Excel文件转换成功，导入脚本应该能够:
- 读取所有4个Excel文件
- 提取活动记录 (预计数千到数万条)
- 自动分类应用
- 生成统计数据
- 输出导入摘要

**预期输出示例**:
```
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

## 总结

**完成度**: 80%
- ✅ 数据库Schema设计完成
- ✅ 导入脚本编写完成
- ✅ 自动分类逻辑实现
- ⚠️ Excel读取遇到技术限制
- ⏳ 需要转换Excel格式

**建议**:
1. 使用Excel打开并重新保存文件 (去除复杂格式)
2. 或者联系ManicTime支持，询问是否有CSV导出选项
3. 转换后重新运行导入脚本

**预计时间**:
- Excel转换: 10-15分钟
- 重新导入: 2-5分钟
- 验证测试: 5分钟
- **总计**: 20-25分钟

---

**报告日期**: 2026-02-24
**状态**: 等待Excel文件转换
