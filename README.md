# Records Backend (Go)

个人数据分析平台后端服务

## 项目简介

本项目是个人数据分析平台的后端服务，使用 Go 语言和 Gin 框架开发，提供 RESTful API 接口。

## 技术栈

- Go 1.21+
- Gin Web Framework
- SQLite3 (WAL 模式)
- JWT 认证
- 纯 RESTful API

## 项目结构

```
go-backend/
├── cmd/
│   └── server/          # 服务器入口
│       └── main.go
├── internal/
│   ├── api/            # API 路由和处理器
│   ├── config/         # 配置管理
│   ├── database/       # 数据库操作
│   ├── middleware/     # 中间件
│   ├── models/         # 数据模型
│   └── services/       # 业务逻辑
├── pkg/                # 公共包
├── scripts/            # Python 辅助脚本
│   ├── tracks/        # 轨迹数据处理
│   └── keyboard/      # 键盘鼠标数据处理
├── data/               # 数据文件和数据库
│   ├── tracks/
│   └── keyboard/
├── main.go            # 简单启动入口
├── go.mod
└── README.md
```

## 核心功能模块

### 1. 轨迹分析 (Tracks)
- GPS 轨迹数据管理
- 轨迹分类和分析
- 行政区划统计
- 停留检测
- **轨迹分析技能 (30/30 已实现 - 100% ✅):**
  - **Phase 1 (5 skills):**
    - footprint_statistics: 足迹统计
    - stay_statistics: 停留统计
    - extreme_events: 极值事件
    - speed_space_coupling: 速度-空间耦合
    - revisit_pattern: 重访模式
  - **Phase 2 (3 skills):**
    - speed_events: 速度事件检测
    - rendering_metadata: 渲染元数据生成
    - stay_annotation: 停留标注与建议
  - **Phase 3 (5 skills):**
    - outlier_detection: 异常点检测 (Z-score & IQR)
    - trajectory_completion: 轨迹补全 (线性插值)
    - transport_mode: 交通方式识别 (速度分类)
    - streak_detection: 连续活动检测
    - grid_system: 网格系统 (Geohash 空间索引)
  - **Phase 4 (12 skills):**
    - admin_crossings: 行政区划穿越检测
    - admin_view_engine: 行政区划视图引擎
    - utilization_efficiency: 空间利用效率
    - altitude_dimension: 海拔维度分析
    - directional_bias: 方向偏好分析
    - spatial_complexity: 空间复杂度分析
    - density_structure: 密度结构分析
    - road_overlap: 道路重叠分析
    - time_axis_map: 时间轴地图
    - trip_construction: 行程构建
    - time_space_slicing: 时空切片
    - time_space_compression: 时空压缩
  - **Phase 5 (5 skills - Python Workers - NEW ✅):**
    - stay_detection: 高级停留检测 (DBSCAN 聚类)
    - density_structure_advanced: 高级密度结构分析 (DBSCAN)
    - trip_construction_advanced: 高级行程构建 (ML 目的推断)
    - spatial_persona: 空间行为画像引擎
    - admin_view_advanced: 高级行政区划分析 (时序趋势)

### 2. 键盘鼠标统计 (Keyboard)
- 键盘鼠标使用数据统计
- 使用习惯分析

### 3. 飞机火车路线 (Flights)
- 飞机航线管理
- 火车路线管理
- 路线数据导入

### 4. 屏幕使用时间 (Screentime)
- 手机屏幕使用时间
- 电脑屏幕使用时间
- 应用使用统计

### 5. Apple健康数据 (Health)
- 步数统计
- 心率数据
- 睡眠分析
- 运动数据

## 运行方式

### 开发环境

```bash
# 安装依赖
go mod download

# 运行服务器（简单方式）
go run main.go

# 运行服务器（标准方式）
go run cmd/server/main.go
```

### 生产构建

```bash
# 构建二进制文件
go build -o records-server cmd/server/main.go

# 运行
./records-server
```

## 环境变量

```bash
PORT=:8080                          # 服务器端口
DB_PATH=./data/records.db          # 数据库路径
JWT_SECRET=your-secret-key         # JWT 密钥
```

## API 接口

### 健康检查
- `GET /health` - 服务健康检查

### 轨迹接口
- `GET /api/v1/tracks` - 获取轨迹列表
- `POST /api/v1/tracks` - 创建轨迹
- **`POST /api/v1/analysis/tasks` - 创建分析任务 (NEW)**
  - 支持的 skill_name: speed_events, rendering_metadata, stay_annotation, footprint_statistics, stay_statistics, extreme_events, speed_space_coupling, revisit_pattern
  - 参数: skill_name, mode (incremental/full_recompute)
- **`GET /api/v1/analysis/tasks/:id` - 查询任务状态 (NEW)**
- **`GET /api/v1/analysis/tasks` - 获取任务列表 (NEW)**

### 键盘鼠标接口
- `GET /api/v1/keyboard/stats` - 获取统计数据

### 飞机火车接口
- `GET /api/v1/flights` - 获取路线列表

### 屏幕使用时间接口
- `GET /api/v1/screentime/stats` - 获取统计数据

### Apple健康数据接口
- `GET /api/v1/health-data/stats` - 获取健康数据统计

## 数据库

- SQLite3 数据库
- WAL 模式开启
- 小连接池配置
- 支持地理空间扩展

**轨迹分析表 (Phase 2 新增):**
- `speed_events`: 高速事件记录（max_speed, avg_speed, duration, location, confidence）
- `stay_annotations`: 停留标注（label, sub_label, note, confirmed）
- `stay_context_cache`: 停留上下文缓存（context_json, suggestions_json）
- `place_anchors`: 地点锚点（HOME/WORK等，grid_id, radius）
- `render_segments_cache`: 渲染缓存（speed_bucket, overlap_rank, style hints）

## 权限模型

- JWT 认证
- 管理员优先级
- 普通用户只读权限
- 请求排队和限流

## 服务器约束

- 2核 2GB 内存
- Go 可用内存 ≤ 1GB
- 预计并发 ≤ 3 req/s
- 计算型接口优化

## Python 辅助脚本

`scripts/` 目录包含 Python 辅助脚本：
- 数据处理
- 测试脚本
- 机器学习模型训练

## 更新日志

### 2026-02-20 - Phase 5: 完成全部30个轨迹分析技能 ✅

**重大里程碑：30/30 技能全部实现 (100%)**

**新增功能（5个Python Worker技能）：**

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

**数据库变更：**
- 新增3个表：density_clusters, spatial_persona, admin_trends
- 增强2个表：stay_segments（+3列），trips（+3列）
- 迁移文件：014_create_phase5_tables.sql, 015_enhance_existing_tables.sql

**技术架构：**
- Python 3.11+ workers with Docker
- 依赖：numpy, scipy, scikit-learn, geopy
- Go wrapper: internal/analysis/python/worker.go
- 通过 exec.Command 调用 Python 脚本
- 任务状态由 Python worker 直接更新数据库

**文档：**
- docs/tracks/phase5-implementation-summary.md

### 2026-02-20 - Phase 3: 迁移5个Python技能到Go

**新增功能：**
1. **outlier_detection (异常点检测)**
   - Z-score 方法：|z| > 3 标记为异常
   - IQR 方法：Q1 - 1.5*IQR 或 Q3 + 1.5*IQR 标记为异常
   - 速度异常：speed > 200 km/h (55.56 m/s)
   - 精度异常：accuracy > 1000m
   - 更新 track_points 表的 outlier_flag 列

2. **trajectory_completion (轨迹补全)**
   - 检测时间间隔 > 5分钟的轨迹间隙
   - 线性插值填补 ≤ 30分钟的间隙
   - 插值字段：latitude, longitude, altitude, speed
   - 标记为 qa_status='interpolated'

3. **transport_mode (交通方式识别)**
   - 基于速度的分类算法
   - 阈值：WALK (0-2 m/s), BIKE (2-8 m/s), CAR (8-40 m/s), TRAIN (40-60 m/s), PLANE (>60 m/s)
   - 生成 segments 表记录
   - 最小段长度：10秒

4. **streak_detection (连续活动检测)**
   - 检测连续活动天数（≥2天）
   - 最小活动阈值：1km/天
   - 输出：start_date, end_date, days_count, total_distance, total_duration
   - 新增 streaks 表

5. **grid_system (网格系统)**
   - Geohash 空间索引（精度 4-7）
   - 聚合统计：visit_count, total_duration, first/last visit
   - 填充 grid_cells 表
   - 支持多精度级别分析

**数据库变更：**
- 新增表：streaks (连续活动记录)
- 迁移文件：012_create_streaks_table.sql

**框架更新：**
- 新增 internal/analysis/foundation/ 包（outlier_detection, trajectory_completion）
- 扩展 internal/analysis/behavior/ 包（transport_mode, streak_detection）
- 扩展 internal/analysis/spatial/ 包（grid_system）
- main.go: 导入 foundation 包

**进度：**
- 13/30 技能完成 (43.3%)
- Phase 1 + Phase 2 + Phase 3 完成
- 准备进入 Phase 4（新 Go 技能实现）

**性能预期：**
- outlier_detection: ~10k points/sec
- trajectory_completion: ~5k points/sec (含插入)
- transport_mode: ~10k points/sec
- streak_detection: ~50k points/sec (聚合)
- grid_system: ~5k points/sec (含 geohash)

### 2026-02-20 - Phase 2: 实现3个轨迹分析技能

**新增功能：**
1. **speed_events (速度事件检测)**
   - 基于 CAR segments 检测高速事件
   - 状态机算法，支持事件级检测（连续高速段）
   - 参数化阈值：min_event_speed=33.33 m/s (120 km/h), min_event_duration=60s
   - 输出：max_speed, avg_speed, duration, location, confidence, reason_codes

2. **rendering_metadata (渲染元数据生成)**
   - 为地图渲染生成可视化元数据
   - 速度分桶（0-5）基于全局百分位数
   - 重叠统计（基于 grid_id）
   - 样式提示：line_weight (1.0-3.0), alpha_hint (0.3-1.0)
   - 支持3个 LOD 级别

3. **stay_annotation (停留标注与建议)**
   - 生成停留上下文卡片（时间、地点、到达/离开上下文）
   - 规则引擎生成标签建议（HOME/WORK/EAT/SLEEP/TRANSIT）
   - 支持历史标注复用和地点锚点机制
   - 输出可解释的建议（confidence + reasons）

**数据库变更：**
- 新增5个表：speed_events, stay_annotations, stay_context_cache, place_anchors, render_segments_cache
- 迁移文件：011_create_phase2_tables.sql

**框架更新：**
- incremental.go: 新增 MarkTaskAsCompleted(支持结果摘要), UpdateTaskProgress 方法
- main.go: 导入所有分析器包以触发注册

**进度：**
- 8/30 技能完成 (26.7%)
- Phase 1 + Phase 2 完成
- 准备进入 Phase 3（中等难度迁移）

**性能预期：**
- speed_events: ~100 segments/sec, <50MB 内存
- rendering_metadata: ~50 segments/sec, <100MB 内存
- stay_annotation: ~20 stays/sec, <50MB 内存

### 2026-02-19
- 初始化项目结构
- 配置 Go 模块和依赖
- 创建基础 API 框架
- 实现路由和中间件
- 配置 SQLite 数据库
