package database

import (
	"fmt"

	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func Connect(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&models.Agent{}, &models.Task{}, &models.Metric{}); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return db, nil
}
