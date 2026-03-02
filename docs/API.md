# Records Backend API 文档

## 概述

Records Backend是个人数据分析平台的后端服务，提供多模块数据分析API。

**Base URL**: `http://localhost:3000/api/v1`

**技术栈**: Go + Gin + SQLite

**支持模块**:
- ScreenTime (屏幕使用时间)
- Health (健康数据)
- Keyboard (键盘使用)
- Flights (航班数据)
- Railway (铁路数据)
- Tracks (GPS轨迹)
- Cross-Module (跨模块分析)

---

## 模块概览

### 1. ScreenTime 模块

**路径前缀**: `/screentime`

**功能**: 手机和电脑屏幕使用时间分析

**API数量**: 24个端点

**详细文档**: 见 `screentime/API.md`

**核心功能**:
- 基础统计 (7个端点)
- 跨设备分析 (7个端点)
- 高级分析 (10个端点)
  - 时间浪费检测
  - 应用依赖度分析
  - 专注力深度分析
  - 专注力对比分析

---

### 2. Health 模块

**路径前缀**: `/health`

**功能**: Apple Health数据分析

**数据来源**: Apple Health导出数据 (710,000条记录)

#### 基础API

##### 获取健康摘要
```
GET /health/summary
```

**响应示例**:
```json
{
  "totalRecords": 710000,
  "totalWorkouts": 156,
  "totalSleepRecords": 1250,
  "dateRangeStart": "2023-01-01",
  "dateRangeEnd": "2026-02-19",
  "availableMetrics": ["HeartRate", "Steps", "Distance", "Sleep"]
}
```

##### 获取健康记录
```
GET /health/records?type=HeartRate&start=2026-01-01&end=2026-02-19&limit=100
```

**查询参数**:
- `type`: 记录类型 (HeartRate/Steps/Distance/Sleep等)
- `start`: 开始日期 (YYYY-MM-DD)
- `end`: 结束日期 (YYYY-MM-DD)
- `limit`: 返回记录数

##### 获取运动记录
```
GET /health/workouts?start=2026-01-01&end=2026-02-19
```

#### 统计API

##### 每日统计
```
GET /health/statistics/daily?metric=steps&start=2026-01-01&end=2026-02-19
```

##### 每周统计
```
GET /health/statistics/weekly?metric=steps&start=2026-01-01&end=2026-02-19
```

##### 每月统计
```
GET /health/statistics/monthly?metric=steps&start=2026-01-01&end=2026-02-19
```

#### 高级分析API

##### 心率分析
```
GET /health/analysis/heartrate/zones
GET /health/analysis/heartrate/anomalies
GET /health/analysis/heartrate/resting
```

**心率区间分析响应示例**:
```json
{
  "zones": [
    {
      "zone": "resting",
      "minBpm": 40,
      "maxBpm": 60,
      "count": 125000,
      "percentage": 35.2
    },
    {
      "zone": "light",
      "minBpm": 60,
      "maxBpm": 100,
      "count": 180000,
      "percentage": 50.7
    }
  ],
  "avgHeartRate": 72.5,
  "restingHeartRate": 58.3
}
```

##### 运动数据分析
```
GET /health/analysis/exercise?start=2026-01-01&end=2026-02-19
```

**响应示例**:
```json
{
  "summary": {
    "totalSteps": 450000,
    "totalDistance": 320.5,
    "totalCalories": 15600,
    "totalWorkouts": 45,
    "avgDailySteps": 9000,
    "activeDays": 50,
    "longestWorkout": 75.5
  },
  "dailyStats": [...],
  "workoutTypes": [...],
  "calorieTrend": [...],
  "intensityAnalysis": {...},
  "achievements": [...],
  "recommendations": [...]
}
```

##### 睡眠质量分析
```
GET /health/analysis/sleep?start=2026-01-01&end=2026-02-19
```

**响应示例**:
```json
{
  "summary": {
    "avgSleepDuration": 7.5,
    "avgDeepSleep": 2.3,
    "avgLightSleep": 4.2,
    "avgREM": 1.0,
    "sleepQualityScore": 78.5
  },
  "dailyStats": [...],
  "sleepPatterns": {...},
  "recommendations": [...]
}
```

##### 体重BMI分析
```
GET /health/analysis/weight-bmi?start=2026-01-01&end=2026-02-19
```

##### 季节性趋势
```
GET /health/analysis/seasonal-trends?metric=steps
```

##### 健康排行榜
```
GET /health/analysis/rankings
```

**响应示例**:
```json
{
  "steps": {
    "topDays": [
      {
        "date": "2026-02-15",
        "value": 18500,
        "rank": 1
      }
    ],
    "personalBest": 18500,
    "avgValue": 9200
  },
  "heartRate": {...},
  "sleep": {...}
}
```

##### 活动模式分析
```
GET /health/analysis/patterns/daily
GET /health/analysis/patterns/weekly
```

#### 跨模块分析

##### 健康×屏幕时间关联
```
GET /health/analysis/health-screentime-correlation
```

**响应示例**:
```json
{
  "correlation": {
    "stepsVsScreentime": -0.45,
    "sleepVsScreentime": -0.62,
    "heartRateVsScreentime": 0.38
  },
  "insights": [
    "屏幕时间越长，步数越少（负相关）",
    "深夜屏幕使用显著影响睡眠质量"
  ],
  "recommendations": [...]
}
```

---

### 3. Keyboard 模块

**路径前缀**: `/keyboard`

**功能**: 键盘和鼠标使用分析

**数据来源**: KMCounter (988天数据)

#### API端点

##### 获取摘要
```
GET /keyboard/summary
```

##### 时间维度统计
```
GET /keyboard/temporal/hourly
GET /keyboard/temporal/daily
GET /keyboard/temporal/weekly
GET /keyboard/temporal/monthly
```

##### 按键分类统计
```
GET /keyboard/category/distribution
GET /keyboard/category/top-keys?limit=20
```

##### 打字行为分析
```
GET /keyboard/typing/backspace-rate
GET /keyboard/typing/delete-rate
GET /keyboard/typing/letter-frequency
```

##### 生产力分析
```
GET /keyboard/productivity/active-days
GET /keyboard/productivity/consistency-score
GET /keyboard/productivity/intensity
```

---

### 4. Flights 模块

**路径前缀**: `/flights`

**功能**: 航班数据分析

#### API端点

##### 航班列表
```
GET /flights?start=2026-01-01&end=2026-02-19
```

##### 航班详情
```
GET /flights/:id
GET /flights/:id/route
```

##### 航班搜索
```
GET /flights/search?query=CA1332
```

##### 统计分析
```
GET /flights/summary
GET /flights/airlines
GET /flights/date-range
GET /flights/statistics/airlines
GET /flights/statistics/enhanced
```

##### 旅行足迹
```
GET /flights/travel-footprint
```

**响应示例**:
```json
{
  "summary": {
    "totalFlights": 45,
    "totalDistance": 125000,
    "citiesVisited": 28,
    "countriesVisited": 5,
    "mostVisitedCity": "北京",
    "farthestFlight": "北京-洛杉矶"
  },
  "visitedCities": [...],
  "visitedCountries": [...],
  "flightRoutes": [...],
  "statistics": {...}
}
```

##### 碳排放计算
```
GET /flights/carbon-emission?start=2026-01-01&end=2026-02-19
```

**响应示例**:
```json
{
  "totalEmission": 5600,
  "avgEmissionPerFlight": 124.4,
  "emissionByAirline": {...},
  "emissionTrend": [...],
  "recommendations": [
    "考虑选择更环保的交通方式",
    "优先选择直飞航班减少碳排放"
  ]
}
```

---

### 5. Railway 模块

**路径前缀**: `/railway`

**功能**: 铁路线路和行程分析

#### API端点

##### 线路列表
```
GET /railway/lines
```

##### 线路详情
```
GET /railway/lines/:id
GET /railway/lines/:id/route
```

##### 行程记录
```
GET /railway/trips?start=2026-01-01&end=2026-02-19
```

##### 统计分析
```
GET /railway/summary
GET /railway/statistics
```

##### KML上传
```
POST /railway/upload-kml
```

---

### 6. Tracks 模块

**路径前缀**: `/tracks`

**功能**: GPS轨迹分析

**数据量**: 408,184个GPS点

#### 基础API

##### 轨迹点查询
```
GET /tracks/points?start=2026-01-01&end=2026-02-19&limit=1000
```

##### 轨迹段查询
```
GET /tracks/segments?start=2026-01-01&end=2026-02-19
```

##### 停留点查询
```
GET /tracks/stays?start=2026-01-01&end=2026-02-19&type=SPATIAL
```

**查询参数**:
- `type`: 停留类型 (SPATIAL/ADMIN_AREA)
- `minDuration`: 最小停留时长(秒)

##### 行程查询
```
GET /tracks/trips?start=2026-01-01&end=2026-02-19
```

#### 统计API

##### 足迹统计
```
GET /tracks/statistics/footprint?start=2026-01-01&end=2026-02-19
```

**响应示例**:
```json
{
  "totalPoints": 408184,
  "totalDistance": 125600,
  "provinces": 15,
  "cities": 48,
  "counties": 125,
  "mostVisitedCity": "北京",
  "farthestDistance": 2500,
  "activeDays": 988
}
```

##### 停留统计
```
GET /tracks/statistics/stays?type=SPATIAL
```

##### 极值事件
```
GET /tracks/statistics/extreme-events
```

##### 行政区划穿越
```
GET /tracks/statistics/admin-crossings
```

#### 高级分析

##### 网格系统分析
```
GET /tracks/analysis/grid-system?resolution=1000
```

##### 道路重叠分析
```
GET /tracks/analysis/road-overlap
```

##### 密度结构分析
```
GET /tracks/analysis/density-structure
```

##### 速度空间耦合
```
GET /tracks/analysis/speed-space-coupling
```

##### 重访模式
```
GET /tracks/analysis/revisit-patterns
```

##### 空间复杂度
```
GET /tracks/analysis/spatial-complexity
```

##### 方向偏好
```
GET /tracks/analysis/directional-bias
```

##### 空间画像
```
GET /tracks/analysis/spatial-persona
```

---

### 7. Cross-Module 跨模块分析

**路径前缀**: `/cross-module`

**功能**: 跨模块数据关联分析

#### 个人效率曲线
```
GET /cross-module/efficiency-curve/hourly?start=2026-01-01&end=2026-02-19
```

**响应示例**:
```json
{
  "hourlyCurve": [
    {
      "hour": 9,
      "efficiencyScore": 82.5,
      "typingSpeed": 75.5,
      "workAppRatio": 0.85,
      "heartRateVariability": 65.2,
      "focusSessionCount": 3
    }
  ],
  "profile": {
    "chronotype": "morning",
    "peakHours": [9, 10, 11],
    "lowHours": [14, 15]
  },
  "insights": [...]
}
```

#### 地理位置行为关联
```
GET /cross-module/location-behavior/locations
GET /cross-module/location-behavior/locations/:id
GET /cross-module/location-behavior/rankings
GET /cross-module/location-behavior/heatmap
GET /cross-module/location-behavior/habits
```

**地点效率排名响应示例**:
```json
{
  "topLocations": [
    {
      "locationId": 1,
      "geohash": "wx4g0e",
      "centerLat": 39.9042,
      "centerLon": 116.4074,
      "label": "OFFICE",
      "labelConfidence": 0.95,
      "efficiencyScore": 85.5,
      "visitCount": 125,
      "avgProductivity": 82.3,
      "avgHealth": 75.2,
      "avgFocus": 88.6
    }
  ],
  "bottomLocations": [...]
}
```

##### 触发分析
```
POST /cross-module/location-behavior/analyze
```

##### 更新地点标注
```
PATCH /cross-module/location-behavior/locations/:id
```

---

## 通用响应格式

### 成功响应
```json
{
  "data": {...},
  "message": "success"
}
```

### 错误响应
```json
{
  "error": "错误描述信息"
}
```

### HTTP状态码
- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误
- `503 Service Unavailable`: 服务不可用

---

## 数据库配置

### 数据库路径

```
data/
├── screentime/
│   └── screentime.db (103MB)
├── health/
│   └── health.db (250MB)
├── keyboard/
│   └── kmcounter.db (15MB)
├── flights/
│   └── flights.db (5MB)
├── railway/
│   └── railway.db (8MB)
└── tracks/
    └── tracks.db (180MB)
```

### 数据库模式

所有数据库使用SQLite，启用WAL模式以提高并发性能。

---

## 性能优化

### 缓存策略
- 统计数据缓存: 1小时
- 分析结果缓存: 30分钟
- 实时数据: 无缓存

### 分页
大多数列表API支持分页:
- `limit`: 每页记录数 (默认100, 最大1000)
- `offset`: 偏移量 (默认0)

### 批量查询
建议使用日期范围限制查询，避免一次性查询过多数据。

---

## 认证与授权

**当前版本**: 无认证 (开发环境)

**计划实现**: JWT认证 + 管理员权限控制

---

## 部署说明

### 开发环境
```bash
cd go-backend
go run cmd/server/main.go
```

服务将在 `http://localhost:3000` 启动

### 生产构建
```bash
go build -o bin/records-backend cmd/server/main.go
./bin/records-backend
```

### 配置文件
```yaml
# config.yaml
server:
  port: 3000
  host: 0.0.0.0

database:
  screentime: data/screentime/screentime.db
  health: data/health/health.db
  keyboard: data/keyboard/kmcounter.db
  flights: data/flights/flights.db
  railway: data/railway/railway.db
  tracks: data/tracks/tracks.db

cors:
  allowOrigins:
    - http://localhost:5173
    - https://record.yzup.top
```

---

## 更新日志

### 2026-03-02
- 新增ScreenTime专注力深度分析API
- 新增ScreenTime专注力对比分析API
- 优化Health模块运动数据分析
- 改进跨模块效率曲线算法

### 2026-03-01
- 新增ScreenTime深夜使用分析API
- 新增ScreenTime应用相关性分析API
- 新增Health×ScreenTime关联分析
- 优化数据库查询性能

### 2026-02-28
- 新增5个ScreenTime高级分析API
- 新增Health健康排行榜API
- 新增Tracks空间画像分析
- 改进错误处理机制

### 2026-02-24
- 完成Keyboard模块11个API端点
- 完成ScreenTime基础功能17个端点
- 实现跨设备分析功能
- 添加CORS支持

---

## 技术支持

**问题反馈**: https://github.com/jengzang/records-backend-go/issues

**文档更新**: 2026-03-02
