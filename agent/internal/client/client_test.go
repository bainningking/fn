package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("localhost:9090", false, "test-agent-id")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.serverAddr != "localhost:9090" {
		t.Errorf("expected serverAddr localhost:9090, got %s", client.serverAddr)
	}
}
