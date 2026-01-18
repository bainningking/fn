package service

import (
	"testing"

	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaskServiceIntegration(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// 迁移数据库
	if err := db.AutoMigrate(&models.Task{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// 创建任务服务
	taskService := NewTaskService(db)

	// 测试创建任务
	task := &models.Task{
		TaskID:  "test-task-001",
		AgentID: "test-agent-001",
		Type:    "shell",
		Script:  "echo 'Hello World'",
		Timeout: 30,
		Status:  "pending",
	}

	if err := taskService.CreateTask(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 测试获取任务
	result, err := taskService.GetTask("test-task-001")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if result.TaskID != "test-task-001" {
		t.Errorf("expected task_id 'test-task-001', got '%s'", result.TaskID)
	}

	if result.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", result.Status)
	}

	// 测试更新任务状态
	if err := taskService.UpdateTaskStatus("test-task-001", "running"); err != nil {
		t.Fatalf("failed to update task status: %v", err)
	}

	result, err = taskService.GetTask("test-task-001")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if result.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", result.Status)
	}

	// 测试更新任务结果
	if err := taskService.UpdateTaskResult("test-task-001", 0, "Hello World\n", ""); err != nil {
		t.Fatalf("failed to update task result: %v", err)
	}

	result, err = taskService.GetTask("test-task-001")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "Hello World\n" {
		t.Errorf("expected stdout 'Hello World\\n', got '%s'", result.Stdout)
	}

	// 测试列出任务
	tasks, err := taskService.ListTasksByAgent("test-agent-001")
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}
