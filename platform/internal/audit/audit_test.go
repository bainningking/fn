package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.AuditLog{})
	assert.NoError(t, err)

	return db
}

func TestService_Log(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	log := &models.AuditLog{
		UserID:    "user-1",
		Action:    "GET /api/v1/agents",
		Resource:  "/api/v1/agents",
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
		Status:    "200",
		CreatedAt: time.Now(),
	}

	err := service.Log(log)
	assert.NoError(t, err)

	var count int64
	db.Model(&models.AuditLog{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestService_Query(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	logs := []models.AuditLog{
		{UserID: "user-1", Action: "GET /agents", CreatedAt: time.Now()},
		{UserID: "user-1", Action: "POST /tasks", CreatedAt: time.Now()},
		{UserID: "user-2", Action: "GET /agents", CreatedAt: time.Now()},
	}

	for _, log := range logs {
		db.Create(&log)
	}

	result, err := service.Query("user-1", "", 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	result, err = service.Query("", "GET /agents", 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	result, err = service.Query("", "", 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}
