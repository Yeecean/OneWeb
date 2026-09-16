import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as api from '@/api/profiles'
import type { Profile, RuntimeType } from '@/api/profiles'

export const useProfileStore = defineStore('profile', () => {
  const profiles = ref<Profile[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loaded = computed(() => profiles.value.length > 0)

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      profiles.value = await api.listProfiles()
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function create(req: {
    id: string
    display_name: string
    confdir: string
    runtime_type: RuntimeType
    runtime_target: string
  }) {
    await api.createProfile(req)
    await fetchAll()
  }

  async function remove(id: string) {
    await api.deleteProfile(id)
    await fetchAll()
  }

  async function discover() {
    loading.value = true
    try {
      profiles.value = await api.discoverProfiles()
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { profiles, loading, error, loaded, fetchAll, create, remove, discover }
})
