# Agent 管理平台

基于 Go 的分布式 Agent 管理平台，支持系统监控、日志采集、任务执行和插件化扩展。

## 项目概述

**目标规模**: 10-100 台机器
**技术栈**: Go + gRPC + React + Ant Design + PostgreSQL + Redis
**架构模式**: 混合架构（Push + Pull）

## 核心功能

### ✅ 已实现功能

**1. Agent 管理**
- gRPC 双向流长连接
- 心跳机制和状态监控
- Agent 注册和注销
- 实时连接状态追踪

**2. 任务执行**
- Shell 脚本远程执行
- Python 脚本远程执行
- 任务状态管理（pending/running/completed/failed）
- 任务结果实时上报
- 超时控制和并发管理

**3. 插件系统**
- 插件化架构（独立进程模式）
- 内置插件：
  - CPU 使用率采集
  - 内存使用率采集
  - 磁盘使用率采集
  - 网络流量监控
  - 日志文件 tail
- 插件生命周期管理（加载/启动/停止/卸载）
- 远程插件安装和卸载

**4. 数据存储和可视化**
- REST API（Gin 框架）
- 指标数据存储和查询
- Web UI（React + Ant Design）
- Agent 列表管理界面
- 实时数据展示

**5. 生产就绪**
- 审计日志系统
- 配置管理（YAML）
- 性能监控（系统指标 + 业务指标）
- 部署脚本（systemd + Docker Compose）
- 完整的部署文档

## 项目结构

```
.
├── agent/                      # Agent 端代码
│   ├── cmd/agent/             # Agent 主程序
│   ├── internal/
│   │   ├── client/            # gRPC 客户端
│   │   ├── config/            # 配置管理
│   │   ├── executor/          # 任务执行器
│   │   └── plugin/            # 插件管理器
│   └── config.example.yaml    # Agent 配置示例
│
├── platform/                   # 管理平台代码
│   ├── cmd/server/            # 服务器主程序
│   ├── internal/
│   │   ├── api/               # REST API
│   │   ├── audit/             # 审计日志
│   │   ├── config/            # 配置管理
│   │   ├── database/          # 数据库连接
│   │   ├── grpc/              # gRPC 服务器
│   │   ├── models/            # 数据模型
│   │   ├── monitor/           # 性能监控
│   │   └── service/           # 业务服务
│   └── config.example.yaml    # 平台配置示例
│
├── web/                        # 前端代码
│   ├── src/
│   │   ├── components/        # React 组件
│   │   ├── pages/             # 页面
│   │   ├── services/          # API 服务
│   │   └── types/             # TypeScript 类型
│   └── package.json
│
├── proto/                      # Protocol Buffers 定义
│   ├── agent.proto            # Agent 消息
│   ├── task.proto             # 任务消息
│   ├── metric.proto           # 指标消息
│   └── plugin.proto           # 插件消息
│
├── plugins/                    # 内置插件
│   ├── cpu/                   # CPU 采集插件
│   ├── memory/                # 内存采集插件
│   ├── disk/                  # 磁盘采集插件
│   ├── network/               # 网络采集插件
│   └── logtail/               # 日志 tail 插件
│
├── scripts/                    # 部署脚本
│   ├── deploy.sh              # 管理平台部署脚本
│   ├── install-agent.sh       # Agent 安装脚本
│   └── backup.sh              # 备份脚本
│
├── docs/                       # 文档
│   ├── plans/                 # 实现计划
│   ├── deployment.md          # 部署文档
│   └── installation.md        # 安装文档
│
├── Makefile                    # 构建配置
├── docker-compose.yml          # Docker Compose 配置
└── README.md                   # 本文件
```

## 快速开始

### 前置条件

- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Redis 7+ (可选)
- protoc 编译器

### 安装步骤

1. **克隆代码仓库**
```bash
git clone <repository-url>
cd agent-platform
```

2. **生成 protobuf 代码**
```bash
make proto
```

3. **编译程序**
```bash
make build-agent
make build-platform
```

4. **配置数据库**
```bash
createdb agent_platform
cp platform/config.example.yaml platform/config.yaml
# 编辑 platform/config.yaml 配置数据库连接
```

5. **启动管理平台**
```bash
./bin/server -config platform/config.yaml
```

6. **启动 Agent**
```bash
cp agent/config.example.yaml agent/config.yaml
# 编辑 agent/config.yaml 配置服务器地址
./bin/agent -config agent/config.yaml
```

7. **启动前端**
```bash
cd web
npm install
npm run dev
```

8. **访问 Web UI**
```
http://localhost:3000
```

## API 文档

### REST API

**Agent 管理**
- `GET /api/v1/agents` - 获取 Agent 列表
- `GET /api/v1/agents/:id` - 获取 Agent 详情
- `DELETE /api/v1/agents/:id` - 删除 Agent

**任务管理**
- `POST /api/v1/tasks` - 创建任务
- `GET /api/v1/tasks` - 获取任务列表
- `GET /api/v1/tasks/:id` - 获取任务详情

**指标查询**
- `GET /api/v1/metrics` - 查询指标数据

**监控**
- `GET /api/v1/monitor/health` - 健康检查
- `GET /api/v1/monitor/metrics` - 获取系统指标

### gRPC API

详见 `proto/` 目录下的 Protocol Buffers 定义文件。

## 部署

### 使用 systemd

```bash
# 部署管理平台
sudo ./scripts/deploy.sh

# 安装 Agent
sudo ./scripts/install-agent.sh <server_address> <agent_id>
```

### 使用 Docker Compose

```bash
docker-compose up -d
```

详细部署文档请参考 [docs/deployment.md](docs/deployment.md)。

## 开发

### 运行测试

```bash
make test
```

### 代码格式化

```bash
go fmt ./...
```

### 清理构建产物

```bash
make clean
```

## 架构设计

### 通信协议

- **gRPC 双向流**: Agent 与管理平台之间的长连接通信
- **Protocol Buffers**: 高效的二进制消息序列化
- **批量上报**: Agent 每 30 秒批量推送指标和日志

### 插件架构

- **独立进程模式**: 插件作为独立可执行文件运行
- **stdin/stdout 通信**: 通过标准输入输出进行 JSON 消息交换
- **生命周期管理**: Agent 负责启动、停止和监控插件进程

### 数据存储

- **PostgreSQL**: 存储 Agent 信息、任务记录、指标数据、审计日志
- **Redis**: 缓存会话数据、实时数据、任务队列（可选）

## 性能指标

- **支持规模**: 10-100 台 Agent
- **心跳间隔**: 30 秒
- **指标采集间隔**: 30 秒
- **任务执行超时**: 可配置（默认 300 秒）
- **并发任务数**: 可配置（默认 5 个）

## 安全特性

- **审计日志**: 记录所有 API 操作
- **配置管理**: 支持环境变量和配置文件
- **进程隔离**: 插件独立进程运行
- **超时控制**: 任务执行超时保护

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue。

---

**项目状态**: ✅ 生产就绪

**最后更新**: 2026-01-18
