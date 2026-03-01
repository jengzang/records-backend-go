# Screentime Module Quick Start Guide

## 快速测试指南

### 前置条件

1. Go 1.21+ 已安装
2. Node.js 18+ 已安装
3. 数据库文件存在: `go-backend/data/screentime/screentime.db` (17MB)

---

## 步骤1: 启动Backend服务

```bash
cd go-backend
go run main.go
```

**预期输出**:
```
Records Backend API is running on :8080
```

**验证Backend**:
```bash
# 测试健康检查
curl http://localhost:8080/health

# 测试screentime API
curl http://localhost:8080/api/v1/screentime/summary
```

**预期响应** (summary):
```json
{
  "totalApps": 368,
  "totalDurationMS": 13999999999,
  "activeDays": 952,
  "avgDailyDuration": 14705882.35,
  "totalLaunches": 329310,
  "totalNotifications": 178021,
  "topApp": "微信",
  "topAppPackage": "com.tencent.mm",
  "topAppDurationMS": 5371200000,
  "dateRange": {
    "start": "20251113",
    "end": "20260219"
  }
}
```

---

## 步骤2: 启动Frontend服务

**打开新终端**:

```bash
cd screentime
npm install
npm run dev
```

**预期输出**:
```
VITE v5.0.8  ready in 500 ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

---

## 步骤3: 访问前端

打开浏览器访问: `http://localhost:5173`

**预期显示**:
1. ✅ 页面标题: "屏幕使用时间分析"
2. ✅ 4个统计卡片:
   - 总应用数: 368
   - 总使用时长: ~3889小时
   - 活跃天数: 952
   - 日均使用: ~4.1小时
3. ✅ 数据时间范围显示
4. ✅ 最常用应用卡片 (微信)
5. ✅ Top 5应用列表
6. ✅ 最近7天趋势图
7. ✅ 底部统计 (启动次数/通知数)

---

## 步骤4: 测试API端点

### 1. 总体统计
```bash
curl http://localhost:8080/api/v1/screentime/summary | jq
```

### 2. 每日统计 (最近7天)
```bash
curl "http://localhost:8080/api/v1/screentime/daily?limit=7" | jq
```

### 3. 应用排名 (Top 10)
```bash
curl "http://localhost:8080/api/v1/screentime/rankings?limit=10" | jq
```

### 4. 类别统计
```bash
curl http://localhost:8080/api/v1/screentime/categories | jq
```

### 5. 小时分布
```bash
curl http://localhost:8080/api/v1/screentime/hourly | jq
```

### 6. 趋势分析 (每日)
```bash
curl "http://localhost:8080/api/v1/screentime/trends?granularity=daily" | jq
```

### 7. 单应用详情 (微信)
```bash
curl http://localhost:8080/api/v1/screentime/app/com.tencent.mm | jq
```

---

## 常见问题

### 问题1: Backend启动失败

**错误**: `failed to open database`

**解决**:
```bash
# 检查数据库文件是否存在
ls -lh go-backend/data/screentime/screentime.db

# 如果不存在,运行导入脚本
cd go-backend/scripts/screentime
python import_screentime.py
```

### 问题2: Frontend无法连接Backend

**错误**: `加载数据失败,请确保后端服务正在运行`

**解决**:
1. 确认Backend正在运行 (端口8080)
2. 检查CORS配置
3. 检查API_BASE URL (应该是 `http://localhost:8080`)

### 问题3: 前端显示空白

**错误**: 控制台显示 `Cannot find module`

**解决**:
```bash
cd screentime
rm -rf node_modules package-lock.json
npm install
npm run dev
```

### 问题4: API返回空数据

**错误**: `totalApps: 0`

**解决**:
```bash
# 检查数据库记录数
cd go-backend
go run -tags="fts5" main.go

# 或者重新导入数据
cd scripts/screentime
python import_screentime.py
```

---

## 验证清单

- [ ] Backend服务启动成功 (端口8080)
- [ ] Frontend服务启动成功 (端口5173)
- [ ] 浏览器能访问 http://localhost:5173
- [ ] 页面显示统计数据 (不是0)
- [ ] Top 5应用列表有数据
- [ ] 趋势图有数据点
- [ ] 无控制台错误
- [ ] API端点全部返回有效数据

---

## 下一步

测试成功后,可以:

1. **添加更多页面**: Rankings, Trends, Patterns
2. **添加图表库**: recharts 或 chart.js
3. **实现路由**: React Router
4. **ManicTime数据**: 导入电脑端数据
5. **设备切换**: 实现多设备支持

---

## 性能测试

### Backend性能
```bash
# 使用ab (Apache Bench) 测试
ab -n 1000 -c 10 http://localhost:8080/api/v1/screentime/summary
```

**预期结果**:
- 请求成功率: 100%
- 平均响应时间: < 100ms
- 吞吐量: > 100 req/s

### Frontend性能
- 首次加载: < 2s
- API请求: < 500ms
- 页面渲染: < 100ms

---

## 日志查看

### Backend日志
Backend会输出请求日志到控制台:
```
[GIN] 2026/02/24 - 10:30:00 | 200 |   50.123ms |   127.0.0.1 | GET  /api/v1/screentime/summary
```

### Frontend日志
打开浏览器开发者工具 (F12) 查看:
- Console: 错误和警告
- Network: API请求和响应
- Performance: 性能分析

---

## 成功标准

✅ 所有API端点返回有效数据
✅ Frontend正确显示统计信息
✅ 无控制台错误
✅ 响应时间 < 500ms
✅ 移动端布局正常

---

## 联系支持

如果遇到问题:
1. 检查日志输出
2. 查看 IMPLEMENTATION_SUMMARY.md
3. 查看 README.md
4. 提交Issue到GitHub
