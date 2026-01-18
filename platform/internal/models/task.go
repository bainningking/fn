package models

import (
	"time"
)

type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"uniqueIndex;not null" json:"task_id"`
	AgentID   string    `gorm:"index;not null" json:"agent_id"`
	Type      string    `json:"type"`  // shell, python
	Script    string    `gorm:"type:text" json:"script"`
	Timeout   int       `json:"timeout"`
	Status    string    `json:"status"`  // pending, running, completed, failed
	ExitCode  int       `json:"exit_code"`
	Stdout    string    `gorm:"type:text" json:"stdout"`
	Stderr    string    `gorm:"type:text" json:"stderr"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	StartedAt *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (Task) TableName() string {
	return "tasks"
}
