package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type TaskHandler struct {
	db *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type CreateTaskRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Content string `json:"content" binding:"required"`
	Timeout int    `json:"timeout"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, err.Error())
		return
	}

	task := &models.Task{
		AgentID: req.AgentID,
		Type:    req.Type,
		Content: req.Content,
		Timeout: req.Timeout,
		Status:  "pending",
	}

	result := h.db.Create(task)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, task)
}

func (h *TaskHandler) List(c *gin.Context) {
	agentID := c.Query("agent_id")

	var tasks []models.Task
	query := h.db.Order("created_at DESC")
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	result := query.Find(&tasks)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, tasks)
}

func (h *TaskHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, "invalid task id")
		return
	}

	var task models.Task
	result := h.db.First(&task, id)
	if result.Error != nil {
		Error(c, 404, "task not found")
		return
	}

	Success(c, task)
}
