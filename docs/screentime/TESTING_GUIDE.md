# 快速测试指南

## 启动服务器

### 方法1: 使用其他端口(推荐)
```bash
cd go-backend
PORT=:8081 ./bin/server.exe
```

### 方法2: 释放端口8080
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <进程ID> /F

# 然后启动
./bin/server.exe
```

---

## API测试命令

### 基础健康检查
```bash
curl http://localhost:8081/health
# 预期: {"status":"ok"}
```

### Screentime基础API
```bash
# 总体统计
curl http://localhost:8081/api/v1/screentime/summary

# 应用排名
curl http://localhost:8081/api/v1/screentime/rankings?limit=10

# 应用分类
curl http://localhost:8081/api/v1/screentime/categories
```

### 跨设备分析API
```bash
# 设备使用对比
curl http://localhost:8081/api/v1/screentime/cross-device/comparison

# 工作生活平衡
curl http://localhost:8081/api/v1/screentime/cross-device/work-life-balance

# 生产力深度分析(验证应用分类)
curl http://localhost:8081/api/v1/screentime/cross-device/productivity-deep

# 应用生态分析(验证跨平台识别)
curl http://localhost:8081/api/v1/screentime/cross-device/app-ecosystem
```

---

## 日志验证

### 启动时日志
应该看到:
```json
{"level":"info","msg":"Logger initialized",...}
{"level":"info","msg":"Starting Records Backend API","port":":8081","version":"1.0.0",...}
{"level":"info","msg":"Database initialized successfully","db_path":"...",...}
{"level":"info","msg":"Initializing device manager",...}
{"level":"info","msg":"Device loaded successfully","device_id":"phone_vivo_x90",...}
{"level":"info","msg":"Device loaded successfully","device_id":"computer_main",...}
{"level":"info","msg":"Server starting","port":":8081",...}
```

### HTTP请求日志
每次API调用应该看到:
```json
{"level":"info","msg":"HTTP Request","method":"GET","path":"/api/v1/screentime/summary","status":200,"duration_ms":45,"client_ip":"127.0.0.1",...}
```

### 调试模式
```bash
LOG_LEVEL=debug ./bin/server.exe
# 会看到更详细的日志,包括设备查询等
```

---

## 应用分类验证

### 检查分类分布
```bash
curl http://localhost:8081/api/v1/screentime/categories | jq
```

预期看到11个类别:
- Productivity
- Development
- Communication
- Entertainment
- Social
- Shopping
- Browser
- Tools
- Office
- Gaming
- System
- Other

### 检查生产力应用
```bash
curl http://localhost:8081/api/v1/screentime/cross-device/productivity-deep | jq '.phoneTopApps'
```

预期看到生产力相关应用,如:
- Microsoft Edge
- WPS Office
- 记账应用
- 等等

---

## 跨平台应用识别验证

```bash
curl http://localhost:8081/api/v1/screentime/cross-device/app-ecosystem | jq '.crossPlatformApps'
```

预期看到跨平台应用列表,如:
- Chrome/Google Chrome
- Edge/Microsoft Edge
- WeChat/微信
- QQ
- 等等

**不应该返回null!**

---

## 常见问题

### Q: 端口8080被占用怎么办?
A: 使用PORT=:8081启动,或者释放8080端口

### Q: 日志没有JSON格式?
A: 检查logger是否正确初始化,查看启动日志

### Q: API返回500错误?
A: 检查日志中的error级别消息,查看具体错误

### Q: 应用分类不准确?
A: 编辑`internal/screentime/app_category_mapper.go`,添加更多关键词

### Q: 跨平台应用识别不准确?
A: 编辑`internal/screentime/app_normalizer.go`,添加更多映射规则

---

## 性能检查

### 响应时间
所有API应该在500ms内响应:
```bash
time curl http://localhost:8081/api/v1/screentime/summary
```

### 内存使用
```bash
# Windows任务管理器查看server.exe进程
# 应该 < 1GB
```

---

## 成功标准

- ✅ 服务器正常启动
- ✅ /health返回200
- ✅ 所有API返回有效JSON
- ✅ 日志格式正确(JSON)
- ✅ 应用分类合理
- ✅ 跨平台应用识别正确
- ✅ 响应时间 < 500ms

---

**更新时间**: 2026-03-01
