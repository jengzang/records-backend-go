# Records Backend (Go)

个人数据分析平台后端服务

## 项目简介

本项目是个人数据分析平台的后端服务，使用 Go 语言和 Gin 框架开发，提供 RESTful API 接口。支持多种个人数据的采集、存储和分析，包括GPS轨迹、键盘鼠标使用、屏幕时间、旅行记录和健康数据。

## 技术栈

- Go 1.21+
- Gin Web Framework
- SQLite3 (WAL 模式)
- Python 3.11+ (数据处理 Workers)
- RESTful API 架构

## 项目结构

```
go-backend/
├── cmd/
│   └── server/          # 服务器入口
│       └── main.go
├── internal/
│   ├── api/            # API 路由和处理器
│   │   └── router.go   # 主路由配置
│   ├── config/         # 配置管理
│   ├── database/       # 数据库操作
│   ├── logger/         # 日志系统
│   ├── tracks/         # 轨迹分析模块
│   ├── keyboard/       # 键盘鼠标统计模块
│   ├── flights/        # 飞机火车路线模块
│   ├── railway/        # 铁路线路模块
│   ├── screentime/     # 屏幕使用时间模块
│   ├── health/         # Apple健康数据模块
│   └── analysis/       # 轨迹分析引擎
│       ├── foundation/ # 基础分析
│       ├── behavior/   # 行为分析
│       ├── spatial/    # 空间分析
│       ├── statistical/# 统计分析
│       ├── advanced/   # 高级分析
│       └── python/     # Python Worker 集成
├── scripts/            # Python 辅助脚本
│   └── tracks/
│       ├── workers/    # 分析 Workers
│       └── migrations/ # 数据库迁移
├── data/               # 数据文件和数据库
├── docs/               # 项目文档
├── main.go            # 简单启动入口
├── go.mod
└── README.md
```

## 核心功能模块

### 1. 轨迹分析 (Tracks) ✅ 100%

**数据规模**: 408,184 个 GPS 点，988 天历史数据

**核心功能**:
- GPS 轨迹数据管理和可视化
- 地理编码 (5级行政区划)
- 异常点检测和轨迹补全
- 交通方式识别 (步行/骑行/汽车/火车/飞机)
- 停留检测 (SPATIAL + ADMIN_AREA 双模式)
- 行程构建和目的推断
- 空间分析 (网格系统、密度结构、道路重叠)
- 统计聚合 (足迹统计、极值事件、行政区划穿越)
- 空间行为画像引擎

**30个分析技能 (100% 完成)**:
- **数据基础与质量控制 (4个)**: 数据导入、异常检测、轨迹补全、行政区划归属
- **行为与分段 (5个)**: 交通方式识别、停留检测、行程构建、连续活动检测、速度事件
- **空间分析 (8个)**: 网格系统、道路重叠、密度结构、速度-空间耦合、重访模式、利用效率、空间复杂度、方向偏好
- **统计聚合 (5个)**: 足迹统计、停留统计、极值事件、行政区划穿越、行政区划视图引擎
- **高级分析 (3个)**: 时空切片、时空压缩、海拔维度
- **可视化与渲染 (3个)**: 渲染元数据、时间轴地图、停留标注
- **集成与画像 (2个)**: 空间行为画像、高级行政区划分析

**API 端点**: 50+ 个

### 2. 键盘鼠标统计 (Keyboard) ✅ 100%

**数据规模**: 988 天使用数据

**核心功能**:
- 键盘鼠标使用数据统计
- 按键分类统计 (字母/数字/功能键/修饰键)
- 时间维度分析 (24小时热力图、星期统计、月份统计)
- 打字行为分析 (退格率、删除率、字母频率)
- 生产力指标 (活跃天数、连续天数、一致性评分)
- 左右手使用平衡分析 ✨
- 工作日vs周末对比分析 ✨
- 键盘×屏幕时间跨模块关联分析 ✨

**API 端点**: 11 个

### 3. 飞机火车路线 (TrainPlane) ✅ 100%

**数据规模**: 航班数据 + 铁路线路数据

**核心功能**:
- 航班路线管理和可视化
- 铁路线路管理 (50+ 条线路)
- 航班详情和轨迹点查询
- 航司统计和排行
- 旅行足迹地图 ✨
  - 全球地图可视化
  - 访问城市/国家统计
  - 航班轨迹动画
- 旅行统计增强 ✨
  - 里程排行榜 (年度/月度)
  - 记录航班 (最远/最短/最频繁)
  - 旅行趋势分析
  - 成就检测系统
- 碳排放计算 ✨
  - 航班/铁路碳排放估算
  - 年度碳排放统计
  - 碳足迹报告
  - 碳中和建议

**API 端点**: 18 个

### 4. 屏幕使用时间 (ScreenTime) ✅ 100%

**数据规模**: 22,013 条手机记录 + 103MB 电脑数据

**核心功能**:
- 手机屏幕使用时间统计
- 电脑屏幕使用时间统计
- 应用使用排名和分类
- 每小时使用分布
- 趋势分析
- 跨设备分析 (手机+电脑)
- 工作生活平衡评分
- 应用生态系统分析
- 用户画像生成
- 时间浪费检测 ✨
  - 识别低价值应用使用
  - 工作时段娱乐应用检测
  - 碎片化使用检测
- 应用依赖度分析 ✨
  - 使用频率和时长占比
  - 连续使用天数 (streak)
  - 依赖度评分 (0-100)
- 工作日vs周末对比 ✨
  - 使用时长和应用分布对比
  - 设备偏好对比
- 生产力/娱乐比例趋势 ✨
  - 周/月/季度趋势分析
  - 堆叠面积图可视化
- 应用切换模式识别 ✨
  - 多任务/专注/碎片化模式识别
  - 切换路径分析
  - 应用切换矩阵
- 深夜使用习惯分析 ✨
  - 22:00-02:00时段使用分析
  - 睡眠影响评估(影响评分、睡眠损失估算)
  - 工作日vs周末对比
  - 个性化改善建议
- 应用使用相关性分析 ✨
  - 应用共现分析(30分钟时间窗口)
  - 关联规则挖掘(置信度、提升度)
  - 频繁模式识别
  - 网络图可视化
- 应用使用热力图 ✨
  - 日期×小时热力图
  - 使用强度可视化
- **专注力深度分析** ✨ NEW
  - 专注会话识别(≥10分钟)
  - 会话质量分类(深度/中度/浅度)
  - 干扰源分析(<5分钟会话)
  - 最佳专注时段识别(24小时分析)
  - 专注力趋势追踪
  - 总体专注分数(0-100)
  - 个性化专注建议
- **专注力对比分析** ✨ NEW
  - 工作日vs周末专注模式对比
  - 工作时段vs休闲时段对比
  - 月度专注力趋势分析
  - 自动洞察生成
  - 一致性评分
  - 最佳时段/月份识别

**API 端点**: 19 个

### 5. Apple健康数据 (Health) ✅ 100%

**数据规模**: 710,000 条记录 (心率 709,990 条)

**核心功能**:
- 心率数据分析
  - 心率区间分布
  - 异常检测
  - 静息心率趋势
  - 每日/每周活动模式
  - 健康评分
- 体重/BMI管理 ✨
  - 体重趋势图
  - BMI 计算和健康范围
  - 体重预测 (线性回归)
- 运动数据分析 ✨
  - 步数/距离/爬楼统计
  - 运动类型分布
  - 卡路里消耗趋势
  - 运动强度分析 (METs)
- 睡眠质量分析 ✨
  - 睡眠时长趋势
  - 睡眠质量评分
  - 睡眠模式分析

**API 端点**: 15 个

### 6. 跨模块分析 (Cross-Module) ✅ NEW

**个人效率曲线 (Personal Efficiency Curve)**

**核心功能**:
- 跨模块数据整合 (键盘+屏幕时间+健康数据)
- 24小时效率评分曲线
  - 综合评分算法: 打字速度(30%) + 工作应用比例(20%) + HRV(20%) + 专注度(15%) + 活动量(15%)
  - 小时级效率评分 (0-100)
  - 数据完整性追踪
- 效率曲线画像
  - 工作日/周末聚合曲线
  - 峰值时段自动检测
  - 生物钟类型分类 (早晨型/夜猫子型/中间型)
- 工作日vs周末对比
  - 平均效率差异分析
  - 24小时逐时对比
  - 自动生成解读建议
- 可操作洞察
  - 峰值生产力时段建议
  - 生物钟类型建议
  - 个性化改进建议

**技术特点**:
- 智能数据分配 (键盘每日数据分配到活跃时段)
- 精确时间重叠计算 (屏幕时间会话毫秒级)
- 分段线性归一化 (支持最优值设定)
- 跨数据库查询 (同时访问3个SQLite数据库)
- 数据缺失容错 (data_completeness标记)

**API 端点**: 5 个
- GET /api/v1/cross-module/efficiency-curve/hourly - 获取小时效率曲线
- GET /api/v1/cross-module/efficiency-curve/profile - 获取效率画像
- GET /api/v1/cross-module/efficiency-curve/comparison - 工作日vs周末对比
- GET /api/v1/cross-module/efficiency-curve/insights - 获取效率洞察
- POST /api/v1/cross-module/efficiency-curve/analyze - 触发效率分析

**数据库表**: 3 个
- hourly_efficiency_scores - 小时效率评分
- efficiency_curve_profiles - 效率曲线画像
- efficiency_insights - 效率洞察

## 运行方式

### 前置要求

- Go 1.21 或更高版本
- Python 3.11+ (用于轨迹分析 Workers)
- SQLite3

### 开发环境

```bash
# 克隆项目
git clone https://github.com/jengzang/records-backend-go.git
cd records-backend-go

# 安装 Go 依赖
go mod download

# 安装 Python 依赖 (用于轨迹分析)
pip install -r scripts/tracks/requirements.txt

# 运行服务器 (简单方式)
go run main.go

# 运行服务器 (标准方式)
go run cmd/server/main.go

# 服务器将在 http://localhost:8080 启动
```

### 生产构建

```bash
# 构建二进制文件
go build -o bin/server.exe cmd/server/main.go

# 运行
./bin/server.exe

# 或者使用 Windows
bin\server.exe
```

### Docker 部署 (可选)

```bash
# 构建镜像
docker build -t records-backend .

# 运行容器
docker run -p 8080:8080 -v $(pwd)/data:/app/data records-backend
```

## 配置说明

### 环境变量

```bash
PORT=:8080                              # 服务器端口
TRACKS_DB_PATH=./data/tracks.db         # 轨迹数据库路径
KEYBOARD_DB_PATH=./data/kmcounter.db    # 键盘数据库路径
FLIGHTS_DB_PATH=./data/flights.db       # 航班数据库路径
RAILWAY_DB_PATH=./data/railway.db       # 铁路数据库路径
SCREENTIME_DB_PATH=./data/screentime.db # 屏幕时间数据库路径
HEALTH_DB_PATH=./data/health.db         # 健康数据库路径
```

### 数据库配置

所有数据库使用 SQLite3 with WAL 模式:
- 自动启用 WAL 模式
- 连接池配置: MaxOpenConns=10, MaxIdleConns=5
- 支持并发读取
- 定期 checkpoint

## API 接口文档

### 健康检查

- `GET /health` - 服务健康检查

### 轨迹分析接口 (Tracks)

**数据查询**:
- `GET /api/v1/tracks/points` - 获取轨迹点列表
- `GET /api/v1/tracks/segments` - 获取轨迹段列表
- `GET /api/v1/tracks/stays` - 获取停留记录
- `GET /api/v1/tracks/trips` - 获取行程记录

**统计分析**:
- `GET /api/v1/tracks/statistics/footprint` - 足迹统计
- `GET /api/v1/tracks/statistics/stays` - 停留统计
- `GET /api/v1/tracks/statistics/extreme-events` - 极值事件
- `GET /api/v1/tracks/statistics/admin-crossings` - 行政区划穿越

**可视化**:
- `GET /api/v1/tracks/visualization/rendering-metadata` - 渲染元数据
- `GET /api/v1/tracks/visualization/time-axis-map` - 时间轴地图

**任务管理**:
- `POST /api/v1/admin/analysis/tasks` - 创建分析任务
- `GET /api/v1/admin/analysis/tasks` - 获取任务列表
- `GET /api/v1/admin/analysis/tasks/:id` - 查询任务状态
- `DELETE /api/v1/admin/analysis/tasks/:id` - 取消任务

### 键盘鼠标接口 (Keyboard)

- `GET /api/v1/keyboard/daily` - 每日统计
- `GET /api/v1/keyboard/scancodes` - 扫描码统计
- `GET /api/v1/keyboard/top-keys` - Top 按键排行
- `GET /api/v1/keyboard/statistics/summary` - 统计摘要
- `GET /api/v1/keyboard/statistics/trends` - 趋势分析
- `GET /api/v1/keyboard/statistics/hand_balance` - 左右手平衡
- `GET /api/v1/keyboard/statistics/weekday_weekend` - 工作日周末对比
- `GET /api/v1/keyboard/cross_module` - 跨模块分析
- `GET /api/v1/keyboard/heatmap/keyboard` - 键盘热力图
- `GET /api/v1/keyboard/heatmap/detailed` - 详细热力图
- `GET /api/v1/keyboard/heatmap/time` - 时间热力图

### 飞机火车接口 (Flights/Railway)

**航班**:
- `GET /api/v1/flights` - 获取航班列表
- `GET /api/v1/flights/:id` - 获取航班详情
- `GET /api/v1/flights/:id/route` - 获取航班路线
- `GET /api/v1/flights/search` - 搜索航班
- `GET /api/v1/flights/summary` - 航班摘要
- `GET /api/v1/flights/airlines` - 航司列表
- `GET /api/v1/flights/statistics/airlines` - 航司统计
- `GET /api/v1/flights/travel-footprint` - 旅行足迹
- `GET /api/v1/flights/statistics/enhanced` - 增强统计
- `GET /api/v1/flights/carbon-emission` - 碳排放分析

**铁路**:
- `GET /api/v1/railway/lines` - 获取线路列表
- `GET /api/v1/railway/lines/:id` - 获取线路详情
- `GET /api/v1/railway/lines/:id/route` - 获取线路路线
- `GET /api/v1/railway/trips` - 获取乘车记录
- `GET /api/v1/railway/trips/:id` - 获取乘车详情
- `POST /api/v1/railway/trips` - 创建乘车记录
- `GET /api/v1/railway/statistics` - 铁路统计
- `POST /api/v1/railway/upload-kml` - 上传 KML 文件

### 屏幕使用时间接口 (ScreenTime)

**基础统计**:
- `GET /api/v1/screentime/summary` - 使用摘要
- `GET /api/v1/screentime/daily` - 每日统计
- `GET /api/v1/screentime/rankings` - 应用排名
- `GET /api/v1/screentime/categories` - 类别统计
- `GET /api/v1/screentime/hourly` - 每小时统计
- `GET /api/v1/screentime/trends` - 趋势分析
- `GET /api/v1/screentime/app/:packageId` - 应用详情

**高级分析**:
- `GET /api/v1/screentime/analysis/time-waste` - 时间浪费检测
- `GET /api/v1/screentime/analysis/app-dependency` - 应用依赖度
- `GET /api/v1/screentime/analysis/weekday-weekend` - 工作日周末对比
- `GET /api/v1/screentime/analysis/productivity-entertainment-trend` - 生产力娱乐趋势
- `GET /api/v1/screentime/analysis/switching-pattern` - 切换模式分析
- `GET /api/v1/screentime/analysis/late-night` - 深夜使用习惯分析 ✨ NEW
- `GET /api/v1/screentime/analysis/app-correlation` - 应用使用相关性分析 ✨ NEW
- `GET /api/v1/screentime/analysis/usage-heatmap` - 应用使用热力图 ✨ NEW

**跨设备分析**:
- `GET /api/v1/screentime/cross-device/comparison` - 设备对比
- `GET /api/v1/screentime/cross-device/work-life-balance` - 工作生活平衡
- `GET /api/v1/screentime/cross-device/app-ecosystem` - 应用生态
- `GET /api/v1/screentime/cross-device/user-profile` - 用户画像

### Apple健康数据接口 (Health)

**基础查询**:
- `GET /api/v1/health/summary` - 健康数据摘要
- `GET /api/v1/health/records` - 健康记录查询
- `GET /api/v1/health/workouts` - 运动记录查询

**统计分析**:
- `GET /api/v1/health/statistics/daily` - 每日统计
- `GET /api/v1/health/statistics/weekly` - 每周统计
- `GET /api/v1/health/statistics/monthly` - 每月统计
- `GET /api/v1/health/statistics/trends` - 趋势分析

**高级分析**:
- `GET /api/v1/health/analysis/exercise` - 运动数据分析 ✅
- `GET /api/v1/health/analysis/sleep` - 睡眠质量分析 ✅
- `GET /api/v1/health/analysis/rankings` - 健康排行榜 ✨ NEW
- `GET /api/v1/health/analysis/weight-bmi` - 体重BMI分析
- `GET /api/v1/health/analysis/seasonal-trends` - 季节趋势
- `GET /api/v1/health/analysis/health-screentime-correlation` - 健康×屏幕时间关联

**心率分析**:
- `GET /api/v1/health/analysis/heartrate/zones` - 心率区间分析
- `GET /api/v1/health/analysis/heartrate/anomalies` - 心率异常检测
- `GET /api/v1/health/analysis/heartrate/resting` - 静息心率分析

**活动模式**:
- `GET /api/v1/health/analysis/patterns/daily` - 每日活动模式
- `GET /api/v1/health/analysis/patterns/weekly` - 每周活动模式

## 数据库架构

### 轨迹分析数据库 (tracks.db)

**核心表**:
- `track_points` - GPS 轨迹点 (408k+ 记录)
- `segments` - 轨迹段 (交通方式分段)
- `stay_segments` - 停留记录 (SPATIAL + ADMIN_AREA)
- `trips` - 行程记录 (含目的推断)
- `grid_cells` - 网格单元 (Geohash 索引)
- `density_clusters` - 密度聚类
- `spatial_persona` - 空间行为画像
- `admin_trends` - 行政区划趋势

**分析表**:
- `speed_events` - 速度事件
- `stay_annotations` - 停留标注
- `render_segments_cache` - 渲染缓存
- `footprint_stats` - 足迹统计
- `stay_stats` - 停留统计
- `extreme_events` - 极值事件

### 其他数据库

- `kmcounter.db` - 键盘鼠标数据
- `flights.db` - 航班数据
- `railway.db` - 铁路数据
- `screentime.db` - 屏幕时间数据
- `health.db` - 健康数据

## 性能指标

### 服务器约束

- CPU: 2 核
- 内存: 2GB (Go 进程 ≤ 1GB)
- 并发: ≤ 3 req/s
- 数据库: ~500MB (当前), ~2GB (预计 5 年)

### 分析性能

- 异常检测: ~10k points/sec
- 轨迹补全: ~5k points/sec
- 交通方式识别: ~10k points/sec
- 停留检测 (DBSCAN): ~1k points/sec
- 网格系统: ~5k points/sec
- 空间行为画像: ~10 sec

## 前端项目

本后端服务对应 5 个前端项目:

1. **tracks** - GPS 轨迹可视化和分析
   - Repository: https://github.com/jengzang/records-frontend-tracks.git

2. **keyboard** - 键盘鼠标使用统计
   - Repository: https://github.com/jengzang/records-frontend-keyboard.git

3. **TrainPlane** - 飞机火车路线管理
   - Repository: https://github.com/jengzang/records-frontend-TrainAndPlane.git

4. **screentime** - 屏幕使用时间分析
   - Repository: https://github.com/jengzang/records-frontend-screentime.git

5. **applehealth** - Apple 健康数据分析
   - Repository: https://github.com/jengzang/record-frontend-AppleHealth.git

## 开发指南

### 添加新的分析技能

1. 在 `internal/analysis/` 下创建分析器
2. 实现 `Analyzer` 接口
3. 在 `init()` 中注册分析器
4. 添加数据库迁移 (如需要)
5. 在 `router.go` 中添加 API 端点
6. 更新文档

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 命名约定
- 添加必要的注释和文档
- 编写单元测试
- 使用 logrus 记录日志

### 提交规范

```
[模块/类型] 简短描述
- 详细说明1
- 详细说明2

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

## 更新日志

### 2026-03-02 (文档完善) - 完整API文档 📚

**文档更新**:
- ✅ 新增完整API文档 (`docs/API.md`)
  - 7个模块完整API说明
  - 113+ API端点详细文档
  - 请求参数和响应示例
  - 错误处理说明
  - 使用示例和最佳实践
- ✅ 更新主README
  - 添加专注力分析功能说明
  - 更新API端点数量
  - 完善功能列表

**ScreenTime模块新增功能** (2个):
1. **专注力深度分析** (`/analysis/focus-depth`)
   - 专注会话识别(≥10分钟)
   - 会话质量分类(深度/中度/浅度)
   - 干扰源分析(<5分钟会话)
   - 最佳专注时段识别(24小时分析)
   - 专注力趋势追踪
   - 总体专注分数(0-100)
   - 个性化专注建议

2. **专注力对比分析** (`/analysis/focus-comparison`)
   - 工作日vs周末专注模式对比
   - 工作时段vs休闲时段对比
   - 月度专注力趋势分析
   - 自动洞察生成
   - 一致性评分
   - 最佳时段/月份识别

**项目完成度**:
- Tracks: 100% (30/30 技能, 50+ 端点)
- Keyboard: 100% (11 个端点)
- ScreenTime: 100% (19 个端点) ⬆️
- TrainPlane: 100% (18 个端点)
- AppleHealth: 100% (15 个端点)
- Cross-Module: 100% (10 个端点)

**总计**: 123+ API 端点，完整文档覆盖

### 2026-03-02 (深夜) - AppleHealth模块功能完成 ✅

**阶段2完成：AppleHealth模块增强 (3个功能)**

**Phase 2.1: 运动数据分析** ✅ (已存在)
- 步数/距离/爬楼统计（日/周/月趋势）
- 运动类型分布分析
- 卡路里消耗趋势
- METs强度分析
- 运动成就检测
- 个性化建议生成

**Phase 2.2: 睡眠质量分析** ✅ (已存在)
- 睡眠时长趋势（深睡/浅睡/REM）
- 入睡/起床时间规律性
- 睡眠质量评分算法
- 睡眠与心率关联分析
- 睡眠效率计算
- 睡眠债务追踪

**Phase 2.3: 健康排行榜** ✨ NEW
1. **6种排行榜 (Top 20)**
   - 步数排行榜
   - 距离排行榜
   - 卡路里排行榜
   - 心率排行榜（最低静息心率）
   - 睡眠排行榜（最长睡眠）
   - 运动排行榜（最长运动时长）

2. **个人最佳记录 (7项)**
   - 最高步数、最远距离、最高卡路里
   - 最低静息心率、最长睡眠、最长运动
   - 单日最多运动次数

3. **排行榜统计**
   - 总追踪天数
   - Top 10天数占比
   - 当前连续天数（≥5000步）
   - 最长连续天数
   - 改进趋势（improving/stable/declining）

4. **前端实现**
   - Rankings.tsx 页面组件
   - Ant Design Table + Tabs 展示
   - 前三名特殊图标（金银铜牌）
   - 当月记录自动标记

**技术实现**:
- 新增文件: internal/health/analysis/rankings.go (555行)
- API端点: GET /api/v1/health/analysis/rankings
- 前端页面: src/pages/Rankings.tsx (352行)
- 连续天数算法: 基于5000步阈值
- 趋势分析: 30天滚动窗口对比

**AppleHealth模块状态**: 100% 完成 ✅

### 2026-03-02 (深夜) - ScreenTime模块功能增强 ✅

**新增功能：应用使用热力图 (Usage Heatmap)**

1. **日期×小时热力图**
   - 可视化应用使用强度分布
   - 支持自定义日期范围查询（默认最近30天）
   - 归一化强度值（0-1）用于前端渲染

2. **统计指标**
   - 最大/最小/平均使用时长
   - 总单元格数统计
   - 日期范围追踪

3. **API端点**
   - `GET /api/v1/screentime/analysis/usage-heatmap?start=20240101&end=20241231`

**ScreenTime模块功能总结**:
- ✅ 时间浪费检测 (工作时间娱乐、深夜社交、碎片化使用)
- ✅ 应用依赖度分析 (依赖评分、使用频率、连续天数)
- ✅ 周末vs工作日对比 (使用模式、类别分布、时段分布)
- ✅ 生产力/娱乐比例趋势 (周/月/季度趋势)
- ✅ 应用切换模式识别 (多任务/专注/碎片化模式)
- ✅ 深夜使用习惯分析 (睡眠影响评估)
- ✅ 应用使用相关性分析 (共现分析、关联规则)
- ✅ 应用使用热力图 (日期×小时可视化)

**技术实现**:
- 新增文件: internal/screentime/usage_heatmap.go
- 修复 efficiency 模块编译错误
- 所有8个高级分析功能已完成

### 2026-03-02 (晚上) - Phase 4.1: 个人效率曲线功能 ✅

**重大功能：跨模块数据整合与效率分析**

**新增功能：个人效率曲线 (Personal Efficiency Curve)**

1. **跨模块数据整合**
   - 整合键盘、屏幕时间、健康数据
   - 同时访问3个SQLite数据库
   - 智能数据获取和归一化

2. **24小时效率评分**
   - 综合评分算法: 打字速度(30%) + 工作应用比例(20%) + HRV(20%) + 专注度(15%) + 活动量(15%)
   - 小时级效率评分 (0-100)
   - 数据完整性追踪 (data_completeness)

3. **效率曲线画像**
   - 工作日/周末聚合曲线
   - 峰值时段自动检测
   - 生物钟类型分类 (早晨型/夜猫子型/中间型)
   - 置信度评分

4. **工作日vs周末对比**
   - 平均效率差异分析
   - 24小时逐时对比
   - 峰值时段差异
   - 自动生成解读建议

5. **可操作洞察**
   - 峰值生产力时段建议
   - 生物钟类型建议
   - 个性化改进建议
   - 优先级分级 (低/中/高)

**技术实现**:
- 新增 5 个 API 端点
- 新增 3 个数据库表
- 新增 6 个 Go 模块文件 (1,500行代码)
- 智能数据分配算法
- 分段线性归一化算法
- 峰值时段检测算法
- 生物钟分类算法

**数据库变更**:
- 新增表: hourly_efficiency_scores (小时效率评分)
- 新增表: efficiency_curve_profiles (效率曲线画像)
- 新增表: efficiency_insights (效率洞察)
- 迁移文件: scripts/applehealth/migrations/002_create_efficiency_curve_tables.sql

**前端集成**:
- 效率曲线仪表盘 (24小时雷达图)
- 工作日vs周末对比页面 (折线图+柱状图)
- TypeScript类型定义 (完整类型安全)
- API服务层封装 (665行代码)

**文档**:
- 完整功能文档 (docs/efficiency-curve/README.md)
- 实施报告 (Day 1-4)
- 最终项目报告 (FINAL_REPORT.md)

**总计**: 118+ API 端点，跨模块分析能力全面提升

### 2026-03-02 (深夜) - 跨模块关联分析与文档更新 ✨

**新增功能**:

1. **健康×屏幕时间关联分析** (`/api/v1/health/analysis/health-screentime-correlation`)
   - 久坐行为检测(屏幕时间>2h + 步数<5000)
   - 风险等级评估(高/中/低风险)
   - 活动相关性分析(Pearson相关系数)
   - 睡眠影响评估(深夜22:00-02:00屏幕使用)
   - 健康平衡评分系统(0-100分)
   - 个性化健康建议

**前端实现**:
- 新增 HealthScreentimeCorrelation.tsx 页面
- 健康平衡评分圆形进度条
- 久坐分析统计卡片
- 活动相关性散点图
- 睡眠影响对比
- 久坐日详情表格

**文档更新**:
- 更新 CLAUDE.md 反映所有模块完整状态
- 确认 ScreenTime 高级分析功能已完整实现
- 总计: 114+ API 端点

### 2026-03-02 (晚) - ScreenTime模块新增2个高级分析功能 ✨

**新增功能**:

1. **深夜使用习惯分析** (`/api/v1/screentime/analysis/late-night`)
   - 深夜时段定义: 22:00-02:00
   - 总体模式分析(总时长、平均时长、活跃天数、频率)
   - 频率分析(每周/每月频率、一致性评分、趋势)
   - 应用使用详情(分类、时长、占比、影响等级)
   - 时间分布(按小时统计)
   - 睡眠影响评估(影响评分0-100、睡眠损失估算、蓝光暴露、风险等级)
   - 工作日vs周末对比
   - 个性化改善建议

2. **应用使用相关性分析** (`/api/v1/screentime/analysis/app-correlation`)
   - 应用共现分析(30分钟时间窗口)
   - 关联指标计算(置信度、提升度、关联评分)
   - 频繁模式识别(3应用组合)
   - 时间窗口统计
   - 网络图构建(节点+边)
   - 关联模式判断(并发/顺序/交替)
   - 典型场景识别(社交/娱乐/工作/购物等)
   - 生成关联洞察

**前端实现**:
- 新增 LateNightUsage.tsx 页面(睡眠影响评估、时段分布、改善建议)
- 新增 AppCorrelation.tsx 页面(关联对表格、频繁模式、网络图、分布统计)
- 使用 ECharts 实现网络图可视化

**技术更新**:
- 新增 2 个 API 端点
- 新增 2 个分析模块文件
- 总计: 113+ API 端点

### 2026-03-02 - 完成所有模块核心功能 ✅

**重大里程碑: 5个模块全部完成核心功能实现**

**新增功能总结**:

1. **Keyboard 模块增强 (3个新功能)**:
   - 左右手使用平衡分析
   - 工作日vs周末详细对比
   - 键盘×屏幕时间跨模块关联分析

2. **ScreenTime 模块增强 (7个新功能)**:
   - 时间浪费检测 (工作时段娱乐应用、深夜使用、碎片化检测)
   - 应用依赖度分析 (使用频率、连续天数、依赖度评分)
   - 工作日vs周末对比 (使用时长、应用分布、设备偏好)
   - 生产力/娱乐比例趋势 (周/月/季度趋势)
   - 应用切换模式识别 (多任务/专注/碎片化模式、切换路径、切换矩阵)
   - 深夜使用习惯分析 (睡眠影响评估、改善建议) ✨ NEW
   - 应用使用相关性分析 (关联规则挖掘、网络图) ✨ NEW

3. **AppleHealth 模块增强 (3个新功能)**:
   - 体重/BMI管理 (趋势图、健康范围、体重预测)
   - 运动数据分析 (步数/距离/卡路里、运动类型、强度分析)
   - 睡眠质量分析 (睡眠时长、质量评分、模式分析)

4. **TrainPlane 模块增强 (3个新功能)**:
   - 旅行足迹地图 (全球地图、城市/国家统计、轨迹动画)
   - 旅行统计增强 (里程排行、记录航班、趋势分析、成就系统)
   - 碳排放计算 (航班/铁路排放、碳足迹报告、碳中和建议)

**技术架构更新**:
- 新增 14 个 API 端点
- 新增 5 个分析模块文件
- 完善跨模块关联分析能力
- 增强数据可视化支持

**项目完成度**:
- Tracks: 100% (30/30 技能)
- Keyboard: 100% (11 个端点)
- ScreenTime: 100% (19 个端点)
- TrainPlane: 100% (18 个端点)
- AppleHealth: 100% (15 个端点)

**总计**: 113+ API 端点，16 个新增高级分析功能

### 2026-02-20 - Phase 5: 完成全部30个轨迹分析技能 ✅

**重大里程碑：30/30 技能全部实现 (100%)**

**新增功能（5个Python Worker技能）**:

1. **stay_detection (高级停留检测)**
   - DBSCAN 聚类算法检测停留
   - 时空约束：200米半径，30分钟最小停留
   - 时间连续性过滤：最大1小时间隙
   - 输出：stay_segments 表增强（cluster_id, cluster_confidence, cluster_type）
   - 性能：~1k points/sec

2. **density_structure_advanced (高级密度结构分析)**
   - DBSCAN 聚类识别高密度区域
   - 聚类分类：HOME/WORK/FREQUENT/OCCASIONAL
   - 凸包面积计算
   - 输出：density_clusters 表（新增）
   - 性能：~500 points/sec

3. **trip_construction_advanced (高级行程构建)**
   - 基于规则的ML目的推断
   - 特征：时间、距离、位置、交通方式
   - 目的分类：COMMUTE/WORK/LEISURE/SHOPPING/TRAVEL/OTHER
   - 输出：trips 表增强（purpose_ml, confidence_ml, features_json）
   - 性能：~100 trips/sec

4. **spatial_persona (空间行为画像)**
   - 4维度评分系统（0-100分）：
     - Mobility Score: 移动性（距离+速度）
     - Exploration Score: 探索性（唯一地点）
     - Routine Score: 规律性（重访模式）
     - Diversity Score: 多样性（交通方式）
   - 中文洞察生成
   - 输出：spatial_persona 表（新增）
   - 性能：~10 sec

5. **admin_view_advanced (高级行政区划分析)**
   - 时序趋势检测：GROWTH/DECLINE/STABLE/SEASONAL
   - 线性回归趋势分析
   - Z-score 异常检测（阈值 > 2.5）
   - 下月访问量预测
   - 输出：admin_trends 表（新增）
   - 性能：~5 sec

**数据库变更**:
- 新增3个表：density_clusters, spatial_persona, admin_trends
- 增强2个表：stay_segments（+3列），trips（+3列）
- 迁移文件：014_create_phase5_tables.sql, 015_enhance_existing_tables.sql

**技术架构**:
- Python 3.11+ workers with Docker
- 依赖：numpy, scipy, scikit-learn, geopy
- Go wrapper: internal/analysis/python/worker.go
- 通过 exec.Command 调用 Python 脚本
- 任务状态由 Python worker 直接更新数据库

### 2026-02-19

- 初始化项目结构
- 配置 Go 模块和依赖
- 创建基础 API 框架
- 实现路由和中间件
- 配置 SQLite 数据库

## 许可证

MIT License

## 联系方式

- GitHub: https://github.com/jengzang/records-backend-go
- 问题反馈: https://github.com/jengzang/records-backend-go/issues

## 致谢

本项目使用了以下开源项目:
- Gin Web Framework
- SQLite
- scikit-learn
- numpy/scipy
- Recharts (前端可视化)
