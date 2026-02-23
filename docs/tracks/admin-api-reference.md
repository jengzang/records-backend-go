# 管理API参考文档

## 概述

本文档详细说明轨迹分析系统的管理API端点，用于任务管理、数据导入和系统维护。

**基础URL**: `http://localhost:8080/api/v1/admin`

**认证方式**: 当前版本无需认证（未来版本将需要管理员JWT令牌）

**通用响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

---

## 1. 地理编码任务管理

### POST /api/v1/admin/geocoding/tasks

创建地理编码任务，将GPS坐标转换为行政区划信息。

**请求体**:
```json
{
  "task_name": "Geocode 2026-02 data",
  "description": "Geocode newly imported GPS points",
  "batch_size": 1000,
  "max_workers": 4
}
```

**请求字段说明**:
- `task_name` (string, 必填): 任务名称
- `description` (string, 可选): 任务描述
- `batch_size` (integer, 可选): 批处理大小 (默认: 1000)
- `max_workers` (integer, 可选): 最大并发worker数 (默认: 4)

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/geocoding/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "task_name": "Geocode February data",
    "batch_size": 500
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "message": "Geocoding task created successfully",
  "data": {
    "task_id": "gc_20260223_001",
    "status": "PENDING",
    "total_points": 10000,
    "processed_points": 0,
    "success_count": 0,
    "failed_count": 0,
    "progress_percentage": 0.0,
    "created_at": "2026-02-23T10:00:00Z",
    "estimated_duration_seconds": 120
  }
}
```

**响应字段说明**:
- `task_id`: 任务唯一标识
- `status`: 任务状态 (PENDING/RUNNING/COMPLETED/FAILED/CANCELLED)
- `total_points`: 待处理GPS点总数
- `processed_points`: 已处理点数
- `success_count`: 成功编码数量
- `failed_count`: 失败数量
- `progress_percentage`: 进度百分比 (0-100)
- `estimated_duration_seconds`: 预计耗时（秒）

---

### GET /api/v1/admin/geocoding/tasks

获取地理编码任务列表。

**查询参数**:
- `status` (string, 可选): 过滤状态 (PENDING/RUNNING/COMPLETED/FAILED/CANCELLED)
- `limit` (integer, 可选): 返回数量 (默认: 50)
- `offset` (integer, 可选): 分页偏移 (默认: 0)

**请求示例**:
```bash
# 获取所有任务
curl "http://localhost:8080/api/v1/admin/geocoding/tasks"

# 获取运行中的任务
curl "http://localhost:8080/api/v1/admin/geocoding/tasks?status=RUNNING"

# 分页查询
curl "http://localhost:8080/api/v1/admin/geocoding/tasks?limit=20&offset=40"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tasks": [
      {
        "task_id": "gc_20260223_001",
        "task_name": "Geocode February data",
        "status": "RUNNING",
        "total_points": 10000,
        "processed_points": 5000,
        "success_count": 4850,
        "failed_count": 150,
        "progress_percentage": 50.0,
        "created_at": "2026-02-23T10:00:00Z",
        "started_at": "2026-02-23T10:00:05Z",
        "eta_seconds": 60
      },
      {
        "task_id": "gc_20260222_005",
        "task_name": "Geocode January data",
        "status": "COMPLETED",
        "total_points": 50000,
        "processed_points": 50000,
        "success_count": 48500,
        "failed_count": 1500,
        "progress_percentage": 100.0,
        "created_at": "2026-02-22T15:00:00Z",
        "started_at": "2026-02-22T15:00:10Z",
        "completed_at": "2026-02-22T15:10:30Z",
        "duration_seconds": 620
      }
    ],
    "total": 25,
    "limit": 50,
    "offset": 0
  }
}
```

---

### GET /api/v1/admin/geocoding/tasks/:id

获取特定地理编码任务的详细信息。

**路径参数**:
- `id` (string): 任务ID

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/admin/geocoding/tasks/gc_20260223_001"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "gc_20260223_001",
    "task_name": "Geocode February data",
    "description": "Geocode newly imported GPS points",
    "status": "RUNNING",
    "total_points": 10000,
    "processed_points": 7500,
    "success_count": 7275,
    "failed_count": 225,
    "progress_percentage": 75.0,
    "batch_size": 500,
    "max_workers": 4,
    "created_at": "2026-02-23T10:00:00Z",
    "started_at": "2026-02-23T10:00:05Z",
    "updated_at": "2026-02-23T10:01:30Z",
    "eta_seconds": 30,
    "error_message": null,
    "logs": [
      "2026-02-23 10:00:05 - Task started",
      "2026-02-23 10:00:15 - Processed batch 1/20 (500 points)",
      "2026-02-23 10:00:25 - Processed batch 2/20 (500 points)",
      "2026-02-23 10:01:30 - Processed batch 15/20 (500 points)"
    ]
  }
}
```

**响应字段说明**:
- `eta_seconds`: 预计剩余时间（秒）
- `error_message`: 错误信息（如果失败）
- `logs`: 任务日志（最近20条）

---

### DELETE /api/v1/admin/geocoding/tasks/:id

取消正在运行的地理编码任务。

**路径参数**:
- `id` (string): 任务ID

**请求示例**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/geocoding/tasks/gc_20260223_001"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "Geocoding task cancelled successfully",
  "data": {
    "task_id": "gc_20260223_001",
    "status": "CANCELLED",
    "processed_points": 7500,
    "cancelled_at": "2026-02-23T10:02:00Z"
  }
}
```

**注意事项**:
- 只能取消状态为 PENDING 或 RUNNING 的任务
- 已完成或已失败的任务无法取消
- 取消操作会等待当前批次处理完成

---

## 2. 分析任务管理

### POST /api/v1/admin/analysis/tasks

创建轨迹分析任务，执行30个分析技能中的任意一个。

**请求体**:
```json
{
  "skill_name": "footprint_statistics",
  "task_type": "INCREMENTAL",
  "parameters": {
    "threshold_profile": "default",
    "force_recompute": false
  }
}
```

**请求字段说明**:
- `skill_name` (string, 必填): 技能名称，可选值:
  - **Foundation**: `outlier_detection`, `trajectory_completion`, `admin_attribution`
  - **Behavior**: `transport_mode`, `stay_detection`, `trip_construction`, `streak_detection`, `speed_events`
  - **Spatial**: `grid_system`, `road_overlap`, `density_structure`, `speed_space_coupling`, `revisit_pattern`, `utilization_efficiency`, `spatial_complexity`, `directional_bias`
  - **Statistics**: `footprint_statistics`, `stay_statistics`, `extreme_events`, `admin_crossings`, `admin_view_engine`
  - **Advanced**: `time_space_slicing`, `time_space_compression`, `altitude_dimension`
  - **Visualization**: `rendering_metadata`, `time_axis_map`, `stay_annotation`
  - **Integration**: `spatial_persona`
- `task_type` (string, 必填): 任务类型
  - `INCREMENTAL`: 增量分析（仅处理新数据）
  - `FULL_RECOMPUTE`: 全量重算（重新计算所有数据）
- `parameters` (object, 可选): 任务参数
  - `threshold_profile`: 阈值配置文件名称 (默认: "default")
  - `force_recompute`: 强制重算 (默认: false)

**请求示例**:
```bash
# 创建足迹统计任务（增量）
curl -X POST http://localhost:8080/api/v1/admin/analysis/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_name": "footprint_statistics",
    "task_type": "INCREMENTAL"
  }'

# 创建空间画像任务（全量重算）
curl -X POST http://localhost:8080/api/v1/admin/analysis/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_name": "spatial_persona",
    "task_type": "FULL_RECOMPUTE",
    "parameters": {
      "threshold_profile": "strict"
    }
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "message": "Analysis task created successfully",
  "data": {
    "task_id": "at_20260223_001",
    "skill_name": "footprint_statistics",
    "task_type": "INCREMENTAL",
    "status": "PENDING",
    "progress_percentage": 0.0,
    "created_at": "2026-02-23T10:05:00Z",
    "estimated_duration_seconds": 30
  }
}
```

---

### GET /api/v1/admin/analysis/tasks

获取分析任务列表。

**查询参数**:
- `skill_name` (string, 可选): 过滤技能名称
- `status` (string, 可选): 过滤状态
- `limit` (integer, 可选): 返回数量 (默认: 50)
- `offset` (integer, 可选): 分页偏移 (默认: 0)

**请求示例**:
```bash
# 获取所有任务
curl "http://localhost:8080/api/v1/admin/analysis/tasks"

# 获取特定技能的任务
curl "http://localhost:8080/api/v1/admin/analysis/tasks?skill_name=spatial_persona"

# 获取运行中的任务
curl "http://localhost:8080/api/v1/admin/analysis/tasks?status=RUNNING"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tasks": [
      {
        "task_id": "at_20260223_001",
        "skill_name": "footprint_statistics",
        "task_type": "INCREMENTAL",
        "status": "COMPLETED",
        "progress_percentage": 100.0,
        "created_at": "2026-02-23T10:05:00Z",
        "started_at": "2026-02-23T10:05:05Z",
        "completed_at": "2026-02-23T10:05:25Z",
        "duration_seconds": 20,
        "result_summary": "Processed 5 provinces, 20 cities"
      },
      {
        "task_id": "at_20260223_002",
        "skill_name": "spatial_persona",
        "task_type": "FULL_RECOMPUTE",
        "status": "RUNNING",
        "progress_percentage": 65.0,
        "created_at": "2026-02-23T10:10:00Z",
        "started_at": "2026-02-23T10:10:10Z",
        "eta_seconds": 15
      }
    ],
    "total": 50,
    "limit": 50,
    "offset": 0
  }
}
```

---

### GET /api/v1/admin/analysis/tasks/:id

获取特定分析任务的详细信息。

**路径参数**:
- `id` (string): 任务ID

**请求示例**:
```bash
curl "http://localhost:8080/api/v1/admin/analysis/tasks/at_20260223_001"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "at_20260223_001",
    "skill_name": "footprint_statistics",
    "task_type": "INCREMENTAL",
    "status": "COMPLETED",
    "progress_percentage": 100.0,
    "parameters": {
      "threshold_profile": "default",
      "force_recompute": false
    },
    "created_at": "2026-02-23T10:05:00Z",
    "started_at": "2026-02-23T10:05:05Z",
    "completed_at": "2026-02-23T10:05:25Z",
    "duration_seconds": 20,
    "result_summary": "Processed 5 provinces, 20 cities, 50 counties",
    "error_message": null,
    "logs": [
      "2026-02-23 10:05:05 - Task started",
      "2026-02-23 10:05:10 - Processing province: 广东省",
      "2026-02-23 10:05:15 - Processing province: 福建省",
      "2026-02-23 10:05:25 - Task completed successfully"
    ],
    "dependencies": [],
    "blocked_by": []
  }
}
```

**响应字段说明**:
- `result_summary`: 结果摘要
- `dependencies`: 依赖的任务ID列表
- `blocked_by`: 阻塞该任务的任务ID列表

---

### DELETE /api/v1/admin/analysis/tasks/:id

取消正在运行的分析任务。

**路径参数**:
- `id` (string): 任务ID

**请求示例**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/analysis/tasks/at_20260223_002"
```

**响应示例**:
```json
{
  "code": 0,
  "message": "Analysis task cancelled successfully",
  "data": {
    "task_id": "at_20260223_002",
    "status": "CANCELLED",
    "progress_percentage": 65.0,
    "cancelled_at": "2026-02-23T10:12:00Z"
  }
}
```

---

### POST /api/v1/admin/analysis/trigger-chain

触发分析任务链，按依赖关系自动执行多个任务。

**请求体**:
```json
{
  "chain_type": "FULL_PIPELINE",
  "task_type": "INCREMENTAL",
  "skills": [
    "outlier_detection",
    "trajectory_completion",
    "transport_mode",
    "stay_detection",
    "trip_construction",
    "footprint_statistics",
    "spatial_persona"
  ]
}
```

**请求字段说明**:
- `chain_type` (string, 必填): 链类型
  - `FULL_PIPELINE`: 完整流水线（所有30个技能）
  - `FOUNDATION_ONLY`: 仅基础层（4个技能）
  - `BEHAVIOR_ONLY`: 仅行为层（5个技能）
  - `STATISTICS_ONLY`: 仅统计层（5个技能）
  - `CUSTOM`: 自定义技能列表
- `task_type` (string, 必填): 任务类型 (INCREMENTAL/FULL_RECOMPUTE)
- `skills` (array, 可选): 自定义技能列表（仅当chain_type=CUSTOM时需要）

**请求示例**:
```bash
# 触发完整流水线
curl -X POST http://localhost:8080/api/v1/admin/analysis/trigger-chain \
  -H "Content-Type: application/json" \
  -d '{
    "chain_type": "FULL_PIPELINE",
    "task_type": "INCREMENTAL"
  }'

# 触发自定义技能链
curl -X POST http://localhost:8080/api/v1/admin/analysis/trigger-chain \
  -H "Content-Type: application/json" \
  -d '{
    "chain_type": "CUSTOM",
    "task_type": "FULL_RECOMPUTE",
    "skills": ["grid_system", "density_structure", "spatial_complexity"]
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "message": "Analysis chain triggered successfully",
  "data": {
    "chain_id": "chain_20260223_001",
    "chain_type": "FULL_PIPELINE",
    "total_tasks": 30,
    "created_tasks": [
      {
        "task_id": "at_20260223_010",
        "skill_name": "outlier_detection",
        "status": "PENDING",
        "dependencies": []
      },
      {
        "task_id": "at_20260223_011",
        "skill_name": "trajectory_completion",
        "status": "PENDING",
        "dependencies": ["at_20260223_010"]
      },
      {
        "task_id": "at_20260223_012",
        "skill_name": "transport_mode",
        "status": "PENDING",
        "dependencies": ["at_20260223_011"]
      }
    ],
    "estimated_total_duration_seconds": 600
  }
}
```

**响应字段说明**:
- `chain_id`: 任务链唯一标识
- `total_tasks`: 总任务数
- `created_tasks`: 创建的任务列表（包含依赖关系）
- `estimated_total_duration_seconds`: 预计总耗时（秒）

**任务依赖关系**:
```
Foundation Layer:
  outlier_detection → trajectory_completion → admin_attribution

Behavior Layer (depends on Foundation):
  transport_mode → stay_detection → trip_construction
  streak_detection (parallel)
  speed_events (parallel)

Spatial Layer (depends on Behavior):
  grid_system → density_structure
  road_overlap (parallel)
  speed_space_coupling (parallel)
  revisit_pattern (parallel)
  utilization_efficiency (parallel)
  spatial_complexity (parallel)
  directional_bias (parallel)

Statistics Layer (depends on Spatial):
  footprint_statistics (parallel)
  stay_statistics (parallel)
  extreme_events (parallel)
  admin_crossings (parallel)
  admin_view_engine (parallel)

Advanced Layer (depends on Statistics):
  time_space_slicing (parallel)
  time_space_compression (parallel)
  altitude_dimension (parallel)

Visualization Layer (depends on Advanced):
  rendering_metadata (parallel)
  time_axis_map (parallel)
  stay_annotation (parallel)

Integration Layer (depends on all):
  spatial_persona
```

---

## 错误处理

### 400 Bad Request

参数错误或格式不正确。

```json
{
  "code": 400,
  "message": "Invalid request body",
  "error": "skill_name is required"
}
```

### 404 Not Found

任务不存在。

```json
{
  "code": 404,
  "message": "Task not found",
  "error": "task_id: at_20260223_999 does not exist"
}
```

### 409 Conflict

任务冲突（如重复创建）。

```json
{
  "code": 409,
  "message": "Task already exists",
  "error": "A running task for skill 'spatial_persona' already exists"
}
```

### 500 Internal Server Error

服务器内部错误。

```json
{
  "code": 500,
  "message": "Failed to create task",
  "error": "database connection error"
}
```

---

## 使用场景

### 场景1: 首次数据导入后的完整分析

```bash
# 1. 创建地理编码任务
curl -X POST http://localhost:8080/api/v1/admin/geocoding/tasks \
  -H "Content-Type: application/json" \
  -d '{"task_name": "Initial geocoding"}'

# 2. 等待地理编码完成（轮询任务状态）
curl "http://localhost:8080/api/v1/admin/geocoding/tasks/gc_20260223_001"

# 3. 触发完整分析流水线
curl -X POST http://localhost:8080/api/v1/admin/analysis/trigger-chain \
  -H "Content-Type: application/json" \
  -d '{
    "chain_type": "FULL_PIPELINE",
    "task_type": "FULL_RECOMPUTE"
  }'
```

### 场景2: 增量数据更新

```bash
# 1. 地理编码新数据
curl -X POST http://localhost:8080/api/v1/admin/geocoding/tasks \
  -H "Content-Type: application/json" \
  -d '{"task_name": "Geocode new data"}'

# 2. 增量分析（仅处理新数据）
curl -X POST http://localhost:8080/api/v1/admin/analysis/trigger-chain \
  -H "Content-Type: application/json" \
  -d '{
    "chain_type": "FULL_PIPELINE",
    "task_type": "INCREMENTAL"
  }'
```

### 场景3: 单个技能重算

```bash
# 重新计算空间画像
curl -X POST http://localhost:8080/api/v1/admin/analysis/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_name": "spatial_persona",
    "task_type": "FULL_RECOMPUTE"
  }'
```

---

## 性能说明

**地理编码性能**:
- 处理速度: ~1,277 points/sec (with GeoHash cache)
- 成功率: ~97%
- 内存占用: ~200MB (Docker worker)

**分析任务性能**:
- Foundation Layer: ~10k points/sec
- Behavior Layer: ~5k points/sec
- Spatial Layer: ~2k points/sec
- Statistics Layer: ~50k points/sec (aggregation)
- Integration Layer: ~10 sec (spatial_persona)

**并发限制**:
- 最大并发地理编码任务: 1
- 最大并发分析任务: 3
- 任务队列长度: 100

---

## 相关文档

- `statistics-api-reference.md` - 统计API文档
- `visualization-api-reference.md` - 可视化API文档
- `api-usage-guide.md` - API使用指南
- `task-management-framework.md` - 任务管理框架详细说明

