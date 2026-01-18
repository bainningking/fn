package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	configContent := `server:
  grpc_port: ":9090"
  http_port: ":8080"

database:
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpass"
  dbname: "testdb"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

log:
  level: "info"
  format: "json"
  output: "stdout"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configContent))
	assert.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, ":9090", cfg.Server.GRPCPort)
	assert.Equal(t, ":8080", cfg.Server.HTTPPort)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.DBName)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.Equal(t, "stdout", cfg.Log.Output)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte("invalid: yaml: content:"))
	assert.NoError(t, err)
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	assert.Error(t, err)
}
