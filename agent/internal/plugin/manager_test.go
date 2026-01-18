package plugin

import (
	"testing"
)

func TestPluginManager(t *testing.T) {
	manager := NewManager("/tmp/test-plugins")

	t.Run("Load Plugin", func(t *testing.T) {
		err := manager.Load("test-plugin")
		if err != nil {
			t.Errorf("Failed to load plugin: %v", err)
		}
	})

	t.Run("List Plugins", func(t *testing.T) {
		plugins := manager.List()
		if len(plugins) == 0 {
			t.Error("Expected at least one plugin")
		}
	})

	t.Run("Unload Plugin", func(t *testing.T) {
		err := manager.Unload("test-plugin")
		if err != nil {
			t.Errorf("Failed to unload plugin: %v", err)
		}

		plugins := manager.List()
		if len(plugins) != 0 {
			t.Error("Expected no plugins after unload")
		}
	})
}
