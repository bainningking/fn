package audit

import (
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Log(log *models.AuditLog) error {
	return s.db.Create(log).Error
}

func (s *Service) Query(userID, action string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := s.db.Order("created_at DESC")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}
