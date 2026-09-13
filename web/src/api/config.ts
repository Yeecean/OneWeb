import client from './client'

export interface OptionSchema {
  key: string
  type: string
  default?: any
  description?: string
  min_version?: string
  constraints?: { min?: number; max?: number; allowed_values?: string[] }
  deprecated?: boolean
  group?: string
}

export interface UIGroup {
  id: string
  label: string
  order: number
}

export interface UIWidget {
  widget: string
  unit?: string
  advanced?: boolean
}

export interface UISchema {
  schema_version: string
  groups: UIGroup[]
  widgets: Record<string, UIWidget>
}

export interface ConfigSchema {
  schema_version: string
  min_client_version: string
  options: OptionSchema[]
}

export interface FileFingerprint {
  mtime: string
  size: number
  sha256: string
}

export interface ConfigResponse {
  file_exists: boolean
  file_config: Record<string, string>
  effective_config: Record<string, string>
  defaults: Record<string, string>
  schema: ConfigSchema
  ui?: UISchema
  version_meta?: FileFingerprint
}

export interface SaveConfigRequest {
  desired: Record<string, string>
  base_sha256: string
  remove_keys?: string[]
  validate_display?: boolean
  validate_dry_run?: boolean
}

export interface ValidationWarning {
  key: string
  level: 'error' | 'warning'
  message: string
}

export async function getConfig(profileId: string): Promise<ConfigResponse> {
  const { data } = await client.get(`/profiles/${profileId}/config`)
  return data
}

export async function saveConfig(
  profileId: string,
  req: SaveConfigRequest,
): Promise<{ status: string; warnings: ValidationWarning[] }> {
  const { data } = await client.put(`/profiles/${profileId}/config`, req)
  return data
}

export async function validateConfig(profileId: string, content: string): Promise<{ status: string; warnings: ValidationWarning[] }> {
  const { data } = await client.post(`/profiles/${profileId}/config/validate`, content, {
    headers: { 'Content-Type': 'text/plain' },
  })
  return data
}
