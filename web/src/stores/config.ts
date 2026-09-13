import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '@/api/config'
import type { ConfigResponse, SaveConfigRequest } from '@/api/config'

export const useConfigStore = defineStore('config', () => {
  const config = ref<ConfigResponse | null>(null)
  const loading = ref(false)
  const conflict = ref(false)
  const lastError = ref<string | null>(null)

  async function load(profileId: string) {
    loading.value = true
    lastError.value = null
    try {
      config.value = await api.getConfig(profileId)
    } catch (e: any) {
      lastError.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function save(profileId: string, req: SaveConfigRequest) {
    conflict.value = false
    lastError.value = null
    try {
      await api.saveConfig(profileId, req)
      await load(profileId)
      return true
    } catch (e: any) {
      if (e.message?.includes('modified')) {
        conflict.value = true
      }
      lastError.value = e.message
      return false
    }
  }

  return { config, loading, conflict, lastError, load, save }
})
