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
