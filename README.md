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

### 2026-02-19
- 初始化项目结构
- 配置 Go 模块和依赖
- 创建基础 API 框架
- 实现路由和中间件
- 配置 SQLite 数据库
