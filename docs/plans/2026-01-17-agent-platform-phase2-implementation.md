# Agent 管理平台 阶段2 实现计划：任务执行

**目标**: 实现 Agent 端任务执行器和管理平台任务下发功能，支持远程执行 Shell 和 Python 脚本

**架构**:
- Agent 端：实现任务执行器，支持脚本执行、超时控制、资源限制
- 管理平台：实现任务管理服务、任务下发、结果存储
- 通信：通过 gRPC 双向流下发任务和返回结果

**技术栈**: Go 1.21+, gRPC, Protocol Buffers, PostgreSQL, GORM

---

## Task 1: 扩展 Protocol Buffers 消息定义

**文件**:
- 修改: `proto/agent.proto` - 添加任务相关消息
- 创建: `proto/task.proto` - 任务消息定义

**步骤 1: 创建任务消息定义**

创建文件 `proto/task.proto`:
```protobuf
syntax = "proto3";

package proto;

option go_package = "github.com/yourusername/agent-platform/proto";

import "proto/common.proto";

// 任务类型枚举
enum TaskType {
  TASK_TYPE_UNSPECIFIED = 0;
  TASK_TYPE_SHELL = 1;
  TASK_TYPE_PYTHON = 2;
}

// 任务请求
message TaskRequest {
  string task_id = 1;
  TaskType type = 2;
  string script = 3;
  int32 timeout = 4;  // 秒
  map<string, string> env = 5;  // 环境变量
}

// 任务执行结果
message TaskResult {
  string task_id = 1;
  int32 exit_code = 2;
  string stdout = 3;
  string stderr = 4;
  Timestamp completed_at = 5;
}

// 任务执行日志（流式）
message TaskLog {
  string task_id = 1;
  string output = 2;
  bool is_stderr = 3;
  Timestamp timestamp = 4;
}
```

**步骤 2: 修改 Agent 消息定义**

修改文件 `proto/agent.proto`，在 `ServerMessage` 中添加任务相关消息：
```protobuf
// 从管理平台到 Agent 的消息
message ServerMessage {
  oneof message {
    Response register_response = 1;
    Response heartbeat_ack = 2;
    TaskRequest task_request = 3;  // 新增：任务下发
  }
}

// 从 Agent 到管理平台的消息
message AgentMessage {
  oneof message {
    AgentRegister register = 1;
    Heartbeat heartbeat = 2;
    TaskResult task_result = 3;  // 新增：任务结果
    TaskLog task_log = 4;  // 新增：任务日志
  }
}
```

**步骤 3: 生成 Protocol Buffers 代码**

运行: `make proto`
预期: 生成 `proto/task.pb.go` 和更新的 `proto/agent.pb.go`

**步骤 4: 提交 Protocol Buffers 更新**

```bash
git add proto/
git commit -m "feat: 扩展 Protocol Buffers 消息定义，添加任务相关消息"
```

---

## Task 2: 实现 Agent 任务执行器

**文件**:
- 创建: `agent/internal/executor/executor.go`
- 创建: `agent/internal/executor/executor_test.go`

**步骤 1: 编写任务执行器测试**

创建文件 `agent/internal/executor/executor_test.go`:
```go
package executor

import (
	"context"
	"testing"
	"time"
)

func TestExecuteShellScript(t *testing.T) {
	executor := NewExecutor()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, "shell", "echo 'hello world'", 10)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "hello world\n" {
		t.Errorf("expected stdout 'hello world\\n', got %q", result.Stdout)
	}
}

func TestExecuteWithTimeout(t *testing.T) {
	executor := NewExecutor()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 执行一个会超时的脚本
	result, err := executor.Execute(ctx, "shell", "sleep 10", 2)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
```

**步骤 2: 运行测试确认失败**

运行: `go test ./agent/internal/executor -v`
预期: FAIL - 找不到 NewExecutor 函数

**步骤 3: 实现任务执行器**

创建文件 `agent/internal/executor/executor.go`:
```go
package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type Executor struct {
	maxConcurrent int
	semaphore     chan struct{}
}

func NewExecutor() *Executor {
	maxConcurrent := 5
	return &Executor{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

func (e *Executor) Execute(ctx context.Context, scriptType, script string, timeoutSeconds int) (*ExecutionResult, error) {
	// 获取信号量，限制并发执行数
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 创建超时上下文
	if timeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()
	}

	var cmd *exec.Cmd
	switch scriptType {
	case "shell":
		cmd = exec.CommandContext(ctx, "sh", "-c", script)
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", script)
	default:
		return nil, fmt.Errorf("unsupported script type: %s", scriptType)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecutionResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}
```

**步骤 4: 运行测试确认通过**

运行: `go test ./agent/internal/executor -v`
预期: PASS

**步骤 5: 提交任务执行器代码**

```bash
git add agent/internal/executor/
git commit -m "feat: 实现 Agent 任务执行器"
```

---

## Task 3: 实现管理平台任务数据库模型

**文件**:
- 创建: `platform/internal/models/task.go`
- 修改: `platform/internal/models/models_test.go`

**步骤 1: 创建任务模型**

创建文件 `platform/internal/models/task.go`:
```go
package models

import (
	"time"
)

type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"uniqueIndex;not null" json:"task_id"`
	AgentID   string    `gorm:"index;not null" json:"agent_id"`
	Type      string    `json:"type"`  // shell, python
	Script    string    `gorm:"type:text" json:"script"`
	Timeout   int       `json:"timeout"`
	Status    string    `json:"status"`  // pending, running, completed, failed
	ExitCode  int       `json:"exit_code"`
	Stdout    string    `gorm:"type:text" json:"stdout"`
	Stderr    string    `gorm:"type:text" json:"stderr"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	StartedAt *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (Task) TableName() string {
	return "tasks"
}
```

**步骤 2: 编写任务模型测试**

修改文件 `platform/internal/models/models_test.go`，添加任务模型测试：
```go
func TestTaskModel(t *testing.T) {
	task := &Task{
		TaskID:   "task-001",
		AgentID:  "agent-001",
		Type:     "shell",
		Script:   "echo 'hello'",
		Timeout:  30,
		Status:   "pending",
	}

	if task.TaskID != "task-001" {
		t.Errorf("expected TaskID task-001, got %s", task.TaskID)
	}

	if task.Status != "pending" {
		t.Errorf("expected Status pending, got %s", task.Status)
	}
}
```

**步骤 3: 运行测试确认通过**

运行: `go test ./platform/internal/models -v`
预期: PASS

**步骤 4: 提交任务模型代码**

```bash
git add platform/internal/models/
git commit -m "feat: 实现管理平台任务数据库模型"
```

---

## Task 4: 实现管理平台任务服务

**文件**:
- 创建: `platform/internal/service/task_service.go`
- 创建: `platform/internal/service/task_service_test.go`

**步骤 1: 编写任务服务测试**

创建文件 `platform/internal/service/task_service_test.go`:
```go
package service

import (
	"testing"

	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

func TestCreateTask(t *testing.T) {
	// 使用内存数据库进行测试
	db := setupTestDB()
	service := NewTaskService(db)

	task := &models.Task{
		TaskID:  "task-001",
		AgentID: "agent-001",
		Type:    "shell",
		Script:  "echo 'hello'",
		Timeout: 30,
		Status:  "pending",
	}

	err := service.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if task.ID == 0 {
		t.Error("expected task ID to be set")
	}
}

func setupTestDB() *gorm.DB {
	// 使用 SQLite 内存数据库进行测试
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Task{})
	return db
}
```

**步骤 2: 运行测试确认失败**

运行: `go test ./platform/internal/service -v`
预期: FAIL - 找不到 NewTaskService 函数

**步骤 3: 实现任务服务**

创建文件 `platform/internal/service/task_service.go`:
```go
package service

import (
	"fmt"

	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) CreateTask(task *models.Task) error {
	if task.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}

	result := s.db.Create(task)
	return result.Error
}

func (s *TaskService) GetTask(taskID string) (*models.Task, error) {
	var task models.Task
	result := s.db.Where("task_id = ?", taskID).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

func (s *TaskService) UpdateTaskStatus(taskID, status string) error {
	result := s.db.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Update("status", status)
	return result.Error
}

func (s *TaskService) UpdateTaskResult(taskID string, exitCode int, stdout, stderr string) error {
	result := s.db.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"exit_code": exitCode,
			"stdout":    stdout,
			"stderr":    stderr,
			"status":    "completed",
		})
	return result.Error
}

func (s *TaskService) ListTasksByAgent(agentID string) ([]models.Task, error) {
	var tasks []models.Task
	result := s.db.Where("agent_id = ?", agentID).Find(&tasks)
	return tasks, result.Error
}
```

**步骤 4: 安装 SQLite 测试依赖**

```bash
go get gorm.io/driver/sqlite
```

**步骤 5: 运行测试确认通过**

运行: `go test ./platform/internal/service -v`
预期: PASS

**步骤 6: 提交任务服务代码**

```bash
git add platform/internal/service/
git commit -m "feat: 实现管理平台任务服务"
```

---

## Task 5: 更新管理平台 gRPC 服务器处理任务

**文件**:
- 修改: `platform/internal/grpc/handler.go` - 添加任务处理逻辑
- 修改: `platform/internal/server/server.go` - 注册任务服务

**步骤 1: 修改 gRPC 处理器**

修改文件 `platform/internal/grpc/handler.go`，添加任务处理方法：
```go
func (h *AgentServiceHandler) handleTaskResult(stream pb.AgentService_ConnectServer, result *pb.TaskResult) error {
	// 更新任务结果
	taskResult := h.db.Model(&models.Task{}).
		Where("task_id = ?", result.TaskId).
		Updates(map[string]interface{}{
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
			"status":    "completed",
		})

	if taskResult.Error != nil {
		return taskResult.Error
	}

	return stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_RegisterResponse{
			RegisterResponse: &pb.Response{
				Success: true,
			},
		},
	})
}
```

**步骤 2: 在 Connect 方法中处理任务结果**

修改 `Connect` 方法的 switch 语句，添加任务结果处理：
```go
case *pb.AgentMessage_TaskResult:
	if err := h.handleTaskResult(stream, m.TaskResult); err != nil {
		return err
	}
```

**步骤 3: 提交 gRPC 处理器更新**

```bash
git add platform/internal/grpc/
git commit -m "feat: 更新 gRPC 服务器处理任务结果"
```

---

## Task 6: 实现管理平台任务 API

**文件**:
- 创建: `platform/internal/api/task_handler.go`
- 创建: `platform/internal/api/task_handler_test.go`

**步骤 1: 创建任务 API 处理器**

创建文件 `platform/internal/api/task_handler.go`:
```go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"github.com/yourusername/agent-platform/platform/internal/service"
	"gorm.io/gorm"
)

type TaskHandler struct {
	taskService *service.TaskService
	agentService *service.AgentService
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{
		taskService: service.NewTaskService(db),
		agentService: service.NewAgentService(db),
	}
}

// CreateTask 创建新任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		AgentID string `json:"agent_id" binding:"required"`
		Type    string `json:"type" binding:"required"`
		Script  string `json:"script" binding:"required"`
		Timeout int    `json:"timeout" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 Agent 存在
	agent, err := h.agentService.GetAgent(req.AgentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	task := &models.Task{
		TaskID:  generateTaskID(),
		AgentID: req.AgentID,
		Type:    req.Type,
		Script:  req.Script,
		Timeout: req.Timeout,
		Status:  "pending",
	}

	if err := h.taskService.CreateTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ListTasks 列出 Agent 的所有任务
func (h *TaskHandler) ListTasks(c *gin.Context) {
	agentID := c.Query("agent_id")

	tasks, err := h.taskService.ListTasksByAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func generateTaskID() string {
	// 简单的 UUID 生成，实际应使用 uuid 库
	return "task-" + time.Now().Format("20060102150405")
}
```

**步骤 2: 安装 Gin 依赖**

```bash
go get github.com/gin-gonic/gin
```

**步骤 3: 提交任务 API 代码**

```bash
git add platform/internal/api/
git commit -m "feat: 实现管理平台任务 API"
```

---

## Task 7: 更新 Agent 客户端处理任务下发

**文件**:
- 修改: `agent/internal/client/client.go` - 添加任务处理逻辑

**步骤 1: 修改 Agent 客户端**

修改文件 `agent/internal/client/client.go`，添加任务处理：
```go
package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/yourusername/agent-platform/agent/internal/executor"
	pb "github.com/yourusername/agent-platform/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	serverAddr string
	useTLS     bool
	conn       *grpc.ClientConn
	executor   *executor.Executor
	agentID    string
}

func NewClient(serverAddr string, useTLS bool, agentID string) *Client {
	return &Client{
		serverAddr: serverAddr,
		useTLS:     useTLS,
		executor:   executor.NewExecutor(),
		agentID:    agentID,
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

	// 启动双向流
	go c.startStream()

	return nil
}

func (c *Client) startStream() {
	client := pb.NewAgentServiceClient(c.conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Printf("Failed to create stream: %v", err)
		return
	}

	// 发送注册消息
	c.sendRegister(stream)

	// 处理来自服务器的消息
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream closed by server")
			return
		}
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}

		c.handleServerMessage(stream, msg)
	}
}

func (c *Client) handleServerMessage(stream pb.AgentService_ConnectServer, msg *pb.ServerMessage) {
	switch m := msg.Message.(type) {
	case *pb.ServerMessage_TaskRequest:
		c.handleTaskRequest(stream, m.TaskRequest)
	}
}

func (c *Client) handleTaskRequest(stream pb.AgentService_ConnectServer, taskReq *pb.TaskRequest) {
	log.Printf("Received task: %s", taskReq.TaskId)

	// 执行任务
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(taskReq.Timeout)*time.Second)
	defer cancel()

	result, err := c.executor.Execute(ctx, taskReq.Type, taskReq.Script, taskReq.Timeout)
	if err != nil {
		log.Printf("Task execution failed: %v", err)
		return
	}

	// 发送结果
	taskResult := &pb.TaskResult{
		TaskId:   taskReq.TaskId,
		ExitCode: int32(result.ExitCode),
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
	}

	msg := &pb.AgentMessage{
		Message: &pb.AgentMessage_TaskResult{
			TaskResult: taskResult,
		},
	}

	if err := stream.Send(msg); err != nil {
		log.Printf("Failed to send task result: %v", err)
	}
}

func (c *Client) sendRegister(stream pb.AgentService_ConnectServer) {
	// 实现注册逻辑
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
```

**步骤 2: 提交 Agent 客户端更新**

```bash
git add agent/internal/client/
git commit -m "feat: 更新 Agent 客户端处理任务下发"
```

---

## Task 8: 集成测试

**文件**:
- 创建: `tests/integration_test.go`

**步骤 1: 创建集成测试**

创建文件 `tests/integration_test.go`:
```go
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/agent-platform/agent/internal/executor"
)

func TestTaskExecution(t *testing.T) {
	executor := executor.NewExecutor()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, "shell", "echo 'test'", 10)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}
```

**步骤 2: 运行集成测试**

运行: `go test ./tests -v`
预期: PASS

**步骤 3: 提交集成测试**

```bash
git add tests/
git commit -m "test: 添加任务执行集成测试"
```

---

## 阶段2 完成

阶段2的任务执行功能已完成。现在系统支持：

✅ 管理平台通过 gRPC 下发任务到 Agent
✅ Agent 执行 Shell 和 Python 脚本
✅ Agent 返回任务执行结果
✅ 管理平台存储任务记录和结果
✅ REST API 创建和查询任务

## 下一步：阶段3 - 插件系统

阶段3将实现：
- 插件管理框架（加载、卸载、配置）
- 内置插件：CPU、内存、磁盘、网络、日志 tail
- 插件远程安装和卸载功能
- Web UI 插件管理界面
