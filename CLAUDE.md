# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

基于 Go 的分布式 Agent 管理平台，支持系统监控、日志采集、任务执行和插件化扩展。

**技术栈**: Go 1.25+ + gRPC + React + Ant Design + PostgreSQL + Redis
**架构**: 混合架构（Push + Pull），gRPC 双向流长连接

## 常用命令

### 构建和编译
```bash
# 生成 protobuf 代码（修改 proto 文件后必须执行）
make proto

# 编译 Agent
make build-agent

# 编译管理平台服务器
make build-platform

# 清理构建产物
make clean
```

### 测试
```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./agent/internal/client
go test -v ./platform/internal/service

# 运行单个测试
go test -v ./platform/internal/service -run TestTaskService_CreateTask
```

### 运行服务
```bash
# 启动管理平台服务器
./bin/server -config platform/config.yaml

# 启动 Agent
./bin/agent -config agent/config.yaml

# 启动前端开发服务器
cd web && npm run dev
```

### 部署
```bash
# 部署管理平台（使用 systemd）
sudo ./scripts/deploy.sh

# 安装 Agent
sudo ./scripts/install-agent.sh <server_address> <agent_id>
```

## 核心架构

### 通信机制
- **gRPC 双向流**: Agent 与 Platform 之间通过 `AgentService.Connect` 建立长连接
- **消息类型**:
  - Agent → Platform: Register, Heartbeat, TaskResult, TaskLog, Plugin 响应
  - Platform → Agent: TaskRequest, InstallPlugin, UninstallPlugin, ListPlugins
- **心跳间隔**: 30 秒（可配置）

### 关键组件

**Agent 端** (`agent/`):
- `internal/client/client.go`: gRPC 客户端，处理与 Platform 的双向流通信
- `internal/executor/executor.go`: 任务执行器，支持 shell 和 python 脚本
- `internal/plugin/manager.go`: 插件管理器，负责插件生命周期（Load/Start/Stop/Unload）
- `internal/plugin/plugin.go`: 插件实例，通过 stdin/stdout 与独立进程通信

**Platform 端** (`platform/`):
- `internal/grpc/handler.go`: gRPC 服务端，处理 Agent 连接和消息路由
- `internal/service/task_service.go`: 任务服务，管理任务的 CRUD 和状态更新
- `internal/api/`: REST API handlers（Gin 框架）
- `internal/models/`: GORM 数据模型（Agent, Task, Metric, AuditLog）

**插件系统** (`plugins/`):
- 独立进程模式：每个插件是独立的可执行文件
- 通信协议：通过 stdin/stdout 交换 JSON 消息
- 内置插件：cpu, memory, disk, network, logtail

### Protocol Buffers 定义
- `proto/agent.proto`: Agent 注册、心跳、连接管理
- `proto/task.proto`: 任务请求和结果
- `proto/plugin.proto`: 插件安装、卸载、列表
- `proto/metric.proto`: 指标数据上报
- `proto/common.proto`: 通用消息类型

### 数据流
1. Agent 启动 → 连接 Platform → 发送 Register 消息
2. Platform 接收注册 → 存储 Agent 信息到数据库 → 返回 RegisterResponse
3. Agent 定期发送 Heartbeat（30s）→ Platform 更新状态
4. Platform 创建任务 → 通过 gRPC 流发送 TaskRequest → Agent 执行 → 返回 TaskResult
5. Agent 插件采集指标 → 批量上报 → Platform 存储到数据库

### 配置文件
- `platform/config.yaml`: 服务器配置（gRPC/HTTP 端口、数据库、Redis）
- `agent/config.yaml`: Agent 配置（服务器地址、Agent ID、采集间隔）

### 数据库模型
- `Agent`: agent_id, hostname, ip, status, last_heartbeat
- `Task`: task_id, agent_id, type, script, status, exit_code, stdout, stderr
- `Metric`: agent_id, metric_name, value, timestamp
- `AuditLog`: user, action, resource, timestamp

## 开发注意事项

### 修改 Protocol Buffers
1. 编辑 `proto/*.proto` 文件
2. 运行 `make proto` 重新生成 Go 代码
3. 更新相关的 handler 和 client 代码

### 添加新插件
1. 在 `plugins/` 下创建新目录
2. 实现独立可执行程序，通过 stdin/stdout 通信
3. 遵循 JSON 消息格式：`{"type": "data", "payload": {...}}`
4. 在 Agent 配置中添加插件名称

### 数据库迁移
- 使用 GORM AutoMigrate 自动创建表结构
- 位置：`platform/internal/database/database.go`

### 测试策略
- 单元测试：测试独立组件逻辑
- 集成测试：`platform/internal/service/integration_test.go` 测试完整流程
- Mock 使用：使用 `go.uber.org/mock` 生成 mock 对象
