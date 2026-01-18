# Agent 管理平台 - 阶段5：生产就绪实现计划

## 概述
本阶段实现审计日志、配置管理优化、部署文档和脚本、性能优化和监控，使系统达到生产就绪状态。

## 实现任务

### 任务1：实现审计日志系统

**文件**:
- 创建: `platform/internal/audit/audit.go`
- 创建: `platform/internal/models/audit_log.go`
- 修改: `platform/internal/api/middleware.go`

**步骤 1: 创建审计日志模型**

创建文件 `platform/internal/models/audit_log.go`:
```go
package models

import "time"

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	Action    string    `gorm:"index" json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Status    string    `json:"status"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
```

**步骤 2: 实现审计日志服务**

创建文件 `platform/internal/audit/audit.go`:
```go
package audit

import (
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Log(log *models.AuditLog) error {
	return s.db.Create(log).Error
}

func (s *Service) Query(userID, action string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := s.db.Order("created_at DESC")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}
```

**步骤 3: 添加审计日志中间件**

修改文件 `platform/internal/api/middleware.go`，添加审计日志中间件：
```go
func AuditLog(auditService *audit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		log := &models.AuditLog{
			UserID:    c.GetString("user_id"),
			Action:    c.Request.Method + " " + c.Request.URL.Path,
			Resource:  c.Request.URL.Path,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Status:    strconv.Itoa(c.Writer.Status()),
		}

		auditService.Log(log)
	}
}
```

**步骤 4: 更新数据库迁移**

修改 `platform/internal/database/database.go`，添加审计日志表迁移：
```go
if err := db.AutoMigrate(&models.Agent{}, &models.Task{}, &models.Metric{}, &models.AuditLog{}); err != nil {
	return nil, fmt.Errorf("failed to migrate: %w", err)
}
```

**步骤 5: 提交审计日志代码**

```bash
git add platform/internal/audit/ platform/internal/models/audit_log.go platform/internal/api/middleware.go platform/internal/database/database.go
git commit -m "feat: 实现审计日志系统"
```

---

### 任务2：优化配置管理

**文件**:
- 创建: `platform/internal/config/config.go`
- 创建: `platform/config.example.yaml`
- 修改: `platform/cmd/server/main.go`
- 修改: `agent/internal/config/config.go`

**步骤 1: 实现管理平台配置管理**

创建文件 `platform/internal/config/config.go`:
```go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	GRPCPort string `yaml:"grpc_port"`
	HTTPPort string `yaml:"http_port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
```

**步骤 2: 创建管理平台示例配置**

创建文件 `platform/config.example.yaml`:
```yaml
server:
  grpc_port: ":9090"
  http_port: ":8080"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "agent_platform"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

log:
  level: "info"
  format: "json"
  output: "stdout"
```

**步骤 3: 更新管理平台主程序使用配置文件**

修改 `platform/cmd/server/main.go`：
```go
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/agent-platform/platform/internal/api"
	"github.com/yourusername/agent-platform/platform/internal/config"
	"github.com/yourusername/agent-platform/platform/internal/database"
	grpcserver "github.com/yourusername/agent-platform/platform/internal/grpc"
)

func main() {
	configPath := flag.String("config", "platform/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	dbCfg := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 启动 gRPC 服务器
	grpcServer := grpcserver.NewServer(cfg.Server.GRPCPort, db)
	go func() {
		log.Printf("Starting gRPC server on %s", cfg.Server.GRPCPort)
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 启动 HTTP API 服务器
	router := api.SetupRouter(db)
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Server.HTTPPort)
		if err := router.Run(cfg.Server.HTTPPort); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Server shutting down")
	grpcServer.Stop()
}
```

**步骤 4: 优化 Agent 配置管理**

修改 `agent/internal/config/config.go`，添加更多配置选项：
```go
type Config struct {
	Server ServerConfig `yaml:"server"`
	Agent  AgentConfig  `yaml:"agent"`
	Log    LogConfig    `yaml:"log"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}
```

**步骤 5: 提交配置管理优化**

```bash
git add platform/internal/config/ platform/config.example.yaml platform/cmd/server/main.go agent/internal/config/config.go
git commit -m "feat: 优化配置管理系统"
```

---

### 任务3：实现性能监控

**文件**:
- 创建: `platform/internal/monitor/monitor.go`
- 创建: `platform/internal/api/monitor_handler.go`

**步骤 1: 实现监控指标收集**

创建文件 `platform/internal/monitor/monitor.go`:
```go
package monitor

import (
	"runtime"
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	// 系统指标
	Goroutines   int
	MemoryAlloc  uint64
	MemorySys    uint64
	GCPauseTotal uint64

	// 业务指标
	ActiveAgents  int
	TotalRequests int64
	FailedRequests int64
	AvgResponseTime float64

	LastUpdate time.Time
}

var globalMetrics = &Metrics{}

func GetMetrics() *Metrics {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	m := &Metrics{}
	*m = *globalMetrics
	return m
}

func UpdateSystemMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.Goroutines = runtime.NumGoroutine()
	globalMetrics.MemoryAlloc = mem.Alloc
	globalMetrics.MemorySys = mem.Sys
	globalMetrics.GCPauseTotal = mem.PauseTotalNs
	globalMetrics.LastUpdate = time.Now()
}

func IncrementRequests() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.TotalRequests++
}

func IncrementFailedRequests() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.FailedRequests++
}

func UpdateActiveAgents(count int) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.ActiveAgents = count
}

func StartMonitoring() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			UpdateSystemMetrics()
		}
	}()
}
```

**步骤 2: 实现监控 API**

创建文件 `platform/internal/api/monitor_handler.go`:
```go
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/monitor"
)

type MonitorHandler struct{}

func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{}
}

func (h *MonitorHandler) GetMetrics(c *gin.Context) {
	metrics := monitor.GetMetrics()
	Success(c, metrics)
}

func (h *MonitorHandler) HealthCheck(c *gin.Context) {
	Success(c, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}
```

**步骤 3: 添加监控路由**

修改 `platform/internal/api/router.go`，添加监控路由：
```go
// 监控和健康检查
monitor := api.Group("/monitor")
{
	handler := NewMonitorHandler()
	monitor.GET("/metrics", handler.GetMetrics)
	monitor.GET("/health", handler.HealthCheck)
}
```

**步骤 4: 在主程序中启动监控**

修改 `platform/cmd/server/main.go`，添加监控启动：
```go
// 启动监控
monitor.StartMonitoring()
```

**步骤 5: 提交性能监控代码**

```bash
git add platform/internal/monitor/ platform/internal/api/monitor_handler.go platform/internal/api/router.go platform/cmd/server/main.go
git commit -m "feat: 实现性能监控系统"
```

---

### 任务4：创建部署文档

**文件**:
- 创建: `docs/deployment.md`
- 创建: `docs/installation.md`

**步骤 1: 创建部署文档**

创建文件 `docs/deployment.md`:
```markdown
# Agent 管理平台部署指南

## 系统要求

### 管理平台服务器
- CPU: 2核心以上
- 内存: 4GB 以上
- 磁盘: 50GB 以上
- 操作系统: Linux (Ubuntu 20.04+ / CentOS 7+)

### Agent 机器
- CPU: 1核心以上
- 内存: 512MB 以上
- 操作系统: Linux / Windows / macOS

## 依赖服务

### PostgreSQL
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# 创建数据库
sudo -u postgres psql
CREATE DATABASE agent_platform;
CREATE USER agent_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE agent_platform TO agent_user;
```

### Redis (可选)
```bash
# Ubuntu/Debian
sudo apt-get install redis-server

# 启动 Redis
sudo systemctl start redis
sudo systemctl enable redis
```

## 部署管理平台

### 1. 编译程序
```bash
# 克隆代码
git clone https://github.com/yourusername/agent-platform.git
cd agent-platform

# 生成 protobuf 代码
make proto

# 编译管理平台
make build-platform

# 编译 Agent
make build-agent
```

### 2. 配置管理平台
```bash
# 复制配置文件
cp platform/config.example.yaml platform/config.yaml

# 编辑配置文件
vim platform/config.yaml
```

配置示例：
```yaml
server:
  grpc_port: ":9090"
  http_port: ":8080"

database:
  host: "localhost"
  port: 5432
  user: "agent_user"
  password: "your_password"
  dbname: "agent_platform"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

log:
  level: "info"
  format: "json"
  output: "/var/log/agent-platform/server.log"
```

### 3. 启动管理平台
```bash
# 创建日志目录
sudo mkdir -p /var/log/agent-platform

# 启动服务
./bin/server -config platform/config.yaml
```

### 4. 使用 systemd 管理服务
创建文件 `/etc/systemd/system/agent-platform.service`:
```ini
[Unit]
Description=Agent Platform Server
After=network.target postgresql.service

[Service]
Type=simple
User=agent
WorkingDirectory=/opt/agent-platform
ExecStart=/opt/agent-platform/bin/server -config /opt/agent-platform/platform/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl start agent-platform
sudo systemctl enable agent-platform
sudo systemctl status agent-platform
```

## 部署 Agent

### 1. 配置 Agent
```bash
# 复制配置文件
cp agent/config.example.yaml agent/config.yaml

# 编辑配置文件
vim agent/config.yaml
```

配置示例：
```yaml
server:
  address: "your-server-ip:9090"
  tls: false

agent:
  id: "agent-001"
  collect_interval: 30
```

### 2. 启动 Agent
```bash
./bin/agent -config agent/config.yaml
```

### 3. 使用 systemd 管理 Agent
创建文件 `/etc/systemd/system/agent.service`:
```ini
[Unit]
Description=Agent Platform Agent
After=network.target

[Service]
Type=simple
User=agent
WorkingDirectory=/opt/agent
ExecStart=/opt/agent/bin/agent -config /opt/agent/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl start agent
sudo systemctl enable agent
sudo systemctl status agent
```

## 前端部署

### 1. 构建前端
```bash
cd web
npm install
npm run build
```

### 2. 使用 Nginx 部署
```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /opt/agent-platform/web/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 验证部署

### 1. 检查服务状态
```bash
# 检查管理平台
curl http://localhost:8080/api/v1/monitor/health

# 检查 Agent 列表
curl http://localhost:8080/api/v1/agents
```

### 2. 查看日志
```bash
# 管理平台日志
sudo journalctl -u agent-platform -f

# Agent 日志
sudo journalctl -u agent -f
```

## 故障排查

### 管理平台无法启动
1. 检查数据库连接
2. 检查端口占用
3. 查看日志文件

### Agent 无法连接
1. 检查网络连接
2. 检查防火墙规则
3. 验证服务器地址配置

### 性能问题
1. 检查数据库性能
2. 调整采集间隔
3. 优化查询语句
```

**步骤 2: 创建安装文档**

创建文件 `docs/installation.md`:
```markdown
# Agent 管理平台安装指南

## 快速开始

### 前置条件
- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- protoc 编译器

### 安装步骤

1. 克隆代码仓库
```bash
git clone https://github.com/yourusername/agent-platform.git
cd agent-platform
```

2. 安装 Go 依赖
```bash
go mod download
```

3. 安装 protobuf 工具
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

4. 生成 protobuf 代码
```bash
make proto
```

5. 编译程序
```bash
make build-agent
make build-platform
```

6. 配置数据库
```bash
# 创建数据库
createdb agent_platform

# 配置连接信息
cp platform/config.example.yaml platform/config.yaml
vim platform/config.yaml
```

7. 启动管理平台
```bash
./bin/server -config platform/config.yaml
```

8. 启动 Agent
```bash
cp agent/config.example.yaml agent/config.yaml
vim agent/config.yaml
./bin/agent -config agent/config.yaml
```

9. 启动前端
```bash
cd web
npm install
npm run dev
```

10. 访问 Web UI
打开浏览器访问 http://localhost:3000

## 开发环境设置

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
```

**步骤 3: 提交部署文档**

```bash
git add docs/deployment.md docs/installation.md
git commit -m "docs: 添加部署和安装文档"
```

---

### 任务5：创建部署脚本

**文件**:
- 创建: `scripts/deploy.sh`
- 创建: `scripts/install-agent.sh`
- 创建: `docker-compose.yml`

**步骤 1: 创建管理平台部署脚本**

创建文件 `scripts/deploy.sh`:
```bash
#!/bin/bash

set -e

echo "=== Agent 管理平台部署脚本 ==="

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
  echo "请使用 root 权限运行此脚本"
  exit 1
fi

# 配置变量
INSTALL_DIR="/opt/agent-platform"
SERVICE_USER="agent"
DB_NAME="agent_platform"
DB_USER="agent_user"
DB_PASSWORD=$(openssl rand -base64 32)

echo "1. 创建服务用户..."
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /bin/false $SERVICE_USER
fi

echo "2. 创建安装目录..."
mkdir -p $INSTALL_DIR
mkdir -p /var/log/agent-platform

echo "3. 安装依赖..."
apt-get update
apt-get install -y postgresql postgresql-contrib

echo "4. 配置数据库..."
sudo -u postgres psql <<EOF
CREATE DATABASE $DB_NAME;
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF

echo "5. 复制程序文件..."
cp bin/server $INSTALL_DIR/
cp platform/config.example.yaml $INSTALL_DIR/config.yaml

echo "6. 更新配置文件..."
sed -i "s/user: \"postgres\"/user: \"$DB_USER\"/" $INSTALL_DIR/config.yaml
sed -i "s/password: \"postgres\"/password: \"$DB_PASSWORD\"/" $INSTALL_DIR/config.yaml
sed -i "s/dbname: \"agent_platform\"/dbname: \"$DB_NAME\"/" $INSTALL_DIR/config.yaml

echo "7. 设置权限..."
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
chown -R $SERVICE_USER:$SERVICE_USER /var/log/agent-platform

echo "8. 创建 systemd 服务..."
cat > /etc/systemd/system/agent-platform.service <<EOF
[Unit]
Description=Agent Platform Server
After=network.target postgresql.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/server -config $INSTALL_DIR/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "9. 启动服务..."
systemctl daemon-reload
systemctl start agent-platform
systemctl enable agent-platform

echo "=== 部署完成 ==="
echo "数据库密码: $DB_PASSWORD"
echo "请保存此密码到安全位置"
echo ""
echo "检查服务状态: systemctl status agent-platform"
echo "查看日志: journalctl -u agent-platform -f"
```

**步骤 2: 创建 Agent 安装脚本**

创建文件 `scripts/install-agent.sh`:
```bash
#!/bin/bash

set -e

echo "=== Agent 安装脚本 ==="

# 检查参数
if [ $# -lt 2 ]; then
    echo "用法: $0 <server_address> <agent_id>"
    echo "示例: $0 192.168.1.100:9090 agent-001"
    exit 1
fi

SERVER_ADDRESS=$1
AGENT_ID=$2

# 配置变量
INSTALL_DIR="/opt/agent"
SERVICE_USER="agent"

echo "1. 创建服务用户..."
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /bin/false $SERVICE_USER
fi

echo "2. 创建安装目录..."
mkdir -p $INSTALL_DIR
mkdir -p $INSTALL_DIR/plugins

echo "3. 复制程序文件..."
cp bin/agent $INSTALL_DIR/
cp agent/config.example.yaml $INSTALL_DIR/config.yaml

echo "4. 更新配置文件..."
sed -i "s/address: \"localhost:9090\"/address: \"$SERVER_ADDRESS\"/" $INSTALL_DIR/config.yaml
sed -i "s/id: \"agent-001\"/id: \"$AGENT_ID\"/" $INSTALL_DIR/config.yaml

echo "5. 设置权限..."
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR

echo "6. 创建 systemd 服务..."
cat > /etc/systemd/system/agent.service <<EOF
[Unit]
Description=Agent Platform Agent
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/agent -config $INSTALL_DIR/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "7. 启动服务..."
systemctl daemon-reload
systemctl start agent
systemctl enable agent

echo "=== 安装完成 ==="
echo "Agent ID: $AGENT_ID"
echo "Server: $SERVER_ADDRESS"
echo ""
echo "检查服务状态: systemctl status agent"
echo "查看日志: journalctl -u agent -f"
```

**步骤 3: 创建 Docker Compose 配置**

创建文件 `docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: agent_platform
      POSTGRES_USER: agent_user
      POSTGRES_PASSWORD: agent_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  server:
    build:
      context: .
      dockerfile: Dockerfile.server
    ports:
      - "8080:8080"
      - "9090:9090"
    depends_on:
      - postgres
      - redis
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=agent_user
      - DB_PASSWORD=agent_password
      - DB_NAME=agent_platform
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      - ./platform/config.yaml:/app/config.yaml

volumes:
  postgres_data:
  redis_data:
```

**步骤 4: 设置脚本执行权限**

```bash
chmod +x scripts/deploy.sh
chmod +x scripts/install-agent.sh
```

**步骤 5: 提交部署脚本**

```bash
git add scripts/ docker-compose.yml
git commit -m "feat: 添加部署脚本和 Docker Compose 配置"
```

---

## 执行计划完成

计划已保存到 `docs/plans/2026-01-18-agent-platform-phase5-implementation.md`。

**执行方式**: 在新会话中使用 `executing-plans` 技能，批量执行并设置检查点。
