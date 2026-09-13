<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useProfileStore } from '@/stores/profile'
import StatusBadge from '@/components/runtime/StatusBadge.vue'

const profileStore = useProfileStore()
const router = useRouter()
const systemInfo = ref<any>(null)

onMounted(async () => {
  profileStore.fetchAll()
  try {
    const { getSystemInfo } = await import('@/api/system')
    systemInfo.value = await getSystemInfo()
  } catch {
    systemInfo.value = null
  }
})
</script>

<template>
  <div>
    <header class="page-head">
      <h1>概览</h1>
      <span class="subtitle">OneWeb — OneDrive Web Control Plane</span>
    </header>

    <section v-if="systemInfo" class="system-card">
      <div class="sys-item"><span class="label">版本</span>{{ systemInfo.version }}</div>
      <div v-if="systemInfo.client_version" class="sys-item">
        <span class="label">onedrive</span>{{ systemInfo.client_version }}
      </div>
      <div v-if="systemInfo.capabilities" class="sys-item">
        <span class="label">能力</span>Schema {{ systemInfo.capabilities.schema_version }}
        <span v-if="systemInfo.capabilities.device_auth" class="chip">Device Auth</span>
        <span v-if="systemInfo.capabilities.sync_list" class="chip">SyncList</span>
      </div>
      <div class="sys-item">
        <span class="label">systemd</span>
        <span v-if="systemInfo.runtimes?.systemd?.available">可用</span>
        <span v-else class="muted">不可用</span>
      </div>
    </section>

    <section class="cards">
      <div v-if="profileStore.profiles.length === 0" class="empty">
        <p>尚未创建任何 Profile</p>
        <button class="btn primary" @click="router.push('/profiles')">创建第一个 Profile</button>
      </div>

      <div
        v-for="p in profileStore.profiles"
        :key="p.id"
        class="profile-card"
        @click="router.push(`/profiles/${p.id}`)"
      >
        <div class="card-head">
          <span class="card-title">{{ p.display_name || p.id }}</span>
          <StatusBadge :state="'stopped'" />
        </div>
        <div class="card-meta">
          <span>{{ p.confdir }}</span>
          <span class="chip">{{ p.runtime_type }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 20px;
}
.page-head h1 {
  font-size: 22px;
  margin: 0;
}
.subtitle {
  color: #64748b;
  font-size: 13px;
}
.system-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 16px 20px;
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
}
.sys-item {
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.sys-item .label {
  color: #64748b;
  font-size: 12px;
}
.chip {
  background: #e0f2fe;
  color: #0369a1;
  border-radius: 20px;
  padding: 2px 10px;
  font-size: 12px;
}
.muted {
  color: #94a3b8;
}
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}
.profile-card {
  background: #fff;
  border: 1px solid var(--oneweb-border);
  border-radius: 10px;
  padding: 18px;
  cursor: pointer;
  transition: box-shadow 0.15s;
}
.profile-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.card-title {
  font-weight: 600;
  font-size: 15px;
}
.card-meta {
  display: flex;
  justify-content: space-between;
  color: #64748b;
  font-size: 13px;
}
.empty {
  grid-column: 1 / -1;
  text-align: center;
  padding: 48px;
  color: #64748b;
}
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: none;
  font-size: 14px;
}
.btn.primary {
  background: var(--oneweb-primary);
  color: #fff;
}
</style>
