import client from './client'

export interface FeatureSupport {
  config_editor: boolean
  sync_list: boolean
  device_auth: boolean
  runtime_ctrl: boolean
  auth_assistant: boolean
  schema_version: string
  recommended: boolean
}

export interface RuntimeCapability {
  available: boolean
  version?: string
  features?: string[]
  warnings?: string[]
}

export interface SystemInfo {
  version: string
  hostname?: string
  client_version?: string
  capabilities?: FeatureSupport
  runtimes: Record<string, RuntimeCapability>
}

export async function getSystemInfo(): Promise<SystemInfo> {
  const { data } = await client.get('/system/info')
  return data
}
