# 数据库Schema设计方案（支持30个Skills）

## 设计原则

1. **点级数据** → 扩展"一生足迹"表
2. **段级数据** → 创建新表（segments, stay_segments, trips）
3. **聚合统计** → 创建新表（各种统计结果表）
4. **空间分析** → 创建新表（grid_cells, spatial_analysis）
5. **配置管理** → 创建新表（threshold_profiles, analysis_tasks）

---

## 1. 扩展现有表："一生足迹"（track_points）

### 当前字段
```sql
-- 基础GPS字段
id INTEGER PRIMARY KEY AUTOINCREMENT
dataTime INTEGER                    -- Unix timestamp (seconds)
longitude REAL
latitude REAL
heading REAL
accuracy REAL
speed REAL                          -- m/s
distance REAL
altitude REAL
time_visually TEXT                  -- 2025/01/22 21:42:18.000
time TEXT                           -- 20250122214218

-- 行政区字段（已添加）
province TEXT
city TEXT
county TEXT
town TEXT
village TEXT

-- 元数据（已添加）
created_at TEXT
updated_at TEXT
algo_version TEXT
```

### 需要新增的字段

```sql
-- === 质量控制层（02_outlier_detection）===
outlier_flag BOOLEAN DEFAULT 0      -- 是否异常点
outlier_reason_codes TEXT           -- 异常原因JSON: ["GPS_DRIFT", "SPEED_JUMP"]
outlier_confidence REAL             -- 异常置信度 0~1
qa_status TEXT                      -- 质量状态: NORMAL/OUTLIER/SUSPICIOUS

-- === 轨迹补全层（03_trajectory_completion）===
is_synthetic BOOLEAN DEFAULT 0      -- 是否补全点（火车/飞机轨迹）
synthetic_source TEXT               -- 补全来源: TRAIN_ROUTE/FLIGHT_ROUTE
synthetic_metadata TEXT             -- 补全元数据JSON

-- === 行为分类层（01_transport_mode_classification）===
mode TEXT                           -- 交通方式: WALK/CAR/TRAIN/FLIGHT/STAY/UNKNOWN
mode_confidence REAL                -- 置信度 0~1
mode_reason_codes TEXT              -- 判定依据JSON
segment_id INTEGER                  -- 所属行为段ID（外键 → segments表）

-- === 停留检测层（02_stay_detection）===
stay_id INTEGER                     -- 所属停留段ID（外键 → stay_segments表）
is_stay_point BOOLEAN DEFAULT 0     -- 是否停留点

-- === 空间分析层（01_grid_system）===
grid_id TEXT                        -- 所属地图区块ID（格式：z12_x1234_y5678）
grid_level INTEGER                  -- 区块层级（8-16）

-- === 可视化层（01_rendering_metadata）===
render_color TEXT                   -- 渲染颜色（基于速度/模式）
render_width INTEGER                -- 线宽（基于重要性）
render_opacity REAL                 -- 透明度
lod_level INTEGER                   -- LOD层级（1-5）
```

**索引优化：**
```sql
CREATE INDEX idx_outlier_flag ON "一生足迹"(outlier_flag);
CREATE INDEX idx_mode ON "一生足迹"(mode);
CREATE INDEX idx_segment_id ON "一生足迹"(segment_id);
CREATE INDEX idx_stay_id ON "一生足迹"(stay_id);
CREATE INDEX idx_grid_id ON "一生足迹"(grid_id);
CREATE INDEX idx_is_synthetic ON "一生足迹"(is_synthetic);
```

---

## 2. 新建表：行为段表（segments）

**用途：** 存储交通方式分类的段级结果

```sql
CREATE TABLE segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 时间范围
    start_ts INTEGER NOT NULL,          -- 开始时间戳
    end_ts INTEGER NOT NULL,            -- 结束时间戳
    duration_s INTEGER NOT NULL,        -- 持续时长（秒）

    -- 行为分类
    mode TEXT NOT NULL,                 -- WALK/CAR/TRAIN/FLIGHT/STAY/UNKNOWN
    mode_confidence REAL,               -- 置信度 0~1
    mode_reason_codes TEXT,             -- 判定依据JSON

    -- 运动学特征
    distance_m REAL,                    -- 总距离（米）
    avg_speed_mps REAL,                 -- 平均速度（m/s）
    max_speed_mps REAL,                 -- 最大速度（m/s）
    avg_heading REAL,                   -- 平均方向
    heading_variance REAL,              -- 方向方差（稳定性）

    -- 空间范围
    start_lat REAL,
    start_lon REAL,
    end_lat REAL,
    end_lon REAL,
    bbox_json TEXT,                     -- 边界框JSON

    -- 行政区归属
    start_province TEXT,
    start_city TEXT,
    start_county TEXT,
    end_province TEXT,
    end_city TEXT,
    end_county TEXT,

    -- 辅助信息
    point_count INTEGER,                -- 包含的轨迹点数量
    supporting_evidence TEXT,           -- 支持证据JSON

    -- 元数据
    threshold_profile_id INTEGER,      -- 使用的阈值配置
    algo_version TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_segments_mode ON segments(mode);
CREATE INDEX idx_segments_time ON segments(start_ts, end_ts);
CREATE INDEX idx_segments_admin ON segments(start_province, start_city);
```

---

## 3. 新建表：停留段表（stay_segments）

**用途：** 存储停留检测的段级结果

```sql
CREATE TABLE stay_segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 时间范围
    start_ts INTEGER NOT NULL,
    end_ts INTEGER NOT NULL,
    duration_s INTEGER NOT NULL,

    -- 停留类型
    stay_type TEXT NOT NULL,            -- SPATIAL（空间停留）/ ADMIN_AREA（行政区停留）

    -- 空间信息
    center_lat REAL,                    -- 停留中心点
    center_lon REAL,
    radius_m REAL,                      -- 停留半径（空间停留）

    -- 行政区归属
    province TEXT,
    city TEXT,
    county TEXT,
    town TEXT,
    village TEXT,

    -- 统计指标
    point_count INTEGER,                -- 包含的轨迹点数量
    distance_radius REAL,               -- 距离半径（用于判定）
    admin_consistency REAL,             -- 行政区一致性（0~1）

    -- 语义标注（03_stay_annotation）
    activity_type TEXT,                 -- 活动类型: HOME/WORK/SHOPPING/DINING/UNKNOWN
    activity_confidence REAL,           -- 置信度
    user_label TEXT,                    -- 用户自定义标签
    poi_name TEXT,                      -- POI名称（如果匹配到）

    -- 配置与元数据
    threshold_profile_id INTEGER,
    confidence REAL,
    supporting_metrics TEXT,            -- 支持指标JSON
    algo_version TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stay_type ON stay_segments(stay_type);
CREATE INDEX idx_stay_time ON stay_segments(start_ts, end_ts);
CREATE INDEX idx_stay_admin ON stay_segments(province, city, county);
CREATE INDEX idx_stay_duration ON stay_segments(duration_s);
CREATE INDEX idx_stay_activity ON stay_segments(activity_type);
```

---

## 4. 新建表：出行表（trips）

**用途：** 存储出行构建的结果

```sql
CREATE TABLE trips (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 时间信息
    trip_date TEXT NOT NULL,            -- 出行日期 YYYY-MM-DD
    start_ts INTEGER NOT NULL,
    end_ts INTEGER NOT NULL,
    duration_s INTEGER NOT NULL,

    -- 起点终点（关联停留段）
    origin_stay_id INTEGER,             -- 起点停留段ID
    dest_stay_id INTEGER,               -- 终点停留段ID

    -- 空间信息
    origin_lat REAL,
    origin_lon REAL,
    dest_lat REAL,
    dest_lon REAL,

    -- 行政区信息
    origin_province TEXT,
    origin_city TEXT,
    origin_county TEXT,
    dest_province TEXT,
    dest_city TEXT,
    dest_county TEXT,

    -- 出行特征
    distance_m REAL,                    -- 总距离
    avg_speed_mps REAL,
    max_speed_mps REAL,
    primary_mode TEXT,                  -- 主要交通方式
    mode_sequence TEXT,                 -- 交通方式序列JSON

    -- 统计指标
    segment_count INTEGER,              -- 包含的行为段数量
    point_count INTEGER,                -- 包含的轨迹点数量

    -- 元数据
    algo_version TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (origin_stay_id) REFERENCES stay_segments(id),
    FOREIGN KEY (dest_stay_id) REFERENCES stay_segments(id)
);

CREATE INDEX idx_trips_date ON trips(trip_date);
CREATE INDEX idx_trips_time ON trips(start_ts, end_ts);
CREATE INDEX idx_trips_origin ON trips(origin_province, origin_city);
CREATE INDEX idx_trips_dest ON trips(dest_province, dest_city);
CREATE INDEX idx_trips_mode ON trips(primary_mode);
```

---

## 5. 新建表：地图区块表（grid_cells）

**用途：** 存储地图区块系统的聚合数据

```sql
CREATE TABLE grid_cells (
    grid_id TEXT PRIMARY KEY,           -- 格式：z12_x1234_y5678

    -- 区块信息
    zoom_level INTEGER NOT NULL,       -- 缩放层级 8-16
    tile_x INTEGER NOT NULL,
    tile_y INTEGER NOT NULL,

    -- 边界框
    bbox_json TEXT NOT NULL,            -- {min_lat, max_lat, min_lon, max_lon}
    center_lat REAL,
    center_lon REAL,

    -- 统计指标
    point_count INTEGER DEFAULT 0,      -- 轨迹点数量
    visit_count INTEGER DEFAULT 0,      -- 访问次数（去重后）
    first_visit_ts INTEGER,             -- 首次访问时间
    last_visit_ts INTEGER,              -- 最后访问时间
    total_duration_s INTEGER,           -- 总停留时长

    -- 行为分布
    mode_distribution TEXT,             -- 交通方式分布JSON
    speed_distribution TEXT,            -- 速度分布JSON

    -- 密度分析（03_density_structure）
    density_level TEXT,                 -- CORE/SECONDARY/EDGE/SPARSE
    density_score REAL,                 -- 密度得分

    -- 重访分析（05_revisit_pattern）
    revisit_count INTEGER DEFAULT 0,    -- 重访次数
    revisit_frequency REAL,             -- 重访频率
    avg_revisit_interval_days REAL,     -- 平均重访间隔

    -- 元数据
    last_updated TEXT DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT
);

CREATE INDEX idx_grid_zoom ON grid_cells(zoom_level);
CREATE INDEX idx_grid_coords ON grid_cells(tile_x, tile_y);
CREATE INDEX idx_grid_density ON grid_cells(density_level);
CREATE INDEX idx_grid_visits ON grid_cells(visit_count);
```

---

## 6. 新建表：阈值配置表（threshold_profiles）

**用途：** 存储不同的分析阈值配置

```sql
CREATE TABLE threshold_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 配置信息
    profile_name TEXT NOT NULL UNIQUE,  -- 配置名称: default/strict/loose
    skill_name TEXT NOT NULL,           -- 适用的skill
    description TEXT,

    -- 参数JSON
    params_json TEXT NOT NULL,          -- 阈值参数JSON

    -- 示例：停留检测参数
    -- {
    --   "min_duration_s": 600,
    --   "spatial_radius_m": 80,
    --   "admin_level": "county",
    --   "merge_gap_s": 120
    -- }

    -- 元数据
    is_default BOOLEAN DEFAULT 0,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_profile_skill ON threshold_profiles(skill_name);
CREATE INDEX idx_profile_default ON threshold_profiles(is_default);
```

---

## 7. 新建表：分析任务表（analysis_tasks）

**用途：** 管理所有数据处理任务（类似geocoding_tasks）

```sql
CREATE TABLE analysis_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 任务信息
    skill_name TEXT NOT NULL,           -- skill名称
    task_type TEXT NOT NULL,            -- 任务类型: FULL/INCREMENTAL/RECOMPUTE
    status TEXT NOT NULL,               -- pending/running/completed/failed

    -- 参数
    threshold_profile_id INTEGER,      -- 使用的阈值配置
    params_json TEXT,                   -- 其他参数JSON

    -- 进度
    total_items INTEGER,
    processed_items INTEGER DEFAULT 0,
    failed_items INTEGER DEFAULT 0,

    -- 时间
    start_time TEXT,
    end_time TEXT,
    eta_seconds INTEGER,

    -- 结果
    result_summary TEXT,                -- 结果摘要JSON
    error_message TEXT,

    -- 元数据
    created_by TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (threshold_profile_id) REFERENCES threshold_profiles(id)
);

CREATE INDEX idx_analysis_status ON analysis_tasks(status);
CREATE INDEX idx_analysis_skill ON analysis_tasks(skill_name);
CREATE INDEX idx_analysis_time ON analysis_tasks(created_at);
```

---

## 8. 新建表：足迹统计表（footprint_statistics）

**用途：** 存储足迹层统计结果（01_footprint_statistics）

```sql
CREATE TABLE footprint_statistics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 统计维度
    stat_type TEXT NOT NULL,            -- PROVINCE/CITY/COUNTY/TOWN/GRID
    stat_key TEXT NOT NULL,             -- 统计键（如省份名称、grid_id）

    -- 时间范围
    time_range TEXT,                    -- ALL/YEAR_2024/MONTH_202401

    -- 统计指标
    point_count INTEGER DEFAULT 0,
    visit_count INTEGER DEFAULT 0,      -- 去重后的访问次数
    total_duration_s INTEGER DEFAULT 0,
    total_distance_m REAL DEFAULT 0,

    -- 时间分布
    first_visit_ts INTEGER,
    last_visit_ts INTEGER,
    visit_dates TEXT,                   -- 访问日期列表JSON

    -- 排名
    rank_by_points INTEGER,
    rank_by_duration INTEGER,
    rank_by_visits INTEGER,

    -- 元数据
    last_updated TEXT DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT,

    UNIQUE(stat_type, stat_key, time_range)
);

CREATE INDEX idx_footprint_type ON footprint_statistics(stat_type);
CREATE INDEX idx_footprint_rank ON footprint_statistics(rank_by_points);
CREATE INDEX idx_footprint_time ON footprint_statistics(time_range);
```

---

## 9. 新建表：停留统计表（stay_statistics）

**用途：** 存储停留层统计结果（02_stay_statistics）

```sql
CREATE TABLE stay_statistics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 统计维度
    stat_type TEXT NOT NULL,            -- PROVINCE/CITY/COUNTY/ACTIVITY_TYPE
    stat_key TEXT NOT NULL,

    -- 时间范围
    time_range TEXT,

    -- 统计指标
    stay_count INTEGER DEFAULT 0,       -- 停留次数
    total_duration_s INTEGER DEFAULT 0, -- 总停留时长
    avg_duration_s REAL,                -- 平均停留时长
    max_duration_s INTEGER,             -- 最长停留时长

    -- 停留类型分布
    spatial_stay_count INTEGER DEFAULT 0,
    admin_stay_count INTEGER DEFAULT 0,

    -- 活动类型分布（如果stat_type=ACTIVITY_TYPE）
    activity_distribution TEXT,         -- JSON

    -- 排名
    rank_by_count INTEGER,
    rank_by_duration INTEGER,

    -- 元数据
    last_updated TEXT DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT,

    UNIQUE(stat_type, stat_key, time_range)
);

CREATE INDEX idx_stay_stat_type ON stay_statistics(stat_type);
CREATE INDEX idx_stay_stat_rank ON stay_statistics(rank_by_duration);
```

---

## 10. 新建表：极值事件表（extreme_events）

**用途：** 存储极值旅行事件（03_extreme_events）

```sql
CREATE TABLE extreme_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 事件类型
    event_type TEXT NOT NULL,           -- HIGHEST_ALTITUDE/EASTMOST/WESTMOST/NORTHMOST/SOUTHMOST

    -- 排名
    rank INTEGER NOT NULL,              -- Top N排名

    -- 位置信息
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    altitude REAL,

    -- 行政区
    province TEXT,
    city TEXT,
    county TEXT,
    town TEXT,

    -- 时间信息
    event_ts INTEGER NOT NULL,
    event_date TEXT,

    -- 上下文
    segment_id INTEGER,                 -- 所属行为段
    mode TEXT,                          -- 交通方式

    -- 指标值
    metric_value REAL,                  -- 指标值（如海拔高度、经纬度）

    -- 元数据
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT,

    UNIQUE(event_type, rank)
);

CREATE INDEX idx_extreme_type ON extreme_events(event_type);
CREATE INDEX idx_extreme_rank ON extreme_events(rank);
```

---

## 11. 新建表：空间分析结果表（spatial_analysis）

**用途：** 存储各种空间分析的结果

```sql
CREATE TABLE spatial_analysis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- 分析类型
    analysis_type TEXT NOT NULL,        -- DENSITY/SPEED_SPACE/REVISIT/EFFICIENCY/COMPLEXITY/DIRECTIONAL

    -- 分析范围
    scope TEXT,                         -- GLOBAL/PROVINCE_XX/CITY_XX/GRID_XX
    time_range TEXT,                    -- ALL/YEAR_2024

    -- 结果JSON
    result_json TEXT NOT NULL,          -- 分析结果JSON

    -- 关键指标（便于查询）
    primary_metric REAL,
    secondary_metric REAL,

    -- 元数据
    threshold_profile_id INTEGER,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    algo_version TEXT,

    UNIQUE(analysis_type, scope, time_range)
);

CREATE INDEX idx_spatial_type ON spatial_analysis(analysis_type);
CREATE INDEX idx_spatial_scope ON spatial_analysis(scope);
```

---

## 总结：表的用途分类

### 扩展旧表（1个）
- ✅ **"一生足迹"** - 轨迹点表，添加点级分析字段

### 核心数据表（3个）
- ✅ **segments** - 行为段（交通方式分类结果）
- ✅ **stay_segments** - 停留段（停留检测结果）
- ✅ **trips** - 出行（出行构建结果）

### 空间分析表（2个）
- ✅ **grid_cells** - 地图区块（空间聚合）
- ✅ **spatial_analysis** - 空间分析结果（通用）

### 统计聚合表（3个）
- ✅ **footprint_statistics** - 足迹统计
- ✅ **stay_statistics** - 停留统计
- ✅ **extreme_events** - 极值事件

### 系统管理表（2个）
- ✅ **threshold_profiles** - 阈值配置
- ✅ **analysis_tasks** - 分析任务管理

### 已有表（1个）
- ✅ **geocoding_tasks** - Geocoding任务管理（已实现）

---

## 下一步：需要您确认

1. **表结构是否合理？** 有没有遗漏或冗余？
2. **字段设计是否完整？** 是否支持所有30个skills？
3. **索引优化是否充分？** 查询性能是否满足需求？
4. **是否需要调整？** 比如合并某些表、拆分某些表？

确认后，我将：
1. 创建完整的migration SQL文件
2. 创建对应的Go models
3. 创建repositories和services
4. 开始实现数据处理pipeline
