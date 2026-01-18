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
