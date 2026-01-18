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
  script: string
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
