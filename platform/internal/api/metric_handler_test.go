package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMetricTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.Metric{})
	assert.NoError(t, err)

	return db
}

func TestMetricHandler_Query(t *testing.T) {
	db := setupMetricTestDB(t)
	handler := NewMetricHandler(db)

	now := time.Now()
	metrics := []models.Metric{
		{AgentID: "agent-1", Name: "cpu", Value: 50.5, Timestamp: now},
		{AgentID: "agent-1", Name: "memory", Value: 80.2, Timestamp: now},
		{AgentID: "agent-2", Name: "cpu", Value: 30.1, Timestamp: now},
	}
	for _, metric := range metrics {
		db.Create(&metric)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", handler.Query)

	req := httptest.NewRequest("GET", "/metrics?agent_id=agent-1&name=cpu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}
