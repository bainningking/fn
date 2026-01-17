# Agent 管理平台 MVP 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标**: 构建 Agent 管理平台的核心基础（MVP），实现 Agent 与管理平台的连接、心跳机制和基本指标采集

**架构**: 使用 gRPC 双向流实现 Agent 与管理平台的长连接通信。Agent 端负责采集指标并上报，管理平台端负责接收连接、存储数据和提供 Web UI。

**技术栈**: Go 1.21+, gRPC, Protocol Buffers, PostgreSQL, Redis, Gin, GORM

---

## Task 1: 项目初始化和依赖管理

**文件**:
- 创建: `go.mod`
- 创建: `Makefile`
- 创建: `.gitignore`

**步骤 1: 初始化 Go 模块**

```bash
go mod init github.com/yourusername/agent-platform
```

**步骤 2: 创建 .gitignore 文件**

创建文件 `.gitignore`:
```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
/agent/cmd/agent/agent
/platform/cmd/server/server

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store

# Config files with secrets
*.local.yaml
config.local.yaml

# Build artifacts
/bin/
/dist/

# Database
*.db
*.sqlite

# Logs
*.log
/logs/

# Plugins
/plugins/*.so
```

**步骤 3: 创建 Makefile**

创建文件 `Makefile`:
```makefile
.PHONY: proto build-agent build-platform test clean

# 生成 protobuf 代码
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# 构建 Agent
build-agent:
	go build -o bin/agent ./agent/cmd/agent

# 构建管理平台
build-platform:
	go build -o bin/server ./platform/cmd/server

# 运行测试
test:
	go test -v ./...

# 清理构建产物
clean:
	rm -rf bin/
```

**步骤 4: 提交初始化代码**

```bash
git add go.mod .gitignore Makefile
git commit -m "chore: 初始化项目结构和构建配置"
```

---

## Task 2: 定义 Protocol Buffers 消息格式

**文件**:
- 创建: `proto/common.proto`
- 创建: `proto/agent.proto`
- 创建: `proto/metric.proto`

**步骤 1: 创建通用消息定义**

创建文件 `proto/common.proto`:
```protobuf
syntax = "proto3";

package proto;

option go_package = "github.com/yourusername/agent-platform/proto";

// 时间戳
message Timestamp {
  int64 seconds = 1;
  int32 nanos = 2;
}

// 通用响应
message Response {
  bool success = 1;
  string message = 2;
  string error = 3;
}
```

**步骤 2: 创建 Agent 消息定义**

创建文件 `proto/agent.proto`:
```protobuf
syntax = "proto3";

package proto;

option go_package = "github.com/yourusername/agent-platform/proto";

import "proto/common.proto";

// Agent 注册信息
message AgentRegister {
  string agent_id = 1;
  string hostname = 2;
  string ip = 3;
  string os = 4;
  string arch = 5;
  string version = 6;
}

// 心跳消息
message Heartbeat {
  string agent_id = 1;
  Timestamp timestamp = 2;
}

// 从管理平台到 Agent 的消息
message ServerMessage {
  oneof message {
    Response register_response = 1;
    Response heartbeat_ack = 2;
  }
}

// 从 Agent 到管理平台的消息
message AgentMessage {
  oneof message {
    AgentRegister register = 1;
    Heartbeat heartbeat = 2;
  }
}

// Agent 服务定义
service AgentService {
  // 双向流连接
  rpc Connect(stream AgentMessage) returns (stream ServerMessage);
}
```

**步骤 3: 创建指标消息定义**

创建文件 `proto/metric.proto`:
```protobuf
syntax = "proto3";

package proto;

option go_package = "github.com/yourusername/agent-platform/proto";

import "proto/common.proto";

// 指标数据点
message MetricPoint {
  string name = 1;
  double value = 2;
  Timestamp timestamp = 3;
  map<string, string> labels = 4;
}

// 批量指标数据
message MetricBatch {
  string agent_id = 1;
  repeated MetricPoint metrics = 2;
}
```

**步骤 4: 安装 protobuf 编译器依赖**

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**步骤 5: 生成 Go 代码**

运行: `make proto`
预期: 在 `proto/` 目录下生成 `.pb.go` 文件

**步骤 6: 提交 protobuf 定义**

```bash
git add proto/ Makefile
git commit -m "feat: 添加 Protocol Buffers 消息定义"
```

---

## Task 3: 实现 Agent 配置管理

**文件**:
- 创建: `agent/internal/config/config.go`
- 创建: `agent/internal/config/config_test.go`
- 创建: `agent/config.example.yaml`

**步骤 1: 编写配置结构测试**

创建文件 `agent/internal/config/config_test.go`:
```go
package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	content := `
server:
  address: "localhost:9090"
  tls: false

agent:
  id: "test-agent"
  collect_interval: 30
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// 测试加载配置
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Address != "localhost:9090" {
		t.Errorf("expected address localhost:9090, got %s", cfg.Server.Address)
	}

	if cfg.Agent.ID != "test-agent" {
		t.Errorf("expected agent_id test-agent, got %s", cfg.Agent.ID)
	}

	if cfg.Agent.CollectInterval != 30 {
		t.Errorf("expected collect_interval 30, got %d", cfg.Agent.CollectInterval)
	}
}
```

**步骤 2: 运行测试确认失败**

运行: `go test ./agent/internal/config -v`
预期: FAIL - 找不到 LoadConfig 函数

**步骤 3: 实现配置加载**

创建文件 `agent/internal/config/config.go`:
```go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Agent  AgentConfig  `yaml:"agent"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	TLS     bool   `yaml:"tls"`
}

type AgentConfig struct {
	ID              string `yaml:"id"`
	CollectInterval int    `yaml:"collect_interval"`
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

**步骤 4: 安装依赖**

```bash
go get gopkg.in/yaml.v3
```

**步骤 5: 运行测试确认通过**

运行: `go test ./agent/internal/config -v`
预期: PASS

**步骤 6: 创建示例配置文件**

创建文件 `agent/config.example.yaml`:
```yaml
server:
  address: "localhost:9090"
  tls: false

agent:
  id: "agent-001"
  collect_interval: 30
```

**步骤 7: 提交配置管理代码**

```bash
git add agent/internal/config/ agent/config.example.yaml
git commit -m "feat: 实现 Agent 配置管理模块"
```

---

## Task 4: 实现 Agent gRPC 客户端基础框架

**文件**:
- 创建: `agent/internal/client/client.go`
- 创建: `agent/internal/client/client_test.go`

**步骤 1: 编写客户端连接测试**

创建文件 `agent/internal/client/client_test.go`:
```go
package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("localhost:9090", false)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.serverAddr != "localhost:9090" {
		t.Errorf("expected serverAddr localhost:9090, got %s", client.serverAddr)
	}
}
```

**步骤 2: 运行测试确认失败**

运行: `go test ./agent/internal/client -v`
预期: FAIL - 找不到 NewClient 函数

**步骤 3: 实现客户端基础结构**

创建文件 `agent/internal/client/client.go`:
```go
package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	serverAddr string
	useTLS     bool
	conn       *grpc.ClientConn
}

func NewClient(serverAddr string, useTLS bool) *Client {
	return &Client{
		serverAddr: serverAddr,
		useTLS:     useTLS,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	var opts []grpc.DialOption
	if !c.useTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, c.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
```

**步骤 4: 安装 gRPC 依赖**

```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

**步骤 5: 运行测试确认通过**

运行: `go test ./agent/internal/client -v`
预期: PASS

**步骤 6: 提交客户端代码**

```bash
git add agent/internal/client/
git commit -m "feat: 实现 Agent gRPC 客户端基础框架"
```

---

## Task 5: 实现管理平台数据库模型

**文件**:
- 创建: `platform/internal/models/agent.go`
- 创建: `platform/internal/models/models_test.go`

**步骤 1: 编写 Agent 模型测试**

创建文件 `platform/internal/models/models_test.go`:
```go
package models

import (
	"testing"
	"time"
)

func TestAgentModel(t *testing.T) {
	agent := &Agent{
		AgentID:       "test-001",
		Hostname:      "test-host",
		IP:            "192.168.1.100",
		OS:            "linux",
		Arch:          "amd64",
		Version:       "1.0.0",
		Status:        "online",
		LastHeartbeat: time.Now(),
	}

	if agent.AgentID != "test-001" {
		t.Errorf("expected AgentID test-001, got %s", agent.AgentID)
	}

	if agent.Status != "online" {
		t.Errorf("expected Status online, got %s", agent.Status)
	}
}
```

**步骤 2: 运行测试确认失败**

运行: `go test ./platform/internal/models -v`
预期: FAIL - 找不到 Agent 结构体

**步骤 3: 实现 Agent 模型**

创建文件 `platform/internal/models/agent.go`:
```go
package models

import (
	"time"
)

type Agent struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AgentID       string    `gorm:"uniqueIndex;not null" json:"agent_id"`
	Hostname      string    `json:"hostname"`
	IP            string    `json:"ip"`
	OS            string    `json:"os"`
	Arch          string    `json:"arch"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Agent) TableName() string {
	return "agents"
}
```

**步骤 4: 运行测试确认通过**

运行: `go test ./platform/internal/models -v`
预期: PASS

**步骤 5: 提交模型代码**

```bash
git add platform/internal/models/
git commit -m "feat: 实现管理平台数据库模型"
```

---

## Task 6: 实现管理平台 gRPC 服务器

**文件**:
- 创建: `platform/internal/grpc/server.go`
- 创建: `platform/internal/grpc/handler.go`

**步骤 1: 实现 gRPC 服务器基础结构**

创建文件 `platform/internal/grpc/server.go`:
```go
package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Server struct {
	addr   string
	db     *gorm.DB
	server *grpc.Server
}

func NewServer(addr string, db *gorm.DB) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()

	// 注册服务
	// proto.RegisterAgentServiceServer(s.server, &AgentServiceHandler{db: s.db})

	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
```

**步骤 2: 实现 Agent 服务处理器**

创建文件 `platform/internal/grpc/handler.go`:
```go
package grpc

import (
	"io"
	"log"
	"time"

	pb "github.com/yourusername/agent-platform/proto"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type AgentServiceHandler struct {
	pb.UnimplementedAgentServiceServer
	db *gorm.DB
}

func (h *AgentServiceHandler) Connect(stream pb.AgentService_ConnectServer) error {
	log.Println("New agent connection")

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch m := msg.Message.(type) {
		case *pb.AgentMessage_Register:
			if err := h.handleRegister(stream, m.Register); err != nil {
				return err
			}
		case *pb.AgentMessage_Heartbeat:
			if err := h.handleHeartbeat(stream, m.Heartbeat); err != nil {
				return err
			}
		}
	}
}

func (h *AgentServiceHandler) handleRegister(stream pb.AgentService_ConnectServer, reg *pb.AgentRegister) error {
	agent := &models.Agent{
		AgentID:       reg.AgentId,
		Hostname:      reg.Hostname,
		IP:            reg.Ip,
		OS:            reg.Os,
		Arch:          reg.Arch,
		Version:       reg.Version,
		Status:        "online",
		LastHeartbeat: time.Now(),
	}

	result := h.db.Where("agent_id = ?", reg.AgentId).FirstOrCreate(agent)
	if result.Error != nil {
		return stream.Send(&pb.ServerMessage{
			Message: &pb.ServerMessage_RegisterResponse{
				RegisterResponse: &pb.Response{
					Success: false,
					Error:   result.Error.Error(),
				},
			},
		})
	}

	log.Printf("Agent registered: %s", reg.AgentId)

	return stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_RegisterResponse{
			RegisterResponse: &pb.Response{
				Success: true,
				Message: "Agent registered successfully",
			},
		},
	})
}

func (h *AgentServiceHandler) handleHeartbeat(stream pb.AgentService_ConnectServer, hb *pb.Heartbeat) error {
	result := h.db.Model(&models.Agent{}).
		Where("agent_id = ?", hb.AgentId).
		Updates(map[string]interface{}{
			"last_heartbeat": time.Now(),
			"status":         "online",
		})

	if result.Error != nil {
		return result.Error
	}

	return stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_HeartbeatAck{
			HeartbeatAck: &pb.Response{
				Success: true,
			},
		},
	})
}
```

**步骤 3: 安装 GORM 依赖**

```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
```

**步骤 4: 提交 gRPC 服务器代码**

```bash
git add platform/internal/grpc/
git commit -m "feat: 实现管理平台 gRPC 服务器"
```

---

## Task 7: 实现 Agent 主程序

**文件**:
- 创建: `agent/cmd/agent/main.go`

**步骤 1: 实现 Agent 主程序**

创建文件 `agent/cmd/agent/main.go`:
```go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/agent-platform/agent/internal/client"
	"github.com/yourusername/agent-platform/agent/internal/config"
	pb "github.com/yourusername/agent-platform/proto"
)

func main() {
	configPath := flag.String("config", "agent/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建客户端
	client := client.NewClient(cfg.Server.Address, cfg.Server.TLS)

	// 连接到服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	log.Printf("Agent %s started", cfg.Agent.ID)

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Agent shutting down")
}
```

**步骤 2: 构建 Agent**

运行: `make build-agent`
预期: 成功生成 `bin/agent`

**步骤 3: 提交 Agent 主程序**

```bash
git add agent/cmd/agent/
git commit -m "feat: 实现 Agent 主程序入口"
```

---

## Task 8: 实现管理平台主程序

**文件**:
- 创建: `platform/cmd/server/main.go`
- 创建: `platform/internal/database/database.go`

**步骤 1: 实现数据库连接**

创建文件 `platform/internal/database/database.go`:
```go
package database

import (
	"fmt"

	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func Connect(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&models.Agent{}); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return db, nil
}
```

**步骤 2: 实现管理平台主程序**

创建文件 `platform/cmd/server/main.go`:
```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/agent-platform/platform/internal/database"
	grpcserver "github.com/yourusername/agent-platform/platform/internal/grpc"
)

func main() {
	// 连接数据库
	dbCfg := &database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "agent_platform",
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 启动 gRPC 服务器
	server := grpcserver.NewServer(":9090", db)

	go func() {
		log.Println("Starting gRPC server on :9090")
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Server shutting down")
	server.Stop()
}
```

**步骤 3: 构建管理平台**

运行: `make build-platform`
预期: 成功生成 `bin/server`

**步骤 4: 提交管理平台主程序**

```bash
git add platform/cmd/server/ platform/internal/database/
git commit -m "feat: 实现管理平台主程序和数据库连接"
```

---

## 执行计划完成

计划已保存到 `docs/plans/2026-01-17-agent-platform-mvp-implementation.md`。

**两种执行选项**:

**1. 子代理驱动（当前会话）** - 我为每个任务分派新的子代理，任务之间进行审查，快速迭代

**2. 并行会话（独立）** - 在新会话中使用 executing-plans，批量执行并设置检查点

您选择哪种方式？
