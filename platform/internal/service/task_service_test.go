package service

import (
	"testing"

	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/sqlite"
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
