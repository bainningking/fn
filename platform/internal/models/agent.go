package models

import (
	"time"
)

type Agent struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AgentID       string    `gorm:"uniqueIndex;not null" json:"agent_id"`
	Hostname      string    `json:"hostname"`
	IP            string    `json:"ip"`
	OS            string    `json:"os"`
	Arch          string    `json:"arch"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Agent) TableName() string {
	return "agents"
}
