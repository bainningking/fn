package plugin

import (
	"fmt"
	"sync"

	pb "github.com/yourusername/agent-platform/proto"
)

type Manager struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
	dataDir string
}

func NewManager(dataDir string) *Manager {
	return &Manager{
		plugins: make(map[string]*Plugin),
		dataDir: dataDir,
	}
}

func (m *Manager) Load(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already loaded", name)
	}

	plugin := NewPlugin(name, m.dataDir)
	m.plugins[name] = plugin
	return nil
}

func (m *Manager) Start(name string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	return plugin.Start()
}

func (m *Manager) Stop(name string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	return plugin.Stop()
}

func (m *Manager) Unload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	if err := plugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	delete(m.plugins, name)
	return nil
}

func (m *Manager) List() []*pb.PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]*pb.PluginInfo, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		infos = append(infos, plugin.Info())
	}
	return infos
}
