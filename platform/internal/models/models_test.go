package models

import (
	"testing"
	"time"
)

func TestAgentModel(t *testing.T) {
	agent := &Agent{
		AgentID:       "test-001",
		Hostname:      "test-host",
		IP:            "192.168.1.100",
		OS:            "linux",
		Arch:          "amd64",
		Version:       "1.0.0",
		Status:        "online",
		LastHeartbeat: time.Now(),
	}

	if agent.AgentID != "test-001" {
		t.Errorf("expected AgentID test-001, got %s", agent.AgentID)
	}

	if agent.Status != "online" {
		t.Errorf("expected Status online, got %s", agent.Status)
	}
}

func TestTaskModel(t *testing.T) {
	task := &Task{
		TaskID:   "task-001",
		AgentID:  "agent-001",
		Type:     "shell",
		Script:   "echo 'hello'",
		Timeout:  30,
		Status:   "pending",
	}

	if task.TaskID != "task-001" {
		t.Errorf("expected TaskID task-001, got %s", task.TaskID)
	}

	if task.Status != "pending" {
		t.Errorf("expected Status pending, got %s", task.Status)
	}
}
