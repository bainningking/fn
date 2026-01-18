package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/service"
)

type PluginHandler struct {
	pluginService *service.PluginService
}

func NewPluginHandler(pluginService *service.PluginService) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginService,
	}
}

type InstallPluginRequest struct {
	AgentID    string            `json:"agent_id" binding:"required"`
	PluginName string            `json:"plugin_name" binding:"required"`
	Config     map[string]string `json:"config"`
}

type UninstallPluginRequest struct {
	AgentID    string `json:"agent_id" binding:"required"`
	PluginName string `json:"plugin_name" binding:"required"`
}

func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	var req InstallPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.pluginService.InstallPlugin(req.AgentID, req.PluginName, req.Config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plugin installed successfully"})
}

func (h *PluginHandler) UninstallPlugin(c *gin.Context) {
	var req UninstallPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.pluginService.UninstallPlugin(req.AgentID, req.PluginName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plugin uninstalled successfully"})
}

func (h *PluginHandler) ListPlugins(c *gin.Context) {
	agentID := c.Query("agent_id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id is required"})
		return
	}

	plugins, err := h.pluginService.ListPlugins(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plugins": plugins})
}
