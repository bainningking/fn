package service

import (
	"fmt"

	pb "github.com/yourusername/agent-platform/proto"
	"gorm.io/gorm"
)

type PluginService struct {
	db *gorm.DB
}

func NewPluginService(db *gorm.DB) *PluginService {
	return &PluginService{db: db}
}

func (s *PluginService) InstallPlugin(agentID, pluginName string, config map[string]string) error {
	// 验证 Agent 是否存在
	var count int64
	if err := s.db.Table("agents").Where("agent_id = ?", agentID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check agent: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// 这里实际上会通过 gRPC 流发送安装指令给 Agent
	// 具体实现在 gRPC handler 中
	return nil
}

func (s *PluginService) UninstallPlugin(agentID, pluginName string) error {
	// 验证 Agent 是否存在
	var count int64
	if err := s.db.Table("agents").Where("agent_id = ?", agentID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check agent: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// 这里实际上会通过 gRPC 流发送卸载指令给 Agent
	return nil
}

func (s *PluginService) ListPlugins(agentID string) ([]*pb.PluginInfo, error) {
	// 验证 Agent 是否存在
	var count int64
	if err := s.db.Table("agents").Where("agent_id = ?", agentID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check agent: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	// 这里实际上会通过 gRPC 流请求 Agent 的插件列表
	return nil, nil
}

func (s *PluginService) UpdatePluginConfig(agentID, pluginName string, config map[string]string) error {
	// 验证 Agent 是否存在
	var count int64
	if err := s.db.Table("agents").Where("agent_id = ?", agentID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check agent: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// 这里实际上会通过 gRPC 流发送配置更新指令给 Agent
	return nil
}
