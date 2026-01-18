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
  create: (data: { agent_id: string; type: string; script: string; timeout?: number }) =>
    api.post<{ data: Task }>('/tasks', data),
  list: (agentId?: string) => api.get<{ data: Task[] }>('/tasks', { params: { agent_id: agentId } }),
  get: (id: number) => api.get<{ data: Task }>(`/tasks/${id}`),
}

export const metricApi = {
  query: (params: { agent_id?: string; name?: string; start_time?: string; end_time?: string }) =>
    api.get<{ data: Metric[] }>('/metrics', { params }),
}
