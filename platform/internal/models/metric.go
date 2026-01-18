package models

import "time"

type Metric struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AgentID   string    `gorm:"index" json:"agent_id"`
	Name      string    `gorm:"index" json:"name"`
	Value     float64   `json:"value"`
	Labels    string    `json:"labels"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

func (Metric) TableName() string {
	return "metrics"
}
