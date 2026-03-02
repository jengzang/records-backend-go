# 健康×屏幕时间关联分析 - 实施报告

## 实施日期
2026-03-02

## 功能概述
实现了跨模块分析功能，关联AppleHealth健康数据和ScreenTime屏幕使用数据，分析久坐行为、活动相关性和睡眠影响。

## 实施内容

### 1. 后端实现

#### 1.1 核心分析逻辑
**文件**: `internal/health/analysis/health_screentime_correlation.go`

**主要功能**:
- 久坐行为检测 (Sedentary Analysis)
- 活动与屏幕时间相关性分析 (Activity Correlation)
- 深夜屏幕使用对睡眠的影响 (Sleep Impact)
- 健康平衡评分 (Balance Score)

**关键算法**:

1. **久坐检测标准**:
   - 屏幕时间 > 2小时 AND 步数 < 5000
   - 风险等级:
     - 高风险: 屏幕时间 > 6小时
     - 中风险: 屏幕时间 > 4小时
     - 低风险: 其他

2. **活动相关性**:
   - 使用Pearson相关系数计算步数与屏幕时间的关系
   - 相关类型:
     - 强负相关: r < -0.5
     - 中等负相关: -0.5 ≤ r < -0.3
     - 弱负相关: -0.3 ≤ r < 0
     - 无相关: r ≈ 0
     - 正相关: r > 0

3. **睡眠影响分析**:
   - 深夜屏幕使用定义: 22:00-02:00 使用 > 30分钟
   - 对比深夜使用日与正常日的平均睡眠时长
   - 影响等级:
     - 严重影响: 差异 > 1小时
     - 中等影响: 差异 > 30分钟
     - 轻微影响: 差异 > 0
     - 无影响: 差异 ≤ 0

4. **健康平衡评分** (0-100):
   - 基础分: 100
   - 久坐率扣分: -久坐率 × 30
   - 相关性扣分: -|相关系数| × 20
   - 睡眠影响扣分: -(睡眠差异/60) × 10

#### 1.2 Service层修改
**文件**: `internal/health/service.go`

**修改内容**:
- 添加 `screentimeDB *sql.DB` 字段
- 添加 `SetScreentimeDB(db *sql.DB)` 方法
- 添加 `GetHealthScreentimeCorrelation()` 方法

#### 1.3 Handler层修改
**文件**: `internal/health/handlers.go`

**修改内容**:
- 添加 `SetScreentimeDB(db *sql.DB)` 方法
- 添加 `GetHealthScreentimeCorrelation(c *gin.Context)` handler

#### 1.4 路由注册
**文件**: `internal/api/router.go`

**修改内容**:
- 在health模块初始化时打开screentimeDB
- 调用 `healthHandler.SetScreentimeDB(screentimeDB)`
- 注册路由: `GET /api/v1/health/analysis/health-screentime-correlation`

### 2. 前端实现

#### 2.1 类型定义
**文件**: `applehealth/src/types/health.ts`

**新增类型**:
- `HealthScreentimeCorrelation` - 主响应类型
- `SedentaryAnalysis` - 久坐分析
- `SedentaryDayDetail` - 久坐日详情
- `ActivityCorrelation` - 活动相关性
- `ActivityDataPoint` - 活动数据点
- `SleepImpact` - 睡眠影响

#### 2.2 API服务
**文件**: `applehealth/src/services/healthApiService.ts`

**新增方法**:
```typescript
async getHealthScreentimeCorrelation(): Promise<HealthScreentimeCorrelation>
```

#### 2.3 前端页面
**文件**: `applehealth/src/pages/HealthScreentimeCorrelation.tsx`

**页面结构**:
1. **健康平衡评分** - 圆形进度条显示
2. **久坐行为分析** - 4个统计卡片
   - 久坐天数/总天数
   - 久坐率
   - 久坐日平均屏幕时间
   - 久坐日平均步数
3. **活动相关性** - 相关系数和散点图
4. **睡眠影响** - 深夜使用对睡眠的影响
5. **健康建议** - 个性化建议列表
6. **详细数据** - 两个Tab
   - 久坐日详情表格
   - 活动散点图

**可视化组件**:
- Ant Design: Card, Statistic, Progress, Table, Tabs, Tag
- Recharts: ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip

#### 2.4 路由集成
**文件**: `applehealth/src/App.tsx`

**修改内容**:
- 导入 `HealthScreentimeCorrelation` 组件
- 添加 `InteractionOutlined` 图标
- 添加菜单项: "健康×屏幕时间"
- 添加路由: `/health-screentime-correlation`

## 数据库依赖

### AppleHealth数据库
- `health_statistics` 表 (stat_type='daily')
  - 步数 (StepCount)
  - 睡眠时长 (SleepAnalysis)

### ScreenTime数据库
- `phone_daily_usage` 表
  - 每日总屏幕时间
- `phone_hourly_usage` 表
  - 小时级屏幕使用数据 (用于深夜使用检测)

## API端点

### GET /api/v1/health/analysis/health-screentime-correlation

**响应示例**:
```json
{
  "balanceScore": 72.5,
  "sedentaryAnalysis": {
    "sedentaryDays": 45,
    "totalDays": 180,
    "sedentaryRate": 0.25,
    "avgScreenTimeOnSedentaryDays": 8.5,
    "avgStepsOnSedentaryDays": 3200,
    "highRiskDays": 15,
    "mediumRiskDays": 20,
    "lowRiskDays": 10,
    "sedentaryDayDetails": [...]
  },
  "activityCorrelation": {
    "correlationCoefficient": -0.42,
    "correlationType": "中等负相关",
    "avgSteps": 7500,
    "avgScreenTime": 5.2,
    "dataPoints": [...]
  },
  "sleepImpact": {
    "avgSleepWithLateScreenTime": 6.5,
    "avgSleepWithoutLateScreenTime": 7.2,
    "sleepDifference": 0.7,
    "lateScreenTimeDays": 30,
    "normalDays": 150,
    "impactLevel": "中等影响"
  },
  "recommendations": [
    "您的久坐率为25.0%，建议增加日常活动量",
    "步数与屏幕时间呈中等负相关，活动越多屏幕时间越少，这是健康的模式",
    "深夜屏幕使用对睡眠有中等影响，建议减少22:00后的屏幕使用"
  ]
}
```

## 测试建议

1. **后端测试**:
   ```bash
   cd go-backend
   go build -o bin/server.exe cmd/server/main.go
   ./bin/server.exe
   curl http://localhost:9000/api/v1/health/analysis/health-screentime-correlation
   ```

2. **前端测试**:
   ```bash
   cd applehealth
   npm install
   npm run dev
   # 访问 http://localhost:5173/health-screentime-correlation
   ```

3. **数据验证**:
   - 检查久坐检测逻辑是否正确
   - 验证相关系数计算
   - 确认睡眠影响分析准确性
   - 测试边界情况 (无数据、单日数据等)

## 已知限制

1. **数据要求**:
   - 需要同时存在健康数据和屏幕时间数据
   - 至少需要7天数据才能进行有意义的分析

2. **性能考虑**:
   - 大数据量时可能需要优化查询
   - 考虑添加缓存机制

3. **功能扩展**:
   - 可添加时间范围筛选
   - 可添加周/月趋势分析
   - 可添加目标设定和追踪

## 下一步计划

1. 测试完整功能流程
2. 优化前端UI和交互
3. 添加数据缓存
4. 实现时间范围筛选
5. 提交代码到仓库

## 文件清单

### 后端文件
- `internal/health/analysis/health_screentime_correlation.go` (NEW)
- `internal/health/service.go` (MODIFIED)
- `internal/health/handlers.go` (MODIFIED)
- `internal/api/router.go` (MODIFIED)

### 前端文件
- `applehealth/src/types/health.ts` (MODIFIED)
- `applehealth/src/services/healthApiService.ts` (MODIFIED)
- `applehealth/src/pages/HealthScreentimeCorrelation.tsx` (NEW)
- `applehealth/src/App.tsx` (MODIFIED)

### 文档文件
- `go-backend/docs/health/health-screentime-correlation-implementation.md` (NEW)
