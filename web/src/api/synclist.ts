import client from './client'

export interface SyncRule {
  raw_text: string
  type: 'include' | 'exclude' | 'comment' | 'blank'
  pattern?: string
  is_rooted?: boolean
  line_number: number
}

export interface ValidationWarning {
  line: number
  rule: string
  level: 'performance' | 'ordering'
  message: string
}

export interface FileFingerprint {
  mtime: string
  size: number
  sha256: string
}

export interface SyncListResponse {
  file_exists: boolean
  source: string
  rules: SyncRule[]
  warnings: ValidationWarning[]
  version_meta?: FileFingerprint
  resync_required: boolean
}

export async function getSyncList(profileId: string): Promise<SyncListResponse> {
  const { data } = await client.get(`/profiles/${profileId}/sync-list`)
  return {
    ...data,
    rules: data?.rules || [],
    warnings: data?.warnings || [],
  }
}

export async function saveSyncList(
  profileId: string,
  req: { source: string; base_sha256: string },
): Promise<SyncListResponse> {
  const { data } = await client.put(`/profiles/${profileId}/sync-list`, req)
  return {
    ...data,
    rules: data?.rules || [],
    warnings: data?.warnings || [],
  }
}

export interface DirNode {
  name: string
  path: string // 相对路径，如 /Documents
  is_dir: boolean
  children?: DirNode[]
}

export async function getSyncListTree(profileId: string, depth = 3): Promise<DirNode[]> {
  const { data } = await client.get(`/profiles/${profileId}/sync-list/tree?depth=${depth}`)
  return data?.tree || []
}
