# Agent 管理平台 - 阶段4：数据存储和可视化实现计划

## 概述
本阶段实现 Web UI、REST API、指标数据可视化和日志查询功能，为用户提供完整的管理界面。

## 技术栈
- 前端：React 18 + Ant Design 5 + TypeScript
- 后端：Go + Gin + GORM
- 数据库：PostgreSQL（时序数据）+ Redis（缓存）
- 实时通信：WebSocket

## 实现任务

### 任务1：初始化前端项目

**文件**:
- 创建: `web/package.json`
- 创建: `web/tsconfig.json`
- 创建: `web/vite.config.ts`
- 创建: `web/src/main.tsx`

**步骤 1: 创建前端项目结构**

```bash
cd /root/go_project/fn
mkdir -p web/src/{components,pages,services,types,utils}
cd web
```

**步骤 2: 初始化 package.json**

创建文件 `web/package.json`:
```json
{
  "name": "agent-platform-web",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "antd": "^5.12.0",
    "axios": "^1.6.0",
    "dayjs": "^1.11.10"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

**步骤 3: 创建 TypeScript 配置**

创建文件 `web/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

**步骤 4: 创建 Vite 配置**

创建文件 `web/vite.config.ts`:
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

**步骤 5: 创建入口文件**

创建文件 `web/src/main.tsx`:
```typescript
import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN}>
      <App />
    </ConfigProvider>
  </React.StrictMode>,
)
```

**步骤 6: 安装依赖**

```bash
cd web
npm install
```

**步骤 7: 提交前端初始化代码**

```bash
git add web/
git commit -m "feat: 初始化前端项目结构"
```

---

### 任务2：实现 REST API 基础框架

**文件**:
- 创建: `platform/internal/api/router.go`
- 创建: `platform/internal/api/middleware.go`
- 创建: `platform/internal/api/response.go`

**步骤 1: 实现统一响应格式**

创建文件 `platform/internal/api/response.go`:
```go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}
```

**步骤 2: 实现中间件**

创建文件 `platform/internal/api/middleware.go`:
```go
package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"duration": duration,
		}).Info("API request")
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

**步骤 3: 实现路由配置**

创建文件 `platform/internal/api/router.go`:
```go
package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.Use(Logger())
	r.Use(CORS())

	api := r.Group("/api/v1")
	{
		// Agent 管理
		agents := api.Group("/agents")
		{
			handler := NewAgentHandler(db)
			agents.GET("", handler.List)
			agents.GET("/:id", handler.Get)
			agents.DELETE("/:id", handler.Delete)
		}

		// 任务管理
		tasks := api.Group("/tasks")
		{
			handler := NewTaskHandler(db)
			tasks.POST("", handler.Create)
			tasks.GET("", handler.List)
			tasks.GET("/:id", handler.Get)
		}

		// 指标查询
		metrics := api.Group("/metrics")
		{
			handler := NewMetricHandler(db)
			metrics.GET("", handler.Query)
		}
	}

	return r
}
```

**步骤 4: 安装 Gin 依赖**

```bash
go get github.com/gin-gonic/gin
go get github.com/sirupsen/logrus
```

**步骤 5: 提交 API 基础框架**

```bash
git add platform/internal/api/
git commit -m "feat: 实现 REST API 基础框架"
```

---

### 任务3：实现 Agent 管理 API

**文件**:
- 创建: `platform/internal/api/agent_handler.go`

**步骤 1: 实现 Agent 处理器**

创建文件 `platform/internal/api/agent_handler.go`:
```go
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type AgentHandler struct {
	db *gorm.DB
}

func NewAgentHandler(db *gorm.DB) *AgentHandler {
	return &AgentHandler{db: db}
}

func (h *AgentHandler) List(c *gin.Context) {
	var agents []models.Agent
	result := h.db.Find(&agents)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, agents)
}

func (h *AgentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, "invalid agent id")
		return
	}

	var agent models.Agent
	result := h.db.First(&agent, id)
	if result.Error != nil {
		Error(c, 404, "agent not found")
		return
	}

	Success(c, agent)
}

func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, "invalid agent id")
		return
	}

	result := h.db.Delete(&models.Agent{}, id)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, nil)
}
```

**步骤 2: 提交 Agent API**

```bash
git add platform/internal/api/agent_handler.go
git commit -m "feat: 实现 Agent 管理 API"
```

---

### 任务4：实现任务管理 API

**文件**:
- 创建: `platform/internal/api/task_handler.go`

**步骤 1: 实现任务处理器**

创建文件 `platform/internal/api/task_handler.go`:
```go
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type TaskHandler struct {
	db *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type CreateTaskRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Content string `json:"content" binding:"required"`
	Timeout int    `json:"timeout"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, err.Error())
		return
	}

	task := &models.Task{
		AgentID: req.AgentID,
		Type:    req.Type,
		Content: req.Content,
		Timeout: req.Timeout,
		Status:  "pending",
	}

	result := h.db.Create(task)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, task)
}

func (h *TaskHandler) List(c *gin.Context) {
	agentID := c.Query("agent_id")

	var tasks []models.Task
	query := h.db.Order("created_at DESC")
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	result := query.Find(&tasks)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, tasks)
}

func (h *TaskHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, "invalid task id")
		return
	}

	var task models.Task
	result := h.db.First(&task, id)
	if result.Error != nil {
		Error(c, 404, "task not found")
		return
	}

	Success(c, task)
}
```

**步骤 2: 提交任务 API**

```bash
git add platform/internal/api/task_handler.go
git commit -m "feat: 实现任务管理 API"
```

---

### 任务5：实现指标数据存储和查询

**文件**:
- 创建: `platform/internal/models/metric.go`
- 创建: `platform/internal/api/metric_handler.go`

**步骤 1: 创建指标数据模型**

创建文件 `platform/internal/models/metric.go`:
```go
package models

import "time"

type Metric struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AgentID   string    `gorm:"index" json:"agent_id"`
	Name      string    `gorm:"index" json:"name"`
	Value     float64   `json:"value"`
	Labels    string    `json:"labels"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

func (Metric) TableName() string {
	return "metrics"
}
```

**步骤 2: 实现指标查询 API**

创建文件 `platform/internal/api/metric_handler.go`:
```go
package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type MetricHandler struct {
	db *gorm.DB
}

func NewMetricHandler(db *gorm.DB) *MetricHandler {
	return &MetricHandler{db: db}
}

func (h *MetricHandler) Query(c *gin.Context) {
	agentID := c.Query("agent_id")
	metricName := c.Query("name")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	query := h.db.Model(&models.Metric{})

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if metricName != "" {
		query = query.Where("name = ?", metricName)
	}
	if startTime != "" {
		t, _ := time.Parse(time.RFC3339, startTime)
		query = query.Where("timestamp >= ?", t)
	}
	if endTime != "" {
		t, _ := time.Parse(time.RFC3339, endTime)
		query = query.Where("timestamp <= ?", t)
	}

	var metrics []models.Metric
	result := query.Order("timestamp DESC").Limit(1000).Find(&metrics)
	if result.Error != nil {
		Error(c, 500, result.Error.Error())
		return
	}

	Success(c, metrics)
}
```

**步骤 3: 更新数据库迁移**

修改文件 `platform/internal/database/database.go`，添加 Metric 模型迁移：
```go
if err := db.AutoMigrate(&models.Agent{}, &models.Task{}, &models.Metric{}); err != nil {
	return nil, fmt.Errorf("failed to migrate: %w", err)
}
```

**步骤 4: 提交指标存储和查询**

```bash
git add platform/internal/models/metric.go platform/internal/api/metric_handler.go platform/internal/database/database.go
git commit -m "feat: 实现指标数据存储和查询 API"
```

---

### 任务6：集成 REST API 到主程序

**文件**:
- 修改: `platform/cmd/server/main.go`

**步骤 1: 更新主程序启动 HTTP 服务器**

修改文件 `platform/cmd/server/main.go`，添加 HTTP 服务器：
```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/agent-platform/platform/internal/api"
	"github.com/yourusername/agent-platform/platform/internal/database"
	grpcserver "github.com/yourusername/agent-platform/platform/internal/grpc"
)

func main() {
	// 连接数据库
	dbCfg := &database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "agent_platform",
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 启动 gRPC 服务器
	grpcServer := grpcserver.NewServer(":9090", db)
	go func() {
		log.Println("Starting gRPC server on :9090")
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 启动 HTTP API 服务器
	router := api.SetupRouter(db)
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Server shutting down")
	grpcServer.Stop()
}
```

**步骤 2: 构建并测试**

```bash
make build-platform
```

**步骤 3: 提交集成代码**

```bash
git add platform/cmd/server/main.go
git commit -m "feat: 集成 REST API 到管理平台主程序"
```

---

### 任务7：实现前端 Agent 列表页面

**文件**:
- 创建: `web/src/pages/AgentList.tsx`
- 创建: `web/src/services/api.ts`
- 创建: `web/src/types/index.ts`

**步骤 1: 定义类型**

创建文件 `web/src/types/index.ts`:
```typescript
export interface Agent {
  id: number
  agent_id: string
  hostname: string
  ip: string
  os: string
  arch: string
  version: string
  status: string
  last_heartbeat: string
  created_at: string
  updated_at: string
}

export interface Task {
  id: number
  agent_id: string
  type: string
  content: string
  status: string
  result: string
  created_at: string
  updated_at: string
}

export interface Metric {
  id: number
  agent_id: string
  name: string
  value: number
  labels: string
  timestamp: string
}
```

**步骤 2: 实现 API 服务**

创建文件 `web/src/services/api.ts`:
```typescript
import axios from 'axios'
import type { Agent, Task, Metric } from '../types'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

export const agentApi = {
  list: () => api.get<{ data: Agent[] }>('/agents'),
  get: (id: number) => api.get<{ data: Agent }>(`/agents/${id}`),
  delete: (id: number) => api.delete(`/agents/${id}`),
}

export const taskApi = {
  create: (data: { agent_id: string; type: string; content: string; timeout?: number }) =>
    api.post<{ data: Task }>('/tasks', data),
  list: (agentId?: string) => api.get<{ data: Task[] }>('/tasks', { params: { agent_id: agentId } }),
  get: (id: number) => api.get<{ data: Task }>(`/tasks/${id}`),
}

export const metricApi = {
  query: (params: { agent_id?: string; name?: string; start_time?: string; end_time?: string }) =>
    api.get<{ data: Metric[] }>('/metrics', { params }),
}
```

**步骤 3: 实现 Agent 列表页面**

创建文件 `web/src/pages/AgentList.tsx`:
```typescript
import React, { useEffect, useState } from 'react'
import { Table, Tag, Button, Space, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { agentApi } from '../services/api'
import type { Agent } from '../types'
import dayjs from 'dayjs'

const AgentList: React.FC = () => {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(false)

  const loadAgents = async () => {
    setLoading(true)
    try {
      const res = await agentApi.list()
      setAgents(res.data.data)
    } catch (error) {
      message.error('加载 Agent 列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadAgents()
    const timer = setInterval(loadAgents, 10000)
    return () => clearInterval(timer)
  }, [])

  const handleDelete = async (id: number) => {
    try {
      await agentApi.delete(id)
      message.success('删除成功')
      loadAgents()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const columns: ColumnsType<Agent> = [
    {
      title: 'Agent ID',
      dataIndex: 'agent_id',
      key: 'agent_id',
    },
    {
      title: '主机名',
      dataIndex: 'hostname',
      key: 'hostname',
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: '操作系统',
      dataIndex: 'os',
      key: 'os',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'online' ? 'green' : 'red'}>
          {status === 'online' ? '在线' : '离线'}
        </Tag>
      ),
    },
    {
      title: '最后心跳',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small">详情</Button>
          <Button type="link" size="small" danger onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Table
        columns={columns}
        dataSource={agents}
        loading={loading}
        rowKey="id"
      />
    </div>
  )
}

export default AgentList
```

**步骤 4: 提交前端 Agent 列表**

```bash
git add web/src/
git commit -m "feat: 实现前端 Agent 列表页面"
```

---

### 任务8：实现前端路由和布局

**文件**:
- 创建: `web/src/App.tsx`
- 创建: `web/src/components/Layout.tsx`
- 创建: `web/index.html`

**步骤 1: 实现布局组件**

创建文件 `web/src/components/Layout.tsx`:
```typescript
import React from 'react'
import { Layout as AntLayout, Menu } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { DashboardOutlined, CloudServerOutlined, FileTextOutlined, BarChartOutlined } from '@ant-design/icons'

const { Header, Sider, Content } = AntLayout

const Layout: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: '概览' },
    { key: '/agents', icon: <CloudServerOutlined />, label: 'Agent 管理' },
    { key: '/tasks', icon: <FileTextOutlined />, label: '任务管理' },
    { key: '/metrics', icon: <BarChartOutlined />, label: '指标监控' },
  ]

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header style={{ color: 'white', fontSize: '20px', fontWeight: 'bold' }}>
        Agent 管理平台
      </Header>
      <AntLayout>
        <Sider width={200} theme="light">
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
          />
        </Sider>
        <Content style={{ padding: '24px', background: '#f0f2f5' }}>
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  )
}

export default Layout
```

**步骤 2: 实现 App 组件**

创建文件 `web/src/App.tsx`:
```typescript
import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import AgentList from './pages/AgentList'

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/agents" replace />} />
          <Route path="agents" element={<AgentList />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
```

**步骤 3: 创建 HTML 入口**

创建文件 `web/index.html`:
```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Agent 管理平台</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

**步骤 4: 创建样式文件**

创建文件 `web/src/index.css`:
```css
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
```

**步骤 5: 提交前端路由和布局**

```bash
git add web/
git commit -m "feat: 实现前端路由和布局组件"
```

---

## 执行计划完成

计划已保存到 `docs/plans/2026-01-18-agent-platform-phase4-implementation.md`。

**执行方式**: 在新会话中使用 `executing-plans` 技能，批量执行并设置检查点。
