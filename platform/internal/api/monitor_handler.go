package api

import (
	"time"

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
