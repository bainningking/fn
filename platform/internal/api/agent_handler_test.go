package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.Agent{})
	assert.NoError(t, err)

	return db
}

func TestAgentHandler_List(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAgentHandler(db)

	// 创建测试数据
	agents := []models.Agent{
		{AgentID: "agent-1", Hostname: "host1", Status: "online"},
		{AgentID: "agent-2", Hostname: "host2", Status: "offline"},
	}
	for _, agent := range agents {
		db.Create(&agent)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/agents", handler.List)

	req := httptest.NewRequest("GET", "/agents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestAgentHandler_Get(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAgentHandler(db)

	agent := models.Agent{AgentID: "agent-1", Hostname: "host1", Status: "online"}
	db.Create(&agent)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/agents/:id", handler.Get)

	req := httptest.NewRequest("GET", "/agents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestAgentHandler_Delete(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAgentHandler(db)

	agent := models.Agent{AgentID: "agent-1", Hostname: "host1", Status: "online"}
	db.Create(&agent)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/agents/:id", handler.Delete)

	req := httptest.NewRequest("DELETE", "/agents/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	db.Model(&models.Agent{}).Count(&count)
	assert.Equal(t, int64(0), count)
}
