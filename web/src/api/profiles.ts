import client from './client'

export type RuntimeType = 'systemd' | 'docker' | 'podman'

export interface Profile {
  id: string
  display_name: string
  confdir: string
  runtime_type: RuntimeType
  runtime_target: string
}

export interface CreateProfileRequest {
  id: string
  display_name?: string
  confdir: string
  runtime_type: RuntimeType
  runtime_target?: string
}

export interface UpdateProfileRequest {
  display_name?: string
  runtime_target?: string
}

export async function listProfiles(): Promise<Profile[]> {
  const { data } = await client.get('/profiles')
  return data
}

export async function createProfile(req: CreateProfileRequest): Promise<Profile> {
  const { data } = await client.post('/profiles', req)
  return data
}

export async function getProfile(id: string): Promise<Profile> {
  const { data } = await client.get(`/profiles/${id}`)
  return data
}

export async function updateProfile(id: string, req: UpdateProfileRequest): Promise<Profile> {
  const { data } = await client.patch(`/profiles/${id}`, req)
  return data
}

export async function deleteProfile(id: string): Promise<void> {
  await client.delete(`/profiles/${id}`)
}

export async function discoverProfiles(): Promise<Profile[]> {
  const { data } = await client.post('/profiles/discover')
  return data
}
