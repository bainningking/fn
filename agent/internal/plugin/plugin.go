package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sync"

	pb "github.com/yourusername/agent-platform/proto"
)

type Plugin struct {
	mu      sync.RWMutex
	info    *pb.PluginInfo
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	config  map[string]interface{}
	dataDir string
	running bool
}

func NewPlugin(name, dataDir string) *Plugin {
	return &Plugin{
		info: &pb.PluginInfo{
			Name:    name,
			Version: "1.0.0",
			Enabled: false,
		},
		config:  make(map[string]interface{}),
		dataDir: dataDir,
	}
}

func (p *Plugin) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("plugin already running")
	}

	pluginPath := filepath.Join(p.dataDir, "plugins", p.info.Name, p.info.Name)
	p.cmd = exec.Command(pluginPath)

	stdin, err := p.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	p.stdin = stdin

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	p.stdout = stdout

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	p.running = true
	p.info.Enabled = true
	return nil
}

func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	if p.stdin != nil {
		p.stdin.Close()
	}

	if p.cmd != nil && p.cmd.Process != nil {
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill plugin process: %w", err)
		}
		p.cmd.Wait()
	}

	p.running = false
	p.info.Enabled = false
	return nil
}

func (p *Plugin) SendConfig(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return fmt.Errorf("plugin not running")
	}

	p.config = config

	msg := map[string]interface{}{
		"type": "config",
		"data": config,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if _, err := p.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (p *Plugin) ReadData() (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return nil, fmt.Errorf("plugin not running")
	}

	reader := bufio.NewReader(p.stdout)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(line, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return data, nil
}

func (p *Plugin) Info() *pb.PluginInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info
}
