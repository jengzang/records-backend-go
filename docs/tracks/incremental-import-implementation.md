# 增量数据导入功能实现文档

## 概述

实现了完整的增量数据导入功能，支持通过Web API上传CSV/Excel文件，自动去重，并可选择性触发后续的地理编码和分析流程。

## 实施内容

### Phase 1: 修改导入脚本支持增量模式 ✅

**修改文件**: `go-backend/scripts/tracks/import/write2sql.py`

**新增功能**:
1. **增量模式支持**: 默认使用 `append` 模式，避免全量替换导致数据丢失
2. **去重逻辑**: 基于 `dataTime` 字段自动去重，避免重复导入
3. **CSV格式支持**: 除Excel外，新增CSV文件格式支持
4. **命令行接口**: 支持非交互式调用，便于Go后端集成
5. **统计信息**: 返回总记录数、新增记录数、重复记录数

**命令行参数**:
```bash
python write2sql.py --help

选项:
  --file FILE           数据文件路径 (Excel或CSV)
  --db DB               数据库路径 (默认: data/tracks.db)
  --table TABLE         表名 (默认: 一生足迹)
  --sheet SHEET         Excel sheet名称
  --mode {append,replace}  导入模式 (默认: append)
  --deduplicate {true,false}  是否去重 (默认: true)
  --gui                 使用GUI模式选择文件
```

**使用示例**:
```bash
# 增量导入CSV文件（自动去重）
python write2sql.py --file data.csv --mode append --deduplicate true

# 全量替换模式
python write2sql.py --file data.xlsx --mode replace

# GUI模式（保持向后兼容）
python write2sql.py --gui
```

### Phase 2: 实现数据导入Web API ✅

**新建文件**:
- `go-backend/internal/models/import_task.go` - 导入任务模型
- `go-backend/internal/service/import_service.go` - 导入服务
- `go-backend/internal/handler/import_handler.go` - 导入处理器
- `go-backend/scripts/tracks/migrations/022_create_import_tasks_table.sql` - 数据库迁移

**API端点**:

#### 1. 上传数据文件
```
POST /api/v1/admin/tracks/import
Content-Type: multipart/form-data

参数:
- file: 文件 (必需，支持 .csv, .xlsx, .xls, .xlsm)
- mode: 导入模式 (可选，默认 "append"，可选 "replace")
- deduplicate: 是否去重 (可选，默认 "true")
- auto_trigger: 是否自动触发地理编码 (可选，默认 "true")

响应:
{
  "task_id": 1,
  "status": "pending",
  "message": "Import task created successfully"
}
```

#### 2. 查询导入任务状态
```
GET /api/v1/admin/tracks/import/:id

响应:
{
  "id": 1,
  "status": "completed",
  "file_path": "/tmp/uploads/1_20260223_150405_data.csv",
  "file_name": "data.csv",
  "file_size": 1024000,
  "mode": "append",
  "deduplicate": true,
  "auto_trigger": true,
  "total_records": 1000,
  "new_records": 850,
  "duplicate_records": 150,
  "created_at": "2026-02-23T15:04:05Z",
  "updated_at": "2026-02-23T15:05:30Z",
  "completed_at": "2026-02-23T15:05:30Z"
}
```

**任务状态**:
- `pending`: 等待执行
- `running`: 正在导入
- `completed`: 导入完成
- `failed`: 导入失败

### 数据库Schema

**import_tasks表**:
```sql
CREATE TABLE import_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    status TEXT NOT NULL DEFAULT 'pending',
    file_path TEXT,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mode TEXT NOT NULL DEFAULT 'append',
    deduplicate INTEGER NOT NULL DEFAULT 1,
    auto_trigger INTEGER NOT NULL DEFAULT 1,
    total_records INTEGER DEFAULT 0,
    new_records INTEGER DEFAULT 0,
    duplicate_records INTEGER DEFAULT 0,
    error_message TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    completed_at DATETIME
);
```

## 使用流程

### 1. 通过API上传数据

```bash
curl -X POST http://localhost:8080/api/v1/admin/tracks/import \
  -F "file=@new_tracks.csv" \
  -F "mode=append" \
  -F "deduplicate=true" \
  -F "auto_trigger=true"
```

### 2. 轮询任务状态

```bash
curl http://localhost:8080/api/v1/admin/tracks/import/1
```

### 3. 查看导入结果

任务完成后，响应中会包含:
- `total_records`: 文件中的总记录数
- `new_records`: 实际导入的新记录数
- `duplicate_records`: 被去重的记录数

## 配置

### 环境变量

- `UPLOAD_DIR`: 上传文件存储目录 (默认: `/tmp/uploads`)
- `TRACKS_DB_PATH`: 轨迹数据库路径 (默认: `data/tracks.db`)

### 文件大小限制

建议在Gin配置中设置文件大小限制:
```go
router.MaxMultipartMemory = 100 << 20  // 100 MB
```

## 测试

### 1. 测试Python脚本

```bash
# 测试增量导入
cd go-backend/scripts/tracks
python import/write2sql.py --file test_data.csv --mode append --deduplicate true

# 查看输出的JSON统计信息
```

### 2. 测试API端点

```bash
# 启动Go后端
cd go-backend
go run cmd/server/main.go

# 上传测试文件
curl -X POST http://localhost:8080/api/v1/admin/tracks/import \
  -F "file=@test_data.csv"

# 查询任务状态
curl http://localhost:8080/api/v1/admin/tracks/import/1
```

## 下一步计划

### Phase 3: 实现端到端自动触发流水线 (P1) ✅

**目标**: 导入完成后自动触发地理编码和分析链

**已实施内容**:
1. ✅ 修改 `ImportService` 添加服务依赖注入
2. ✅ 实现 `triggerPipeline()` 方法在导入完成后自动触发地理编码
3. ✅ 实现 `waitAndTriggerAnalysis()` 方法轮询地理编码状态
4. ✅ 实现 `triggerAnalysisChain()` 方法触发完整分析链
5. ✅ 在路由中注入 `GeocodingService` 和 `AnalysisTaskService` 依赖
6. ✅ 创建统一流水线API端点 `POST /api/v1/admin/pipeline/trigger`

**流水线设计**:
```
Import Task → Geocoding Task → Analysis Chain (8 skills)
   ↓              ↓                  ↓
 pending      pending           pending
 running      running           running
completed    completed         completed
```

**自动触发逻辑**:
1. 导入任务完成后，检查 `auto_trigger` 标志和 `new_records > 0`
2. 如果启用，创建地理编码任务
3. 后台轮询地理编码任务状态（每10秒检查一次，最多等待30分钟）
4. 地理编码完成后，自动触发分析任务链
5. 分析链包含8个核心技能（按依赖顺序执行）

**API端点**:

#### 统一流水线触发
```
POST /api/v1/admin/pipeline/trigger
Content-Type: multipart/form-data

参数:
- file: 文件 (必需)
- mode: 导入模式 (可选，默认 "append")
- deduplicate: 是否去重 (可选，默认 "true")

响应:
{
  "task_id": 1,
  "status": "pending",
  "message": "Import task created successfully. Pipeline will be triggered automatically after import completes."
}
```

**分析技能执行顺序**:
1. `outlier_detection` - 异常检测
2. `transport_mode` - 交通方式分类
3. `stay_detection` - 停留检测
4. `trip_construction` - 行程构建
5. `grid_system` - 网格系统
6. `footprint_statistics` - 足迹统计
7. `stay_statistics` - 停留统计
8. `rendering_metadata` - 渲染元数据

**轮询机制**:
- 轮询间隔：10秒
- 最大等待时间：30分钟
- 状态检查：pending → running → completed/failed

**错误处理**:
- 地理编码失败：中止流水线，不触发分析
- 地理编码超时：中止流水线
- 分析任务创建失败：记录错误日志，继续执行其他任务

### Phase 4: 实现前端上传界面 (P2)

**目标**: 在React前端添加文件上传组件

**实施内容**:
1. 创建文件上传组件 (`FileUpload.tsx`)
2. 创建管理页面 (`DataImport.tsx`)
3. 实现拖拽上传功能
4. 实现任务监控面板
5. 显示实时进度和统计信息

## 关键文件清单

**已修改的文件**:
- `go-backend/scripts/tracks/import/write2sql.py` - 增量模式和CSV支持
- `go-backend/internal/api/router.go` - 添加导入路由

**已创建的文件**:
- `go-backend/internal/models/import_task.go`
- `go-backend/internal/service/import_service.go`
- `go-backend/internal/handler/import_handler.go`
- `go-backend/scripts/tracks/migrations/022_create_import_tasks_table.sql`

## 验证清单

- [x] Python脚本支持增量模式
- [x] Python脚本支持CSV格式
- [x] Python脚本支持命令行参数
- [x] Python脚本返回JSON统计信息
- [x] 数据库迁移创建import_tasks表
- [x] Go服务实现ImportService
- [x] Go服务实现ImportHandler
- [x] 路由添加导入端点
- [x] 自动触发地理编码（Phase 3）
- [x] 自动触发分析链（Phase 3）
- [x] 统一流水线API端点（Phase 3）
- [ ] 端到端测试（需要启动Go服务）
- [ ] 前端上传界面（Phase 4）

## 性能考虑

- **批量处理**: pandas批量读取和插入，性能优于逐行处理
- **去重优化**: 使用集合（set）进行内存去重，避免重复查询数据库
- **后台执行**: Go服务使用goroutine后台执行导入，避免阻塞API响应
- **文件清理**: 建议定期清理 `/tmp/uploads` 目录中的旧文件

## 故障排查

### 问题1: Python脚本执行失败

**检查**:
- Python3是否安装
- pandas库是否安装 (`pip install pandas openpyxl`)
- 文件路径是否正确
- 数据库文件是否存在

### 问题2: 去重不生效

**检查**:
- `deduplicate` 参数是否设置为 `true`
- 数据文件中是否包含 `dataTime` 字段
- `dataTime` 字段值是否为有效的Unix时间戳

### 问题3: 导入任务一直处于pending状态

**检查**:
- Go服务是否正常运行
- Python脚本路径是否正确
- 查看Go服务日志中的错误信息

## 更新日志

### 2026-02-23
- ✅ 实现Phase 1: 修改导入脚本支持增量模式
- ✅ 实现Phase 2: 实现数据导入Web API
- ✅ 创建数据库迁移022
- ✅ 更新路由配置
- ✅ 实现Phase 3: 端到端自动触发流水线
  - 添加服务依赖注入机制
  - 实现自动触发地理编码
  - 实现轮询机制等待地理编码完成
  - 实现自动触发分析链
  - 创建统一流水线API端点
- 📝 编写实现文档
