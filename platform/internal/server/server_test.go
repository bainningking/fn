package server

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	srv := NewServer(":50051")
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServerStart(t *testing.T) {
	srv := NewServer(":0")
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}

	// 测试启动和停止
	go func() {
		if err := srv.Start(); err != nil {
			t.Errorf("Start failed: %v", err)
		}
	}()

	srv.Stop()
}
