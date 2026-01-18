package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"github.com/yourusername/agent-platform/platform/internal/service"
	"gorm.io/gorm"
)

type TaskHandler struct {
	taskService *service.TaskService
	db          *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{
		taskService: service.NewTaskService(db),
		db:          db,
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
	// 简单的任务 ID 生成，实际应使用 uuid 库
	return "task-" + time.Now().Format("20060102150405")
}
