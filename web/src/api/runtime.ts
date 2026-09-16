import client from './client'

export interface RuntimeStatus {
  state: string
  sub_state?: string
  pid?: number
  started_at?: string
  memory_bytes?: number
}

export interface LogEntry {
  timestamp: string
  stream: string
  level?: string
  message: string
}

export async function getRuntimeStatus(profileId: string): Promise<RuntimeStatus> {
  const { data } = await client.get(`/profiles/${profileId}/runtime`)
  return data
}

export async function runtimeAction(profileId: string, action: 'start' | 'stop' | 'restart'): Promise<RuntimeStatus> {
  const { data } = await client.post(`/profiles/${profileId}/runtime/actions`, { action })
  return data
}

export async function getRuntimeLogs(profileId: string, lines = 100): Promise<LogEntry[]> {
  const { data } = await client.get(`/profiles/${profileId}/runtime/logs?lines=${lines}`)
  if (Array.isArray(data)) {
    return data
  }
  if (typeof data === 'string') {
    return data
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean)
      .map((l) => {
        try {
          return JSON.parse(l)
        } catch {
          return null
        }
      })
      .filter((e): e is LogEntry => e !== null)
  }
  return []
}
