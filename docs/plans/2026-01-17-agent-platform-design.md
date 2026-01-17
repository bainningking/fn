# Agent 管理平台设计文档

**日期**: 2026-01-17
**版本**: 1.0
**状态**: 设计完成

## 项目概述

开发一个基于 Go 的 Agent 管理平台，支持系统监控、日志采集、自动化运维和安全审计。系统采用混合架构，使用 gRPC 双向流实现高效通信，支持插件化扩展。

**目标规模**: 10-100 台机器
**数据存储**: 短期存储（数天）
**技术栈**: Go + React + Ant Design + PostgreSQL + Redis

## 整体架构

### 核心组件

**1. Agent（Go）**
- 轻量级守护进程，运行在每台被管理的机器上
- 使用 gRPC 双向流与管理平台保持长连接
- 插件化架构：通过独立进程方式加载采集插件
- 内置插件：CPU/内存/磁盘/网络指标采集、日志文件 tail、进程监控等
- 支持动态加载/卸载插件，无需重启 Agent

**2. 管理平台（Go）**
- 单体应用，包含 API 服务器和 gRPC 服务器
- 维护与所有 Agent 的长连接（使用 goroutine pool 管理）
- Web UI（React + Ant Design）
- 数据存储：PostgreSQL（存储 Agent 信息、任务记录、时序指标）+ Redis（缓存和会话）
- 任务调度器：支持立即执行和定时任务

**3. 通信协议**
- gRPC 双向流：Agent 连接后保持长连接，用于心跳、任务下发、实时事件
- Protocol Buffers 定义消息格式：高效的二进制序列化
- 批量上报：Agent 每 10-30 秒批量推送指标和日志，减少网络开销
- 断线重连：Agent 自动重连，支持指数退避策略

## 任务执行模块

### Agent 任务执行器

**支持的任务类型**：
- Shell 脚本执行（bash/sh）
- Python 脚本执行（需要目标机器安装 Python）
- 内置命令（如文件操作、进程管理等）

**执行流程**：
1. 管理平台通过 gRPC 长连接下发任务（包含脚本内容、执行参数、超时时间等）
2. Agent 接收任务后，创建隔离的执行环境（使用 os/exec 包）
3. 实时捕获标准输出和标准错误
4. 通过长连接流式返回执行日志（可选：批量返回或实时流式）
5. 任务完成后返回退出码和完整输出

**安全措施**：
- 脚本执行使用受限用户权限（可配置）
- 支持超时控制，防止任务无限运行
- 支持任务取消（通过 context.Context）
- 可配置黑名单命令（如 rm -rf /）
- 脚本内容通过 gRPC 加密传输

**并发控制**：
- 支持同时执行多个任务（可配置最大并发数）
- 使用 goroutine pool 管理任务执行
- 任务队列：当达到并发上限时，新任务进入队列等待

## 数据流设计

### 1. Agent 启动和连接流程
- Agent 启动时读取配置文件（管理平台地址、认证 Token、采集间隔等）
- 使用 gRPC 双向流连接到管理平台，发送注册消息（包含主机名、IP、OS 信息等）
- 管理平台验证 Token，将 Agent 信息存入数据库，返回确认消息
- 连接建立后，Agent 每 30 秒发送心跳包，管理平台更新最后心跳时间

### 2. 指标采集和上报流程
- Agent 每 10-30 秒（可配置）触发一次采集周期
- 并发调用所有已加载的指标采集插件（CPU、内存、磁盘、网络等）
- 将采集结果批量打包成 Protobuf 消息
- 通过 gRPC 双向流推送到管理平台
- 管理平台接收后写入 PostgreSQL，返回 ACK

### 3. 日志采集和上报流程
- 日志采集插件使用 tail -f 方式实时监听日志文件
- 新日志行缓存在内存队列中（带大小限制，防止内存溢出）
- 每 5-10 秒或队列达到阈值时，批量推送到管理平台
- 管理平台可选择存储到数据库或转发到外部日志系统（如 Elasticsearch）

### 4. 任务下发和执行流程
- 用户在 Web UI 创建任务（选择目标 Agent、脚本内容、参数等）
- 管理平台通过 gRPC 双向流推送任务到目标 Agent
- Agent 接收任务，创建执行上下文，启动脚本执行
- 执行过程中，实时或批量返回输出日志
- 执行完成后，返回退出码和完整结果，管理平台更新任务状态

## 插件架构设计

### 插件类型
1. **指标采集插件**（Metric Collector）：采集 CPU、内存、磁盘、网络等系统指标
2. **日志采集插件**（Log Collector）：监听和采集日志文件
3. **自定义插件**（Custom Plugin）：用户自定义的采集逻辑

### 插件实现方式（独立进程模式）
- 插件作为独立的可执行文件（Go 编译的二进制）
- Agent 通过 stdin/stdout 与插件通信（使用 JSON 或 Protobuf）
- 优点：插件崩溃不影响 Agent 主进程，易于开发和调试，支持任意语言编写插件
- Agent 使用 os/exec 启动插件进程，通过 context 控制生命周期

### 插件接口定义
```json
// 插件标准输入：配置信息
{"action": "collect", "config": {...}}

// 插件标准输出：采集结果
{"metrics": [...], "logs": [...], "error": null}
```

### 内置插件
- `cpu_collector`：采集 CPU 使用率、负载等
- `memory_collector`：采集内存使用情况
- `disk_collector`：采集磁盘 IO、空间使用
- `network_collector`：采集网络流量、连接数
- `log_tail`：tail 日志文件，支持多行日志合并
- `process_monitor`：监控指定进程状态

### 插件管理
- 插件配置存储在 Agent 配置文件中（YAML 格式）
- 支持通过管理平台动态下发插件配置（无需重启 Agent）
- 插件安装：管理平台下发插件二进制文件，Agent 保存到本地目录
- 插件卸载：删除插件文件和配置

## 插件安装和分发

### 插件打包格式
- 插件以独立的二进制文件形式分发（Go 编译后的可执行文件）
- 每个插件包含元数据文件（JSON 格式）：插件名称、版本、依赖、配置模板等
- 插件存储在管理平台的文件系统或对象存储中（如 MinIO）

### 插件安装方式

**方式1：通过管理平台远程安装（推荐）**
1. 管理员在 Web UI 上传插件二进制文件到管理平台
2. 管理平台验证插件（检查签名、扫描恶意代码等）
3. 管理员选择目标 Agent，点击"安装插件"
4. 管理平台通过 gRPC 长连接下发安装命令和插件二进制数据
5. Agent 接收后保存到本地插件目录（如 `/opt/agent/plugins/`）
6. Agent 设置可执行权限，加载插件配置
7. Agent 返回安装结果（成功/失败）

**方式2：Agent 本地安装**
- Agent 启动时扫描插件目录，自动加载所有插件
- 管理员可以手动将插件文件复制到 Agent 机器的插件目录
- Agent 检测到新插件后自动加载（通过文件监听或定期扫描）

### 插件目录结构
```
/opt/agent/
├── bin/agent           # Agent 主程序
├── config/agent.yaml   # Agent 配置
└── plugins/            # 插件目录
    ├── cpu_collector   # CPU 采集插件
    ├── memory_collector
    ├── disk_collector
    └── custom_plugin_v1.0.0
```

### 插件版本管理
- 插件文件名包含版本号（如 `custom_plugin_v1.0.0`）
- 支持多版本共存，Agent 配置中指定使用哪个版本
- 管理平台支持插件升级：下发新版本，Agent 停止旧版本，启动新版本

### 安全措施
- 插件二进制文件使用数字签名验证（防止篡改）
- 管理平台可配置插件白名单（只允许安装特定插件）
- Agent 使用受限权限运行插件（通过 setuid/setgid）

## 错误处理和容错机制

### Agent 端容错
- **断线重连**：Agent 与管理平台断开后，使用指数退避策略自动重连（1s, 2s, 4s, 8s...最大60s）
- **数据缓存**：网络断开时，指标和日志缓存在本地队列（带大小限制，如100MB），重连后批量上传
- **插件崩溃恢复**：插件进程崩溃时，Agent 自动重启插件（最多重试3次，失败后标记插件为不可用）
- **任务超时**：每个任务设置超时时间，超时后强制终止进程（使用 context.WithTimeout）
- **资源限制**：限制插件的 CPU 和内存使用（通过 cgroup 或进程监控）

### 管理平台容错
- **连接管理**：维护 Agent 连接池，定期清理僵尸连接（超过5分钟无心跳）
- **任务重试**：任务执行失败时，支持自动重试（可配置重试次数和间隔）
- **数据持久化**：任务状态和结果持久化到数据库，防止服务重启丢失
- **优雅关闭**：服务关闭时，通知所有 Agent 断开连接，等待正在执行的任务完成

### 监控和告警
- Agent 离线告警（超过3分钟无心跳）
- 任务执行失败告警
- 插件崩溃告警
- 系统资源异常告警（CPU/内存/磁盘）

### 日志和审计
- 所有任务执行记录保存到数据库（包含执行者、时间、脚本内容、结果）
- Agent 操作日志（安装、卸载、配置变更）
- 管理平台操作日志（用户登录、任务创建、配置修改）

## 安全设计

### 认证和授权
- **Agent 认证**：每个 Agent 使用唯一的 Token（UUID）进行认证，Token 在 Agent 安装时生成
- **用户认证**：管理平台支持用户登录（用户名/密码 + JWT），可选集成 LDAP/OAuth
- **权限控制**：基于角色的访问控制（RBAC）：管理员、操作员、只读用户
- **API 认证**：管理平台 REST API 使用 JWT Token 认证

### 通信安全
- **TLS 加密**：gRPC 连接强制使用 TLS 1.3，防止中间人攻击
- **证书管理**：管理平台使用自签名证书或 Let's Encrypt，Agent 信任管理平台证书
- **双向认证**：可选启用 mTLS（mutual TLS），Agent 和管理平台互相验证证书

### 脚本执行安全
- **沙箱执行**：脚本在受限用户权限下执行（非 root）
- **命令白名单/黑名单**：可配置禁止执行的危险命令（如 `rm -rf /`、`mkfs` 等）
- **资源限制**：使用 cgroup 限制脚本的 CPU、内存、磁盘 IO
- **审计日志**：所有脚本执行记录完整的审计日志（谁、何时、执行了什么、结果如何）

### 数据安全
- **敏感数据加密**：数据库中的敏感信息（如 Token）使用 AES-256 加密存储
- **日志脱敏**：自动检测和脱敏日志中的敏感信息（密码、密钥、信用卡号等）
- **数据传输**：所有数据传输使用 TLS 加密

### 防护措施
- **速率限制**：防止 API 滥用和 DDoS 攻击
- **输入验证**：严格验证所有用户输入，防止注入攻击
- **最小权限原则**：Agent 和管理平台都以最小必要权限运行

## 技术栈选型

### Agent 端
- **核心语言**：Go 1.21+
- **gRPC 框架**：`google.golang.org/grpc` - 双向流通信
- **配置管理**：`gopkg.in/yaml.v3` - YAML 配置文件解析
- **进程管理**：`os/exec` - 插件和脚本执行
- **日志库**：`go.uber.org/zap` - 高性能结构化日志
- **系统指标采集**：`github.com/shirou/gopsutil` - 跨平台系统信息采集
- **文件监听**：`github.com/fsnotify/fsnotify` - 日志文件 tail

### 管理平台后端
- **核心语言**：Go 1.21+
- **Web 框架**：`github.com/gin-gonic/gin` - REST API
- **gRPC 框架**：`google.golang.org/grpc` - Agent 通信
- **数据库 ORM**：`gorm.io/gorm` - 数据库操作
- **数据库**：PostgreSQL 14+
- **缓存**：Redis 7+
- **Redis 客户端**：`github.com/redis/go-redis/v9`

### 管理平台前端
- **框架**：React 18+
- **UI 组件库**：Ant Design 5.x
- **状态管理**：React Context + Hooks 或 Zustand
- **HTTP 客户端**：Axios
- **路由**：React Router 6
- **图表库**：ECharts 或 Recharts
- **实时更新**：WebSocket 或 Server-Sent Events

### 通信协议
- **Protocol Buffers**：定义消息格式
- **gRPC**：双向流通信，支持 TLS

### 数据存储方案

**PostgreSQL 表结构**：
- `agents` - Agent 信息（ID、主机名、IP、状态、最后心跳时间）
- `tasks` - 任务记录（ID、类型、脚本内容、目标 Agent、状态、结果）
- `metrics` - 时序指标（时间戳、Agent ID、指标类型、值）
- `logs` - 日志记录（时间戳、Agent ID、日志内容、级别）
- `plugins` - 插件信息（名称、版本、二进制文件路径）
- `users` - 用户信息（用户名、密码哈希、角色）

**Redis 使用场景**：
- Session 存储：`session:{token}` - 用户会话
- Agent 在线状态缓存：`agent:online:{agent_id}` - TTL 5分钟
- 实时指标缓存：`metrics:latest:{agent_id}` - 最新指标数据
- 任务队列：`task:queue` - 待执行任务队列（可选）

## 部署方案

### Agent 部署
- **安装方式**：提供安装脚本（shell script）自动化部署
- **安装脚本功能**：
  - 下载 Agent 二进制文件
  - 创建配置文件（需要输入管理平台地址和 Token）
  - 创建 systemd 服务（Linux）或 Windows 服务
  - 启动 Agent 服务
  - 验证连接成功
- **配置文件位置**：`/etc/agent/config.yaml`（Linux）或 `C:\Program Files\Agent\config.yaml`（Windows）
- **日志位置**：`/var/log/agent/`（Linux）或 `C:\ProgramData\Agent\logs\`（Windows）
- **卸载方式**：通过管理平台远程卸载或本地运行卸载脚本

### 管理平台部署

**部署方式1：Docker Compose（推荐用于开发和小规模生产）**
```yaml
services:
  postgres:
    image: postgres:14
  redis:
    image: redis:7
  platform:
    build: .
    ports:
      - "8080:8080"  # Web UI
      - "9090:9090"  # gRPC
```

**部署方式2：二进制部署（生产环境）**
- 编译 Go 后端为单个二进制文件
- 构建 React 前端为静态文件，嵌入到 Go 二进制中（使用 embed）
- 使用 systemd 管理服务
- Nginx 反向代理（可选，用于 HTTPS 和负载均衡）

### 初始化流程
1. 启动 PostgreSQL 和 Redis
2. 运行数据库迁移（创建表结构）
3. 创建管理员账户
4. 启动管理平台服务
5. 访问 Web UI，登录并生成 Agent Token
6. 在目标机器上安装 Agent

### 监控和维护
- 管理平台自身的监控（使用 Prometheus + Grafana 或简单的健康检查端点）
- 日志轮转和归档
- 数据库备份策略
- 定期清理过期数据（如30天前的指标和日志）

## 项目结构

```
agent-platform/
├── agent/                    # Agent 端代码
│   ├── cmd/
│   │   └── agent/           # Agent 主程序入口
│   ├── internal/
│   │   ├── collector/       # 指标采集模块
│   │   ├── executor/        # 任务执行模块
│   │   ├── plugin/          # 插件管理
│   │   ├── client/          # gRPC 客户端
│   │   └── config/          # 配置管理
│   ├── plugins/             # 内置插件源码
│   │   ├── cpu/
│   │   ├── memory/
│   │   ├── disk/
│   │   └── logtail/
│   └── scripts/             # 安装/卸载脚本
├── platform/                # 管理平台后端
│   ├── cmd/
│   │   └── server/          # 服务器主程序入口
│   ├── internal/
│   │   ├── api/             # REST API 处理器
│   │   ├── grpc/            # gRPC 服务实现
│   │   ├── models/          # 数据模型
│   │   ├── service/         # 业务逻辑
│   │   ├── repository/      # 数据访问层
│   │   └── middleware/      # 中间件（认证、日志等）
│   └── migrations/          # 数据库迁移脚本
├── web/                     # 前端代码
│   ├── src/
│   │   ├── components/      # React 组件
│   │   ├── pages/           # 页面
│   │   ├── services/        # API 调用
│   │   ├── store/           # 状态管理
│   │   └── utils/           # 工具函数
│   └── public/
├── proto/                   # Protocol Buffers 定义
│   ├── agent.proto          # Agent 消息定义
│   ├── task.proto           # 任务消息定义
│   └── metric.proto         # 指标消息定义
├── docs/                    # 文档
│   ├── architecture.md      # 架构文档
│   ├── api.md              # API 文档
│   └── deployment.md       # 部署文档
├── docker/                  # Docker 相关文件
│   ├── Dockerfile.agent
│   ├── Dockerfile.platform
│   └── docker-compose.yml
├── Makefile                 # 构建脚本
├── go.mod
└── README.md
```

## 开发路线图

### 阶段1：核心基础（MVP）
- 定义 Protocol Buffers 消息格式
- 实现 Agent 基础框架：配置管理、gRPC 客户端、心跳机制
- 实现管理平台基础框架：gRPC 服务器、连接管理、数据库模型
- 实现一个简单的指标采集插件（如 CPU 监控）
- 实现基本的 Web UI：Agent 列表、在线状态显示
- 目标：Agent 能连接到管理平台并上报心跳和基本指标

### 阶段2：任务执行
- 实现 Agent 端任务执行器（支持 Shell 脚本）
- 实现管理平台任务下发和管理
- 实现 Web UI 任务创建和执行界面
- 实现任务执行日志实时查看
- 目标：能够远程执行脚本并查看结果

### 阶段3：插件系统
- 实现插件管理框架（加载、卸载、配置）
- 开发内置插件：CPU、内存、磁盘、网络、日志 tail
- 实现插件远程安装和卸载功能
- 实现 Web UI 插件管理界面
- 目标：完整的插件化采集系统

### 阶段4：数据存储和可视化
- 实现指标数据存储（PostgreSQL）
- 实现日志数据存储
- 实现 Web UI 指标可视化（图表）
- 实现日志查询和过滤界面
- 目标：完整的数据查看和分析能力

### 阶段5：安全和生产就绪
- 实现 TLS 加密通信
- 实现用户认证和权限控制
- 实现审计日志
- 完善错误处理和容错机制
- 编写部署文档和安装脚本
- 目标：生产环境可用

## 总结

本设计文档详细描述了一个完整的 Agent 管理平台系统，涵盖了架构设计、数据流、插件系统、安全机制、技术栈选型和部署方案。系统采用混合架构，使用 gRPC 双向流实现高效通信，支持插件化扩展，适合 10-100 台机器的小规模部署场景。

设计遵循业界最佳实践，参考了 Prometheus、Telegraf、Datadog Agent 等成熟产品的设计理念，同时针对具体需求进行了优化和简化。
