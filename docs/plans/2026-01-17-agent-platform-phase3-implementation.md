# Agent 管理平台 - 阶段3：插件系统实现计划

## 概述
本阶段实现 Agent 的插件系统，包括插件加载、生命周期管理、配置系统以及远程安装卸载功能。

## 架构设计

### 插件通信模型
- 插件作为独立进程运行
- 通过 stdin/stdout 进行 JSON 消息通信
- Agent 负责启动、停止和监控插件进程

### 插件消息格式
```json
// Agent -> Plugin
{
  "type": "config",
  "data": { "interval": 60 }
}

// Plugin -> Agent
{
  "type": "metric",
  "data": {
    "name": "cpu_usage",
    "value": 45.2,
    "timestamp": 1234567890
  }
}
```

## 实现任务

### 任务1：扩展 Protocol Buffers 定义
**文件**: `proto/plugin.proto`

创建新的 proto 文件定义插件相关消息：
- PluginInfo：插件元数据（名称、版本、描述）
- PluginConfig：插件配置
- InstallPluginRequest/Response：安装插件
- UninstallPluginRequest/Response：卸载插件
- ListPluginsRequest/Response：列出插件

**依赖**: 无

---

### 任务2：实现插件管理器
**文件**: `agent/internal/plugin/manager.go`

实现插件生命周期管理：
```go
type Manager struct {
    plugins map[string]*Plugin
    dataDir string
}

func (m *Manager) Load(name string) error
func (m *Manager) Start(name string) error
func (m *Manager) Stop(name string) error
func (m *Manager) Unload(name string) error
func (m *Manager) List() []PluginInfo
```

**依赖**: 任务1

---

### 任务3：实现插件进程通信
**文件**: `agent/internal/plugin/plugin.go`

实现单个插件的进程管理和通信：
```go
type Plugin struct {
    info    PluginInfo
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    config  map[string]interface{}
}

func (p *Plugin) Start() error
func (p *Plugin) Stop() error
func (p *Plugin) SendConfig(config map[string]interface{}) error
func (p *Plugin) ReadData() (PluginData, error)
```

**依赖**: 任务1

---

### 任务4：开发内置插件 - CPU 监控
**文件**: `plugins/cpu/main.go`

实现 CPU 使用率监控插件：
- 读取 `/proc/stat` 获取 CPU 数据
- 计算使用率百分比
- 通过 stdout 输出 JSON 格式指标

**依赖**: 无

---

### 任务5：开发内置插件 - 内存监控
**文件**: `plugins/memory/main.go`

实现内存使用监控插件：
- 读取 `/proc/meminfo` 获取内存数据
- 计算使用率和可用内存
- 通过 stdout 输出 JSON 格式指标

**依赖**: 无

---

### 任务6：实现管理平台插件服务
**文件**: `platform/internal/service/plugin_service.go`

实现插件管理服务：
```go
type PluginService struct {
    db *gorm.DB
}

func (s *PluginService) InstallPlugin(agentID, pluginName string, config map[string]interface{}) error
func (s *PluginService) UninstallPlugin(agentID, pluginName string) error
func (s *PluginService) ListPlugins(agentID string) ([]PluginInfo, error)
func (s *PluginService) UpdatePluginConfig(agentID, pluginName string, config map[string]interface{}) error
```

**依赖**: 任务1

---

### 任务7：更新 gRPC 服务器处理插件操作
**文件**: `platform/internal/grpc/handler.go`

在双向流处理中添加插件操作：
- 接收插件安装/卸载请求
- 通过流发送给对应 Agent
- 接收 Agent 的插件操作结果

**依赖**: 任务1, 任务6

---

### 任务8：实现管理平台插件 API
**文件**: `platform/internal/api/plugin_handler.go`

实现 REST API 端点：
- POST `/api/agents/:id/plugins` - 安装插件
- DELETE `/api/agents/:id/plugins/:name` - 卸载插件
- GET `/api/agents/:id/plugins` - 列出插件
- PUT `/api/agents/:id/plugins/:name/config` - 更新配置

**依赖**: 任务6

---

### 任务9：更新 Agent 客户端处理插件操作
**文件**: `agent/internal/client/client.go`

在双向流处理中添加：
- 接收插件安装/卸载指令
- 调用插件管理器执行操作
- 返回操作结果给管理平台

**依赖**: 任务2, 任务3

---

### 任务10：集成测试
**文件**: `test/plugin_test.go`

测试场景：
1. 启动 Agent 和管理平台
2. 通过 API 安装 CPU 插件
3. 验证插件进程启动
4. 验证指标数据上报
5. 通过 API 卸载插件
6. 验证插件进程停止

**依赖**: 所有前置任务

---

## 技术细节

### 插件目录结构
```
agent/
  plugins/
    cpu/
      cpu           # 可执行文件
      config.json   # 配置文件
    memory/
      memory
      config.json
```

### 插件配置示例
```yaml
# agent/config.yaml
plugins:
  cpu:
    enabled: true
    interval: 60
  memory:
    enabled: true
    interval: 60
```

### 错误处理
- 插件崩溃自动重启（最多3次）
- 插件无响应超时检测（30秒）
- 插件输出格式错误记录日志

## 验收标准
- [ ] 插件可以独立进程运行
- [ ] 支持动态安装/卸载插件
- [ ] CPU 和内存插件正常采集数据
- [ ] 管理平台可以远程管理插件
- [ ] 插件崩溃能自动重启
- [ ] 所有集成测试通过

## 预计工作量
10个任务，建议使用并行会话执行。
