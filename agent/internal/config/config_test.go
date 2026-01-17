package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	content := `
server:
  address: "localhost:9090"
  tls: false

agent:
  id: "test-agent"
  collect_interval: 30
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// 测试加载配置
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Server.Address != "localhost:9090" {
		t.Errorf("expected address localhost:9090, got %s", cfg.Server.Address)
	}

	if cfg.Agent.ID != "test-agent" {
		t.Errorf("expected agent_id test-agent, got %s", cfg.Agent.ID)
	}

	if cfg.Agent.CollectInterval != 30 {
		t.Errorf("expected collect_interval 30, got %d", cfg.Agent.CollectInterval)
	}
}
