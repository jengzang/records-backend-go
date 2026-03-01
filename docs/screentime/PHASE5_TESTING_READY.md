# Phase 5: 功能测试准备完成

**日期**: 2026-03-01
**状态**: ✅ 服务器已启动,等待用户测试
**耗时**: 30分钟

---

## 服务器状态

### 后端服务器
- **状态**: ✅ 运行中
- **端口**: 18080
- **URL**: http://localhost:18080
- **日志**: /tmp/backend_final.log

**启动信息**:
```
[GIN-debug] Listening and serving HTTP on :18080
Server starting on port :18080
```

**已注册的ScreenTime API端点** (18个):
- 7个单设备API
- 1个设备管理API
- 10个跨设备分析API

### 前端服务器
- **状态**: ✅ 运行中
- **端口**: 5176
- **URL**: http://localhost:5176/screentime/
- **日志**: /tmp/frontend.log

**启动信息**:
```
VITE v5.4.21 ready in 2127 ms
➜  Local:   http://localhost:5176/screentime/
```

---

## 配置修改

### 1. 后端端口配置
**文件**: `go-backend/internal/config/config.go`
**修改**: 默认端口从8080改为18080
```go
// 修改前
port = ":8080"

// 修改后
port = ":18080"
```

**原因**: Windows权限限制,8080端口无法绑定

### 2. 前端API配置
**文件**: `screentime/src/services/screentimeApi.ts`
**修改**: API_BASE保持18080端口
```typescript
// 最终配置
const API_BASE = 'http://localhost:18080/api/v1/screentime';
```

---

## 测试清单

### 基础页面测试 (4个)

#### 1. Home页面
**URL**: http://localhost:5176/screentime/
**测试项**:
- [ ] 统计卡片显示正确(总时长、应用数、日均使用)
- [ ] Top 5应用列表加载
- [ ] 数据格式化正确(小时/分钟)
- [ ] 设备选择器工作正常

#### 2. Rankings页面
**URL**: http://localhost:5176/screentime/rankings
**测试项**:
- [ ] 应用排名表格显示
- [ ] 排序功能(按时长/启动次数/通知数)
- [ ] 类别筛选功能
- [ ] 分页功能

#### 3. Trends页面
**URL**: http://localhost:5176/screentime/trends
**测试项**:
- [ ] 趋势图表显示
- [ ] 时间粒度切换(日/周/月)
- [ ] 日期范围筛选
- [ ] 图表交互功能

#### 4. Categories页面
**URL**: http://localhost:5176/screentime/categories
**测试项**:
- [ ] 9个类别卡片显示
- [ ] 类别统计数据正确
- [ ] 百分比计算准确
- [ ] 类别图标显示

---

### 跨设备页面测试 (5个)

#### 5. CrossDeviceAnalysis页面
**URL**: http://localhost:5176/screentime/cross-device
**API**: `/cross-device/comparison`
**测试项**:
- [ ] 手机vs电脑对比图表
- [ ] 总时长统计
- [ ] 应用数对比
- [ ] 设备使用占比

#### 6. WorkLifeBalance页面
**URL**: http://localhost:5176/screentime/work-life-balance
**API**: `/cross-device/work-life-balance`
**测试项**:
- [ ] 工作生活平衡评分显示
- [ ] 工作时长统计
- [ ] 生活时长统计
- [ ] 平衡建议显示

#### 7. TimeAllocation页面
**URL**: http://localhost:5176/screentime/time-allocation
**API**: `/cross-device/time-allocation`
**测试项**:
- [ ] 24小时时间分配图表
- [ ] 手机vs电脑时间分布
- [ ] 高峰时段识别
- [ ] 百分比计算

#### 8. UserProfile页面
**URL**: http://localhost:5176/screentime/user-profile
**API**: `/cross-device/user-profile`
**测试项**:
- [ ] 用户画像雷达图
- [ ] 设备依赖度评分
- [ ] 工作模式识别
- [ ] 个性化建议

#### 9. AppEcosystem页面
**URL**: http://localhost:5176/screentime/app-ecosystem
**API**: `/cross-device/app-ecosystem`
**测试项**:
- [ ] 跨平台应用识别
- [ ] 手机专属应用列表
- [ ] 电脑专属应用列表
- [ ] 应用生态系统分析

---

### 设备切换测试

**测试项**:
- [ ] DeviceSelector组件显示
- [ ] 设备列表加载(phone_vivo_x90, computer_main)
- [ ] 切换设备后数据更新
- [ ] 跨设备页面不受设备切换影响

---

### API端点测试

可以使用浏览器或curl测试以下端点:

#### 单设备API
```bash
curl http://localhost:18080/api/v1/screentime/summary
curl http://localhost:18080/api/v1/screentime/daily
curl http://localhost:18080/api/v1/screentime/rankings
curl http://localhost:18080/api/v1/screentime/categories
```

#### 设备管理API
```bash
curl http://localhost:18080/api/v1/screentime/devices
```

#### 跨设备API
```bash
curl http://localhost:18080/api/v1/screentime/cross-device/comparison
curl http://localhost:18080/api/v1/screentime/cross-device/work-life-balance
curl http://localhost:18080/api/v1/screentime/cross-device/user-profile
curl http://localhost:18080/api/v1/screentime/cross-device/productivity-deep
curl http://localhost:18080/api/v1/screentime/cross-device/focus-analysis
curl http://localhost:18080/api/v1/screentime/cross-device/recommendations
```

---

## 已知问题

### 1. 未使用的Import警告
**文件**:
- `src/pages/AppEcosystem.tsx`
- `src/pages/CrossDeviceAnalysis.tsx`
- `src/pages/WorkLifeBalance.tsx`

**问题**: TypeScript警告未使用的图标import
**影响**: 无,仅警告,不影响功能
**修复**: 可选,删除未使用的import

### 2. 端口冲突
**问题**: 8080端口被Windows权限阻止
**解决**: 已改用18080端口
**状态**: ✅ 已解决

---

## 数据验证

### 手机数据 (phone_vivo_x90)
- **数据库**: `./data/screentime/phone_vivo_x90.db`
- **大小**: 17MB
- **记录数**: 22,013条
- **日期范围**: 952天
- **应用数**: 368个
- **总使用时长**: 3,889小时

### 电脑数据 (computer_main)
- **数据库**: `./data/screentime/manictime_computer.db`
- **大小**: 103MB
- **数据来源**: ManicTime
- **状态**: ✅ 已导入

### 设备注册
- **数据库**: `./data/screentime/devices.db`
- **设备数**: 2个
- **设备列表**:
  1. phone_vivo_x90 (手机)
  2. computer_main (电脑)

---

## 测试方法

### 方法1: 浏览器手动测试 (推荐)
1. 打开浏览器访问: http://localhost:5176/screentime/
2. 依次访问9个页面
3. 测试各项功能
4. 检查控制台是否有错误

### 方法2: API直接测试
1. 使用curl或Postman测试API端点
2. 验证返回的JSON数据格式
3. 检查数据准确性

### 方法3: 开发者工具
1. 打开浏览器开发者工具(F12)
2. 查看Network标签,检查API请求
3. 查看Console标签,检查JavaScript错误
4. 查看Response数据

---

## 下一步

### 如果测试通过
- 进入Phase 6: 优化与完善
  - 添加图表库(recharts)
  - 实现日期范围筛选
  - 添加应用搜索功能
  - 优化UI/UX

### 如果测试失败
- 检查错误日志
- 修复发现的问题
- 重新测试

---

## 服务器管理

### 停止服务器
```bash
# 停止后端
ps aux | grep server.exe | grep -v grep | awk '{print $2}' | xargs kill

# 停止前端
ps aux | grep vite | grep -v grep | awk '{print $2}' | xargs kill
```

### 重启服务器
```bash
# 重启后端
cd C:/Users/joengzaang/CodeProject/records/go-backend
./bin/server.exe > /tmp/backend.log 2>&1 &

# 重启前端
cd C:/Users/joengzaang/CodeProject/records/screentime
npm run dev > /tmp/frontend.log 2>&1 &
```

### 查看日志
```bash
# 后端日志
tail -f /tmp/backend_final.log

# 前端日志
tail -f /tmp/frontend.log
```

---

## 总结

✅ **Phase 5准备完成**
- 后端服务器运行在18080端口
- 前端服务器运行在5176端口
- 18个API端点已注册
- 9个页面已实现
- 等待用户进行功能测试

**完成度**: ScreenTime模块 → 97%

**访问地址**: http://localhost:5176/screentime/

---

**报告完成时间**: 2026-03-01 23:15
**报告作者**: Claude Sonnet 4.5
