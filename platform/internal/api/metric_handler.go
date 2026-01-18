package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type MetricHandler struct {
	db *gorm.DB
}

func NewMetricHandler(db *gorm.DB) *MetricHandler {
	return &MetricHandler{db: db}
}

func (h *MetricHandler) Query(c *gin.Context) {
	agentID := c.Query("agent_id")
	metricName := c.Query("name")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	query := h.db.Model(&models.Metric{})

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if metricName != "" {
		query = query.Where("name = ?", metricName)
	}
	if startTime != "" {
		t, _ := time.Parse(time.RFC3339, startTime)
		query = query.Where("timestamp >= ?", t)
	}
	if endTime != "" {
		t, _ := time.Parse(time.RFC3339, endTime)
		query = query.Where("timestamp <= ?", t)
	}

	var metrics []models.Metric
	result := query.Order("timestamp DESC").Limit(1000).Find(&metrics)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, metrics)
}
