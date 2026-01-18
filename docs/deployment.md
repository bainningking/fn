# Agent 管理平台部署指南

## 系统要求

- Go 1.21 或更高版本
- PostgreSQL 12 或更高版本（生产环境）
- Redis 6.0 或更高版本（可选，用于缓存）
- 至少 2GB RAM
- 至少 10GB 磁盘空间

## 部署步骤

### 1. 准备环境

```bash
# 安装 Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装 PostgreSQL
sudo apt-get update
sudo apt-get install postgresql postgresql-contrib

# 安装 Redis（可选）
sudo apt-get install redis-server
```

### 2. 配置数据库

```bash
# 创建数据库和用户
sudo -u postgres psql
CREATE DATABASE agent_platform;
CREATE USER agent_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE agent_platform TO agent_user;
\q
```

### 3. 配置应用

复制配置文件模板：

```bash
cp platform/config.example.yaml platform/config.yaml
```

编辑 `platform/config.yaml`：

```yaml
server:
  grpc_port: ":50051"
  http_port: ":8080"

database:
  host: localhost
  port: 5432
  user: agent_user
  password: your_password
  dbname: agent_platform
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

log:
  level: info
  file: /var/log/agent-platform/app.log
```

### 4. 构建应用

```bash
# 构建平台服务器
go build -o bin/platform-server platform/cmd/server/main.go

# 构建 Agent
go build -o bin/agent agent/cmd/main.go
```

### 5. 运行应用

```bash
# 启动平台服务器
./bin/platform-server -config platform/config.yaml

# 启动 Agent
./bin/agent -server localhost:50051 -id agent-001
```

## 使用 systemd 管理服务

### 平台服务器服务

创建 `/etc/systemd/system/agent-platform.service`：

```ini
[Unit]
Description=Agent Management Platform
After=network.target postgresql.service

[Service]
Type=simple
User=agent
WorkingDirectory=/opt/agent-platform
ExecStart=/opt/agent-platform/bin/platform-server -config /opt/agent-platform/config.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### Agent 服务

创建 `/etc/systemd/system/agent@.service`：

```ini
[Unit]
Description=Agent %i
After=network.target agent-platform.service

[Service]
Type=simple
User=agent
WorkingDirectory=/opt/agent-platform
ExecStart=/opt/agent-platform/bin/agent -server localhost:50051 -id %i
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### 启动服务

```bash
# 重载 systemd 配置
sudo systemctl daemon-reload

# 启动平台服务器
sudo systemctl start agent-platform
sudo systemctl enable agent-platform

# 启动 Agent
sudo systemctl start agent@agent-001
sudo systemctl enable agent@agent-001
```

## 监控和维护

### 健康检查

```bash
# 检查平台健康状态
curl http://localhost:8080/api/v1/monitor/health

# 查看性能指标
curl http://localhost:8080/api/v1/monitor/metrics
```

### 日志管理

```bash
# 查看平台日志
sudo journalctl -u agent-platform -f

# 查看 Agent 日志
sudo journalctl -u agent@agent-001 -f
```

### 数据库备份

```bash
# 备份数据库
pg_dump -U agent_user agent_platform > backup_$(date +%Y%m%d).sql

# 恢复数据库
psql -U agent_user agent_platform < backup_20260118.sql
```

## 安全建议

1. 使用强密码保护数据库和 Redis
2. 启用 PostgreSQL SSL 连接
3. 配置防火墙规则限制访问
4. 定期更新系统和依赖
5. 启用审计日志记录
6. 使用 HTTPS 保护 HTTP API

## 故障排查

### 平台无法启动

- 检查配置文件是否正确
- 确认数据库连接是否正常
- 查看日志文件获取详细错误信息

### Agent 无法连接

- 确认平台服务器正在运行
- 检查网络连接和防火墙规则
- 验证 gRPC 端口是否正确

### 性能问题

- 检查数据库查询性能
- 监控系统资源使用情况
- 考虑启用 Redis 缓存
- 调整数据库连接池大小
