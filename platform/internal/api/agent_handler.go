package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type AgentHandler struct {
	db *gorm.DB
}

func NewAgentHandler(db *gorm.DB) *AgentHandler {
	return &AgentHandler{db: db}
}

func (h *AgentHandler) List(c *gin.Context) {
	var agents []models.Agent
	result := h.db.Find(&agents)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, agents)
}

func (h *AgentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, "invalid agent id")
		return
	}

	var agent models.Agent
	result := h.db.First(&agent, id)
	if result.Error != nil {
		Error(c, 404, "agent not found")
		return
	}

	Success(c, agent)
}

func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, "invalid agent id")
		return
	}

	result := h.db.Delete(&models.Agent{}, id)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, nil)
}
