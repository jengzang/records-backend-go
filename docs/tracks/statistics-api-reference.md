# 统计API参考文档

## 概述

本文档详细说明轨迹分析系统的统计API端点，包含16个端点组（38个handler方法）。

**基础URL**: `http://localhost:8080/api/v1/stats`

**认证方式**: 当前版本无需认证（公开只读访问）

**通用响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**错误响应格式**:
```json
{
  "code": 400,
  "message": "Invalid query parameters",
  "error": "详细错误信息"
}
```

**通用查询参数**:
- `limit` (integer): 返回结果数量限制，默认值因端点而异
- `bucket` (string): 时间分桶类型，可选值: `all`, `year`, `month`, `week`

---

## 1. 足迹统计 (Footprint Statistics)

### GET /api/v1/stats/footprint/rankings

获取行政区划足迹排行榜。

**查询参数**:
- `stat_type` (string): 统计类型，可选值: `PROVINCE`, `CITY`, `COUNTY`, `TOWN` (默认: `PROVINCE`)
- `time_range` (string): 时间范围，可选值: `all`, `year`, `month` (默认: `all`)
- `order_by` (string): 排序字段，可选值: `points`, `duration`, `visits` (默认: `points`)
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/footprint/rankings?stat_type=CITY&order_by=points&limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "stat_type": "CITY",
        "area_key": "广东省-深圳市",
        "province": "广东省",
        "city": "深圳市",
        "county": null,
        "town": null,
        "point_count": 150000,
        "visit_count": 500,
        "total_duration_seconds": 2592000,
        "first_visit": 1704067200,
        "last_visit": 1737552138,
        "created_at": "2026-02-22T00:00:00Z",
        "updated_at": "2026-02-22T00:00:00Z"
      }
    ],
    "count": 20
  }
}
```

**响应字段说明**:
- `stat_type`: 统计类型 (PROVINCE/CITY/COUNTY/TOWN)
- `area_key`: 区域唯一标识 (格式: "省-市-区-镇")
- `point_count`: GPS点数量
- `visit_count`: 访问次数
- `total_duration_seconds`: 总停留时长（秒）
- `first_visit`: 首次访问时间 (Unix timestamp)
- `last_visit`: 最后访问时间 (Unix timestamp)

---

## 2. 停留统计 (Stay Statistics)

### GET /api/v1/stats/stay/rankings

获取停留地点排行榜。

**查询参数**:
- `stat_type` (string): 统计类型 (默认: `PROVINCE`)
- `time_range` (string): 时间范围 (默认: `all`)
- `order_by` (string): 排序字段，可选值: `count`, `duration` (默认: `count`)
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/stay/rankings?stat_type=CITY&order_by=duration&limit=10"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "stat_type": "CITY",
        "area_key": "广东省-深圳市",
        "province": "广东省",
        "city": "深圳市",
        "stay_count": 120,
        "total_duration_seconds": 864000,
        "avg_duration_seconds": 7200,
        "longest_stay_seconds": 86400,
        "created_at": "2026-02-22T00:00:00Z"
      }
    ],
    "count": 10
  }
}
```

**响应字段说明**:
- `stay_count`: 停留次数
- `total_duration_seconds`: 总停留时长（秒）
- `avg_duration_seconds`: 平均停留时长（秒）
- `longest_stay_seconds`: 最长单次停留时长（秒）

---

## 3. 极值事件 (Extreme Events)

### GET /api/v1/stats/extreme-events

获取极值事件记录（最快速度、最高海拔等）。

**查询参数**:
- `eventType` (string): 事件类型，可选值: `FASTEST_SPEED`, `HIGHEST_ALTITUDE`, `LONGEST_DISTANCE`
- `eventCategory` (string): 事件分类，可选值: `SPEED`, `ALTITUDE`, `DISTANCE`
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/extreme-events?eventType=FASTEST_SPEED&limit=10"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "event_type": "FASTEST_SPEED",
        "event_category": "SPEED",
        "value": 850.5,
        "unit": "km/h",
        "location_lat": 22.5431,
        "location_lon": 114.0579,
        "province": "广东省",
        "city": "深圳市",
        "event_time": 1737552138,
        "confidence": 0.98,
        "reason_codes": "PLANE_MODE,HIGH_ALTITUDE",
        "created_at": "2026-02-22T00:00:00Z"
      }
    ],
    "count": 10
  }
}
```

**响应字段说明**:
- `event_type`: 事件类型
- `event_category`: 事件分类
- `value`: 事件数值
- `unit`: 单位
- `confidence`: 置信度 (0-1)
- `reason_codes`: 原因代码（逗号分隔）

---

## 4. 行政区划跨越 (Admin Crossings)

### GET /api/v1/stats/admin-crossings

获取行政区划边界跨越统计。

**查询参数**:
- `crossing_type` (string): 跨越类型，可选值: `PROVINCE`, `CITY`, `COUNTY`
- `from` (string): 起始区域名称
- `to` (string): 目标区域名称
- `start_time` (integer): 开始时间 (Unix timestamp, 默认: 0)
- `end_time` (integer): 结束时间 (Unix timestamp, 默认: 0)
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/admin-crossings?crossing_type=CITY&limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "id": 1,
        "crossing_type": "CITY",
        "from_province": "广东省",
        "from_city": "深圳市",
        "to_province": "广东省",
        "to_city": "广州市",
        "crossing_count": 25,
        "first_crossing": 1704067200,
        "last_crossing": 1737552138,
        "created_at": "2026-02-22T00:00:00Z"
      }
    ],
    "count": 20
  }
}
```

---

## 5. 行政区域视图 (Admin View)

### GET /api/v1/stats/admin-view

获取特定行政区域的详细统计信息。

**查询参数**:
- `admin_level` (string): 行政级别，可选值: `province`, `city`, `county`, `town`
- `admin_name` (string): 行政区域名称
- `parent_name` (string): 父级区域名称（用于过滤）
- `sort_by` (string): 排序字段 (默认: `visit_count`)
- `limit` (integer): 返回数量 (默认: 50)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/admin-view?admin_level=city&admin_name=深圳市"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "admin_level": "city",
        "province": "广东省",
        "city": "深圳市",
        "visit_count": 500,
        "point_count": 150000,
        "total_duration": 2592000,
        "first_visit": 1704067200,
        "last_visit": 1737552138
      }
    ],
    "count": 1
  }
}
```

---

## 6. 速度-空间耦合 (Speed-Space Coupling)

### GET /api/v1/stats/speed-space

获取速度与空间分布的耦合统计。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型，可选值: `province`, `city`, `county`
- `area_name` (string): 区域名称
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/speed-space?area_type=city&limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "bucket_type": "all",
      "area_type": "city",
      "area_key": "广东省-深圳市",
      "avg_speed": 35.5,
      "max_speed": 120.0,
      "speed_variance": 450.2,
      "high_speed_ratio": 0.15,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/speed-space/high-speed-zones

获取高速区域排行榜。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `limit` (integer): 返回数量 (默认: 50)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/speed-space/high-speed-zones?limit=10"
```

### GET /api/v1/stats/speed-space/slow-life-zones

获取慢生活区域排行榜（低速高停留时长）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `limit` (integer): 返回数量 (默认: 50)

---

## 7. 方向偏好 (Directional Bias)

### GET /api/v1/stats/directional-bias

获取方向偏好统计（东西南北移动倾向）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `area_key` (string): 区域标识
- `mode` (string): 交通方式过滤，可选值: `ALL`, `WALK`, `BIKE`, `CAR`, `TRAIN`, `PLANE` (默认: `ALL`)
- `limit` (integer): 返回数量 (默认: 50)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/directional-bias?mode=CAR&limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "bucket_type": "all",
      "area_type": "city",
      "area_key": "广东省-深圳市",
      "north_ratio": 0.25,
      "south_ratio": 0.20,
      "east_ratio": 0.30,
      "west_ratio": 0.25,
      "dominant_direction": "EAST",
      "bias_strength": 0.65,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/directional-bias/top-areas

获取方向偏好最强的区域。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 10)

### GET /api/v1/stats/directional-bias/bidirectional

获取双向移动模式（往返频繁的路线）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 10)

---

## 8. 重访模式 (Revisit Patterns)

### GET /api/v1/stats/revisit-patterns

获取地点重访模式统计。

**查询参数**:
- `min_visits` (integer): 最小访问次数 (默认: 2)
- `habitual_only` (boolean): 仅返回习惯性地点 (默认: false)
- `periodic_only` (boolean): 仅返回周期性地点 (默认: false)
- `limit` (integer): 返回数量 (默认: 50)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/revisit-patterns?min_visits=5&limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "grid_id": "L5_12345_67890",
      "center_lat": 22.5431,
      "center_lon": 114.0579,
      "visit_count": 50,
      "total_duration": 360000,
      "avg_interval_days": 7.5,
      "is_habitual": true,
      "is_periodic": true,
      "periodicity_score": 0.85,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/revisit-patterns/top-locations

获取访问次数最多的地点。

**查询参数**:
- `limit` (integer): 返回数量 (默认: 20)

### GET /api/v1/stats/revisit-patterns/habitual

获取习惯性访问地点（高频率、规律性强）。

**查询参数**:
- `limit` (integer): 返回数量 (默认: 20)

### GET /api/v1/stats/revisit-patterns/periodic

获取周期性访问地点（有明显时间周期）。

**查询参数**:
- `limit` (integer): 返回数量 (默认: 20)

---

## 9. 空间利用效率 (Spatial Utilization)

### GET /api/v1/stats/spatial-utilization

获取空间利用效率统计（目的地、走廊、深度参与区域）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `area_key` (string): 区域标识
- `limit` (integer): 返回数量 (默认: 20)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/spatial-utilization?limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "bucket_type": "all",
      "area_type": "city",
      "area_key": "广东省-深圳市",
      "utilization_type": "DESTINATION",
      "efficiency_score": 0.85,
      "visit_density": 150.5,
      "time_density": 3600.0,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/spatial-utilization/destinations

获取目的地区域（高停留时长、低移动）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `limit` (integer): 返回数量 (默认: 10)

### GET /api/v1/stats/spatial-utilization/corridors

获取交通走廊（高移动、低停留）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `limit` (integer): 返回数量 (默认: 10)

### GET /api/v1/stats/spatial-utilization/deep-engagement

获取深度参与区域（高频率、长时间、多样化活动）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `limit` (integer): 返回数量 (默认: 10)

---

## 10. 密度结构 (Density Structure)

### GET /api/v1/stats/density

获取空间密度网格统计。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `level` (string): 密度级别，可选值: `HIGH`, `MEDIUM`, `LOW`
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/density?level=HIGH&limit=50"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "bucket_type": "all",
      "grid_id": "L5_12345_67890",
      "center_lat": 22.5431,
      "center_lon": 114.0579,
      "density_level": "HIGH",
      "point_density": 500.5,
      "time_density": 7200.0,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/density/core

获取核心活动区域（最高密度）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 50)

### GET /api/v1/stats/density/rare

获取稀疏访问区域（低密度）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 50)

### GET /api/v1/stats/density/clusters

获取密度聚类结果（DBSCAN聚类）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 20)

---

## 11. 海拔维度 (Altitude Dimension)

### GET /api/v1/stats/altitude

获取海拔统计信息。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `area_key` (string): 区域标识
- `limit` (integer): 返回数量 (默认: 50)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/altitude?limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "bucket_type": "all",
      "area_type": "province",
      "area_key": "西藏自治区",
      "avg_altitude": 4500.0,
      "max_altitude": 5500.0,
      "min_altitude": 3000.0,
      "altitude_span": 2500.0,
      "vertical_intensity": 0.75,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/altitude/highest-spans

获取海拔跨度最大的区域。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 10)

### GET /api/v1/stats/altitude/highest-intensity

获取垂直移动强度最高的区域。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 10)

---

## 12. 时空压缩 (Time-Space Compression)

### GET /api/v1/stats/time-space-compression

获取时空压缩统计（单位时间内的移动距离）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `area_type` (string): 区域类型
- `area_key` (string): 区域标识
- `limit` (integer): 返回数量 (默认: 50)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/time-space-compression?limit=20"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "bucket_type": "all",
      "time_window": "1h",
      "avg_distance_per_hour": 50.5,
      "max_distance_per_hour": 850.0,
      "compression_intensity": 0.85,
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/time-space-compression/highest-intensity

获取移动强度最高的时间段。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 10)

### GET /api/v1/stats/time-space-compression/burst-periods

获取爆发性移动时段（短时间内大距离移动）。

**查询参数**:
- `bucket` (string): 时间分桶 (默认: `all`)
- `limit` (integer): 返回数量 (默认: 10)

---

## 13. 时空切片 (Time-Space Slicing)

### GET /api/v1/stats/time-space-slices

获取时空切片统计（按时间维度切分的空间分布）。

**查询参数**:
- `slice_type` (string): 切片类型，可选值: `HOURLY`, `DAILY`, `WEEKLY`, `MONTHLY`
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/time-space-slices?slice_type=HOURLY&limit=24"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "slice_type": "HOURLY",
      "time_key": "08:00-09:00",
      "point_count": 5000,
      "unique_locations": 150,
      "avg_speed": 35.5,
      "dominant_mode": "CAR",
      "created_at": "2026-02-22T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/stats/time-space-slices/weekly-pattern

获取每周活动模式（周一到周日的空间分布）。

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/time-space-slices/weekly-pattern"
```

### GET /api/v1/stats/time-space-slices/hourly-pattern

获取每日时段模式（24小时的空间分布）。

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/time-space-slices/hourly-pattern"
```

---

## 14. 空间复杂度 (Spatial Complexity)

### GET /api/v1/stats/spatial-complexity

获取空间行为复杂度评分。

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/spatial-complexity"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "complexity_score": 0.85,
    "unique_locations": 500,
    "location_entropy": 8.5,
    "mode_diversity": 0.75,
    "temporal_irregularity": 0.65,
    "spatial_spread": 1500.5,
    "created_at": "2026-02-22T00:00:00Z"
  }
}
```

**响应字段说明**:
- `complexity_score`: 综合复杂度评分 (0-1)
- `unique_locations`: 唯一地点数量
- `location_entropy`: 地点熵（多样性指标）
- `mode_diversity`: 交通方式多样性 (0-1)
- `temporal_irregularity`: 时间不规律性 (0-1)
- `spatial_spread`: 空间广度（公里）

---

## 15. 道路重叠 (Road Overlap)

### GET /api/v1/stats/road-overlap

获取道路重叠统计（重复经过的路段）。

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/road-overlap"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "total_segments": 10000,
    "overlapped_segments": 3500,
    "overlap_ratio": 0.35,
    "avg_overlap_count": 2.5,
    "max_overlap_count": 50,
    "created_at": "2026-02-22T00:00:00Z"
  }
}
```

**响应字段说明**:
- `total_segments`: 总路段数
- `overlapped_segments`: 重叠路段数
- `overlap_ratio`: 重叠比例 (0-1)
- `avg_overlap_count`: 平均重叠次数
- `max_overlap_count`: 最大重叠次数

---

## 16. 空间画像 (Spatial Persona)

### GET /api/v1/stats/spatial-persona

获取综合空间行为画像（个人空间指数PSI）。

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/stats/spatial-persona"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "psi": 63.40,
    "behavior_type": "COMMUTER",
    "stability": 0.75,
    "footprint_diversity": 0.0,
    "footprint_spread": 12.92,
    "movement_intensity": 3.51,
    "movement_burst": 75.0,
    "spatial_complexity": 0.95,
    "spatial_entropy": 8.36,
    "temporal_regularity": 0.98,
    "temporal_coverage": 1.0,
    "road_overlap": 0.88,
    "features_json": "{...}",
    "algo_version": "v1",
    "created_at": "2026-02-22T18:20:41Z"
  }
}
```

**响应字段说明**:
- `psi`: 个人空间指数 (0-100)，综合评分
- `behavior_type`: 行为类型，可选值:
  - `EXPLORER`: 探索者（高多样性、广范围）
  - `COMMUTER`: 通勤者（规律性强、固定路线）
  - `HOMEBODY`: 宅家者（低移动、小范围）
  - `TRAVELER`: 旅行者（长距离、低频率）
  - `WANDERER`: 漫游者（不规律、多样化）
  - `BALANCED`: 平衡型（各维度均衡）
- `stability`: 稳定性 (0-1)，行为模式的一致性
- `footprint_diversity`: 足迹多样性，唯一地点数量的归一化值
- `footprint_spread`: 足迹广度（公里），活动范围的空间跨度
- `movement_intensity`: 移动强度，平均每天移动距离
- `movement_burst`: 爆发强度 (0-100)，短时间内大距离移动的频率
- `spatial_complexity`: 空间复杂度 (0-1)，行为模式的复杂程度
- `spatial_entropy`: 空间熵，地点分布的均匀程度
- `temporal_regularity`: 时间规律性 (0-1)，活动时间的规律程度
- `temporal_coverage`: 时间覆盖度 (0-1)，活跃时段的比例
- `road_overlap`: 道路重叠度 (0-1)，重复路线的比例
- `features_json`: 详细特征JSON字符串
- `algo_version`: 算法版本

---

## 错误处理

### 400 Bad Request

参数错误或格式不正确。

```json
{
  "code": 400,
  "message": "Invalid query parameters",
  "error": "limit must be a positive integer"
}
```

### 404 Not Found

请求的资源不存在。

```json
{
  "code": 404,
  "message": "No spatial persona data found",
  "error": null
}
```

### 500 Internal Server Error

服务器内部错误。

```json
{
  "code": 500,
  "message": "Failed to get footprint rankings",
  "error": "database connection error"
}
```

---

## 性能说明

**响应时间**:
- 简单查询（单表）: <100ms
- 复杂聚合（多表JOIN）: <500ms
- 大数据量查询（100k+ points）: <2s

**并发限制**:
- 最大并发: 3 req/s（全局限流）
- 单IP限流: 3 req/s

**缓存策略**:
- 统计结果缓存: 1小时
- 排行榜缓存: 30分钟
- 实时数据: 无缓存

---

## 使用建议

1. **分页查询**: 使用 `limit` 参数控制返回数量，避免一次性获取大量数据
2. **时间过滤**: 使用 `bucket` 参数进行时间分桶，提高查询效率
3. **区域过滤**: 使用 `area_type` 和 `area_key` 参数缩小查询范围
4. **排序优化**: 使用 `order_by` 参数指定排序字段，避免客户端排序
5. **错误重试**: 遇到 500 错误时，建议等待 1-2 秒后重试

---

## 相关文档

- `api-endpoints.md` - 完整API端点列表
- `visualization-api-reference.md` - 可视化API文档
- `admin-api-reference.md` - 管理API文档
- `data-models.md` - 数据模型详细说明
- `api-usage-guide.md` - API使用指南

