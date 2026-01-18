package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/agent-platform/platform/internal/monitor"
)

func TestMonitorHandler_GetMetrics(t *testing.T) {
	handler := NewMonitorHandler()

	monitor.UpdateSystemMetrics()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", handler.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestMonitorHandler_HealthCheck(t *testing.T) {
	handler := NewMonitorHandler()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.NotNil(t, resp.Data)
}
