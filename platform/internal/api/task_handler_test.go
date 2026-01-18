package api

import (
	"bytes"
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

func setupTaskTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.Task{})
	assert.NoError(t, err)

	return db
}

func TestTaskHandler_Create(t *testing.T) {
	db := setupTaskTestDB(t)
	handler := NewTaskHandler(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", handler.Create)

	reqBody := CreateTaskRequest{
		AgentID: "agent-1",
		Type:    "shell",
		Script:  "echo test",
		Timeout: 30,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestTaskHandler_List(t *testing.T) {
	db := setupTaskTestDB(t)
	handler := NewTaskHandler(db)

	tasks := []models.Task{
		{AgentID: "agent-1", Type: "shell", Script: "test1", Status: "pending"},
		{AgentID: "agent-2", Type: "shell", Script: "test2", Status: "completed"},
	}
	for _, task := range tasks {
		db.Create(&task)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks", handler.List)

	req := httptest.NewRequest("GET", "/tasks?agent_id=agent-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestTaskHandler_Get(t *testing.T) {
	db := setupTaskTestDB(t)
	handler := NewTaskHandler(db)

	task := models.Task{AgentID: "agent-1", Type: "shell", Script: "test", Status: "pending"}
	db.Create(&task)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks/:id", handler.Get)

	req := httptest.NewRequest("GET", "/tasks/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}
