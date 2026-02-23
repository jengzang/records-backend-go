# 可视化API参考文档

## 概述

本文档详细说明轨迹分析系统的可视化API端点，用于地图渲染、热力图生成和时间轴可视化。

**基础URL**: `http://localhost:8080/api/v1/viz`

**认证方式**: 当前版本无需认证（公开只读访问）

**通用响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

---

## 1. 网格单元查询

### GET /api/v1/viz/grid-cells

获取指定区域的网格单元数据，用于地图可视化。

**查询参数**:
- `level` (integer): 网格级别 (1-15, 默认: 3)
  - Level 1: 省级 (~100km cells)
  - Level 2: 市级 (~50km cells)
  - Level 3: 区级 (~10km cells)
  - Level 4: 镇级 (~5km cells)
  - Level 5: 路级 (~1km cells)
  - Level 6-15: 更精细级别
- `minLat` (float): 最小纬度（边界框）
- `maxLat` (float): 最大纬度（边界框）
- `minLon` (float): 最小经度（边界框）
- `maxLon` (float): 最大经度（边界框）
- `minDensity` (integer): 最小点密度过滤（可选）

**请求示例**:
```bash
# 获取深圳地区的Level 3网格
curl "http://localhost:8080/api/v1/viz/grid-cells?level=3&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"

# 过滤低密度网格
curl "http://localhost:8080/api/v1/viz/grid-cells?level=5&minLat=22.5&maxLat=22.6&minLon=114.0&maxLon=114.1&minDensity=100"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [
      {
        "grid_id": "L3_1234_5678",
        "level": 3,
        "center_lat": 22.5432,
        "center_lon": 113.9234,
        "min_lat": 22.5400,
        "max_lat": 22.5464,
        "min_lon": 113.9200,
        "max_lon": 113.9268,
        "point_count": 1250,
        "visit_count": 45,
        "total_duration_seconds": 86400,
        "first_visit": 1704067200,
        "last_visit": 1737552138,
        "modes_json": "[\"walk\",\"bike\",\"car\"]",
        "created_at": "2026-02-22T00:00:00Z",
        "updated_at": "2026-02-22T00:00:00Z"
      },
      {
        "grid_id": "L3_1234_5679",
        "level": 3,
        "center_lat": 22.5496,
        "center_lon": 113.9234,
        "min_lat": 22.5464,
        "max_lat": 22.5528,
        "min_lon": 113.9200,
        "max_lon": 113.9268,
        "point_count": 850,
        "visit_count": 30,
        "total_duration_seconds": 43200,
        "first_visit": 1704153600,
        "last_visit": 1737465738,
        "modes_json": "[\"car\"]",
        "created_at": "2026-02-22T00:00:00Z",
        "updated_at": "2026-02-22T00:00:00Z"
      }
    ],
    "count": 150
  }
}
```

**响应字段说明**:
- `grid_id`: 网格唯一标识（格式: L{level}_{x}_{y}）
- `level`: 网格级别 (1-15)
- `center_lat/center_lon`: 网格中心坐标
- `min_lat/max_lat/min_lon/max_lon`: 网格边界框
- `point_count`: 网格内GPS点数量
- `visit_count`: 访问次数（去重后的访问）
- `total_duration_seconds`: 总停留时长（秒）
- `first_visit/last_visit`: 首次/最后访问时间 (Unix timestamp)
- `modes_json`: 交通方式JSON数组字符串

**使用场景**:
- 地图瓦片渲染
- 区域热力图
- 活动范围可视化
- 密度分析

**性能说明**:
- 最大返回10,000个网格单元
- 建议使用边界框限制查询范围
- Level 1-3适合全局视图，Level 4-7适合区域视图，Level 8+适合街道级视图

---

## 2. 热力图数据

### GET /api/v1/viz/heatmap

获取热力图数据，返回归一化的强度值用于可视化。

**查询参数**:
- `level` (integer): 网格级别 (1-15, 默认: 3)
- `metric` (string): 度量指标 (默认: `point_count`)
  - `point_count`: GPS点数量
  - `duration`: 停留时长
  - `visit_count`: 访问次数
- `minLat` (float): 最小纬度（边界框）
- `maxLat` (float): 最大纬度（边界框）
- `minLon` (float): 最小经度（边界框）
- `maxLon` (float): 最大经度（边界框）

**请求示例**:
```bash
# 基础热力图（点密度）
curl "http://localhost:8080/api/v1/viz/heatmap?level=3&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"

# 停留时长热力图
curl "http://localhost:8080/api/v1/viz/heatmap?level=3&metric=duration&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"

# 访问频率热力图
curl "http://localhost:8080/api/v1/viz/heatmap?level=3&metric=visit_count&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "points": [
      {
        "lat": 22.5432,
        "lng": 113.9234,
        "intensity": 0.85,
        "value": 1250,
        "metric": "point_count"
      },
      {
        "lat": 22.5567,
        "lng": 113.9456,
        "intensity": 0.42,
        "value": 620,
        "metric": "point_count"
      },
      {
        "lat": 22.5123,
        "lng": 113.8901,
        "intensity": 1.0,
        "value": 1500,
        "metric": "point_count"
      }
    ],
    "count": 150,
    "max_value": 1500,
    "min_value": 10,
    "metric": "point_count",
    "grid_level": 3
  }
}
```

**响应字段说明**:
- `points`: 热力图点数组
  - `lat/lng`: 点坐标
  - `intensity`: 归一化强度值 (0-1)，用于颜色映射
  - `value`: 原始数值
  - `metric`: 度量指标名称
- `count`: 返回点数量
- `max_value/min_value`: 数值范围（用于归一化）
- `metric`: 使用的度量指标
- `grid_level`: 网格级别

**前端集成示例** (Leaflet.js):
```javascript
// 使用Leaflet.heat插件
fetch('http://localhost:8080/api/v1/viz/heatmap?level=3&minLat=22.0&maxLat=23.5&minLon=113.0&maxLon=114.5')
  .then(res => res.json())
  .then(data => {
    const heatData = data.data.points.map(p => [p.lat, p.lng, p.intensity]);
    L.heatLayer(heatData, {
      radius: 25,
      blur: 15,
      maxZoom: 17,
      max: 1.0,
      gradient: {
        0.0: 'blue',
        0.5: 'lime',
        1.0: 'red'
      }
    }).addTo(map);
  });
```

**使用场景**:
- 活动热力图
- 停留时长可视化
- 访问频率分析
- 区域对比

**性能说明**:
- 最大返回10,000个点
- 强度值已归一化，无需客户端计算
- 建议使用边界框限制查询范围
- 空结果表示该区域无网格数据（需先运行grid_system分析）

---

## 3. 渲染元数据

### GET /api/v1/viz/rendering

获取轨迹渲染元数据，用于地图上的轨迹线条渲染。

**查询参数**:
- `start_time` (integer): 开始时间 (Unix timestamp, 可选)
- `end_time` (integer): 结束时间 (Unix timestamp, 可选)
- `lod` (integer): 细节级别 (Level of Detail, 1-3, 默认: 2)
  - LOD 1: 低细节（仅主要路段）
  - LOD 2: 中等细节（常用路段）
  - LOD 3: 高细节（所有路段）
- `minLat` (float): 最小纬度（边界框，可选）
- `maxLat` (float): 最大纬度（边界框，可选）
- `minLon` (float): 最小经度（边界框，可选）
- `maxLon` (float): 最大经度（边界框，可选）
- `limit` (integer): 返回数量 (默认: 1000)

**请求示例**:
```bash
# 获取中等细节渲染数据
curl "http://localhost:8080/api/v1/viz/rendering?lod=2&limit=500"

# 获取特定时间范围的渲染数据
curl "http://localhost:8080/api/v1/viz/rendering?start_time=1704067200&end_time=1737552138&lod=3"

# 获取特定区域的渲染数据
curl "http://localhost:8080/api/v1/viz/rendering?minLat=22.5&maxLat=22.6&minLon=114.0&maxLon=114.1&lod=2"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "segments": [
      {
        "segment_id": 12345,
        "start_time": 1737552138,
        "end_time": 1737555738,
        "start_lat": 22.5431,
        "start_lon": 114.0579,
        "end_lat": 22.5567,
        "end_lon": 114.0723,
        "transport_mode": "car",
        "speed_bucket": 3,
        "overlap_rank": 5,
        "line_weight": 2.5,
        "alpha_hint": 0.7,
        "color_hint": "#FF5733",
        "lod_level": 2
      },
      {
        "segment_id": 12346,
        "start_time": 1737555738,
        "end_time": 1737559338,
        "start_lat": 22.5567,
        "start_lon": 114.0723,
        "end_lat": 22.5789,
        "end_lon": 114.0901,
        "transport_mode": "walk",
        "speed_bucket": 0,
        "overlap_rank": 1,
        "line_weight": 1.0,
        "alpha_hint": 0.3,
        "color_hint": "#33FF57",
        "lod_level": 3
      }
    ],
    "count": 500,
    "lod_level": 2
  }
}
```

**响应字段说明**:
- `segment_id`: 路段唯一标识
- `start_time/end_time`: 路段起止时间 (Unix timestamp)
- `start_lat/start_lon`: 起点坐标
- `end_lat/end_lon`: 终点坐标
- `transport_mode`: 交通方式 (walk/bike/car/train/plane)
- `speed_bucket`: 速度分桶 (0-5)，基于全局百分位数
  - 0: 0-20% (最慢)
  - 1: 20-40%
  - 2: 40-60%
  - 3: 60-80%
  - 4: 80-95%
  - 5: 95-100% (最快)
- `overlap_rank`: 重叠等级 (1-10)，表示该路段被重复经过的次数
- `line_weight`: 线条粗细建议 (1.0-3.0)
- `alpha_hint`: 透明度建议 (0.3-1.0)
- `color_hint`: 颜色建议（十六进制）
- `lod_level`: 该路段的细节级别

**前端集成示例** (Leaflet.js):
```javascript
fetch('http://localhost:8080/api/v1/viz/rendering?lod=2&limit=500')
  .then(res => res.json())
  .then(data => {
    data.data.segments.forEach(seg => {
      const polyline = L.polyline(
        [[seg.start_lat, seg.start_lon], [seg.end_lat, seg.end_lon]],
        {
          color: seg.color_hint,
          weight: seg.line_weight,
          opacity: seg.alpha_hint
        }
      ).addTo(map);

      polyline.bindPopup(`
        Mode: ${seg.transport_mode}<br>
        Speed Bucket: ${seg.speed_bucket}<br>
        Overlap: ${seg.overlap_rank}x
      `);
    });
  });
```

**使用场景**:
- 轨迹线条渲染
- 速度可视化（颜色编码）
- 重复路线高亮
- 多级细节渲染（LOD）

**性能说明**:
- LOD 1: ~100-500 segments（快速预览）
- LOD 2: ~500-2000 segments（常规使用）
- LOD 3: ~2000-10000 segments（详细分析）
- 建议根据地图缩放级别动态调整LOD

---

## 4. 时间切片数据

### GET /api/v1/viz/time-slices

获取时间轴切片数据，用于时间轴地图可视化。

**查询参数**:
- `slice_type` (string): 切片类型 (默认: `HOURLY`)
  - `HOURLY`: 按小时切片（24个切片）
  - `DAILY`: 按日期切片
  - `WEEKLY`: 按星期切片（7个切片）
  - `MONTHLY`: 按月份切片（12个切片）
- `start_time` (integer): 开始时间 (Unix timestamp, 可选)
- `end_time` (integer): 结束时间 (Unix timestamp, 可选)
- `limit` (integer): 返回数量 (默认: 100)

**请求示例**:
```bash
# 获取每小时活动分布
curl "http://localhost:8080/api/v1/viz/time-slices?slice_type=HOURLY"

# 获取每周活动分布
curl "http://localhost:8080/api/v1/viz/time-slices?slice_type=WEEKLY"

# 获取特定时间范围的日活动分布
curl "http://localhost:8080/api/v1/viz/time-slices?slice_type=DAILY&start_time=1704067200&end_time=1737552138"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "slices": [
      {
        "slice_key": "08:00-09:00",
        "slice_index": 8,
        "point_count": 5000,
        "unique_locations": 150,
        "avg_speed": 35.5,
        "dominant_mode": "car",
        "top_provinces": ["广东省", "福建省"],
        "top_cities": ["深圳市", "广州市", "厦门市"]
      },
      {
        "slice_key": "09:00-10:00",
        "slice_index": 9,
        "point_count": 3500,
        "unique_locations": 80,
        "avg_speed": 15.2,
        "dominant_mode": "walk",
        "top_provinces": ["广东省"],
        "top_cities": ["深圳市"]
      }
    ],
    "slice_type": "HOURLY",
    "count": 24
  }
}
```

**响应字段说明**:
- `slice_key`: 切片标识（如"08:00-09:00", "Monday", "2026-02"）
- `slice_index`: 切片索引（0-23 for HOURLY, 0-6 for WEEKLY, etc.）
- `point_count`: 该时间段的GPS点数量
- `unique_locations`: 唯一地点数量
- `avg_speed`: 平均速度 (m/s)
- `dominant_mode`: 主要交通方式
- `top_provinces`: 主要活动省份（最多3个）
- `top_cities`: 主要活动城市（最多3个）

**前端集成示例** (Chart.js):
```javascript
fetch('http://localhost:8080/api/v1/viz/time-slices?slice_type=HOURLY')
  .then(res => res.json())
  .then(data => {
    const ctx = document.getElementById('timeChart').getContext('2d');
    new Chart(ctx, {
      type: 'bar',
      data: {
        labels: data.data.slices.map(s => s.slice_key),
        datasets: [{
          label: 'Activity Level',
          data: data.data.slices.map(s => s.point_count),
          backgroundColor: 'rgba(54, 162, 235, 0.5)'
        }]
      },
      options: {
        scales: {
          y: { beginAtZero: true }
        }
      }
    });
  });
```

**使用场景**:
- 时间轴地图
- 活动模式分析
- 时段对比
- 周期性检测

**性能说明**:
- HOURLY: 返回24个切片
- WEEKLY: 返回7个切片
- MONTHLY: 返回12个切片
- DAILY: 返回实际天数（受时间范围限制）

---

## 错误处理

### 400 Bad Request

参数错误或格式不正确。

```json
{
  "code": 400,
  "message": "Invalid query parameters",
  "error": "level must be between 1 and 15"
}
```

### 404 Not Found

请求的资源不存在或无数据。

```json
{
  "code": 404,
  "message": "No grid data found",
  "error": "Run grid_system analyzer first"
}
```

### 500 Internal Server Error

服务器内部错误。

```json
{
  "code": 500,
  "message": "Failed to get heatmap data",
  "error": "database query error"
}
```

---

## 性能优化建议

1. **使用边界框**: 始终提供 `minLat/maxLat/minLon/maxLon` 参数限制查询范围
2. **选择合适的网格级别**:
   - 全局视图: Level 1-3
   - 区域视图: Level 4-7
   - 街道视图: Level 8-12
3. **LOD动态调整**: 根据地图缩放级别动态调整LOD参数
4. **分页加载**: 使用 `limit` 参数分批加载数据
5. **缓存策略**: 可视化数据变化较少，建议客户端缓存

---

## 相关文档

- `statistics-api-reference.md` - 统计API文档
- `admin-api-reference.md` - 管理API文档
- `api-usage-guide.md` - API使用指南
- `data-models.md` - 数据模型详细说明

