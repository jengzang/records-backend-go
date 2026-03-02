# Personal Efficiency Curve - 个人效率曲线

## 概述

个人效率曲线是一个跨模块分析功能，通过整合键盘、屏幕时间和健康数据，生成24小时效率评分曲线，帮助用户识别最佳工作时段和生物钟类型。

## 功能特性

### 1. 小时级效率评分
- 综合评分算法：打字速度(30%) + 工作应用比例(20%) + 心率变异性(20%) + 专注度(15%) + 活动量(15%)
- 数据完整性追踪：标记每小时可用的数据源
- 归一化处理：所有指标归一化到0-100范围

### 2. 效率曲线画像
- **工作日画像**：基于周一至周五的数据生成
- **周末画像**：基于周六日的数据生成
- 峰值时段检测：自动识别连续高效时段
- 低效时段检测：识别需要改进的时段

### 3. 生物钟类型分类
- **早晨型**：峰值效率在6-10am
- **夜猫子型**：峰值效率在20-24点
- **中间型**：峰值效率在10-18点
- 置信度评分：0-1范围

### 4. 工作日vs周末对比
- 平均效率差异
- 峰值时段差异
- 24小时逐时对比
- 自动生成解读建议

### 5. 可操作洞察
- 峰值生产力时段建议
- 生物钟类型建议
- 工作日周末差异建议
- 改进建议

## 数据库架构

### 表1: hourly_efficiency_scores
存储每小时的效率评分和原始指标。

**关键字段**:
- `date`, `hour`: 时间标识
- `typing_speed`, `work_app_ratio`, `avg_heart_rate`, `step_count`: 原始指标
- `efficiency_score`: 综合效率评分 (0-100)
- `data_completeness`: 数据完整性 (0-1)

### 表2: efficiency_curve_profiles
存储工作日和周末的聚合效率曲线。

**关键字段**:
- `profile_type`: 'workday' 或 'weekend'
- `hour_0` ~ `hour_23`: 24小时效率曲线
- `peak_hour`, `peak_score`: 峰值时段
- `chronotype`: 生物钟类型
- `avg_efficiency`, `std_efficiency`: 统计指标

### 表3: efficiency_insights
存储自动生成的洞察和建议。

**关键字段**:
- `insight_type`: 洞察类型
- `priority`: 优先级 (0=低, 1=中, 2=高)
- `title`, `description`, `recommendation`: 内容
- `confidence`: 置信度 (0-1)

## API端点

### 1. GET /api/v1/cross-module/efficiency-curve/hourly
获取小时级效率曲线数据。

**查询参数**:
- `start_date`: 开始日期 (YYYY-MM-DD)，默认30天前
- `end_date`: 结束日期 (YYYY-MM-DD)，默认今天

**响应示例**:
```json
{
  "scores": [
    {
      "date": "2026-03-01",
      "hour": 9,
      "typing_speed": 1200.5,
      "work_app_ratio": 0.85,
      "avg_heart_rate": 72.3,
      "efficiency_score": 78.5,
      "data_completeness": 1.0
    }
  ],
  "stats": {
    "total_hours": 720,
    "avg_efficiency": 65.2,
    "max_efficiency": 92.1,
    "min_efficiency": 28.4,
    "data_completeness": 0.87
  }
}
```

### 2. GET /api/v1/cross-module/efficiency-curve/profile
获取效率曲线画像（工作日或周末）。

**查询参数**:
- `profile_type`: 'workday' 或 'weekend'

**响应示例**:
```json
{
  "profile_type": "workday",
  "hourly_curve": [45.2, 42.1, ..., 58.9],
  "peak_hour": 10,
  "peak_score": 82.5,
  "peak_start_hour": 9,
  "peak_end_hour": 11,
  "chronotype": "morning",
  "chronotype_confidence": 0.78,
  "avg_efficiency": 65.3,
  "total_samples": 120
}
```

### 3. GET /api/v1/cross-module/efficiency-curve/comparison
获取工作日vs周末对比。

**响应示例**:
```json
{
  "workday": { ... },
  "weekend": { ... },
  "diff": {
    "avg_diff": 12.5,
    "peak_hour_diff": 2,
    "hourly_diff": [5.2, 3.1, ..., -2.4],
    "interpretation": "工作日效率显著高于周末，建议保持工作日的良好习惯"
  }
}
```

### 4. GET /api/v1/cross-module/efficiency-curve/insights
获取可操作洞察。

**响应示例**:
```json
[
  {
    "insight_type": "peak_hours",
    "priority": 2,
    "title": "你的峰值生产力在9-11am",
    "description": "数据显示你在上午9-11点效率最高，平均效率评分82.5",
    "recommendation": "建议将重要任务安排在这个时段",
    "confidence": 0.85
  }
]
```

### 5. POST /api/v1/cross-module/efficiency-curve/analyze
触发效率分析（生成小时级评分和画像）。

**查询参数**:
- `start_date`: 开始日期 (YYYY-MM-DD)
- `end_date`: 结束日期 (YYYY-MM-DD)

**响应示例**:
```json
{
  "message": "Efficiency analysis completed successfully",
  "start_date": "2026-02-01",
  "end_date": "2026-03-01"
}
```

## 使用流程

### 1. 首次使用
```bash
# 1. 运行迁移创建数据库表
cd go-backend
python run_migration_simple.py

# 2. 触发效率分析（生成数据）
curl -X POST "http://localhost:18080/api/v1/cross-module/efficiency-curve/analyze?start_date=2026-02-01&end_date=2026-03-01"

# 3. 查看效率曲线
curl "http://localhost:18080/api/v1/cross-module/efficiency-curve/hourly?start_date=2026-02-01&end_date=2026-03-01"

# 4. 查看工作日画像
curl "http://localhost:18080/api/v1/cross-module/efficiency-curve/profile?profile_type=workday"

# 5. 查看对比分析
curl "http://localhost:18080/api/v1/cross-module/efficiency-curve/comparison"
```

### 2. 定期更新
建议每周运行一次分析，更新效率数据：
```bash
curl -X POST "http://localhost:18080/api/v1/cross-module/efficiency-curve/analyze?start_date=2026-02-24&end_date=2026-03-02"
```

## 数据依赖

### 必需数据源
- **键盘数据** (`keyboard/kmcounter.db`): 打字速度
- **屏幕时间数据** (`screentime/screentime.db`): 应用使用模式
- **健康数据** (`applehealth/health.db`): 心率、HRV、步数

### 数据缺失处理
- 如果某个数据源缺失，该小时的效率评分会基于可用数据计算
- `data_completeness` 字段标记数据完整性 (0-1)
- 建议数据完整性 ≥ 0.6 时才使用效率评分

## 算法说明

### 效率评分计算
```
efficiency_score = (
    typing_speed_normalized * 0.30 +
    work_app_ratio_normalized * 0.20 +
    hrv_normalized * 0.20 +
    focus_normalized * 0.15 +
    activity_normalized * 0.15
) / total_weight * 100
```

### 归一化方法
使用Z-score归一化，然后映射到0-100范围：
```
normalized = (value - mean) / std_dev
normalized = max(0, min(100, normalized * 15 + 50))
```

### 峰值时段检测
1. 计算平均效率
2. 设置阈值为平均值的110%
3. 查找连续超过阈值的最长时段

### 生物钟分类
- 计算早晨(6-10am)、中午(10-18点)、晚上(20-24点)的平均效率
- 比较三个时段，差异>10分则分类为对应类型
- 置信度 = 差异 / 最高效率

## 性能考虑

### 数据量估算
- 每天24小时 × 365天 = 8,760条记录/年
- 每条记录约200字节
- 1年数据约1.7MB

### 分析性能
- 30天数据分析：约2-3秒
- 90天数据分析：约5-8秒
- 建议分批分析，避免一次性分析超过180天

### 缓存策略
- 效率曲线画像缓存在 `efficiency_curve_profiles` 表
- 无需每次重新计算，只需增量更新

## 前端集成

前端代码将放在 `applehealth/` 仓库中。

### 推荐可视化组件
1. **24小时雷达图**：展示效率曲线
2. **折线图**：工作日vs周末对比
3. **热力图**：日期×小时效率热力图
4. **卡片组件**：峰值时段、生物钟类型、洞察建议

### TypeScript接口定义
```typescript
interface HourlyEfficiencyScore {
  date: string;
  hour: number;
  efficiency_score: number;
  data_completeness: number;
  // ... 其他字段
}

interface EfficiencyCurveProfile {
  profile_type: 'workday' | 'weekend';
  hourly_curve: number[]; // 24个元素
  peak_hour: number;
  chronotype: 'morning' | 'evening' | 'intermediate';
  // ... 其他字段
}
```

## 未来扩展

### 待实现功能
1. **数据获取方法**：
   - `fetchKeyboardData()`: 从keyboard数据库获取小时级打字数据
   - `fetchScreenTimeData()`: 从screentime数据库获取小时级应用使用数据
   - `fetchHealthData()`: 从health数据库获取小时级健康数据

2. **洞察生成**：
   - 自动生成个性化建议
   - 异常检测（效率突然下降）
   - 趋势预测

3. **高级分析**：
   - 季节性效率变化
   - 效率与天气关联
   - 效率与睡眠质量关联

## 故障排查

### 问题1: 分析失败
**症状**: POST /analyze 返回500错误

**可能原因**:
- 数据库连接失败
- 数据源缺失
- 日期格式错误

**解决方法**:
```bash
# 检查数据库文件是否存在
ls -la data/applehealth/health.db
ls -la data/keyboard/kmcounter.db
ls -la data/screentime/screentime.db

# 检查日志
tail -f logs/server.log
```

### 问题2: 效率评分为0
**症状**: 所有小时的efficiency_score都是0

**可能原因**:
- 数据获取方法未实现（返回nil）
- 数据库中无对应时段的数据

**解决方法**:
- 检查 `data_completeness` 字段
- 实现 `fetchKeyboardData`, `fetchScreenTimeData`, `fetchHealthData` 方法

## 更新日志

### 2026-03-02
- ✅ 创建数据库迁移（3个表）
- ✅ 实现Repository层（数据访问）
- ✅ 实现Service层（业务逻辑）
- ✅ 实现Handler层（API端点）
- ✅ 集成到主服务器路由
- ⏳ 待实现：数据获取方法
- ⏳ 待实现：前端可视化组件
